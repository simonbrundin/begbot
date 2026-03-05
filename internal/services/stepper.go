package services

import (
	"fmt"
	"strings"
)

// StepLogger is a simple logging function used by Stepper
type StepLogger func(format string, args ...interface{})

// Stepper tracks and logs steps for a single run (e.g. per-ad)
type Stepper struct {
	steps []string
	idx   map[string]int
	total int
	cur   int
	logf  StepLogger
}

// NewStepper creates a Stepper with a canonical ordered list of step names and a logging function
func NewStepper(steps []string, logf StepLogger) *Stepper {
	idx := make(map[string]int, len(steps))
	for i, s := range steps {
		idx[strings.ToLower(s)] = i + 1
	}
	return &Stepper{
		steps: steps,
		idx:   idx,
		total: len(steps),
		cur:   len(steps),
		logf:  logf,
	}
}

// Markf logs a formatted message for the named step. If the step name exists in the
// canonical list its canonical index is used. Otherwise an incremental index is used.
func (sp *Stepper) Markf(name string, format string, args ...interface{}) {
	nameKey := strings.ToLower(name)
	i := 0
	if pos, ok := sp.idx[nameKey]; ok {
		i = pos
	} else {
		sp.cur++
		i = sp.cur
	}
	msg := fmt.Sprintf(format, args...)
	if sp.logf != nil {
		sp.logf("[%d/%d %s] %s", i, sp.total, strings.ToUpper(name), msg)
	}
}

// Mark is like Markf but accepts args similar to Printf style
func (sp *Stepper) Mark(name string, args ...interface{}) {
	sp.Markf(name, "%v", args...)
}
