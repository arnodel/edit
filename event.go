package main

import (
	"errors"
	"log"
	"strings"

	"github.com/gdamore/tcell"
)

type EventType int

const (
	NoEvent EventType = iota
	Rune
	Key
	Mouse
	Resize
)

func (t EventType) Name() string {
	switch t {
	case Rune:
		return "Rune"
	case Key:
		return "Key"
	case Mouse:
		return "Mouse"
	case Resize:
		return "Resize"
	default:
		return ""
	}
}

type Modifiers int

const (
	Shift Modifiers = 1 << iota
	Control
	Alt
	Meta
)

type MouseButtons int

const (
	Button1 MouseButtons = 1 << iota
	Button2
	Button3
	WheelDown
	WheelUp
	WheelLeft
	WheelRight
)

func (mb MouseButtons) Has(mb1 MouseButtons) bool {
	return mb&mb1 == mb1
}

func (mb MouseButtons) Empty() bool {
	return mb == 0
}

func (mb MouseButtons) Name() string {
	switch {
	case mb.Has(Button1):
		return "Button1"
	case mb.Has(Button2):
		return "Button2"
	case mb.Has(Button3):
		return "Button3"
	case mb.Has(WheelDown):
		return "WheelUp"
	case mb.Has(WheelUp):
		return "WheelDown"
	case mb.Has(WheelLeft):
		return "WheelLeft"
	case mb.Has(WheelRight):
		return "WheelRight"
	default:
		return ""
	}
}

type Position struct {
	X, Y int
}

type MouseData struct {
	Position
	Buttons         MouseButtons
	ButtonsPressed  MouseButtons
	ButtonsReleased MouseButtons
}

type KeyData tcell.Key

type Size struct {
	W, H int
}

type Event struct {
	EventType
	KeyData
	Rune rune
	Modifiers
	MouseData
	Size
}

func (e Event) nameWithoutMod() string {
	switch e.EventType {
	case Rune:
		if e.Rune == ' ' {
			return "Space"
		}
		return string(e.Rune)
	case Key:
		return tcell.KeyNames[tcell.Key(e.KeyData)]
	case Mouse:
		if !e.ButtonsPressed.Empty() {
			return "MousePress-" + e.ButtonsPressed.Name()
		}
		if !e.ButtonsReleased.Empty() {
			return "MouseRelease-" + e.ButtonsReleased.Name()
		}
		return "MouseMove"
	default:
		return ""
	}
}

func (e Event) Name() string {
	name := e.nameWithoutMod()
	if e.Modifiers&Control != 0 {
		name = addPrefix(name, "Ctrl-")
	}
	if e.Modifiers&Meta != 0 {
		name = addPrefix(name, "Meta+")
	}
	if e.Modifiers&Alt != 0 {
		name = addPrefix(name, "Alt+")
	}
	if e.Modifiers&Shift != 0 {
		name = addPrefix(name, "Shift+")
	}
	return name
}

func addPrefix(s, p string) string {
	if strings.HasPrefix(s, p) {
		return s
	}
	return p + s
}

type Action func(win *Window)

type ActionMaker func([]Event) Action

type stateDef struct {
	action      ActionMaker
	transitions map[string]string
}

type EventHandler struct {
	states       map[string]stateDef
	currentState string
	events       []Event
}

func NewEventHandler() *EventHandler {
	return &EventHandler{
		states: map[string]stateDef{},
	}
}

// HandleEvent transitions the state of the event handler and returns any action
// that may be triggered.
func (h *EventHandler) HandleEvent(evt Event) (Action, error) {
	sDef := h.states[h.currentState]
	evtName := evt.Name()
	trans := evtName
	newState, ok := sDef.transitions[trans]
	if !ok {
		trans = evt.EventType.Name()
		newState, ok = sDef.transitions[trans]
	}
	if !ok {
		if evt.EventType == Mouse && evtName == "MouseMove" {
			// Discard a mouse event that has no button presses or releases
			return nil, nil
		}
		h.currentState = ""
		log.Printf("No transition for event %s", evt.Name())
		return nil, errors.New("no known transition for event")
	}
	log.Printf("Transition: %s, new state: %s", trans, newState)
	sDef = h.states[newState]
	h.currentState = newState
	h.events = append(h.events, evt)
	var action Action
	if sDef.action != nil {
		action = sDef.action(h.events)
		h.currentState = ""
		h.events = nil
		log.Printf("Action found, back to initial state")
	}
	return action, nil
}

// RegisterAction associates the given action with the event sequence
// represented by seq. E.g.  RegisterAction("x y z", action) will create the
// following transitions:
//
// ""  -x-> x
// x   -y-> x:y
// x:y -z-> x:y:z
//
// It will also record that the state x:y:z triggers action x:y:z triggers
// action
func (h *EventHandler) RegisterAction(seq string, action ActionMaker) error {
	events := strings.Split(seq, " ")
	s := ""
	for _, event := range events {
		sDef, ok := h.states[s]
		if !ok {
			break
		}
		if sDef.action != nil {
			return errors.New("action already exists")
		}
		s += ":" + event
	}
	sDef, ok := h.states[s]
	if ok && len(sDef.transitions) > 0 {
		return errors.New("seq is prefix to existing seq")
	}
	// All is well
	s = ""
	for _, event := range events {
		sDef := h.states[s]
		if sDef.transitions == nil {
			sDef.transitions = map[string]string{}
			h.states[s] = sDef
		}
		s += ":" + event
		sDef.transitions[event] = s
	}
	sDef = h.states[s]
	sDef.action = action
	h.states[s] = sDef
	return nil
}
