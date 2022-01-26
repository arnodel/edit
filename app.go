package edit

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
	running       bool
	window        *Window
	cmdWindow     *Window
	focusedWindow *Window
	screenSize    Size
	eventHandler  *EventHandler
	lua           *runtime.Runtime
	logWindow     *Window
	cmdCallback   func(string)
	eventHandlers map[string]*EventHandler
}

func NewApp(win *Window) *App {

	logBuf := &FileBuffer{
		lines: make([]Line, 1),
	}
	logWin := &Window{
		buffer:  logBuf,
		tabSize: 4,
	}

	cmdBuf := &FileBuffer{
		// 		lines: []Line{nil},
	}
	// cmdEventHandler := NewEventHandler()
	// cmdEventHandler.RegisterAction("Enter", func(w *Window) {
	// 	cmd := w.CurrentLine().String()
	// 	if a.cmdCallback != nil {
	// 		a.cmdCallback(cmd)
	// 	}
	// })
	cmdWin := &Window{
		buffer:  cmdBuf,
		tabSize: 4,
		//		eventHandler: cmdEventHandler,
	}
	evtHandler := NewEventHandler()

	win.MoveCursorToEnd()
	app := &App{
		window:        win,
		logWindow:     logWin,
		focusedWindow: win,
		cmdWindow:     cmdWin,
		eventHandlers: map[string]*EventHandler{
			"app": evtHandler,
		},
		eventHandler: evtHandler,
		running:      true,
	}
	app.lua = runtime.New(app)
	for _, b := range defaultBindings {
		if err := evtHandler.RegisterAction(b.seq, b.action); err != nil {
			app.Logf("Unable to register %s: %s", b.seq, err)
		}
	}
	lib.LoadAll(app.lua)
	app.lua.PushContext(runtime.RuntimeContextDef{
		MessageHandler: debuglib.Traceback,
	})
	win.RegisterWithApp(app)
	return app
}

func (a *App) InitLuaFile(initfile string) error {
	luaCode, err := ioutil.ReadFile(initfile)
	if err != nil {
		a.Logf("Cannot read init file: %s", err)
		return err
	}
	return a.InitLuaCode(initfile, luaCode)
}

func (a *App) InitLuaCode(initfile string, luaCode []byte) error {
	chunk, err := a.lua.CompileAndLoadLuaChunk(initfile, luaCode, runtime.TableValue(a.lua.GlobalEnv()))
	if err != nil {
		a.Logf("Error compiling init file: %s", err)
		return err
	}
	initFunc, err2 := runtime.Call1(a.lua.MainThread(), runtime.FunctionValue(chunk))
	if err2 != nil {
		a.Logf("Error running init chunk: %s", err2)
		return err
	}
	_, err2 = runtime.Call1(a.lua.MainThread(), initFunc, golib.NewGoValue(a.lua, a))
	if err2 != nil {
		a.Logf("Error running init function: %s", err2)
		return err
	}
	return nil
}

func (a *App) Resize(w, h int) {
	a.focusedWindow.Resize(w, h-1)
	a.cmdWindow.Resize(w, 1)
	a.screenSize = Size{W: w, H: h}
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
	screen.Show()
}

func (a *App) SwitchWindow() {
	if a.focusedWindow == a.window {
		a.focusedWindow = a.logWindow
	} else {
		a.focusedWindow = a.window
	}
	a.focusedWindow.Resize(a.screenSize.W, a.screenSize.H-1)
}

func (a *App) Write(p []byte) (int, error) {
	log.Print(string(p))
	return len(p), nil
}

func (a *App) Log(msg string) {
	log.Print(msg)
	a.logWindow.buffer.AppendLine(NewLineFromString(msg, nil))
}

func (a *App) Logf(format string, args ...interface{}) {
	a.Log(fmt.Sprintf(format, args...))
}

func (a *App) Quit() {
	a.running = false
}

func (a *App) Running() bool {
	return a.running
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

func (a *App) BindEvents(name, seq string, f runtime.Value) {
	h := a.GetEventHandler(name)
	err := h.RegisterAction(seq, a.LuaActionMaker(f))
	if err != nil {
		a.Logf("Error binding events: %s", err)
	}
}

func (a *App) GetEventHandler(name string) *EventHandler {
	h, ok := a.eventHandlers[name]
	if !ok {
		h = NewEventHandler()
		a.eventHandlers[name] = h
	}
	return h
}
