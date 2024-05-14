package util

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// constants from remote target parameters supported by Kustomize
// https://github.com/kubernetes-sigs/kustomize/blob/master/examples/remoteBuild.md
const (
	QueryStringRef        = "ref"
	QueryStringVersion    = "version"
	QueryStringTimeout    = "timeout"
	QueryStringSubmodules = "submodules"

	RepoUrlDelimiter = "//"
	SCPDelimiter     = ":"
	UserDelimiter    = "@"

	defaultTimeout        = time.Second * 27
	defaultCloneSubmodule = true

	errMsgUrlUnsupported = "url must have // after the repository url. example: https://github.com/kubernetes-sigs/kustomize//examples"
	errMsgUrlColon       = "first path segment in URL cannot contain colon"
)

type KustomizeRemote struct {
	raw string

	Scheme   string
	User     string
	Password string
	Host     string
	Port     string
	RepoPath string

	FilePath string

	Ref        string
	Submodules bool
	Timeout    time.Duration
}

func (g *KustomizeRemote) CloneUrl() string {
	sb := strings.Builder{}
	if g.Scheme != "" {
		sb.WriteString(fmt.Sprintf("%s://", g.Scheme))
	}
	if g.User != "" {
		sb.WriteString(g.User)
		if g.Password != "" {
			sb.WriteString(fmt.Sprintf(":%s", g.Password))
		}
		sb.Write([]byte(UserDelimiter))
	}

	sb.WriteString(g.Host)
	if g.Port != "" {
		sb.WriteString(fmt.Sprintf(":%s", g.Port))
	}
	if g.Scheme == "" {
		sb.WriteString(":")
	} else {
		sb.WriteString("/")
	}

	sb.WriteString(g.RepoPath)
	return sb.String()
}

func (g *KustomizeRemote) Path() string {
	return g.FilePath
}

func (g *KustomizeRemote) parseQuery() error {
	_, query, _ := strings.Cut(g.raw, "?")
	values, err := url.ParseQuery(query)

	if err != nil {
		return err
	}

	// if empty, it means we checkout the default branch
	version := values.Get(QueryStringVersion)
	ref := values.Get(QueryStringRef)
	// ref has higher priority per kustomize doc
	if ref != "" {
		version = ref
	}

	duration := defaultTimeout
	timeoutString := values.Get(QueryStringTimeout)
	if timeoutString != "" {
		timeoutInt, sErr := strconv.Atoi(timeoutString)
		if sErr == nil {
			duration = time.Duration(timeoutInt) * time.Second
		} else {
			t, sErr := time.ParseDuration(timeoutString)
			if sErr == nil {
				duration = t
			}
		}
	}

	cloneSubmodules := defaultCloneSubmodule
	submodule := values.Get(QueryStringSubmodules)
	if submodule != "" {
		v, pErr := strconv.ParseBool(submodule)
		if pErr == nil {
			cloneSubmodules = v
		}
	}

	g.Ref = version
	g.Submodules = cloneSubmodules
	g.Timeout = duration

	return nil
}

func (g *KustomizeRemote) parse() error {
	parsed, err := url.Parse(g.raw)
	if err != nil {
		if strings.Contains(err.Error(), errMsgUrlColon) {
			return g.parseSCPStyle()
		}
		return err
	}

	g.Scheme, g.User, g.Host = parsed.Scheme, parsed.User.Username(), parsed.Host
	p, ok := parsed.User.Password()
	if ok {
		g.Password = p
	}

	err = g.parseQuery()
	if err != nil {
		return fmt.Errorf("parsing query parameters in package url: %s: %w", g.raw, err)
	}

	return g.parsePath(parsed.Path)
}

func (g *KustomizeRemote) parseSCPStyle() error {
	// example git@github.com:owner/repo
	cIndex := strings.Index(g.raw, SCPDelimiter)
	if cIndex == -1 {
		return fmt.Errorf("not a valid SCP style URL")
	}

	uIndex := strings.Index(g.raw[:cIndex], UserDelimiter)
	if uIndex != -1 {
		g.User = g.raw[:uIndex]
	}
	g.Host = g.raw[uIndex+1 : cIndex]
	err := g.parseQuery()
	if err != nil {
		return fmt.Errorf("parsing query parameters in package url: %s: %w", g.raw, err)
	}

	pathEnd := len(g.raw)
	qIndex := strings.Index(g.raw, "?")
	if qIndex != -1 {
		pathEnd = qIndex
	}
	return g.parsePath(g.raw[cIndex+1 : pathEnd])
}

func (g *KustomizeRemote) parsePath(path string) error {
	// example kubernetes-sigs/kustomize//examples/multibases/dev/
	index := strings.Index(path, RepoUrlDelimiter)
	if index == -1 {
		return fmt.Errorf(errMsgUrlUnsupported)
	}

	g.RepoPath = strings.TrimPrefix(path[:index], "/")
	g.FilePath = strings.TrimSuffix(path[index+2:], "/")
	return nil
}

func NewKustomizeRemote(uri string) (*KustomizeRemote, error) {
	r := &KustomizeRemote{raw: uri}
	return r, r.parse()
}
