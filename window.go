package main

import (
	"log"
)

// A Window is a view to a buffer.  It maintains a cursor position, and a
// rectangle view into the buffer.
type Window struct {
	buffer           IBuffer
	l, c             int // line and column of the cursor
	topLine, leftCol int // Index of the topmost visible line, column of the leftmost visibile column
	tabSize          int
	width, height    int
	eventHandler     *EventHandler
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
	l := y + w.topLine
	if l >= w.buffer.LineCount() {
		l = w.buffer.LineCount() - 1
	}
	line, _ := w.buffer.GetLine(l, 0)
	w.c = w.getPrinter().LineIndex(line, x)
	w.l = l
}

// MoveCursorToLineStart moves the cursor to the start of the current line.
func (w *Window) MoveCursorToLineStart() {
	w.c = 0
}

// MoveCursorToLineEnd moves the cursor to the end of the current line.
func (w *Window) MoveCursorToLineEnd() {
	line, err := w.buffer.GetLine(w.l, w.c)
	if err == nil {
		w.c = line.Len()
	}
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
func (w *Window) DeleteRune() error {
	l, c := w.buffer.AdvancePos(w.l, w.c, 0, -1)
	if w.c > 0 {
		w.c--
		w.buffer.DeleteRuneAt(l, c)
	} else {
		w.buffer.MergeLineWithPrevious(w.l)
	}
	w.c = c
	w.l = l
	return nil
}

// SplitLine splits the current line at the cursor position. If move is true,
// the cursor is moved down otherwise it stays in the same position.
func (w *Window) SplitLine(move bool) {
	err := w.buffer.SplitLine(w.l, w.c)
	if err != nil {
		log.Printf("error splitting line: %s", err)
		return
	}
	if move {
		w.c = 0
		w.l++
	}
}

//
// Buffer inspection methods
//

// CurrentLine returns the line the cursor is on currently.
func (w *Window) CurrentLine() (Line, error) {
	return w.buffer.GetLine(w.l, w.c)
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
		line, _ := w.buffer.GetLine(p.Y+w.topLine, 0)
		lp.Print(screen, p, line)
	}
}

// DrawCursor highlights the cursor if it is visible.
func (w *Window) DrawCursor(screen ScreenWriter) {
	line, _ := w.buffer.GetLine(w.l, w.c)
	screen.Reverse(Position{
		X: w.getPrinter().LineCol(line[:w.c]),
		Y: w.l - w.topLine,
	})
}

// FocusCursor adjusts the visible rectangle of the window if necessary to make
// the cursor visible.
func (w *Window) FocusCursor(screen ScreenWriter) {
	sz := screen.Size()
	line, _ := w.buffer.GetLine(w.l, w.c)
	x := w.getPrinter().LineCol(line[:w.c])
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
