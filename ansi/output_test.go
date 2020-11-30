package ansi

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func writer() (*bytes.Buffer, *Writer) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	return &buf, w
}

func TestWriterCreateNil(t *testing.T) {
	w := NewWriter(nil)
	assert.NotNil(t, w)
}

func TestWriterWithOutput(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	assert.NotNil(t, w)
}

func TestOperations(t *testing.T) {
	out, w := writer()
	tests := map[string]struct {
		f    func()
		want string
	}{
		"Origin":               {f: w.Origin, want: "\x1b[H"},
		"ClearLineRight":       {f: w.ClearLineRight, want: "\x1b[K"},
		"ClearLineLeft":        {f: w.ClearLineLeft, want: "\x1b[1K"},
		"ClearLine":            {f: w.ClearLine, want: "\x1b[2K"},
		"ClearDown":            {f: w.ClearDown, want: "\x1b[J"},
		"ClearUp":              {f: w.ClearUp, want: "\x1b[1J"},
		"ClearAll":             {f: w.ClearAll, want: "\x1b[2J"},
		"CursorHide":           {f: w.CursorHide, want: "\x1b[?25l"},
		"CursorShow":           {f: w.CursorShow, want: "\x1b[?25h"},
		"CursorSave":           {f: w.CursorSave, want: "\x1b[s"},
		"CursorRestore":        {f: w.CursorRestore, want: "\x1b[u"},
		"CursorPosition":       {f: w.CursorPosition, want: "\x1b[6n"},
		"CursorBlinker":        {f: w.CursorBlinker, want: "\x1b[0 q"},
		"CursorSteady":         {f: w.CursorSteady, want: "\x1b[2 q"},
		"CursorUnderlineBlink": {f: w.CursorUnderlineBlink, want: "\x1b[3 q"},
		"CursorUnderline":      {f: w.CursorUnderline, want: "\x1b[4 q"},
		"CursorIBlink":         {f: w.CursorIBlink, want: "\x1b[5 q"},
		"CursorI":              {f: w.CursorI, want: "\x1b[6 q"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			out.Reset()
			tc.f()
			assert.Equal(t, tc.want, out.String())
		})
	}
}

func TestMovement(t *testing.T) {
	out, w := writer()
	tests := map[string]struct {
		f    func(int)
		n    int
		want string
	}{
		"Up":    {f: w.Up, n: 4, want: "\x1b[4A"},
		"Down":  {f: w.Down, n: 3, want: "\x1b[3B"},
		"Left":  {f: w.Left, n: 100, want: "\x1b[100D"},
		"Right": {f: w.Right, n: 20, want: "\x1b[20C"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			out.Reset()
			tc.f(tc.n)
			assert.Equal(t, tc.want, out.String())
		})
	}
}

func TestMoveTo(t *testing.T) {
	out, w := writer()

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
			out.Reset()
			w.MoveTo(tc.x, tc.y)
			assert.Equal(t, tc.want, out.String())
		})
	}

}

func TestColumn(t *testing.T) {
	out, w := writer()

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
			out.Reset()
			w.Column(tc.n)
			assert.Equal(t, tc.want, out.String())
		})
	}

}

func TestBuffered(t *testing.T) {
	var out bytes.Buffer // stand-in for stdout or err

	buf := bufio.NewWriterSize(&out, 8192)
	w := NewWriter(buf)

	w.ClearAll()
	assert.Empty(t, out.String())
	assert.True(t, buf.Buffered() > 0)
	n := buf.Buffered()

	w.Origin()
	assert.Empty(t, out.String())
	assert.True(t, buf.Buffered() > n)
	n = buf.Buffered()

	w.CursorIBlink()
	assert.Empty(t, out.String())
	assert.True(t, buf.Buffered() > n)

	buf.Flush()
	assert.Equal(t, "\x1b[2J\x1b[H\x1b[5 q", out.String())
}

func TestMouseEnable(t *testing.T) {
	out, w := writer()

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
			out.Reset()
			w.MouseEnable(tc.mode)
			assert.Equal(t, tc.want, out.String())
		})
	}
}

func TestMouseDisable(t *testing.T) {
	out, w := writer()

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
			out.Reset()
			w.MouseDisable(tc.mode)
			assert.Equal(t, tc.want, out.String())
		})
	}
}

func TestScreenModeAlt(t *testing.T) {
	out, w := writer()
	w.Screen(Alt)
	assert.Equal(t, "\x1b[?1049h", out.String())
}

func TestScreenModeNormal(t *testing.T) {
	out, w := writer()
	w.Screen(Normal)
	assert.Equal(t, "\x1b[?1049l", out.String())
}

func TestBufStyle(t *testing.T) {
	out, w := writer()
	w.Style(Underline)
	assert.Equal(t, "\x1b[4m", out.String())
}

func TestBufStyleMultiple(t *testing.T) {
	out, w := writer()
	w.Style(Underline, Blink)
	assert.Equal(t, "\x1b[4;5m", out.String())
}

func TestBufColorBasic(t *testing.T) {
	out, w := writer()
	w.Color(Red)
	assert.Equal(t, "\x1b[31m", out.String())
}

func TestBufColor8Bit(t *testing.T) {
	out, w := writer()
	w.Color(EBColor(212))
	assert.Equal(t, "\x1b[38;5;212m", out.String())
}

func TestBufColor24Bit(t *testing.T) {
	out, w := writer()
	w.Color(T(38, 50, 127))
	assert.Equal(t, "\x1b[38;2;38;50;127m", out.String())
}

func TestCursorCmds(t *testing.T) {
	assert.Equal(t, "\x1b[A", Up.String())

	tests := map[string]struct {
		cmd CursorCmd
		val string
	}{
		"Up":                   {cmd: Up, val: "A"},
		"Down":                 {cmd: Down, val: "B"},
		"Right":                {cmd: Right, val: "C"},
		"Left":                 {cmd: Left, val: "D"},
		"Origin":               {cmd: Origin, val: "H"},
		"Column0":              {cmd: Column0, val: "G"},
		"ClearLineRight":       {cmd: ClearLineRight, val: "K"},
		"ClearLineLeft":        {cmd: ClearLineLeft, val: "1K"},
		"ClearLine":            {cmd: ClearLine, val: "2K"},
		"ClearDown":            {cmd: ClearDown, val: "J"},
		"ClearUp":              {cmd: ClearUp, val: "1J"},
		"ClearAll":             {cmd: ClearAll, val: "2J"},
		"CursorHide":           {cmd: CursorHide, val: "?25l"},
		"CursorShow":           {cmd: CursorShow, val: "?25h"},
		"CursorSave":           {cmd: CursorSave, val: "s"},
		"CursorRestore":        {cmd: CursorRestore, val: "u"},
		"CursorPosition":       {cmd: CursorPosition, val: "6n"},
		"CursorBlinker":        {cmd: CursorBlinker, val: "0 q"},
		"CursorSteady":         {cmd: CursorSteady, val: "2 q"},
		"CursorUnderlineBlink": {cmd: CursorUnderlineBlink, val: "3 q"},
		"CursorUnderline":      {cmd: CursorUnderline, val: "4 q"},
		"CursorIBlink":         {cmd: CursorIBlink, val: "5 q"},
		"CursorI":              {cmd: CursorI, val: "6 q"},
		"ScreenModeAlt":        {cmd: ScreenModeAlt, val: "?1049h"},
		"ScreenModenNormal":    {cmd: ScreenModenNormal, val: "?1049l"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, "\x1b["+tc.val, tc.cmd.String())
		})
	}

}
