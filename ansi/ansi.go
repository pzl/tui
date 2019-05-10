package ansi

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

const csi = "\x1b["

// ---- @TODO: change outputbuffer idea into more go-ish stream? Something that has Read() or Bytes() or String() or Write() at the end to perform the operation.. ?

/*
The ansi.OutputBuffer is the primary way to send out the control characters (aside from text effects). Calls are not printed immediately. They are queued up until buffer.Flush() is called.
*/
type OutputBuffer struct {
	buffer string
	sink   io.Writer
}

//Creates an output buffer that prints to Stderr. This will work fine with an application that prints its text to stdout.
func NewOutputBuffer() *OutputBuffer { return &OutputBuffer{sink: os.Stderr} }

//Creates an output buffer with a user-provided writer. This could be Stdout if you prefer that over Stderr. It could be a file (if you want ansi sequences in them). A stream, network, whatever.
func NewOutputBufferSink(w io.Writer) *OutputBuffer { return &OutputBuffer{sink: w} }

// Writes all queued calls immediately to the writer (stderr by default)
func (o *OutputBuffer) Flush() {
	if len(o.buffer) > 0 {
		fmt.Fprintf(o.sink, o.buffer)
		o.buffer = ""
	}
}

func (o *OutputBuffer) csi(s string) { o.write(csi + s) }

func (o *OutputBuffer) write(s string) {
	// handle non-displayable chars
	bytes := []byte(s)
	runes := []rune{}
	for len(bytes) > 0 {
		r, sz := utf8.DecodeRune(bytes)
		crlf := r == '\n' || r == '\r'
		if r >= 32 || r == '\x1b' || crlf {
			if r == utf8.RuneError {
				runes = append(runes, ' ') //skip
			} else {
				runes = append(runes, r)
			}
		}
		bytes = bytes[sz:]
	}
	o.buffer += string(runes)
}

/* -- actual commands -- */

// Movement

func (o *OutputBuffer) Up(n int)        { o.csi(strconv.Itoa(n) + "A") }
func (o *OutputBuffer) Down(n int)      { o.csi(strconv.Itoa(n) + "B") }
func (o *OutputBuffer) Right(n int)     { o.csi(strconv.Itoa(n) + "C") }
func (o *OutputBuffer) Left(n int)      { o.csi(strconv.Itoa(n) + "D") }
func (o *OutputBuffer) Origin()         { o.csi("H") }
func (o *OutputBuffer) MoveTo(x, y int) { o.csi(strconv.Itoa(y) + ";" + strconv.Itoa(x) + "H") } // note these are swapped
func (o *OutputBuffer) Column(n int)    { o.csi(strconv.Itoa(n) + "G") }

// Clearing

func (o *OutputBuffer) ClearLineRight() { o.csi("K") }
func (o *OutputBuffer) ClearLineLeft()  { o.csi("1K") }
func (o *OutputBuffer) ClearLine()      { o.csi("2K") }
func (o *OutputBuffer) ClearDown()      { o.csi("J") }
func (o *OutputBuffer) ClearUp()        { o.csi("1J") }
func (o *OutputBuffer) ClearAll()       { o.csi("2J") }

// Cursor

func (o *OutputBuffer) CursorHide()           { o.csi("?25l") }
func (o *OutputBuffer) CursorShow()           { o.csi("?25h") }
func (o *OutputBuffer) CursorSave()           { o.csi("s") }  // rarely supported
func (o *OutputBuffer) CursorRestore()        { o.csi("u") }  // rarely supported
func (o *OutputBuffer) CursorPosition()       { o.csi("6n") } // you gotta be ready to read here ...
func (o *OutputBuffer) CursorBlinker()        { o.csi("0 q") }
func (o *OutputBuffer) CursorSteady()         { o.csi("2 q") }
func (o *OutputBuffer) CursorUnderlineBlink() { o.csi("3 q") }
func (o *OutputBuffer) CursorUnderline()      { o.csi("4 q") }
func (o *OutputBuffer) CursorIBlink()         { o.csi("5 q") }
func (o *OutputBuffer) CursorI()              { o.csi("6 q") }

// Screen

type ScreenMode rune

const (
	Alt    ScreenMode = 'h'
	Normal ScreenMode = 'l'
)

