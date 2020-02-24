package main

import "github.com/gdamore/tcell"

type ScreenWriter interface {
	Size() Size
	SetRune(Position, rune)
	Reverse(Position)
	SubScreen(Rectangle) ScreenWriter
}

type Screen struct {
	tcellScreen tcell.Screen
}

func (s Screen) Fill(c rune) {
	s.tcellScreen.Fill(' ', tcell.StyleDefault)
}

func (s Screen) Size() Size {
	w, h := s.tcellScreen.Size()
	return Size{W: w, H: h}
}

func (s Screen) SetRune(p Position, c rune) {
	s.tcellScreen.SetContent(p.X, p.Y, c, nil, tcell.StyleDefault)
}

func (s Screen) Reverse(p Position) {
	mainc, combc, style, _ := s.tcellScreen.GetContent(p.X, p.Y)
	s.tcellScreen.SetContent(p.X, p.Y, mainc, combc, style.Reverse(true))

}

func (s Screen) SubScreen(rect Rectangle) ScreenWriter {
	return &SubScreen{
		rect:   rect,
		screen: s,
	}
}

type SubScreen struct {
	rect   Rectangle
	screen Screen
}

func (s SubScreen) Size() Size {
	return s.rect.Size
}

func (s SubScreen) SetRune(p Position, c rune) {
	if s.rect.Size.Contains(p) {
		s.screen.SetRune(p.MoveBy(s.rect.Position), c)
	}
}

func (s SubScreen) Reverse(p Position) {
	if s.rect.Size.Contains(p) {
		s.screen.Reverse(p.MoveBy(s.rect.Position))
	}
}

func (s SubScreen) SubScreen(rect Rectangle) ScreenWriter {
	return &SubScreen{
		rect:   rect.Intersect(s.rect),
		screen: s.screen,
	}
}

// A Printer knows how to print lines on a screen
type Printer struct {
	Offset   int
	TabWidth int
}

// Print the line to the screen starting at coordinates (p.X, p.Y)
func (lp Printer) Print(s ScreenWriter, p Position, l Line) {
	sz := s.Size()
	if p.Y < 0 || p.Y >= sz.H {
		return
	}
	col := -lp.Offset
	for _, c := range l {
		if col >= 0 {
			s.SetRune(p.MoveByX(col), c)
		}
		if c == '\t' {
			col += lp.TabWidth
		} else {
			col++
		}
		if p.X+col >= sz.W {
			return
		}
	}

}

func (p Printer) LineCol(l Line) int {
	col := -p.Offset
	for _, c := range l {
		if c == '\t' {
			col += p.TabWidth
		} else {
			col++
		}
	}
	return col
}

func (p Printer) LineIndex(l Line, targetCol int) int {
	col := -p.Offset
	for i, c := range l {
		if c == '\t' {
			col += p.TabWidth
		} else {
			col++
		}
		if col > targetCol {
			return i
		}
	}
	return l.Len()
}
