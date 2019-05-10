package tui_test

import (
	"fmt"
	"os"

	"github.com/pzl/tui"
)

func Example() {
	events, restore, err := tui.GetInput(nil, int(os.Stdin.Fd())) // start capturing keyboard
	defer restore()                                               // sets input mode of terminal back to normal
	if err != nil {
		panic(err)
	}

	for ev := range events {
		// handle incoming input

		switch ev.Type {
		case tui.KeyPrint: // printable keys (alphanum, punctuation)
			fmt.Print(string(ev.Key))
		case tui.KeySpecial: // key sequences, shortcuts, non printables like Home, Esc, F2, etc
			switch ev.Key {
			case tui.CtrlC:
				return
				// handle any other keys you want ...
			}
		}
	}
}
