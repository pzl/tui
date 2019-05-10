`tui` is a low-level terminal interface library. It is mostly just a wrapper around ANSI terminal escape codes, and terminal input handling. See the [docs](https://godoc.org/github.com/pzl/tui) or look at the [demo programs](_demos)

There is no concept of state, or repainting. That is to be implemented in your apps. 

This library tries to do very little _for_ you. This means more manual work if you use it, but ultimate flexibility.


the `tui` top-level package provides keyboard/mouse event handling if your program chooses to take input control. The `ansi` sub-package is just for outputting things (color, text effects, clearing, cursor movement, etc).

LICENSE
-------

MIT License (c) Dan Panzarella 2019