//Change terminal to the "Alternate" screen and back. This is usually what full-screen TUIs do
func (o *OutputBuffer) Screen(m ScreenMode) { o.csi("?1049" + string(m)) }

type textEffect interface{ effect() string }

// Apply colors and/or styles via Effect.
//
// This is not required, you can usually just use the color and styles directly with the %s specifier
func Effect(t ...textEffect) string {
	s := make([]string, len(t))
	for i := range t {
		s[i] = t[i].effect()
	}
	return csi + strings.Join(s, ";") + "m"
}

func (o *OutputBuffer) Effect(t ...textEffect) { o.write(Effect(t...)) }

// styles, e.g. bold, underline, blink
type TextStyle int

const (
	Reset TextStyle = iota
	Bold
	Dim
	It
	Underline
	Blink
	FastBlink
	Reverse
	Hidden
	Strikethrough
	// rare and usually not supported
	Fraktur   TextStyle = 20
	Dunder    TextStyle = 21
	Framed    TextStyle = 51
	Encircled TextStyle = 52
	Overlined TextStyle = 53
)

// convenience function for printing the style directly, e.g. in a printf %s specifier
func (s TextStyle) String() string { return Style(s) }
func (s TextStyle) effect() string { return strconv.Itoa(int(s)) }

//Apply TextStyles directly to a string, that you can print as you desire
func Style(ts ...TextStyle) string {
	s := make([]string, len(ts))
	for i := range ts {
		s[i] = ts[i].effect()
	}
	return csi + strings.Join(s, ";") + "m"
}

func (o *OutputBuffer) Style(s ...TextStyle) { o.write(Style(s...)) }

/* ---------- Mouse -------- */

// Type of motion events the terminal can send us
type MouseMotion int

const (
	MotionNone   MouseMotion = 0
	MotionOnDrag MouseMotion = 2
	MotionAll    MouseMotion = 3
)

func getMotion(m MouseMotion) string {
	if int(m) < 0 || int(m) == 1 || int(m) > 3 {
		m = MotionNone
	}
	return "?100" + strconv.Itoa(int(m))
}

func (o *OutputBuffer) MouseEnable(m MouseMotion)  { o.csi(getMotion(m) + "h") }
func (o *OutputBuffer) MouseDisable(m MouseMotion) { o.csi(getMotion(m) + "l") }

// Color provider. Either the basic 16 colors, or 8-bit and 24-bit-truecolor
type colorable interface{ color() string }

// Returns the escape sequence string for the given color
func Color(c colorable) string { return c.color() }

// Standard 16 terminal colors
type BasicColor uint8

const (
	Black BasicColor = 30 + iota
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

// support for using BasicColor directly with %s specifier
func (c BasicColor) String() string { return csi + c.effect() + "m" }
func (c BasicColor) color() string  { return c.String() }           // colorable
func (c BasicColor) effect() string { return strconv.Itoa(int(c)) } // texteffect

func (o *OutputBuffer) Color(c colorable) { o.write(c.color()) }

//func (o *OutputBuffer) ColorBasicBg(b BasicColor) { o.csi(strconv.Itoa(int(b)+10) + "m") }

// Eight bit color. first 16 identical to BasicColor. 216 colors as color cube. 24 greyscale
// https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit
type EBColor uint8

func (e EBColor) String() string { return csi + e.effect() }
func (e EBColor) color() string  { return e.String() }                           //colorable
func (e EBColor) effect() string { return "38;5;" + strconv.Itoa(int(e)) + "m" } //texteffect
// ^ background is 48

// 24-bit True Color rendering. Terminal support for this is spotty. And detection is VERY hard
type TrueColor int

// convenience function for creating a color with (r,g,b)
func T(r uint8, g uint8, b uint8) TrueColor { return TrueColor(int(r)<<16 | int(g)<<8 | int(b)) }

func (t TrueColor) String() string { return csi + t.effect() }
func (t TrueColor) color() string  { return t.String() } //colorable
func (t TrueColor) effect() string {
	//bg is 48
	return "38;2;" + strconv.Itoa(int(t)>>16) + ";" + strconv.Itoa(int(t)>>8&0xff) + ";" + strconv.Itoa(int(t)&0xff) + "m"
}
