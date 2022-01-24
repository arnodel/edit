package main

import (
	"errors"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
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

type MouseData struct {
	Position
	Buttons         MouseButtons
	ButtonsPressed  MouseButtons
	ButtonsReleased MouseButtons
}

type KeyData tcell.Key

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

type ActionMaker func([]interface{}) Action

type stateDef struct {
	action      ActionMaker
	eventFields []string
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

func convertEvents(events []Event, fields []string) []interface{} {
	var out []interface{}
	for i, evt := range events {
		switch fields[i] {
		case "Rune":
			out = append(out, evt.Rune)
		case "Position":
			out = append(out, evt.Position)
		case "Size":
			out = append(out, evt.Size)

		}
	}
	return out
}

// HandleEvent transitions the state of the event handler and returns any action
// that may be triggered.
func (h *EventHandler) HandleEvent(evt Event) (Action, error) {
	if h == nil {
		return nil, nil
	}
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
		h.Reset()
		return nil, errors.New("no known transition for event")
	}
	sDef = h.states[newState]
	h.currentState = newState
	h.events = append(h.events, evt)
	var action Action
	if sDef.action != nil {
		action = sDef.action(convertEvents(h.events, sDef.eventFields))
		h.Reset()
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
	if h == nil {
		return errors.New("cannot register an action on a nil handler")
	}
	events := strings.Split(seq, " ")
	s := ""
	var eventFields []string
	var eventNames []string
	for _, event := range events {
		parts := strings.SplitN(event, ".", 2)
		eventName := parts[0]
		eventField := ""
		if len(parts) == 2 {
			eventField = parts[1]
		}
		sDef := h.states[s]
		if sDef.action != nil {
			return errors.New("action already exists")
		}
		s = childState(s, eventName)
		eventNames = append(eventNames, eventName)
		eventFields = append(eventFields, eventField)
	}
	sDef, ok := h.states[s]
	if ok && len(sDef.transitions) > 0 {
		return errors.New("seq is prefix to existing seq")
	}
	// All is well
	s = ""
	log.Printf("Action for %s: %v", seq, eventNames)
	for _, eventName := range eventNames {
		sDef := h.states[s]
		if sDef.transitions == nil {
			sDef.transitions = map[string]string{}
			h.states[s] = sDef
		}
		s = childState(s, eventName)
		sDef.transitions[eventName] = s
	}
	sDef = h.states[s]
	sDef.action = action
	sDef.eventFields = eventFields
	h.states[s] = sDef
	return nil
}

// Reset the handler so any ongoing sequence is aborted.
func (h *EventHandler) Reset() {
	if h == nil {
		return
	}
	h.currentState = ""
	h.events = nil
}

func childState(s, t string) string {
	return s + " " + t
}
