package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
)

type IBuffer interface {
	LineCount() int
	GetLine(l, c int) (Line, error)
	InsertRune(r rune, l, c int) error
	InsertLine(l int, line Line)
	DeleteLine(l int)
	MergeLineWithPrevious(l int)
	SplitLine(l, c int) error
	DeleteRuneAt(l, c int) error
	AdvancePos(l, c, dl, dc int) (int, int)
	EndPos() (int, int)
}

// A Buffer maintains the data for a file.
type Buffer struct {
	lines    []Line
	filename string
}

func NewBufferFromFile(filename string) *Buffer {
	buf := &Buffer{
		filename: filename,
	}
	file, err := os.Open(filename)
	defer file.Close()

	if err != nil {
		buf.InsertLine(0, Line{})
		return buf
	}
	reader := bufio.NewReader(file)
	var lines []Line
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		lines = append(lines, NewLineFromString(line[:len(line)-1]))
	}
	buf.lines = lines
	if len(lines) == 0 {
		buf.InsertLine(0, Line{})
	}
	return buf
}

func (b *Buffer) Save() error {
	err := os.Rename(b.filename, b.filename+"~")
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Unable to back up file: %s", err)
		return err
	}
	file, err := os.Create(b.filename)
	defer file.Close()
	writer := bufio.NewWriter(file)
	for _, line := range b.lines {
		_, err = writer.WriteString(line.String())
		if err != nil {
			return err
		}
		_, err = writer.WriteRune('\n')
		if err != nil {
			return err
		}
	}
	writer.Flush()
	return nil
}

func (b *Buffer) LineCount() int {
	return len(b.lines)
}

func (b *Buffer) GetLine(l, c int) (Line, error) {
	if len(b.lines) <= l {
		return nil, fmt.Errorf("have %d lines, wamt to get line %d", len(b.lines), l)
	}
	line := b.lines[l]
	if line.Len() < c {
		return nil, errors.New("line too short")
	}
	return line, nil
}

func (b *Buffer) InsertRune(r rune, l, c int) error {
	line, err := b.GetLine(l, c)
	if err != nil {
		return err
	}
	b.lines[l] = line.InsertRune(r, c)
	return nil
}

func (b *Buffer) InsertLine(l int, line Line) {
	if l < 0 || l > len(b.lines) {
		return
	}
	switch {
	case l == len(b.lines):
		b.lines = append(b.lines, line)
	case len(b.lines) < cap(b.lines):
		tail := b.lines[l:]
		b.lines = b.lines[:len(b.lines)+1]
		copy(b.lines[l+1:], tail)
		b.lines[l] = line
	default:
		newLines := make([]Line, len(b.lines)+1, len(b.lines)+10)
		copy(newLines[:l], b.lines[:l])
		newLines[l] = line
		copy(newLines[l+1:], b.lines[l:])
		b.lines = newLines
	}
}

func (b *Buffer) DeleteLine(l int) {
	if l < 0 || l >= len(b.lines) {
		return
	}
	copy(b.lines[l:], b.lines[l+1:])
	b.lines = b.lines[:len(b.lines)-1]
	return
}

func (b *Buffer) MergeLineWithPrevious(l int) {
	if l < 1 || l >= len(b.lines) {
		return
	}
	b.lines[l-1] = b.lines[l-1].MergeWith(b.lines[l])
	copy(b.lines[l:], b.lines[l+1:])
	b.lines = b.lines[:len(b.lines)-1]
	return
}

func (b *Buffer) SplitLine(l, c int) error {
	line, err := b.GetLine(l, c)
	if err != nil {
		return err
	}
	l1, l2 := line.SplitAt(c)
	b.lines[l] = l1
	b.InsertLine(l+1, l2)
	return nil
}

func (b *Buffer) DeleteRuneAt(l, c int) error {
	line, err := b.GetLine(l, c)
	if err != nil {
		return err
	}
	if line.Len() == 0 {
		copy(b.lines[l:], b.lines[l+1:])
		b.lines = b.lines[:len(b.lines)-1]
		return nil
	}
	b.lines[l] = line.DeleteAt(c)
	return nil
}

func (b *Buffer) AdvancePos(l, c, dl, dc int) (int, int) {
	if l < 0 {
		return 0, 0
	}
	if l >= len(b.lines) {
		b.EndPos()
	}
	c += dc
	for c < 0 && l > 0 {
		l--
		c += b.lines[l].Len() + 1
	}
	if c < 0 {
		return 0, 0
	}
	for l < len(b.lines) && c > len(b.lines[l]) {
		c -= b.lines[l].Len() + 1
		l++
	}
	if l >= len(b.lines) {
		return b.EndPos()
	}
	l += dl
	if l < 0 {
		return 0, 0
	}
	if l >= len(b.lines) {
		return b.EndPos()
	}
	if c > len(b.lines[l]) {
		return l, len(b.lines[l])
	}
	return l, c
}

func (b *Buffer) NearestPos(l, c int) (int, int) {
	if l < 0 {
		l = 0
	} else if l >= len(b.lines) {
		l = len(b.lines) - 1
	}
	line := b.lines[l]
	if c < 0 {
		c = 0
	} else if c > len(line) {
		c = len(line)
	}
	return l, c
}

func (b *Buffer) EndPos() (int, int) {
	return len(b.lines) - 1, len(b.lines[len(b.lines)-1])
}
