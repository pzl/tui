/*
tui is a very thin wrapper around terminal events. The tui package itself provides for mostly just keyboard + mouse input from the terminal. It handles input.

Output is handled mostly through the ansi subpackage here. It handles:
  - text formatting like color, underline
  - cursor movement
  - screen control (clearing, scrolling, etc)

a full TUI program should ideally make use of both parts. Unless you're just pretty-printing some colors to the terminal, then just ansi might work for you. But there are other go packages out there with better cross-platform support for just colorizing things.

For good examples of usage together, there are two small example programs in the _demos folder. One is a small event debugger that prints input events to the top of the screen on keypress and mouse. The second demo program is a tiny editor, a cross-between vim and nano (interface only, no actual file/saving)
*/
package tui
