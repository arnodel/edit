package main

import "unicode/utf8"

type Line []rune

func NewLineFromString(s string) Line {
	line := make(Line, utf8.RuneCountInString(s))
	i := 0
	for _, r := range s {
		line[i] = r
		i++
	}
	return line
}

func (l Line) Len() int {
	return len(l)
}

func (l Line) SplitAt(c int) (Line, Line) {
	switch {
	case c <= 0:
		return Line(nil), l
	case c >= len(l):
		return l, Line(nil)
	default:
		left := append(Line(nil), l[:c]...)
		right := append(Line(nil), l[c:]...)
		return left, right
	}
}

func (l Line) InsertRune(r rune, c int) Line {
	switch {
	case c < 0 || c > len(l):
		return l
	case c == len(l):
		return append(l, r)
	case cap(l) > len(l):
		tail := l[c:]
		l = l[:len(l)+1]
		copy(l[c+1:], tail)
		l[c] = r
		return l
	default:
		newL := make(Line, len(l)+1, len(l)+10)
		copy(newL[:c], l[:c])
		newL[c] = r
		copy(newL[c+1:], l[c:])
		return newL
	}
}

func (l Line) DeleteAt(c int) Line {
	switch {
	case c < 0 || c >= len(l):
		return l
	case c == len(l)-1:
		return l[:c]
	default:
		copy(l[c:], l[c+1:])
		return l[:len(l)-1]
	}
}

func (l Line) MergeWith(l2 Line) Line {
	return append(l, l2...)
}

func (l Line) String() string {
	return string(l)
}
