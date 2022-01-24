package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/arnodel/golua/lib"
	"github.com/arnodel/golua/lib/debuglib"
	"github.com/arnodel/golua/lib/golib"
	"github.com/arnodel/golua/runtime"
)

type App struct {
	window        *Window
	cmdWindow     *Window
	focusedWindow *Window
	screenSize    Size
	eventHandler  *EventHandler
	lua           *runtime.Runtime
	logWindow     *Window
	cmdCallback   func(string)
}

func NewApp(filename string) *App {
	var buf *FileBuffer
	if filename == "" {
		buf = &FileBuffer{
			lines: []Line{nil},
		}
	} else {
		buf = NewBufferFromFile(filename)
	}

	win := &Window{
		buffer:  buf,
		tabSize: 4,
	}

	logBuf := &FileBuffer{
		lines: []Line{nil},
	}
	logWin := &Window{
		buffer:  logBuf,
		tabSize: 4,
	}

	// cmdBuf := &Buffer{
	// 	lines: []Line{nil},
	// }
	// cmdEventHandler := NewEventHandler()
	// cmdEventHandler.RegisterAction("Enter", func(w *Window) {
	// 	cmd := w.CurrentLine().String()
	// 	if a.cmdCallback != nil {
	// 		a.cmdCallback(cmd)
	// 	}
	// })
	// cmdWin := &Window{
	// 	buffer:       cmdBuf,
	// 	tabSize:      4,
	// 	eventHandler: cmdEventHandler,
	// }
	evtHandler := NewEventHandler()

	app := &App{
		window:        win,
		logWindow:     logWin,
		focusedWindow: win,
		eventHandler:  evtHandler,
	}
	for _, b := range defaultBindings {
		if err := evtHandler.RegisterAction(b.seq, b.action); err != nil {
			app.Logf("Unable to register %s: %s", b.seq, err)
		}
	}

	app.lua = runtime.New(app)
	lib.LoadAll(app.lua)
	app.lua.PushContext(runtime.RuntimeContextDef{
		MessageHandler: debuglib.Traceback,
	})
	return app
}

func (a *App) Resize(w, h int) {
	a.focusedWindow.Resize(w, h-1)
	a.cmdWindow.Resize(w, 1)
	a.screenSize = Size{W: w, H: h}
}

func (a *App) HandleMouseEvent(evt Event) {

}

func (a *App) HandleEvent(evt Event) {
	action, err := a.focusedWindow.eventHandler.HandleEvent(evt)
	if action == nil {
		action, err = a.eventHandler.HandleEvent(evt)
	}
	if action != nil {
		a.focusedWindow.eventHandler.Reset()
		a.eventHandler.Reset()
	}
	if err != nil {
		a.Logf("Error handling event: %s", err)
	}
	if action != nil {
		action(a.focusedWindow)
	}
}

func (a *App) CommandInput() {
	focusedWindow := a.focusedWindow
	cmdCallback := a.cmdCallback
	a.focusedWindow = a.cmdWindow
	a.cmdCallback = func(cmd string) {
		a.focusedWindow = focusedWindow
		a.cmdCallback = cmdCallback
	}
}

func (a *App) Draw(screen *Screen) {
	screen.Fill(' ')
	sz := screen.Size()
	wscreen := screen.SubScreen(Rectangle{
		Size: Size{W: sz.W, H: sz.H - 1},
	})
	a.focusedWindow.FocusCursor(wscreen)
	a.focusedWindow.Draw(wscreen)
	a.focusedWindow.DrawCursor(wscreen)
}

func (a *App) SwitchWindow() {
	if a.focusedWindow == a.window {
		a.focusedWindow = a.logWindow
	} else {
		a.focusedWindow = a.window
	}
	a.focusedWindow.Resize(a.screenSize.W, a.screenSize.H)
}

func (a *App) Write(p []byte) (int, error) {
	log.Print(string(p))
	return len(p), nil
}

func (a *App) Log(msg string) {
	log.Print(msg)
	a.logWindow.buffer.AppendLine(NewLineFromString(msg))
}

func (a *App) Logf(format string, args ...interface{}) {
	a.Log(fmt.Sprintf(format, args...))
}

func (a *App) Init() {
	initFile, err := ioutil.ReadFile("init.lua")
	if err != nil {
		a.Logf("Cannot read init file: %s", err)
		return
	}
	chunk, err := a.lua.CompileAndLoadLuaChunk("test", initFile, runtime.TableValue(a.lua.GlobalEnv()))
	if err != nil {
		a.Logf("Error compiling init file: %s", err)
		return
	}
	initFunc, err2 := runtime.Call1(a.lua.MainThread(), runtime.FunctionValue(chunk))
	if err2 != nil {
		a.Logf("Error running init chunk: %s", err2)
		return
	}
	_, err2 = runtime.Call1(a.lua.MainThread(), initFunc, golib.NewGoValue(a.lua, a))
	if err2 != nil {
		a.Logf("Error running init function: %s", err2)
		return
	}
}

func (a *App) LuaActionMaker(f runtime.Value) ActionMaker {
	return func(args []interface{}) Action {
		luaArgs := make([]runtime.Value, len(args)+1)
		for i, arg := range args {
			luaArgs[i+1] = golib.NewGoValue(a.lua, arg)
		}
		return func(win *Window) {
			luaArgs[0] = golib.NewGoValue(a.lua, win)
			_, err := runtime.Call1(a.lua.MainThread(), f, luaArgs...)
			if err != nil {
				a.Logf("Lua error: %s", err)
			}
		}
	}
}

type LuaBinding struct {
	Seq    string
	Action func(...interface{})
}

func (a *App) BindEvents(seq string, f runtime.Value) {
	err := a.eventHandler.RegisterAction(seq, a.LuaActionMaker(f))
	if err != nil {
		a.Logf("Error binding events: %s", err)
	}
}
