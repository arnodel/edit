package main

import "github.com/gdamore/tcell"

type Screen struct {
	tcellScreen tcell.Screen
}

func (s Screen) Fill(c rune) {
	s.tcellScreen.Fill(' ', tcell.StyleDefault)
}

func (s Screen) Size() (int, int) {
	return s.tcellScreen.Size()
}

func (s Screen) SetRune(x, y int, c rune) {
	s.tcellScreen.SetContent(x, y, c, nil, tcell.StyleDefault)
}

func (s Screen) Reverse(x, y int) {
	mainc, combc, style, _ := s.tcellScreen.GetContent(x, y)
	s.tcellScreen.SetContent(x, y, mainc, combc, style.Reverse(true))

}

type Printer struct {
	X, Y        int
	Offset, End int
	TabWidth    int
}

func (p Printer) Print(s *Screen, l Line) {
	sw, sh := s.Size()
	if p.Y < 0 || p.Y >= sh {
		return
	}
	col := -p.Offset
	for _, c := range l {
		if p.End > 0 && col >= p.End {
			return
		}
		if col >= 0 {
			s.SetRune(p.X+col, p.Y, c)
		}
		if c == '\t' {
			col += p.TabWidth
		} else {
			col++
		}
		if p.X+col >= sw {
			return
		}
	}

}

func (p Printer) LineCol(l Line) int {
	col := -p.Offset
	for _, c := range l {
		if p.End > 0 && col >= p.End {
			return col
		}
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
		if p.End > 0 && col >= p.End {
			return i
		}
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
