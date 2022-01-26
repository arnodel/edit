package edit

import (
	"unicode/utf8"
)

type ILine interface {
	Len() int
	InsertRune(r rune, c int) ILine
	DeleteAt(c int) Line
	MergeWith(l2 ILine) ILine
	String() string
	Iter(c int) LineIter
}

type LineIter interface {
	Next() rune
	HasNext() bool
}

type StyledLineIter interface {
	Next() (rune, Style)
	HasNext() bool
}

type constStyleLineIter struct {
	LineIter
	style Style
}

func NewConstStyleLineIter(iter LineIter, style Style) StyledLineIter {
	return &constStyleLineIter{
		LineIter: iter,
		style:    style,
	}
}

var _ StyledLineIter = (*constStyleLineIter)(nil)

func (i *constStyleLineIter) Next() (rune, Style) {
	return i.LineIter.Next(), i.style
}

type Line struct {
	Runes []rune
	Meta  interface{}
}

type lineIter struct {
	c     int
	runes []rune
}

func (i *lineIter) HasNext() bool {
	return len(i.runes) > i.c
}

func (i *lineIter) Next() rune {
	r := i.runes[i.c]
	i.c++
	return r
}

func NewLineFromString(s string, meta interface{}) Line {
	runes := make([]rune, utf8.RuneCountInString(s))
	i := 0
	for _, r := range s {
		runes[i] = r
		i++
	}
	return Line{Runes: runes, Meta: meta}
}

func (l Line) Len() int {
	return len(l.Runes)
}

func (l Line) SplitAt(c int) (Line, Line) {
	switch {
	case c <= 0:
		return Line{Meta: l.Meta}, l
	case c >= l.Len():
		return l, Line{Meta: l.Meta}
	default:
		left := Line{Runes: append([]rune(nil), l.Runes[:c]...), Meta: l.Meta}
		right := Line{Runes: append([]rune(nil), l.Runes[c:]...), Meta: l.Meta}
		return left, right
	}
}

func (l Line) InsertRune(r rune, c int) Line {
	runes := l.Runes
	switch {
	case c < 0 || c > len(runes):
		return l
	case c == len(runes):
		return Line{Runes: append(runes, r), Meta: l.Meta}
	case cap(runes) > len(runes):
		tail := runes[c:]
		runes = runes[:len(runes)+1]
		copy(runes[c+1:], tail)
		runes[c] = r
		return Line{Runes: runes, Meta: l.Meta}
	default:
		newRunes := make([]rune, len(runes)+1, len(runes)+10)
		copy(newRunes[:c], runes[:c])
		newRunes[c] = r
		copy(newRunes[c+1:], runes[c:])
		return Line{Runes: newRunes, Meta: l.Meta}
	}
}

func (l Line) DeleteAt(c int) Line {
	runes := l.Runes
	switch {
	case c < 0 || c >= len(runes):
		return l
	case c == len(runes)-1:
		return Line{Runes: runes[:c], Meta: l.Meta}
	default:
		copy(runes[c:], runes[c+1:])
		return Line{Runes: runes[:len(runes)-1], Meta: l.Meta}
	}
}

func (l Line) MergeWith(l2 Line) Line {
	return Line{Runes: append(l.Runes, l2.Runes...), Meta: l.Meta}
}

func (l Line) String() string {
	return string(l.Runes)
}

func (l Line) Iter(c int) LineIter {
	return &lineIter{
		c:     c,
		runes: l.Runes,
	}
}
