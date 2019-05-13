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

/*
The ansi.Writer is the primary use case for the ansi library. Text effects (including colors) can be accessed separately as strings. (see: `ansi.Efect()`)

You must create a writer with ansi.NewWriter() prior to use. A nil argument is accepted, and a new ansi.Writer will be created using os.Stderr as the default output. Commands are issued to the provided io.Writer immediately. In some situations, this is undesired as a user may see the cursor flash around the screen. It may be preferred to buffer the output and write it all at once. You can accomplish that with bufio.
*/
type Writer struct {
	w io.Writer
}

// Creates an ansi.Writer that will use the provided io.Writer. If nil is provided, Stderr is used. This could be any writer, however: Stdout if you prefer that over Stderr. It could be a file (if you want ansi sequences in them). A stream, network, whatever.
func NewWriter(w io.Writer) *Writer {
	if w == nil {
		w = os.Stderr
	}
	return &Writer{w: w}
}

func (w *Writer) csi(s string) { w.write(csi + s) }

func (w *Writer) write(s string) {
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
	fmt.Fprint(w.w, string(runes))
}

/* -- actual commands -- */

// Movement

func (w *Writer) Up(n int)        { w.csi(strconv.Itoa(n) + "A") }
func (w *Writer) Down(n int)      { w.csi(strconv.Itoa(n) + "B") }
func (w *Writer) Right(n int)     { w.csi(strconv.Itoa(n) + "C") }
func (w *Writer) Left(n int)      { w.csi(strconv.Itoa(n) + "D") }
func (w *Writer) Origin()         { w.csi("H") }
func (w *Writer) MoveTo(x, y int) { w.csi(strconv.Itoa(y) + ";" + strconv.Itoa(x) + "H") } // note these are swapped
func (w *Writer) Column(n int)    { w.csi(strconv.Itoa(n) + "G") }

// Clearing

func (w *Writer) ClearLineRight() { w.csi("K") }
func (w *Writer) ClearLineLeft()  { w.csi("1K") }
func (w *Writer) ClearLine()      { w.csi("2K") }
func (w *Writer) ClearDown()      { w.csi("J") }
func (w *Writer) ClearUp()        { w.csi("1J") }
func (w *Writer) ClearAll()       { w.csi("2J") }

// Cursor

func (w *Writer) CursorHide()           { w.csi("?25l") }
func (w *Writer) CursorShow()           { w.csi("?25h") }
func (w *Writer) CursorSave()           { w.csi("s") }  // rarely supported
func (w *Writer) CursorRestore()        { w.csi("u") }  // rarely supported
func (w *Writer) CursorPosition()       { w.csi("6n") } // you gotta be ready to read here ...
func (w *Writer) CursorBlinker()        { w.csi("0 q") }
func (w *Writer) CursorSteady()         { w.csi("2 q") }
func (w *Writer) CursorUnderlineBlink() { w.csi("3 q") }
func (w *Writer) CursorUnderline()      { w.csi("4 q") }
func (w *Writer) CursorIBlink()         { w.csi("5 q") }
func (w *Writer) CursorI()              { w.csi("6 q") }

// Screen

type ScreenMode rune

const (
	Alt    ScreenMode = 'h'
	Normal ScreenMode = 'l'
)

//Change terminal to the "Alternate" screen and back. This is usually what full-screen TUIs do
func (w *Writer) Screen(m ScreenMode) { w.csi("?1049" + string(m)) }

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

func (w *Writer) Effect(t ...textEffect) { w.write(Effect(t...)) }

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

func (w *Writer) Style(s ...TextStyle) { w.write(Style(s...)) }

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

func (w *Writer) MouseEnable(m MouseMotion)  { w.csi(getMotion(m) + "h") }
func (w *Writer) MouseDisable(m MouseMotion) { w.csi(getMotion(m) + "l") }

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

func (w *Writer) Color(c colorable) { w.write(c.color()) }

//func (w *Writer) ColorBasicBg(b BasicColor) { w.csi(strconv.Itoa(int(b)+10) + "m") }

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

/* ---- Duplicate things for top-level package ----- */

/*
func Up(n int)                   { NewWriter(nil).Up(n) }
func Down(n int)                 { NewWriter(nil).Down(n) }
func Right(n int)                { NewWriter(nil).Right(n) }
func Left(n int)                 { NewWriter(nil).Left(n) }
func Origin()                    { NewWriter(nil).Origin() }
func MoveTo(x, y int)            { NewWriter(nil).MoveTo(x, y) }
func Column(n int)               { NewWriter(nil).Column(n) }
func ClearLineRight()            { NewWriter(nil).ClearLineRight() }
func ClearLineLeft()             { NewWriter(nil).ClearLineLeft() }
func ClearLine()                 { NewWriter(nil).ClearLine() }
func ClearDown()                 { NewWriter(nil).ClearDown() }
func ClearUp()                   { NewWriter(nil).ClearUp() }
func ClearAll()                  { NewWriter(nil).ClearAll() }
func CursorHide()                { NewWriter(nil).CursorHide() }
func CursorShow()                { NewWriter(nil).CursorShow() }
func CursorSave()                { NewWriter(nil).CursorSave() }
func CursorRestore()             { NewWriter(nil).CursorRestore() }
func CursorPosition()            { NewWriter(nil).CursorPosition() }
func CursorBlinker()             { NewWriter(nil).CursorBlinker() }
func CursorSteady()              { NewWriter(nil).CursorSteady() }
func CursorUnderlineBlink()      { NewWriter(nil).CursorUnderlineBlink() }
func CursorUnderline()           { NewWriter(nil).CursorUnderline() }
func CursorIBlink()              { NewWriter(nil).CursorIBlink() }
func CursorI()                   { NewWriter(nil).CursorI() }
func Screen(m ScreenMode)        { NewWriter(nil).Screen(m) }
func MouseEnable(m MouseMotion)  { NewWriter(nil).MouseEnable(m) }
func MouseDisable(m MouseMotion) { NewWriter(nil).MouseDisable(m) }
*/
