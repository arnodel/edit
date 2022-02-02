package edit

import (
	"errors"
	"fmt"
	"log"
	"math"
)

// A Window is a view to a buffer.  It maintains a cursor position, and a
// rectangle view into the buffer.
type Window struct {
	buffer           Buffer
	l, c             int // line and column of the cursor
	topLine, leftCol int // Index of the topmost visible line, column of the leftmost visibile column
	tabSize          int
	width, height    int
	eventHandler     *EventHandler
	app              *App

	copyStartL, copyStartC int
	copyEndL, copyEndC     int

	regionFirstL, regionFirstC int
	regionLastL, regionLastC   int
}

func NewWindow(buf Buffer) *Window {
	return &Window{
		buffer:     buf,
		tabSize:    4,
		copyStartL: -1,
		copyStartC: -1,
		copyEndL:   -1,
		copyEndC:   -1,
	}
}

//
// General
//

func (w *Window) RegisterWithApp(app *App) {
	w.eventHandler = app.GetEventHandler(w.buffer.Kind())
	w.app = app
}

func (w *Window) App() *App {
	return w.app
}

func (w *Window) HandleEvent(evt Event) (Action, error) {
	if evt.EventType == Key || evt.EventType == Rune {
		w.ResetHighlightRegion()
	}
	return w.eventHandler.HandleEvent(evt)
}

//
// Movement methods
//

// MoveCursor moves the cursor by a number of lines and columns.
func (w *Window) MoveCursor(dl, dc int) {
	w.l, w.c = w.buffer.AdvancePos(w.l, w.c, dl, dc)
}

// MoveCursorTo moves the cursor to a given screen position.
func (w *Window) MoveCursorTo(x, y int) {
	w.l, w.c = w.GetLineCol(x, y)
}

func (w *Window) GetLineCol(x, y int) (int, int) {
	l := y + w.topLine
	if l >= w.buffer.LineCount() {
		l = w.buffer.LineCount() - 1
	}
	line, _ := w.buffer.GetLine(l, 0)
	return w.buffer.AdvancePos(l, w.getPrinter().LineIndex(line, x), 0, 0)
}

// MoveCursorToLineStart moves the cursor to the start of the current line.
func (w *Window) MoveCursorToLineStart() {
	w.l, w.c = w.buffer.AdvancePos(w.l, 0, 0, 0)
}

// MoveCursorToLineEnd moves the cursor to the end of the current line.
func (w *Window) MoveCursorToLineEnd() {
	line, err := w.buffer.GetLine(w.l, w.c)
	if err == nil {
		w.l, w.c = w.buffer.AdvancePos(w.l, line.Len(), 0, 0)
	}
}

func (w *Window) MoveCursorToEnd() {
	l, c := w.buffer.EndPos()
	w.l, w.c = w.buffer.AdvancePos(l, c, 0, 0)
}

// PageDown moves the cursor down by n pages (or up by -n pages if n < 0).
func (w *Window) PageDown(n int) {
	if n == 0 {
		return
	}
	lineOffset := (w.height - 2) + (n-1)*w.height
	w.MoveCursor(lineOffset, 0)
}

// ScrollDown scrolls down by n lines, attempting to keep the cursor on the same
// buffer line.
func (w *Window) ScrollDown(n int) {
	if n <= 0 {
		return
	}
	w.topLine += n
	maxTopLine := w.buffer.LineCount() - 1
	if w.topLine > maxTopLine {
		w.topLine = maxTopLine
	}
	dl := w.l - w.topLine
	if dl < 0 {
		w.MoveCursor(-dl, 0)
	}
}

// ScrollUp scrolls up by n lines, attempting to keep the cursor on the same
// buffer line.
func (w *Window) ScrollUp(n int) {
	if n <= 0 {
		return
	}
	w.topLine -= n
	if w.topLine < 0 {
		w.topLine = 0
	}
	dl := w.topLine + w.height - 1 - w.l
	if dl < 0 {
		w.MoveCursor(dl, 0)
	}
}

func (w *Window) StartHighlightRegion(x, y int) {
	w.copyStartL, w.copyStartC = w.GetLineCol(x, y)
	w.copyEndL, w.copyEndC = -1, -1
}

func (w *Window) MoveHighlightRegion(x, y int) {
	w.copyEndL, w.copyEndC = w.GetLineCol(x, y)
}

func (w *Window) StopHightlightRegion(x, y int) bool {
	w.copyEndL, w.copyEndC = w.GetLineCol(x, y)
	if w.copyEndC == w.copyStartC && w.copyEndL == w.copyStartL {
		w.ResetHighlightRegion()
		return false
	}
	return true
}

func (w *Window) ResetHighlightRegion() {
	w.copyStartL, w.copyStartC = -1, -1
	w.copyEndL, w.copyEndC = -1, -1
}

func (w *Window) GetHighlightedString() (string, error) {
	if w.copyEndC == -1 {
		return "", fmt.Errorf("no highlight region")
	}
	return w.buffer.StringFromRegion(w.copyStartL, w.copyStartC, w.copyEndL, w.copyEndC)
}

