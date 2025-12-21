package status

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
)

// ANSI color codes
const (
	Reset         = "\033[0m"
	Green         = "\033[32m"
	Yellow        = "\033[33m"
	Blue          = "\033[34m"
	Red           = "\033[31m"
	Gray          = "\033[90m"
	Bold          = "\033[1m"
	ClearLine     = "\033[2K"
	CursorUp      = "\033[1A"
	SaveCursor    = "\033[s"
	RestoreCursor = "\033[u"
)

// State represents the current state of a workflow step
type State int

const (
	StatePending State = iota
	StateRunning
	StateComplete
	StateFailed
)

// Step represents a single step in the workflow
type Step struct {
	Name        string
	Description string
	State       State
	StartTime   time.Time
	EndTime     time.Time
	SubSteps    []SubStep
}

// SubStep represents a sub-task within a step
type SubStep struct {
	Name        string
	Description string
	State       State
}

// Reporter provides inline status reporting for CLI operations
type Reporter struct {
	steps      []Step
	currentIdx int
	writer     io.Writer
	mu         sync.Mutex
	colored    bool
	simpleMode bool
	lastOutput string
}

// NewReporter creates a new status reporter
func NewReporter(colored bool) *Reporter {
	return &Reporter{
		steps:      []Step{},
		writer:     os.Stdout,
		colored:    colored,
		simpleMode: false,
	}
}

// SetSimpleMode enables or disables simple mode (no inline updates)
func (r *Reporter) SetSimpleMode(simple bool) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.simpleMode = simple
}

// GetSteps returns a copy of the current steps (for testing)
func (r *Reporter) GetSteps() []Step {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	// Return a copy to prevent external modification
	steps := make([]Step, len(r.steps))
	copy(steps, r.steps)
	return steps
}

// AddStep adds a new step to the workflow
func (r *Reporter) AddStep(name, description string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.steps = append(r.steps, Step{
		Name:        name,
		Description: description,
		State:       StatePending,
	})
}

// StartStep marks a step as running
func (r *Reporter) StartStep(name string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.steps {
		if r.steps[i].Name == name {
			r.steps[i].State = StateRunning
			r.steps[i].StartTime = time.Now()
			r.currentIdx = i
			r.render()
			return
		}
	}
}

// CompleteStep marks a step as complete
func (r *Reporter) CompleteStep(name string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.steps {
		if r.steps[i].Name == name {
			r.steps[i].State = StateComplete
			r.steps[i].EndTime = time.Now()
			r.render()
			return
		}
	}
}

// FailStep marks a step as failed
func (r *Reporter) FailStep(name string, err error) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.steps {
		if r.steps[i].Name == name {
			r.steps[i].State = StateFailed
			r.steps[i].EndTime = time.Now()
			r.render()
			if err != nil {
				fmt.Fprintf(r.writer, "\n%sError: %v%s\n", r.color(Red), err, r.color(Reset))
			}
			return
		}
	}
}

// AddSubStep adds a sub-step to a parent step
func (r *Reporter) AddSubStep(parentName, subStepName, description string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.steps {
		if r.steps[i].Name == parentName {
			r.steps[i].SubSteps = append(r.steps[i].SubSteps, SubStep{
				Name:        subStepName,
				Description: description,
				State:       StatePending,
			})
			r.render()
			return
		}
	}
}

// UpdateSubStep updates the state of a sub-step
func (r *Reporter) UpdateSubStep(parentName, subStepName string, state int) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.steps {
		if r.steps[i].Name == parentName {
			for j := range r.steps[i].SubSteps {
				if r.steps[i].SubSteps[j].Name == subStepName {
					r.steps[i].SubSteps[j].State = State(state)
					r.render()
					return
				}
			}
		}
	}
}

// render updates the display with current status
func (r *Reporter) render() {
	isTerminal := r.isTerminal()

	// In simple mode, only output on state changes, no inline updates
	if r.simpleMode {
		output := r.buildSimpleOutput()
		if output != "" {
			fmt.Fprint(r.writer, output)
		}
		return
	}

	// Clear previous output if in interactive mode
	// Count the actual lines that need to be cleared
	if r.lastOutput != "" && isTerminal {
		// Count lines in previous output to clear properly
		lineCount := strings.Count(r.lastOutput, "\n")
		for i := 0; i < lineCount; i++ {
			fmt.Fprintf(r.writer, "%s%s\r", CursorUp, ClearLine)
		}
	}

	output := r.buildOutput()
	fmt.Fprint(r.writer, output)
	r.lastOutput = output
}

