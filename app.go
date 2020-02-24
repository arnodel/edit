package main

import "log"

type App struct {
	window        *Window
	cmdWindow     *Window
	focusedWindow *Window
	screenSize    Size
	eventHandler  *EventHandler
}

func (a *App) Resize(w, h int) {
	a.window.Resize(w, h-1)
	a.cmdWindow.Resize(w, 1)
	a.screenSize = Size{W: w, H: h}
}

func (a *App) HandleMouseEvent(evt Event) {

}

func (a *App) HandleEvent(evt Event) {
	action, err := a.eventHandler.HandleEvent(evt)
	if err != nil {
		log.Printf("Error handling event: %s", err)
	}
	if action != nil {
		action(a.window)
	}
}

func (a *App) Draw(screen *Screen) {
	screen.Fill(' ')
	sz := screen.Size()
	wscreen := screen.SubScreen(Rectangle{
		Size: Size{W: sz.W, H: sz.H - 1},
	})
	a.window.FocusCursor(wscreen)
	a.window.Draw(wscreen)
	a.window.DrawCursor(wscreen)
}
