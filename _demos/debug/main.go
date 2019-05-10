/*
This is a simple debug program.
It prints an underlined "Welcome"
to roughly the center of the terminal.
and it displays key pressed, keycodes,
and mouse events received in the top left.
*/
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pzl/tui"
	"github.com/pzl/tui/ansi"
)

func main() {
	buf := ansi.NewOutputBuffer()
	w, h := tui.TermSize(0)

	msg := "Welcome"

	x := w/2 - len(msg)/2
	y := h / 2

	motion := ansi.MotionAll

	defer func() {
		buf.MouseDisable(motion)
		buf.Screen(ansi.Normal)
		buf.CursorShow()
		buf.Flush()
	}()

	buf.Screen(ansi.Alt)
	buf.MoveTo(x, y)
	buf.CursorHide()
	buf.MouseEnable(motion)
	buf.Style(ansi.Underline)
	buf.Flush()

	fmt.Printf(string(ansi.Underline) + msg + ansi.Style(ansi.Reset))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events, restore, err := tui.GetInput(ctx, int(os.Stdin.Fd()))
	defer restore()
	if err != nil {
		panic(err)
	}

	i := 0
	for ev := range events {
		buf.Origin()
		buf.ClearLineRight()
		buf.Flush()

		i++
		switch ev.Type {
		case tui.KeyPrint:
			fmt.Printf("%06d %s", i, string(ev.Key))
		case tui.KeySpecial:
			fmt.Printf("%06d %s", i, keyMap[ev.Key])
			if ev.Key == tui.CtrlC || ev.Key == tui.ESC {
				return
			}
		case tui.EventInvalid:
			fmt.Printf("%06d invalid ev %v", i, ev.M)
		case tui.Mouse:
			buf.Down(1)
			fmt.Printf("%06d Mouse %v", i, ev.M)
		}
	}
}

