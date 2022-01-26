package edit

import "github.com/gdamore/tcell/v2"

type ScreenWriter interface {
	Size() Size
	SetRune(Position, rune, tcell.Style)
	Reverse(Position)
	SubScreen(Rectangle) ScreenWriter
}

type Screen struct {
	tcellScreen    tcell.Screen
	eventConverter TcellEventConverter
}

func NewScreen() (*Screen, error) {
	tcellScreen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	err = tcellScreen.Init()
	if err != nil {
		return nil, err
	}
	tcellScreen.EnableMouse()
	tcellScreen.EnablePaste()
	return &Screen{tcellScreen: tcellScreen}, nil
}

func (s *Screen) Cleanup() {
	s.tcellScreen.Fini()
}

func (s *Screen) PollEvent() Event {
	return s.eventConverter.EventFromTcell(s.tcellScreen.PollEvent())
}

func (s *Screen) Fill(c rune) {
	s.tcellScreen.Fill(' ', tcell.StyleDefault)
}

func (s *Screen) Show() {
	s.tcellScreen.Show()
}

func (s *Screen) Size() Size {
	w, h := s.tcellScreen.Size()
	return Size{W: w, H: h}
}

func (s *Screen) SetRune(p Position, c rune, style tcell.Style) {
	s.tcellScreen.SetContent(p.X, p.Y, c, nil, style)
}

func (s *Screen) Reverse(p Position) {
	mainc, combc, style, _ := s.tcellScreen.GetContent(p.X, p.Y)
	s.tcellScreen.SetContent(p.X, p.Y, mainc, combc, style.Reverse(true))

}

func (s *Screen) SubScreen(rect Rectangle) ScreenWriter {
	return &SubScreen{
		rect:   rect,
		screen: s,
	}
}

type SubScreen struct {
	rect   Rectangle
	screen *Screen
}

func (s SubScreen) Size() Size {
	return s.rect.Size
}

func (s SubScreen) SetRune(p Position, c rune, style tcell.Style) {
	if s.rect.Size.Contains(p) {
		s.screen.SetRune(p.MoveBy(s.rect.Position), c, style)
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
func (lp Printer) Print(s ScreenWriter, p Position, iter StyledLineIter) {
	sz := s.Size()
	if p.Y < 0 || p.Y >= sz.H {
		return
	}
	col := -lp.Offset
	for iter.HasNext() {
		r, style := iter.Next()
		if col >= 0 {
			s.SetRune(p.MoveByX(col), r, style)
		}
		if r == '\t' {
			col += lp.TabWidth
		} else {
			col++
		}
		if p.X+col >= sz.W {
			return
		}
	}

}

func (p Printer) LineCol(l Line, i int) int {
	col := -p.Offset
	for iter := l.Iter(0); iter.HasNext() && i > 0; i-- {
		if iter.Next() == '\t' {
			col += p.TabWidth
		} else {
			col++
		}
	}
	return col
}

func (p Printer) LineIndex(l Line, targetCol int) int {
	col := -p.Offset
	for i, iter := 0, l.Iter(0); iter.HasNext(); i++ {
		if iter.Next() == '\t' {
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
