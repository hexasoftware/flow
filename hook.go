package flow

import (
	"sync"
	"time"
)

// Hooks for node life cycle
type Hooks struct {
	sync.Mutex
	hooks []Hook
}

// Hook funcs to handle certain events on the flow
type Hook struct {
	Wait   func(op Operation, triggerTime time.Time)
	Start  func(op Operation, triggerTime time.Time)
	Finish func(op Operation, triggerTime time.Time, res interface{})
	Error  func(op Operation, triggerTime time.Time, err error)
	Any    func(name string, op Operation, triggerTime time.Time, extra ...interface{})
}

// Trigger a hook
func (hs *Hooks) Trigger(name string, op Operation, extra ...Data) {
	hs.Lock()
	defer hs.Unlock()

	for _, h := range hs.hooks {
		if h.Any != nil {
			h.Any(name, op, time.Now(), extra...)
		}
		switch name {
		case "Wait":
			if h.Wait != nil {
				h.Wait(op, time.Now())
			}
		case "Start":
			if h.Start != nil {
				h.Start(op, time.Now())
			}
		case "Finish":
			if h.Finish != nil {
				h.Finish(op, time.Now(), extra[0])
			}
		case "Error":
			if h.Error != nil {
				h.Error(op, time.Now(), extra[0].(error))
			}
		}
	}
}

func (hs *Hooks) wait(op Operation)             { hs.Trigger("Wait", op) }
func (hs *Hooks) start(op Operation)            { hs.Trigger("Start", op) }
func (hs *Hooks) finish(op Operation, res Data) { hs.Trigger("Finish", op, res) }
func (hs *Hooks) error(op Operation, err error) { hs.Trigger("Error", op, err) }

// Attach attach a hook
func (hs *Hooks) Attach(h Hook) {
	hs.Lock()
	defer hs.Unlock()
	hs.hooks = append(hs.hooks, h)

}
