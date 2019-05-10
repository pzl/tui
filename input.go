package tui

import (
	"context"
	"fmt"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/ssh/terminal"
)

const (
	// Basic keycodes, map directly to their ANSI numbering
	// order strictly matters
	Null rune = iota // ctrl-space
	CtrlA
	CtrlB
	CtrlC
	CtrlD // aka EOF
	CtrlE
	CtrlF
	CtrlG
	CtrlH // \b  backsp But actual backspace is 127. CtrlBackspace is 8
	Tab   // \t aka Ctrl-I
	CtrlJ // \n, sometimes enter
	CtrlK
	CtrlL
	CtrlM // aka \r, also sometimes enter
	CtrlN
	CtrlO
	CtrlP
	CtrlQ
	CtrlR
	CtrlS
	CtrlT
	CtrlU
	CtrlV
	CtrlW
	CtrlX
	CtrlY
	CtrlZ
	ESC             // ctrl-[
	CtrlFwdSlash    // ctrl-\
	CtrlBackBracket // ctrl-]
	CtrlCaret       // ctrl-6  (^)
	CtrlUnderscore  // shift-ctrl -

	/*
		32-126 are the printable keyboard keys
		For regular pressing of these keys, the
		following table list isn't used. The keys
		are returned as KeyPrintable with their rune
		value. Where KeySpecial for the same rune value,
		it is equivalent to Alt[key].
	*/

	AltSpace // [ESC, 32] = Alt + space
	AltBang  // [ESC, 33] i.e. shift-alt-1
	AltDQuo
	AltHash
	AltDollar
	AltPercent
	AltAmpersand
	AltSQuo
	AltOpenParen
	AltCloseParen
	AltStar
	AltPlus
	AltComma
	AltMinus
	AltPeriod
	AltSlash
	Alt0
	Alt1
	Alt2
	Alt3
	Alt4
	Alt5
	Alt6
	Alt7
	Alt8
	Alt9
	AltColon
	AltSemicolon
	AltLT
	AltEqual
	AltGT
	AltQuestion
	AltAt
	AltA
	AltB
	AltC
	AltD
	AltE
	AltF
	AltG
	AltH
	AltI
	AltJ
	AltK
	AltL
	AltM
	AltN
	AltO
	AltP
	AltQ
	AltR
	AltS
	AltT
	AltU
	AltV
	AltW
	AltX
	AltY
	AltZ
	AltOpenBracket
	AltFwdSlash
	AltCloseBracket
	AltCaret
	AltUnderscore
	AltGrave
	Alta
	Altb
	Altc
	Altd
	Alte
	Altf
	Altg
	Alth
	Alti
	Altj
	Altk
	Altl
	Altm
	Altn
	Alto
	Altp
	Altq
	Altr
	Alts
	Altt
	Altu
	Altv
	Altw
	Altx
	Alty
	Altz
	AltOpenCurly
	AltPipe
	AltCloseCurly
	AltTilde
	AltBS
	//End printable

	// actual number values below are arbitrary, are not in order.

	// Extended keyCodes
	Del
	AltDel

	BTab
	BSpace

	PgUp
	PgDn

	Up
	Down
	Right
	Left
	Home
	End

	// /relative/ order of the next 8 matter
	SUp // Shift
	SDown
	SRight
	SLeft
	CtrlUp
	CtrlDown
	CtrlRight
	CtrlLeft

	// don't actually need to be in order
	F1
	F2
	F3
	F4
	F5
	F6
	F7
	F8
	F9
	F10
	F11
	F12

	//sequential calculated. Keep in relative order
	CtrlAlta
	CtrlAltb
	CtrlAltc
	CtrlAltd
	CtrlAlte
	CtrlAltf
	CtrlAltg
	CtrlAlth
	CtrlAlti
	CtrlAltj
	CtrlAltk
	CtrlAltl
	CtrlAltm
	CtrlAltn
	CtrlAlto
	CtrlAltp
	CtrlAltq
	CtrlAltr
	CtrlAlts
	CtrlAltt
	CtrlAltu
	CtrlAltv
	CtrlAltw
	CtrlAltx
	CtrlAlty
	CtrlAltz
)

type EvType uint8

const (
	EventInvalid EvType = iota
	KeySpecial
	KeyPrint
	Mouse
)

/*
How to determine mouse action:
	Mousedown: Type=Mouse && Btn != 3 && !Motion
	Mouseup:   Type=Mouse && Btn == 3 && !Motion
	Mousedrag: Type=Mouse && Btn != 3 && Motion
	Mousemove: Type=Mouse && Btn == 3 && Motion
	ScrollUp:  Type=Mouse && Btn=4
	ScrollDn:  Type=Mouse && Btn=5
*/
type MouseEvent struct {
	Y      int
	X      int
	Btn    int // 0=Primary, 1=Middle, 2=Right, 3=Release, 4=ScrUp, 5=ScrDown
	Shift  bool
	Meta   bool
	Ctrl   bool
	Motion bool
	buf    []byte
}

