package main

type Position struct {
	X, Y int
}
type Size struct {
	W, H int
}

type Rectangle struct {
	Position
	Size
}

func (r Rectangle) BottomRight() Position {
	return Position{r.X + r.W, r.Y + r.H}
}

func (p Position) MoveBy(q Position) Position {
	return Position{
		p.X + q.X,
		p.Y + q.Y,
	}
}

func (p Position) MoveByX(x int) Position {
	return Position{
		p.X + x,
		p.Y,
	}
}

func (r Rectangle) Intersect(s Rectangle) Rectangle {
	r1 := r.BottomRight()
	s1 := s.BottomRight()
	if r.X < s.X {
		r.X = s.X
	}
	if r.Y < s.Y {
		r.Y = s.Y
	}
	if r1.X > s1.X {
		r1.X = s1.X
	}
	if r1.Y > s1.Y {
		r1.Y = s1.Y
	}
	r.W = s1.X - r.X
	if r.W < 0 {
		r.W = 0
	}
	r.H = s1.Y - r.Y
	if r.H < 0 {
		r.H = 0
	}
	return r
}

func (s Size) Contains(p Position) bool {
	return p.X >= 0 && p.Y >= 0 && p.X < s.W && p.Y < s.H
}
