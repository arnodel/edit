package main

import "github.com/gdamore/tcell/v2"

type TcellEventConverter struct {
	lastButtonMask tcell.ButtonMask
}

func (c *TcellEventConverter) EventFromTcell(tcevt tcell.Event) Event {
	switch tevt := tcevt.(type) {
	case *tcell.EventResize:
		w, h := tevt.Size()
		return Event{
			EventType: Resize,
			Size: Size{
				W: w,
				H: h,
			},
		}
	case *tcell.EventMouse:
		buttonMask := tevt.Buttons()
		mx, my := tevt.Position()
		return Event{
			EventType: Mouse,
			MouseData: MouseData{
				Position: Position{
					X: mx,
					Y: my,
				},
				Buttons:         buttonsFromTcell(buttonMask),
				ButtonsPressed:  buttonsFromTcell(buttonMask &^ c.lastButtonMask),
				ButtonsReleased: buttonsFromTcell(c.lastButtonMask &^ buttonMask),
			},
			Modifiers: modifiersFromTcell(tevt.Modifiers()),
		}
	case *tcell.EventKey:
		if tevt.Key() == tcell.KeyRune {
			return Event{
				EventType: Rune,
				Rune:      tevt.Rune(),
				Modifiers: modifiersFromTcell(tevt.Modifiers()),
			}
		}
		return Event{
			EventType: Key,
			Rune:      tevt.Rune(),
			KeyData:   KeyData(tevt.Key()),
			Modifiers: modifiersFromTcell(tevt.Modifiers()),
		}
	default:
		return Event{}
	}
}

func modifiersFromTcell(mod tcell.ModMask) Modifiers {
	var m Modifiers
	if mod&tcell.ModAlt != 0 {
		m |= Alt
	}
	if mod&tcell.ModShift != 0 {
		m |= Shift
	}
	if mod&tcell.ModCtrl != 0 {
		m |= Control
	}
	if mod&tcell.ModMeta != 0 {
		m |= Meta
	}
	return m
}

func buttonsFromTcell(btn tcell.ButtonMask) MouseButtons {
	var b MouseButtons
	if btn&tcell.Button1 != 0 {
		b |= Button1
	}
	if btn&tcell.Button2 != 0 {
		b |= Button2
	}
	if btn&tcell.Button3 != 0 {
		b |= Button3
	}
	if btn&tcell.WheelDown != 0 {
		b |= WheelDown
	}
	if btn&tcell.WheelUp != 0 {
		b |= WheelUp
	}
	if btn&tcell.WheelLeft != 0 {
		b |= WheelLeft
	}
	if btn&tcell.WheelRight != 0 {
		b |= WheelRight
	}
	return b
}
