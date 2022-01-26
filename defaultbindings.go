package edit

func CmdInsertRune(r rune) Action {
	return func(w *Window) { w.InsertRune(r) }
}

func CmdCursorLeft(w *Window)  { w.MoveCursor(0, -1) }
func CmdCursorRight(w *Window) { w.MoveCursor(0, 1) }
func CmdCursorUp(w *Window)    { w.MoveCursor(-1, 0) }
func CmdCursorDown(w *Window)  { w.MoveCursor(1, 0) }

func CmdDeletePrevRune(w *Window)  { w.DeleteRune() }
func CmdCarriageReturn(w *Window)  { w.SplitLine(true) }
func CmdMoveToLineStart(w *Window) { w.MoveCursorToLineStart() }
func CmdMoveToLineEnd(w *Window)   { w.MoveCursorToLineEnd() }

func CmdPageDown(w *Window) { w.PageDown(1) }
func CmdPageUp(w *Window)   { w.PageDown(-1) }

func CmdScrollDown(w *Window) { w.ScrollDown(1) }
func CmdScrollUp(w *Window)   { w.ScrollUp(1) }

func CmdResize(w, h int) Action {
	return func(win *Window) { win.App().Resize(w, h) }
}

func CmdMoveCursorTo(pos Position) Action {
	return func(w *Window) { w.MoveCursorTo(pos.X, pos.Y) }
}

func CmdQuit(w *Window) { w.App().Quit() }

func CmdSaveBuffer(w *Window) { w.buffer.Save() }

func SimpleActionMaker(f Action) ActionMaker {
	return func(args []interface{}) Action { return f }
}

var defaultBindings = []struct {
	seq    string
	action ActionMaker
}{
	{
		seq: "Rune.Rune",
		action: func(args []interface{}) Action {
			return CmdInsertRune(args[0].(rune))
		},
	},
	{
		seq:    "Tab",
		action: SimpleActionMaker(CmdInsertRune('\t')),
	},
	{
		seq:    "Left",
		action: SimpleActionMaker(CmdCursorLeft),
	},
	{
		seq:    "Right",
		action: SimpleActionMaker(CmdCursorRight),
	},
	{
		seq:    "Ctrl-F",
		action: SimpleActionMaker(CmdCursorRight),
	},
	{
		seq:    "Up",
		action: SimpleActionMaker(CmdCursorUp),
	},
	{
		seq:    "Ctrl-P",
		action: SimpleActionMaker(CmdCursorUp),
	},
	{
		seq:    "Down",
		action: SimpleActionMaker(CmdCursorDown),
	},
	{
		seq:    "Ctrl-N",
		action: SimpleActionMaker(CmdCursorDown),
	},
	{
		seq:    "Backspace2", // TODO
		action: SimpleActionMaker(CmdDeletePrevRune),
	},
	{
		seq:    "Enter",
		action: SimpleActionMaker(CmdCarriageReturn),
	},
	{
		seq:    "Ctrl-A",
		action: SimpleActionMaker(CmdMoveToLineStart),
	},
	{
		seq:    "Ctrl-E",
		action: SimpleActionMaker(CmdMoveToLineEnd),
	},
	{
		seq:    "Ctrl-V",
		action: SimpleActionMaker(CmdPageDown),
	},
	{
		seq:    "Alt+Ctrl-V",
		action: SimpleActionMaker(CmdPageUp),
	},
	{
		seq: "Resize.Size",
		action: func(args []interface{}) Action {
			size := args[0].(Size)
			return CmdResize(size.W, size.H)
		},
	},
	{
		seq: "MousePress-Button1.Position",
		action: func(args []interface{}) Action {
			return CmdMoveCursorTo(args[0].(Position))
		},
	},
	{
		seq:    "MousePress-WheelDown",
		action: SimpleActionMaker(CmdScrollDown),
	},
	{
		seq:    "MousePress-WheelUp",
		action: SimpleActionMaker(CmdScrollUp),
	},
	{
		seq:    "Ctrl-X Ctrl-S",
		action: SimpleActionMaker(CmdSaveBuffer),
	},
	{
		seq:    "Ctrl-C",
		action: SimpleActionMaker(CmdQuit),
	},
}
