package ansi

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func sinkbuf() (*bytes.Buffer, *OutputBuffer) {
	var buf bytes.Buffer
	op := NewOutputBufferSink(&buf)
	return &buf, op
}

func TestOutputBufferCreation(t *testing.T) {
	op := NewOutputBuffer()
	assert.NotNil(t, op)
}

func TestOutputBufferSinkCreation(t *testing.T) {
	var buf bytes.Buffer
	op := NewOutputBufferSink(&buf)
	assert.NotNil(t, op)
}

func TestOperations(t *testing.T) {
	buf, op := sinkbuf()
	tests := map[string]struct {
		f    func()
		want string
	}{
		"Origin":               {f: op.Origin, want: "\x1b[H"},
		"ClearLineRight":       {f: op.ClearLineRight, want: "\x1b[K"},
		"ClearLineLeft":        {f: op.ClearLineLeft, want: "\x1b[1K"},
		"ClearLine":            {f: op.ClearLine, want: "\x1b[2K"},
		"ClearDown":            {f: op.ClearDown, want: "\x1b[J"},
		"ClearUp":              {f: op.ClearUp, want: "\x1b[1J"},
		"ClearAll":             {f: op.ClearAll, want: "\x1b[2J"},
		"CursorHide":           {f: op.CursorHide, want: "\x1b[?25l"},
		"CursorShow":           {f: op.CursorShow, want: "\x1b[?25h"},
		"CursorSave":           {f: op.CursorSave, want: "\x1b[s"},
		"CursorRestore":        {f: op.CursorRestore, want: "\x1b[u"},
		"CursorPosition":       {f: op.CursorPosition, want: "\x1b[6n"},
		"CursorBlinker":        {f: op.CursorBlinker, want: "\x1b[0 q"},
		"CursorSteady":         {f: op.CursorSteady, want: "\x1b[2 q"},
		"CursorUnderlineBlink": {f: op.CursorUnderlineBlink, want: "\x1b[3 q"},
		"CursorUnderline":      {f: op.CursorUnderline, want: "\x1b[4 q"},
		"CursorIBlink":         {f: op.CursorIBlink, want: "\x1b[5 q"},
		"CursorI":              {f: op.CursorI, want: "\x1b[6 q"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			buf.Reset()
			tc.f()
			op.Flush()
			assert.Equal(t, tc.want, buf.String())
		})
	}
}

func TestMovement(t *testing.T) {
	buf, op := sinkbuf()
	tests := map[string]struct {
		f    func(int)
		n    int
		want string
	}{
		"Up":    {f: op.Up, n: 4, want: "\x1b[4A"},
		"Down":  {f: op.Down, n: 3, want: "\x1b[3B"},
		"Left":  {f: op.Left, n: 100, want: "\x1b[100D"},
		"Right": {f: op.Right, n: 20, want: "\x1b[20C"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			buf.Reset()
			tc.f(tc.n)
			op.Flush()
			assert.Equal(t, tc.want, buf.String())
		})
	}
}

func TestMoveTo(t *testing.T) {
	buf, op := sinkbuf()

	tests := []struct {
		x    int
		y    int
		want string
	}{
		{x: 0, y: 0, want: "\x1b[0;0H"},
		{x: 1, y: 0, want: "\x1b[0;1H"},
		{x: 0, y: 1, want: "\x1b[1;0H"},
		{x: 20, y: 18, want: "\x1b[18;20H"},
		{x: 100, y: 9, want: "\x1b[9;100H"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%d,%d", tc.x, tc.y), func(t *testing.T) {
			buf.Reset()
			op.MoveTo(tc.x, tc.y)
			op.Flush()
			assert.Equal(t, tc.want, buf.String())
		})
	}

}

func TestColumn(t *testing.T) {
	buf, op := sinkbuf()

	tests := []struct {
		n    int
		want string
	}{
		{n: 0, want: "\x1b[0G"},
		{n: 1, want: "\x1b[1G"},
		{n: 20, want: "\x1b[20G"},
		{n: 100, want: "\x1b[100G"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%d", tc.n), func(t *testing.T) {
			buf.Reset()
			op.Column(tc.n)
			op.Flush()
			assert.Equal(t, tc.want, buf.String())
		})
	}

}

func TestBufferMultiple(t *testing.T) {
	buf, op := sinkbuf()

	op.ClearAll()
	assert.Empty(t, buf.String())

	op.Origin()
	assert.Empty(t, buf.String())

	op.CursorIBlink()
	assert.Empty(t, buf.String())

	buf.Reset() // clear just in case

	op.Flush()
	assert.Equal(t, "\x1b[2J\x1b[H\x1b[5 q", buf.String())
}

func TestMouseEnable(t *testing.T) {
	buf, op := sinkbuf()

	tests := []struct {
		name string
		mode MouseMotion
		want string
	}{
		{name: "none", mode: MotionNone, want: "\x1b[?1000h"},
		{name: "onDrag", mode: MotionOnDrag, want: "\x1b[?1002h"},
		{name: "All", mode: MotionAll, want: "\x1b[?1003h"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			op.MouseEnable(tc.mode)
			op.Flush()
			assert.Equal(t, tc.want, buf.String())
		})
	}
}

func TestMouseDisable(t *testing.T) {
	buf, op := sinkbuf()

	tests := []struct {
		name string
		mode MouseMotion
		want string
	}{
		{name: "none", mode: MotionNone, want: "\x1b[?1000l"},
		{name: "onDrag", mode: MotionOnDrag, want: "\x1b[?1002l"},
		{name: "All", mode: MotionAll, want: "\x1b[?1003l"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			op.MouseDisable(tc.mode)
			op.Flush()
			assert.Equal(t, tc.want, buf.String())
		})
	}
}

func TestScreenModeAlt(t *testing.T) {
	buf, op := sinkbuf()
	op.Screen(Alt)
	op.Flush()
	assert.Equal(t, "\x1b[?1049h", buf.String())
}

func TestScreenModeNormal(t *testing.T) {
	buf, op := sinkbuf()
	op.Screen(Normal)
	op.Flush()
	assert.Equal(t, "\x1b[?1049l", buf.String())
}

func TestBufStyle(t *testing.T) {
	buf, op := sinkbuf()
	op.Style(Underline)
	op.Flush()
	assert.Equal(t, "\x1b[4m", buf.String())
}

func TestBufStyleMultiple(t *testing.T) {
	buf, op := sinkbuf()
	op.Style(Underline, Blink)
	op.Flush()
	assert.Equal(t, "\x1b[4;5m", buf.String())
}

func TestBufColorBasic(t *testing.T) {
	buf, op := sinkbuf()
	op.Color(Red)
	op.Flush()
	assert.Equal(t, "\x1b[31m", buf.String())
}

func TestBufColor8Bit(t *testing.T) {
	buf, op := sinkbuf()
	op.Color(EBColor(212))
	op.Flush()
	assert.Equal(t, "\x1b[38;5;212m", buf.String())
}

func TestBufColor24Bit(t *testing.T) {
	buf, op := sinkbuf()
	op.Color(T(38, 50, 127))
	op.Flush()
	assert.Equal(t, "\x1b[38;2;38;50;127m", buf.String())
}
