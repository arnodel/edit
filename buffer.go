package edit

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"
)

type Buffer interface {
	LineCount() int
	GetLine(l, c int) (Line, error)
	InsertRune(r rune, l, c int) error
	InsertString(s string, l, c int) (int, int, error)
	InsertLine(l int, line Line) error
	DeleteLine(l int) error
	MergeLineWithPrevious(l int) error
	SplitLine(l, c int) error
	DeleteRuneAt(l, c int) error
	AdvancePos(l, c, dl, dc int) (int, int)
	EndPos() (int, int)
	AppendLine(Line)
	Save() error
	StyledLineIter(l, c int) StyledLineIter
	Kind() string
	StringFromRegion(l0, c0, l1, c1 int) (string, error)
}

// A FileBuffer maintains the data for a file.
type FileBuffer struct {
	lines    []Line
	filename string
	readOnly bool
}

var _ Buffer = (*FileBuffer)(nil)

func NewEmptyFileBuffer() *FileBuffer {
	return &FileBuffer{
		lines: []Line{NewLineFromString("", nil)},
	}
}

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
		lines = append(lines, NewLineFromString(line[:len(line)-1], nil))
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

func (b *FileBuffer) Line(l int) Line {
	return b.lines[l]
}

func (b *FileBuffer) GetLine(l, c int) (Line, error) {
	if l < 0 || len(b.lines) <= l {
		return Line{}, fmt.Errorf("out of range")
	}
	line := b.lines[l]
	if line.Len() < c {
		return Line{}, errors.New("line too short")
	}
	return line, nil
}

func (b *FileBuffer) SetLine(l int, line Line) error {
	if l < 0 || len(b.lines) <= l {
		return fmt.Errorf("out of range")
	}
	b.lines[l] = line
	return nil
}

func (b *FileBuffer) InsertRune(r rune, l, c int) error {
	line, err := b.GetLine(l, c)
	if err != nil {
		return err
	}
	b.lines[l] = line.InsertRune(r, c)
	return nil
}

func (b *FileBuffer) InsertString(s string, l, c int) (int, int, error) {
	for i, part := range splitString(s) {
		if i > 0 {
			if err := b.SplitLine(l, c); err != nil {
				return l, c, err
			}
			l, c = b.AdvancePos(l+1, 0, 0, 0)
		}
		b.lines[l] = b.lines[l].InsertString(part, c)
		l, c = b.AdvancePos(l, c, 0, utf8.RuneCountInString(part))
	}
	return l, c, nil
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

func (b *FileBuffer) AppendLine(line Line) {
	b.lines = append(b.lines, line)
}

func (b *FileBuffer) DeleteLine(l int) error {
	if l < 0 || l >= len(b.lines) {
		return fmt.Errorf("out of range")
	}
	copy(b.lines[l:], b.lines[l+1:])
	b.lines = b.lines[:len(b.lines)-1]
	return nil
}

func (b *FileBuffer) Truncate(count int) {
	b.lines = b.lines[:count]
}

func (b *FileBuffer) MergeLineWithPrevious(l int) error {
	if l < 1 || l >= len(b.lines) {
		return fmt.Errorf("out of range")
	}
	b.lines[l-1] = b.lines[l-1].MergeWith(b.lines[l])
	copy(b.lines[l:], b.lines[l+1:])
	b.lines = b.lines[:len(b.lines)-1]
	return nil
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
		return l, b.lines[l].Len()
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
	return len(b.lines) - 1, b.lines[len(b.lines)-1].Len()
}

func (b *FileBuffer) StyledLineIter(l, c int) StyledLineIter {
	return NewConstStyleLineIter(b.lines[l].Iter(c), DefaultStyle)
}

func (b *FileBuffer) Kind() string {
	return "plain"
}

func (b *FileBuffer) StringFromRegion(l0, c0, l1, c1 int) (string, error) {
	if l1 < l0 || (l0 == l1 && c1 < c0) {
		l0, c0, l1, c1 = l1, c1, l0, c0
	}
	var builder strings.Builder
	for l := l0; l <= l1; l++ {
		line, err := b.GetLine(l, -1)
		runes := line.Runes
		if err != nil {
			return "", err
		}
		if l == l1 && c1 < len(runes) {
			runes = runes[:c1+1]
		}
		if l == l0 {
			runes = runes[c0:]
		} else {
			builder.WriteByte('\n')
		}
		for _, r := range runes {
			builder.WriteRune(r)
		}
	}
	return builder.String(), nil
}

func splitString(s string) []string {
	return newLines.Split(s, -1)
}

var newLines = regexp.MustCompile(`(?s)\r\n|\n\r|\r`)
