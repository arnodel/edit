package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
)

type Buffer interface {
	LineCount() int
	GetLine(l, c int) (Line, error)
	InsertRune(r rune, l, c int) error
	InsertLine(l int, line Line) error
	DeleteLine(l int) error
	MergeLineWithPrevious(l int)
	SplitLine(l, c int) error
	DeleteRuneAt(l, c int) error
	AdvancePos(l, c, dl, dc int) (int, int)
	EndPos() (int, int)
	AppendLine(Line) error
	Save() error
}

// A FileBuffer maintains the data for a file.
type FileBuffer struct {
	lines    []Line
	filename string
	readOnly bool
}

var _ Buffer = (*FileBuffer)(nil)

func NewBufferFromFile(filename string) *FileBuffer {
	buf := &FileBuffer{
		filename: filename,
	}
	file, err := os.Open(filename)
	if err != nil {
		buf.InsertLine(0, Line{})
		return buf
	}
	defer file.Close()
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

func (b *FileBuffer) Save() error {
	if b.readOnly {
		return errors.New("Cannot save a read only buffer")
	}
	err := os.Rename(b.filename, b.filename+"~")
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	file, err := os.Create(b.filename)
	if err != nil {
		return err
	}
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

func (b *FileBuffer) LineCount() int {
	return len(b.lines)
}

func (b *FileBuffer) GetLine(l, c int) (Line, error) {
	if len(b.lines) <= l {
		return nil, fmt.Errorf("have %d lines, want to get line %d", len(b.lines), l)
	}
	line := b.lines[l]
	if line.Len() < c {
		return nil, errors.New("line too short")
	}
	return line, nil
}

func (b *FileBuffer) InsertRune(r rune, l, c int) error {
	line, err := b.GetLine(l, c)
	if err != nil {
		return err
	}
	b.lines[l] = line.InsertRune(r, c)
	return nil
}

func (b *FileBuffer) InsertLine(l int, line Line) error {
	if l < 0 || l > len(b.lines) {
		return fmt.Errorf("out of range")
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
	return nil
}

func (b *FileBuffer) AppendLine(line Line) error {
	return b.InsertLine(len(b.lines), line)
}

func (b *FileBuffer) DeleteLine(l int) error {
	if l < 0 || l >= len(b.lines) {
		return fmt.Errorf("out of range")
	}
	copy(b.lines[l:], b.lines[l+1:])
	b.lines = b.lines[:len(b.lines)-1]
	return nil
}

func (b *FileBuffer) MergeLineWithPrevious(l int) {
	if l < 1 || l >= len(b.lines) {
		return
	}
	b.lines[l-1] = b.lines[l-1].MergeWith(b.lines[l])
	copy(b.lines[l:], b.lines[l+1:])
	b.lines = b.lines[:len(b.lines)-1]
}

func (b *FileBuffer) SplitLine(l, c int) error {
	line, err := b.GetLine(l, c)
	if err != nil {
		return err
	}
	l1, l2 := line.SplitAt(c)
	b.lines[l] = l1
	b.InsertLine(l+1, l2)
	return nil
}

func (b *FileBuffer) DeleteRuneAt(l, c int) error {
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

func (b *FileBuffer) AdvancePos(l, c, dl, dc int) (int, int) {
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
	for l < len(b.lines) && c > b.lines[l].Len() {
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
	if c > b.lines[l].Len() {
		return l, len(b.lines[l])
	}
	return l, c
}

func (b *FileBuffer) NearestPos(l, c int) (int, int) {
	if l < 0 {
		l = 0
	} else if l >= len(b.lines) {
		l = len(b.lines) - 1
	}
	line := b.lines[l]
	if c < 0 {
		c = 0
	} else if c > line.Len() {
		c = line.Len()
	}
	return l, c
}

func (b *FileBuffer) EndPos() (int, int) {
	return len(b.lines) - 1, len(b.lines[len(b.lines)-1])
}