type Event struct {
	Type EvType
	Key  rune
	M    *MouseEvent
}

// returns true if a specific character int(rune) is a printable character (alphanumeric, punctuation)
func Printable(i int) bool { return i >= 32 && i <= 126 }

/*
Grabs Stdin(or whatever passed fd) to listen for keyboard input. Returns 3 things:
 - a channel to listen for key events on
 - a terminal restore function. Always safe to call, especially when error is set
 - error condition

This is the primary use of the top-level tui package, if you intend to capture input, or mouse events
*/
func GetInput(ctx context.Context, fd int) (<-chan Event, func() error, error) {
	ch := make(chan Event, 1000)
	st, err := terminal.GetState(fd)
	if err != nil {
		return nil, func() error { return nil }, err
	}
	restore := func() error { return terminal.Restore(fd, st) }

	_, err = terminal.MakeRaw(fd)
	if err != nil {
		return nil, restore, err
	}

	ib := inputBuf{b: make([]byte, 0, 9)}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case ev := <-ib.readEvent(fd):
				ch <- ev
			}
		}
	}()
	return ch, restore, nil
}

type inputBuf struct {
	b  []byte
	mu sync.Mutex
}

func (ib *inputBuf) readEvent(fd int) <-chan Event {
	ch := make(chan Event)

	go func() {
		ib.mu.Lock()
		defer func() {
			ib.mu.Unlock()
		}()

		for ; len(ib.b) == 0; ib.b = fillBuf(fd, ib.b) {
		}

		if len(ib.b) == 0 {
			close(ch)
			return
		}

		sz := 1
		defer func() {
			ib.b = ib.b[sz:]
		}()

		switch ib.b[0] {
		case byte(CtrlC), byte(CtrlG), byte(CtrlQ):
			ch <- Event{KeySpecial, rune(ib.b[0]), nil}
			return
		case 127:
			ch <- Event{KeySpecial, BSpace, nil}
			return
		case 0:
			ch <- Event{KeySpecial, Null, nil} // Ctrl-space?
			return
		case byte(ESC):
			ch <- ib.escSequence(&sz)
			return
		}

		if ib.b[0] < 32 { // Ctrl-A_Z
			ch <- Event{KeySpecial, rune(ib.b[0]), nil}
			return
		}
		char, rsz := utf8.DecodeRune(ib.b)
		if char == utf8.RuneError {
			ch <- Event{KeySpecial, ESC, nil}
			return
		}
		sz = rsz
		ch <- Event{KeyPrint, char, nil}
	}()
	return ch
}

/*
 * Gets first byte, blocking to do so.
 * Tries to get any extra bytes within a 100ms timespan
 * like esc key sequences (arrows, etc)
 *
 */
func fillBuf(fd int, buf []byte) []byte {
	const pollInt = 5 //ms
	const span = 100  //ms -- reflected via retries*pollInt

	c, ok := getchar(fd, false)
	if !ok {
		return buf
	}
	buf = append(buf, byte(c))
	retries := 0
	if c == int(ESC) {
		retries = span / pollInt // 20
	}

	pc := c
	for {
		c, ok := getchar(fd, true)
		if !ok {
			if retries > 0 {
				retries--
				time.Sleep(pollInt * time.Millisecond)
				continue
			}
			break
		} else if c == int(ESC) && pc != c {
			retries = span / pollInt // got the next char, keep going
		} else {
			retries = 0
		}
		buf = append(buf, byte(c))
		pc = c
	}
	return buf
}

func getchar(fd int, nonblock bool) (int, bool) {
	b := make([]byte, 1)
	err := syscall.SetNonblock(fd, nonblock)
	if err != nil {
		return 0, false
	}
	if n, err := syscall.Read(fd, b); err != nil || n < 1 {
		return 0, false
	}
	return int(b[0]), true
}