func (w *Window) PasteString(s string) (err error) {
	w.l, w.c, err = w.buffer.InsertString(s, w.l, w.c)
	return
}

//
// Editing methods
//

// InsertRune inserts a character into the buffer at the cursor position.
func (w *Window) InsertRune(r rune) {
	err := w.buffer.InsertRune(r, w.l, w.c)
	if err != nil {
		log.Printf("error inserting rune: %s", err)
		return
	}
	w.c++
}

// DeleteRune deletes the character to the left of the cursor position.
func (w *Window) DeleteRune() (err error) {
	if w.l == 0 && w.c == 0 {
		return errors.New("start of buffer")
	}
	l, c := w.buffer.AdvancePos(w.l, w.c, 0, -1)
	if l == w.l {
		err = w.buffer.DeleteRuneAt(l, c)
	} else {
		err = w.buffer.MergeLineWithPrevious(w.l)
	}
	w.c = c
	w.l = l
	return
}

// SplitLine splits the current line at the cursor position. If move is true,
// the cursor is moved down otherwise it stays in the same position.
func (w *Window) SplitLine(move bool) error {
	err := w.buffer.SplitLine(w.l, w.c)
	if err != nil {
		log.Printf("error splitting line: %s", err)
		return err
	}
	if move {
		w.l, w.c = w.buffer.AdvancePos(w.l+1, 0, 0, 0)
	}
	return nil
}

//
// Buffer inspection methods
//

// CurrentLine returns the line the cursor is on currently.
func (w *Window) CurrentLine() (Line, error) {
	return w.buffer.GetLine(w.l, w.c)
}

func (w *Window) Buffer() Buffer {
	return w.buffer
}

func (w *Window) CursorLine() int {
	return w.l
}

func (w *Window) CursorPos() (int, int) {
	return w.l, w.c
}

//
// Drawing methods
//

// Resize changes the size of the window.
func (w *Window) Resize(width, height int) {
	w.width = width
	w.height = height
}

// Draw draws the contents of the window on the screen.
func (w *Window) Draw(screen ScreenWriter) {
	w.orderRegion()
	sh := screen.Size().H
	linesAvail := w.buffer.LineCount() - w.topLine
	if linesAvail <= 0 {
		return
	}
	if sh > linesAvail {
		sh = linesAvail
	}
	lp := w.getPrinter()
	for p := (Position{}); p.Y < sh; p.Y++ {
		lp.Print(screen, p, w.StyledLineIter(p.Y+w.topLine, 0))
	}
}

// DrawCursor highlights the cursor if it is visible.
func (w *Window) DrawCursor(screen ScreenWriter) {
	line, _ := w.buffer.GetLine(w.l, w.c)
	screen.Reverse(Position{
		X: w.getPrinter().LineCol(line, w.c),
		Y: w.l - w.topLine,
	})
}

// FocusCursor adjusts the visible rectangle of the window if necessary to make
// the cursor visible.
func (w *Window) FocusCursor(screen ScreenWriter) {
	sz := screen.Size()
	line, _ := w.buffer.GetLine(w.l, w.c)
	x := w.getPrinter().LineCol(line, w.c)
	if x < 0 {
		w.leftCol += x
	} else if x >= sz.W {
		w.leftCol += x - sz.W + 1
	}
	y := w.l - w.topLine
	if y < 0 {
		w.topLine = w.l
	} else if y >= sz.H {
		w.topLine = w.l - sz.H + 1
	}
}

func (w *Window) getPrinter() Printer {
	return Printer{
		TabWidth: w.tabSize,
		Offset:   w.leftCol,
	}
}

func (w *Window) orderRegion() {
	l0, c0 := w.copyStartL, w.copyStartC
	l1, c1 := w.copyEndL, w.copyEndC
	if l1 < l0 || (l0 == l1 && c1 < c0) {
		l0, c0, l1, c1 = l1, c1, l0, c0
	}
	w.regionFirstL, w.regionFirstC = l0, c0
	w.regionLastL, w.regionLastC = l1, c1

}
func (w *Window) StyledLineIter(l, c int) StyledLineIter {
	iter := w.buffer.StyledLineIter(l, c)
	if w.copyEndL >= 0 && l >= w.regionFirstL && l <= w.regionLastL {
		c1, c2 := 0, math.MaxInt
		if l == w.regionFirstL {
			c1 = w.regionFirstC - c
		}
		if l == w.regionLastL {
			c2 = w.regionLastC - c
		}
		iter = &highlightIter{
			iter: iter,
			c1:   c1,
			c2:   c2,
		}
	}
	return iter
}

type highlightIter struct {
	iter   StyledLineIter
	c1, c2 int
}

var _ StyledLineIter = (*highlightIter)(nil)

func (i *highlightIter) Next() (rune, Style) {
	r, s := i.iter.Next()
	if i.c1 <= 0 && i.c2 >= 0 {
		s = s.Reverse(true)
	}
	i.c1--
	i.c2--
	return r, s
}

func (i *highlightIter) HasNext() bool {
	return i.iter.HasNext()
}