// buildOutput creates the status display
func (r *Reporter) buildOutput() string {
	var output string

	// Title
	output += fmt.Sprintf("\n%s%sIDPBuilder Progress%s\n", r.color(Bold), r.color(Blue), r.color(Reset))

	// Steps
	for i, step := range r.steps {
		symbol := r.getSymbol(step.State)
		color := r.getColor(step.State)

		status := ""
		if step.State == StateRunning {
			status = fmt.Sprintf(" %s(in progress)%s", r.color(Gray), r.color(Reset))
		} else if step.State == StateComplete && !step.EndTime.IsZero() && !step.StartTime.IsZero() {
			duration := step.EndTime.Sub(step.StartTime).Round(time.Millisecond)
			status = fmt.Sprintf(" %s(%s)%s", r.color(Gray), duration, r.color(Reset))
		}

		// Format: [✓] Step description (status)
		output += fmt.Sprintf("  %s%s%s %s%s\n",
			r.color(color),
			symbol,
			r.color(Reset),
			step.Description,
			status)

		// Add separator after current running step
		if i == r.currentIdx && step.State == StateRunning {
			output += fmt.Sprintf("  %s│%s\n", r.color(Blue), r.color(Reset))

			// Show sub-steps if any
			if len(step.SubSteps) > 0 {
				for _, subStep := range step.SubSteps {
					subSymbol := r.getSymbol(subStep.State)
					subColor := r.getColor(subStep.State)
					output += fmt.Sprintf("  %s│%s   %s%s%s %s\n",
						r.color(Blue),
						r.color(Reset),
						r.color(subColor),
						subSymbol,
						r.color(Reset),
						subStep.Description)
				}
			}
		}
	}

	return output
}

// buildSimpleOutput creates simple status output (one line per state change)
func (r *Reporter) buildSimpleOutput() string {
	// Only show the current step that changed
	if r.currentIdx >= 0 && r.currentIdx < len(r.steps) {
		step := r.steps[r.currentIdx]
		symbol := r.getSymbol(step.State)
		color := r.getColor(step.State)

		status := ""
		if step.State == StateRunning {
			status = "..."
		} else if step.State == StateComplete && !step.EndTime.IsZero() && !step.StartTime.IsZero() {
			duration := step.EndTime.Sub(step.StartTime).Round(time.Millisecond)
			status = fmt.Sprintf(" (%s)", duration)
		} else if step.State == StateFailed {
			status = " (failed)"
		}

		return fmt.Sprintf("%s%s%s %s%s\n",
			r.color(color),
			symbol,
			r.color(Reset),
			step.Description,
			status)
	}
	return ""
}

// getSymbol returns the symbol for a state
func (r *Reporter) getSymbol(state State) string {
	switch state {
	case StatePending:
		return "○"
	case StateRunning:
		return "●"
	case StateComplete:
		return "✓"
	case StateFailed:
		return "✗"
	default:
		return "○"
	}
}

// getColor returns the color for a state
func (r *Reporter) getColor(state State) string {
	switch state {
	case StatePending:
		return Gray
	case StateRunning:
		return Blue
	case StateComplete:
		return Green
	case StateFailed:
		return Red
	default:
		return Reset
	}
}

// color returns the ANSI color code if colored output is enabled
func (r *Reporter) color(code string) string {
	if r.colored {
		return code
	}
	return ""
}

// isTerminal checks if output is a terminal
func (r *Reporter) isTerminal() bool {
	if f, ok := r.writer.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// Summary prints a final summary
func (r *Reporter) Summary() {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	// Ensure we have a clean final render
	r.render()

	// Count states
	completed := 0
	failed := 0
	for _, step := range r.steps {
		if step.State == StateComplete {
			completed++
		} else if step.State == StateFailed {
			failed++
		}
	}

	if failed > 0 {
		fmt.Fprintf(r.writer, "\n%s%s✗ Build failed: %d/%d steps completed%s\n",
			r.color(Bold), r.color(Red), completed, len(r.steps), r.color(Reset))
	} else {
		fmt.Fprintf(r.writer, "\n%s%s✓ Build completed successfully: %d/%d steps%s\n",
			r.color(Bold), r.color(Green), completed, len(r.steps), r.color(Reset))
	}
}
