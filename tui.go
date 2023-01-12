package tui

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/mattn/go-isatty"
	"github.com/mattn/go-runewidth"
	"github.com/pzl/tui/ansi"
	"golang.org/x/crypto/ssh/terminal"
)

// width and height of the terminal. Falls back to defaults (124,80) if it can't be determined
//
// fd should be where you intend to print to. For Stdout: int(os.Stdout.Fd())
func TermSize(fd int) (int, int) {
	const defaultWidth = 124
	const defaultHeight = 80

	w, h, err := terminal.GetSize(fd)
	if err != nil {
		w = env("COLUMNS", defaultWidth)
		h = env("LINES", defaultHeight)
	}
	return w, h
}

func env(k string, def int) int {
	if wd := os.Getenv(k); len(wd) != 0 {
		return atoi(wd, def)
	}
	return def
}

func atoi(s string, def int) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return def
}

func IsTTY(fd uintptr) bool { return isatty.IsTerminal(fd) }
func RuneWidth(r rune) int  { return runewidth.RuneWidth(r) }

// Returns the coordinates of the cursor. fd should almost always be 0 for stdin
// Or you can use int(os.Stdin.Fd())
func CursorPosition(fd int) (x int, y int, err error) {
	if !IsTTY(uintptr(fd)) {
		return 0, 0, errors.New("input is not a TTY")
	}
	state, err := terminal.GetState(fd)
	if err != nil {
		return 0, 0, fmt.Errorf("error getting terminal state: %v", err)
	}
	defer func() {
		err = terminal.Restore(fd, state)
	}()

	err = setNonBlock(fd, false)
	if err != nil {
		return 0, 0, fmt.Errorf("error setting terminal as blocking: %v", err)
	}
	_, err = terminal.MakeRaw(fd)
	if err != nil {
		return 0, 0, fmt.Errorf("error putting terminal in raw mode: %v", err)
	}

	// prints 6n to stderr
	ansi.NewWriter(nil).CursorPosition()

	//@todo: probably read char-by-char.
	// a random, blocking, 13-byte buffer is odd

	// read response from stdin
	b := make([]byte, 13)
	_, err = sysRead(fd, b)
	if err != nil {
		return 0, 0, fmt.Errorf("error reading input: %v", err)
	}

	n, err := fmt.Sscanf(string(b), "\x1b[%d;%dR", &x, &y)
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing terminal response: %v", err)
	}
	if n < 2 {
		return 0, 0, fmt.Errorf("error parsing cursor position: %v", err)
	}
	return x, y, nil
}