//@todo: more shift/ctrl/alt of extended keys like Home, F#, PgUp
//http://www.manmrk.net/tutorials/ISPF/XE/xehelp/html/HID00000594.htm
//this is the ugliest, code ever. to check the seemingly most random
//assignment of codes to meaningful keys
func (ib *inputBuf) escSequence(sz *int) Event {
	if len(ib.b) < 2 {
		return Event{KeySpecial, ESC, nil}
	}

	*sz = 2

	switch ib.b[1] {
	case byte(ESC):
		return Event{KeySpecial, ESC, nil}
	case 127:
		return Event{KeySpecial, AltBS, nil}
	case 91, 79: // [, O
		if len(ib.b) < 3 {
			if ib.b[1] == '[' {
				return Event{KeySpecial, AltOpenBracket, nil}
			} else if ib.b[1] == 'O' {
				return Event{KeySpecial, AltO, nil}
			}
			return debugEv(ib.b)
		}
		*sz = 3
		switch ib.b[2] {
		case 65:
			return Event{KeySpecial, Up, nil}
		case 66:
			return Event{KeySpecial, Down, nil}
		case 67:
			return Event{KeySpecial, Right, nil}
		case 68:
			return Event{KeySpecial, Left, nil}
		case 90:
			return Event{KeySpecial, BTab, nil}
		case 72:
			return Event{KeySpecial, Home, nil}
		case 70:
			return Event{KeySpecial, End, nil}
		case 77:
			return ib.mouseSequence(sz)
		case 80:
			return Event{KeySpecial, F1, nil}
		case 81:
			return Event{KeySpecial, F2, nil}
		case 82:
			return Event{KeySpecial, F3, nil}
		case 49, 50, 51, 52, 53, 54:
			if len(ib.b) < 4 {
				return debugEv(ib.b)
			}
			*sz = 4
			switch ib.b[2] {
			case 50:
				if len(ib.b) == 5 && ib.b[4] == 126 {
					*sz = 5
					switch ib.b[3] {
					case 48:
						return Event{KeySpecial, F9, nil}
					case 49:
						return Event{KeySpecial, F10, nil}
					// @todo: WTF does 50 mean?
					case 51:
						return Event{KeySpecial, F11, nil}
					case 52:
						return Event{KeySpecial, F12, nil}
					}
				} else if ib.b[3] == '0' && (ib.b[4] == '0' || ib.b[4] == '1') && ib.b[5] == '~' {
					// bracketed paste mode. \e[200~ .. \e[201~
					//discard seq from buffer
					ib.b = ib.b[6:]
					*sz = 0
					return Event{KeySpecial, Null, nil}
				}
				return debugEv(ib.b)
			case 51:
				if len(ib.b) >= 6 && ib.b[3] == 59 && ib.b[4] == 51 && ib.b[5] == '~' {
					*sz = 6
					return Event{KeySpecial, AltDel, nil}
				}
				return Event{KeySpecial, Del, nil}
			case 52:
				return Event{KeySpecial, End, nil}
			case 53:
				return Event{KeySpecial, PgUp, nil}
			case 54:
				return Event{KeySpecial, PgDn, nil}
			case 49: //'1'
				switch ib.b[3] {
				case 126:
					return Event{KeySpecial, Home, nil}
				case 53, 55, 56, 57:
					if len(ib.b) == 5 && ib.b[4] == 126 {
						*sz = 5
						switch ib.b[3] {
						case 53:
							return Event{KeySpecial, F5, nil}
						case 55:
							return Event{KeySpecial, F6, nil}
						case 56:
							return Event{KeySpecial, F7, nil}
						case 57:
							return Event{KeySpecial, F8, nil}
						}
					}
					return debugEv(ib.b)
				case ';': //59
					if len(ib.b) != 6 {
						return debugEv(ib.b)
					}
					*sz = 6
					if ib.b[4] != '2' && ib.b[4] != '5' {
						return debugEv(ib.b)
					}
					if ib.b[5] < 'A' || ib.b[5] > 'D' {
						return debugEv(ib.b)
					}

					//ESC[1;2A == shift-up
					//ESC[1;5A == ctrl-up
					k := SUp
					if ib.b[4] == '5' { // move to up Ctrl*
						k += 4
					}
					k += rune(int(ib.b[5]) - int('A')) // set arrow direction
					return Event{KeySpecial, k, nil}
				}
			}
		}
	}

	// ESC-0 ~ ESC-26 == ctrl-alt-[key]
	if ib.b[1] >= 1 && ib.b[1] <= 'z'-'a'+1 {
		return Event{KeySpecial, rune(int(CtrlAlta) + int(ib.b[1]) - 1), nil}
	}

	// ESC-32 ~ ESC-126 == alt-[key]
	if Printable(int(ib.b[1])) {
		return Event{KeySpecial, rune(ib.b[1]), nil}
	}

	return debugEv(ib.b)
}

// mouse stuff

func debugEv(buf []byte) Event {
	b := make([]byte, len(buf))
	copy(b, buf)
	return Event{EventInvalid, 0, &MouseEvent{0, 0, 0, false, false, false, false, b}}
}

// https://www.xfree86.org/current/ctlseqs.html#Mouse%20Tracking
// \x1b[M<button><x+33><y+33>
func (ib *inputBuf) mouseSequence(sz *int) Event {
	b := make([]byte, len(ib.b))
	copy(b, ib.b)
	if len(ib.b) < 6 {
		fmt.Printf("short mouse seq: %v\n", ib.b)
		return debugEv(ib.b)
	}
	*sz = 6

	evCode := int(b[3] - 32)

	bNum := evCode & 0x3 // low two bits, 00=MB1, 01=MB2, 10=MB3, 11=Release
	if evCode&(1<<6) != 0 {
		bNum += 4 // scroll buttons set a high bit (+32)
	}
	shift := evCode&(1<<2) != 0  // 4
	meta := evCode&(1<<3) != 0   // 8
	ctrl := evCode&(1<<4) != 0   // 16
	motion := evCode&(1<<5) != 0 //32, motion indicator

	x := int(ib.b[4] - 33)
	y := int(ib.b[5] - 33) // - yoffset if any
	return Event{Mouse, 0, &MouseEvent{y, x, bNum, shift, meta, ctrl, motion, b}}
}