var keyMap = map[rune]string{
	tui.Null:            "Null",
	tui.CtrlA:           "CtrlA",
	tui.CtrlB:           "CtrlB",
	tui.CtrlC:           "CtrlC",
	tui.CtrlD:           "CtrlD",
	tui.CtrlE:           "CtrlE",
	tui.CtrlF:           "CtrlF",
	tui.CtrlG:           "CtrlG",
	tui.CtrlH:           "CtrlH",
	tui.Tab:             "Tab",
	tui.CtrlJ:           "CtrlJ",
	tui.CtrlK:           "CtrlK",
	tui.CtrlL:           "CtrlL",
	tui.CtrlM:           "CtrlM",
	tui.CtrlN:           "CtrlN",
	tui.CtrlO:           "CtrlO",
	tui.CtrlP:           "CtrlP",
	tui.CtrlQ:           "CtrlQ",
	tui.CtrlR:           "CtrlR",
	tui.CtrlS:           "CtrlS",
	tui.CtrlT:           "CtrlT",
	tui.CtrlU:           "CtrlU",
	tui.CtrlV:           "CtrlV",
	tui.CtrlW:           "CtrlW",
	tui.CtrlX:           "CtrlX",
	tui.CtrlY:           "CtrlY",
	tui.CtrlZ:           "CtrlZ",
	tui.ESC:             "ESC",
	tui.CtrlFwdSlash:    "CtrlFwdSlash",
	tui.CtrlBackBracket: "CtrlBackBracket",
	tui.CtrlCaret:       "CtrlCaret",
	tui.CtrlUnderscore:  "CtrlUnderscore",
	tui.AltSpace:        "AltSpace",
	tui.AltBang:         "AltBang",
	tui.AltDQuo:         "AltDQuo",
	tui.AltHash:         "AltHash",
	tui.AltDollar:       "AltDollar",
	tui.AltPercent:      "AltPercent",
	tui.AltAmpersand:    "AltAmpersand",
	tui.AltSQuo:         "AltSQuo",
	tui.AltOpenParen:    "AltOpenParen",
	tui.AltCloseParen:   "AltCloseParen",
	tui.AltStar:         "AltStar",
	tui.AltPlus:         "AltPlus",
	tui.AltComma:        "AltComma",
	tui.AltMinus:        "AltMinus",
	tui.AltPeriod:       "AltPeriod",
	tui.AltSlash:        "AltSlash",
	tui.Alt0:            "Alt0",
	tui.Alt1:            "Alt1",
	tui.Alt2:            "Alt2",
	tui.Alt3:            "Alt3",
	tui.Alt4:            "Alt4",
	tui.Alt5:            "Alt5",
	tui.Alt6:            "Alt6",
	tui.Alt7:            "Alt7",
	tui.Alt8:            "Alt8",
	tui.Alt9:            "Alt9",
	tui.AltColon:        "AltColon",
	tui.AltSemicolon:    "AltSemicolon",
	tui.AltLT:           "AltLT",
	tui.AltEqual:        "AltEqual",
	tui.AltGT:           "AltGT",
	tui.AltQuestion:     "AltQuestion",
	tui.AltAt:           "AltAt",
	tui.AltA:            "AltA",
	tui.AltB:            "AltB",
	tui.AltC:            "AltC",
	tui.AltD:            "AltD",
	tui.AltE:            "AltE",
	tui.AltF:            "AltF",
	tui.AltG:            "AltG",
	tui.AltH:            "AltH",
	tui.AltI:            "AltI",
	tui.AltJ:            "AltJ",
	tui.AltK:            "AltK",
	tui.AltL:            "AltL",
	tui.AltM:            "AltM",
	tui.AltN:            "AltN",
	tui.AltO:            "AltO",
	tui.AltP:            "AltP",
	tui.AltQ:            "AltQ",
	tui.AltR:            "AltR",
	tui.AltS:            "AltS",
	tui.AltT:            "AltT",
	tui.AltU:            "AltU",
	tui.AltV:            "AltV",
	tui.AltW:            "AltW",
	tui.AltX:            "AltX",
	tui.AltY:            "AltY",
	tui.AltZ:            "AltZ",
	tui.AltOpenBracket:  "AltOpenBracket",
	tui.AltFwdSlash:     "AltFwdSlash",
	tui.AltCloseBracket: "AltCloseBracket",
	tui.AltCaret:        "AltCaret",
	tui.AltUnderscore:   "AltUnderscore",
	tui.AltGrave:        "AltGrave",
	tui.Alta:            "Alta",
	tui.Altb:            "Altb",
	tui.Altc:            "Altc",
	tui.Altd:            "Altd",
	tui.Alte:            "Alte",
	tui.Altf:            "Altf",
	tui.Altg:            "Altg",
	tui.Alth:            "Alth",
	tui.Alti:            "Alti",
	tui.Altj:            "Altj",
	tui.Altk:            "Altk",
	tui.Altl:            "Altl",
	tui.Altm:            "Altm",
	tui.Altn:            "Altn",
	tui.Alto:            "Alto",
	tui.Altp:            "Altp",
	tui.Altq:            "Altq",
	tui.Altr:            "Altr",
	tui.Alts:            "Alts",
	tui.Altt:            "Altt",
	tui.Altu:            "Altu",
	tui.Altv:            "Altv",
	tui.Altw:            "Altw",
	tui.Altx:            "Altx",
	tui.Alty:            "Alty",
	tui.Altz:            "Altz",
	tui.AltOpenCurly:    "AltOpenCurly",
	tui.AltPipe:         "AltPipe",
	tui.AltCloseCurly:   "AltCloseCurly",
	tui.AltTilde:        "AltTilde",
	tui.AltBS:           "AltBS",
	tui.Del:             "Del",
	tui.BTab:            "BTab",
	tui.BSpace:          "BSpace",

	tui.PgUp: "PgUp",
	tui.PgDn: "PgDn",

	tui.Up:    "Up",
	tui.Down:  "Down",
	tui.Right: "Right",
	tui.Left:  "Left",
	tui.Home:  "Home",
	tui.End:   "End",

	// relative order of the next 8 matter
	tui.SUp:    "SUp",
	tui.SDown:  "SDown",
	tui.SRight: "SRight",
	tui.SLeft:  "SLeft",

	tui.CtrlUp:    "CtrlUp",
	tui.CtrlDown:  "CtrlDown",
	tui.CtrlRight: "CtrlRight",
	tui.CtrlLeft:  "CtrlLeft",

	tui.F1:  "F1",
	tui.F2:  "F2",
	tui.F3:  "F3",
	tui.F4:  "F4",
	tui.F5:  "F5",
	tui.F6:  "F6",
	tui.F7:  "F7",
	tui.F8:  "F8",
	tui.F9:  "F9",
	tui.F10: "F10",
	tui.F11: "F11",
	tui.F12: "F12",

	tui.AltDel: "AltDel",

	tui.CtrlAlta: "CtrlAlta",
	tui.CtrlAltb: "CtrlAltb",
	tui.CtrlAltc: "CtrlAltc",
	tui.CtrlAltd: "CtrlAltd",
	tui.CtrlAlte: "CtrlAlte",
	tui.CtrlAltf: "CtrlAltf",
	tui.CtrlAltg: "CtrlAltg",
	tui.CtrlAlth: "CtrlAlth",
	tui.CtrlAlti: "CtrlAlti",
	tui.CtrlAltj: "CtrlAltj",
	tui.CtrlAltk: "CtrlAltk",
	tui.CtrlAltl: "CtrlAltl",
	tui.CtrlAltm: "CtrlAltm",
	tui.CtrlAltn: "CtrlAltn",
	tui.CtrlAlto: "CtrlAlto",
	tui.CtrlAltp: "CtrlAltp",
	tui.CtrlAltq: "CtrlAltq",
	tui.CtrlAltr: "CtrlAltr",
	tui.CtrlAlts: "CtrlAlts",
	tui.CtrlAltt: "CtrlAltt",
	tui.CtrlAltu: "CtrlAltu",
	tui.CtrlAltv: "CtrlAltv",
	tui.CtrlAltw: "CtrlAltw",
	tui.CtrlAltx: "CtrlAltx",
	tui.CtrlAlty: "CtrlAlty",
	tui.CtrlAltz: "CtrlAltz",
}
