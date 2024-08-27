package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"slices"
	"sync"
	"time"
)

// https://en.wikipedia.org/wiki/ANSI_escape_code
const (
	Reset              = "\033[0m"
	White              = "\033[37m"
	WhiteDim           = "\033[37;2m"
	Green              = "\033[32m"
	GreenDimUnderlined = "\033[32;2;4m"
	Magenta            = "\033[35m"
	BrightRed          = "\033[91m"
	BrightYellow       = "\033[93m"
	Cyan               = "\033[36m"
	CyanDim            = "\033[36;2m"
	// this mirrors the limit value from the internal slog package
	maxBufferSize = 16384
	dateFormat    = time.Stamp
)

var bufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 2048)
		return &b
	},
}

type Options struct {
	AddSource  bool
	Colored    bool
	Level      slog.Leveler
	TimeFormat string
}

// Handler is very similar to slog's commonHandler
type Handler struct {
	opts              Options
	json              bool
	preformattedAttrs []byte
	groupPrefix       string
	groups            []string
	unopenedGroups    []string
	nOpenGroups       int
	mu                *sync.Mutex
	w                 io.Writer
}

func NewHandler(out io.Writer, opts Options) *Handler {
	return &Handler{
		opts:              opts,
		preformattedAttrs: make([]byte, 0),
		unopenedGroups:    make([]string, 0),
		nOpenGroups:       0,
		mu:                &sync.Mutex{},
		w:                 out,
	}
}

func (h *Handler) clone() *Handler {
	return &Handler{
		opts:              h.opts,
		json:              h.json,
		preformattedAttrs: slices.Clip(h.preformattedAttrs),
		groupPrefix:       h.groupPrefix,
		groups:            slices.Clip(h.groups),
		nOpenGroups:       h.nOpenGroups,
		w:                 h.w,
		mu:                h.mu,
	}
}

func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

func (h *Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	h2 := h.clone()
	h2.unopenedGroups = make([]string, len(h.unopenedGroups)+1)
	copy(h2.unopenedGroups, h.unopenedGroups)
	h2.unopenedGroups[len(h2.unopenedGroups)-1] = name
	return h2
}

func (h *Handler) WithAttrs(as []slog.Attr) slog.Handler {
	if len(as) == 0 {
		return h
	}
	h2 := h.clone()
	h2.preformattedAttrs = h2.appendUnopenedGroups(h2.preformattedAttrs)
	h2.unopenedGroups = nil

	for _, a := range as {
		h2.preformattedAttrs = h2.appendAttr(h2.preformattedAttrs, a)
	}
	return h2
}

func (h *Handler) appendUnopenedGroups(buf []byte) []byte {
	for _, g := range h.unopenedGroups {
		buf = fmt.Appendf(buf, "%s ", g)
	}
	return buf
}

func (h *Handler) appendAttr(buf []byte, a slog.Attr) []byte {
	a.Value = a.Value.Resolve()
	if a.Equal(slog.Attr{}) {
		return buf
	}

	switch a.Value.Kind() {
	case slog.KindGroup:
		attrs := a.Value.Group()
		if len(attrs) == 0 {
			return buf
		}
		if a.Key != "" {
			for _, ga := range attrs {
				buf = h.appendAttr(buf, ga)
			}
		}
	default:
		if a.Key == "" || a.Value.String() == "" {
			return buf
		}
		buf = h.appendKeyValuePair(buf, a)
	}
	return buf
}

func (h *Handler) Handle(ctx context.Context, record slog.Record) error {
	bufp := bufPool.Get().(*[]byte)
	buf := *bufp

	defer func() {
		*bufp = buf
		free(bufp)
	}()

	// append time, level, then message.
	if h.opts.Colored {
		buf = fmt.Appendf(buf, WhiteDim)
		buf = slog.Time(slog.TimeKey, record.Time).Value.Time().AppendFormat(buf, fmt.Sprintf("%s%s ", dateFormat, Reset))

		var color string
		switch record.Level {
		case slog.LevelDebug:
			color = Magenta
		case slog.LevelInfo:
			color = Green
		case slog.LevelWarn:
			color = BrightYellow
		case slog.LevelError:
			color = BrightRed
		default:
			color = Magenta
		}
		buf = fmt.Appendf(buf, "%s%s%s ", color, record.Level.String(), Reset)

		buf = fmt.Appendf(buf, "%s%s%s ", Cyan, record.Message, Reset)
	} else {
		buf = slog.Time(slog.TimeKey, record.Time).Value.Time().AppendFormat(buf, fmt.Sprintf("%s ", dateFormat))
		buf = fmt.Appendf(buf, "%s ", record.Level)
		buf = fmt.Appendf(buf, "%s ", record.Message)
	}

	if h.opts.AddSource {
		buf = h.appendAttr(buf, slog.Any(slog.SourceKey, source(record)))
	}

	buf = append(buf, h.preformattedAttrs...)
	if record.NumAttrs() > 0 {
		buf = h.appendUnopenedGroups(buf)
		record.Attrs(func(a slog.Attr) bool {
			buf = h.appendAttr(buf, a)
			return true
		})
	}
	buf = append(buf, "\n"...)
	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write(buf)
	return err
}

func (h *Handler) appendKeyValuePair(buf []byte, a slog.Attr) []byte {
	if h.opts.Colored {
		if a.Key == "err" {
			return fmt.Appendf(buf, "%s%s=%v%s ", BrightRed, a.Key, a.Value.String(), Reset)
		}
		return fmt.Appendf(buf, "%s%s=%s%s%s%s ", WhiteDim, a.Key, Reset, White, a.Value.String(), Reset)
	}
	return fmt.Appendf(buf, "%s=%v ", a.Key, a.Value.String())
}

func free(b *[]byte) {
	if cap(*b) <= maxBufferSize {
		*b = (*b)[:0]
		bufPool.Put(b)
	}
}

func source(r slog.Record) *slog.Source {
	fs := runtime.CallersFrames([]uintptr{r.PC})
	f, _ := fs.Next()
	return &slog.Source{
		Function: f.Function,
		File:     f.File,
		Line:     f.Line,
	}
}
