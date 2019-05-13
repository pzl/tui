/*
This is a tiny editor in ~200 lines
It is a gross half-breed of nano and vim.
It's almost entirely nano, but ESC puts you in command mode
where you can move with arrows OR hjkl.
i for insert mode, where you can just type as normal like nano.
Paste seems to work.
Scrolling does *NOT*. So don't write more lines than you have terminal lines.
It also doesn't save or read anything. It's just a UI scratch pad.
*/
package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pzl/tui"
	"github.com/pzl/tui/ansi"
)

const MOTION = ansi.MotionOnDrag

func main() {
	e := NewEditor()
	defer e.Cleanup()
	err := e.Start()

	if err != nil {
		panic(err)
	}
}

type EditMode int

const (
	ModeCommand EditMode = iota
	ModeInsert
)

type Editor struct {
	scr         *ansi.Writer
	mode        EditMode
	buf         [][]byte
	curLine     int
	curX        int
	w           int
	h           int
	done        chan struct{}
	inputDone   func()
	restoreTerm func() error
}

func NewEditor() *Editor {
	w, h := tui.TermSize(0) // fd 0 = stdin

	buf := make([][]byte, 1)
	buf[0] = newline()

	return &Editor{
		scr:     ansi.NewWriter(nil),
		mode:    ModeCommand,
		curLine: 0,
		curX:    0,
		w:       w,
		h:       h,
		done:    make(chan struct{}),
		buf:     buf,
	}
}

func (e *Editor) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	e.inputDone = cancel
	events, restore, err := tui.GetInput(ctx, int(os.Stdin.Fd()))
	e.restoreTerm = restore
	if err != nil {
		return err
	}

	e.scr.Screen(ansi.Alt)
	e.scr.Origin()
	e.scr.MouseEnable(MOTION)

	for {
		select {
		case ev := <-events:
			e.handle(ev)
			e.Redraw()
		case <-e.done:
			return nil
		}
	}
}

func (e *Editor) Redraw() {
	s := bytes.Join(e.buf, []byte("\r\n"))

	// clear & print
	e.scr.ClearAll()
	e.scr.Origin()
	fmt.Print(string(s))

	// status bar
	e.scr.MoveTo(0, e.h-1)
	fmt.Printf("% 3d, % 3d  lines:%03d%s", e.curLine, e.curX, len(e.buf), strings.Repeat(" ", 40))

	// place cursor
	if e.mode == ModeInsert {
		e.scr.MoveTo(e.curX+1, e.curLine+1)
	} else {
		e.scr.MoveTo(e.curX, e.curLine+1)
	}
}

func (e *Editor) Cleanup() {
	if e.inputDone != nil {
		e.inputDone()
	}
	if e.restoreTerm != nil {
		e.restoreTerm()
	}

	e.scr.MouseDisable(MOTION)
	e.scr.Screen(ansi.Normal)
	e.scr.CursorBlinker()
	e.scr.CursorShow()
}

func (e *Editor) handle(ev tui.Event) {
	// same regardless of mode
	switch ev.Key {
	case tui.ESC:
		e.mode = ModeCommand
		e.scr.CursorBlinker()
		return
	case tui.CtrlC:
		close(e.done)
		return
	case tui.Up, tui.Down, tui.Left, tui.Right:
		e.move(ev.Key)
		return
	}

	switch e.mode {
	case ModeInsert:
		switch ev.Key {
		case tui.CtrlJ, tui.CtrlM:
			if e.curX == 0 && len(e.buf[e.curLine]) != 0 { // sitting at the start of a non-blank line
				e.insertLine(newline(), e.curLine)
			} else if e.atEOL() { // simple add a line
				e.insertLine(newline(), e.curLine+1)
				e.curX = 0
			} else { // enter in the middle of a line
				// add new line with existing content
				n := make([]byte, len(e.buf[e.curLine][e.curX:]))
				copy(n, e.buf[e.curLine][e.curX:])
				e.insertLine(n, e.curLine+1)
				e.buf[e.curLine] = e.buf[e.curLine][:e.curX] // truncate current line
				e.curX = 0
			}
			e.curLine++
		case tui.Tab:
			e.insertBytes([]byte("  ")) // bake-in 2-spaces as tab?
		case tui.BSpace:
			if e.curLine == 0 && e.curX == 0 {
				return
			}
			if e.curX == 0 && len(e.buf[e.curLine]) == 0 {
				// delete empty line
				e.delLine(e.curLine)
				e.curLine--
				e.curX = len(e.buf[e.curLine])
			} else if e.curX == 0 {
				//combine lines
				x := len(e.buf[e.curLine-1])
				e.buf[e.curLine-1] = append(e.buf[e.curLine-1], e.buf[e.curLine]...)
				e.delLine(e.curLine)

				//update indexes
				e.curLine--
				e.curX = x
			} else {
				e.buf[e.curLine] = append(e.buf[e.curLine][:e.curX-1], e.buf[e.curLine][e.curX:]...)
				e.curX--
			}
		}

		if ev.Type == tui.KeyPrint {
			e.insertBytes([]byte(string(ev.Key)))
		}
	case ModeCommand:
		switch ev.Key {
		case 'i':
			e.mode = ModeInsert
			e.scr.CursorIBlink()
		case 'h', 'j', 'k', 'l':
			e.move(ev.Key)
		}
	}
}

func (e *Editor) move(key rune) {
	switch key {
	case tui.Up, 'k':
		if e.curLine > 0 {
			if len(e.buf[e.curLine-1]) < e.curX {
				e.curX = len(e.buf[e.curLine-1])
			}
			e.curLine--
		}
	case tui.Down, 'j':
		if len(e.buf)-1 > e.curLine {
			if len(e.buf[e.curLine+1]) < e.curX {
				e.curX = len(e.buf[e.curLine+1])
			}
			e.curLine++
		}
	case tui.Left, 'h':
		if e.curX > 0 {
			e.curX--
		}
	case tui.Right, 'l':
		if len(e.buf[e.curLine]) > e.curX {
			e.curX++
		}
	}
}
func (e *Editor) atEOL() bool { return e.curX == len(e.buf[e.curLine]) }
func (e *Editor) insertBytes(b []byte) {
	if e.atEOL() {
		e.buf[e.curLine] = append(e.buf[e.curLine], b...)
	} else {
		e.buf[e.curLine] = append(e.buf[e.curLine][:e.curX], append(b, e.buf[e.curLine][e.curX:]...)...)
	}
	e.curX += len(b)
}

// https://github.com/golang/go/wiki/SliceTricks#insert
func (e *Editor) insertLine(line []byte, at int) {
	e.buf = append(e.buf[:at], append([][]byte{line}, e.buf[at:]...)...) //@todo: leaves garbage
}
func (e *Editor) delLine(at int) { e.buf = append(e.buf[:at], e.buf[at+1:]...) } //@todo: leaves garbage

func newline() []byte { return make([]byte, 0, 80) }
