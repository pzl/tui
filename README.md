[![GoDoc](https://godoc.org/github.com/pzl/tui?status.svg)](https://godoc.org/github.com/pzl/tui)

`tui` is a low-level terminal interface library. It is mostly just a wrapper around ANSI terminal escape codes, and terminal input handling.

Getting started with fancy screen features is easy:

```go
package main

import (
    "fmt"
    "strings"
    "time"
    "github.com/pzl/tui/ansi"
)

func main() {
    w := ansi.NewWriter(nil)
    w.CursorHide()
    defer w.CursorShow()

    const max = 50
    for i := 0; i <= max; i++ {
        w.Column(0)
        w.ClearLineRight()
        fmt.Printf("%s[%s%s>%s%s]%s % 3.0f%%",
            ansi.Cyan,
            ansi.Green,
            strings.Repeat("=", i),
            strings.Repeat(" ", max-i),
            ansi.Cyan,
            ansi.Reset,
            (float64(i)/max)*100,
        )
        time.Sleep(70 * time.Millisecond)
    }
    fmt.Print("\n")
}
```

![progress-bar](https://gist.githubusercontent.com/pzl/a485be3d54e3b9364e4a49a5c7450d3f/raw/7d1cdb005c43decb8c1c5f1bc17006ff3cfd8e0e/progressbar.svg?sanitize=true)


Or you can present full-screen apps with keyboard and mouse control. There are more examples in the [`_demos`](_demos) folder. You can also check out the [docs](https://godoc.org/github.com/pzl/tui) on godoc.


This library tries to do very little _for_ you. This means more manual work if you use it, but ultimate flexibility. There is no concept of state, or repainting in `tui`. That is to be implemented in your apps.


the `tui` top-level package provides keyboard/mouse event handling if your program chooses to take input control. The `ansi` sub-package is just for outputting things (color, text effects, clearing, cursor movement, etc).

LICENSE
-------

MIT License (c) Dan Panzarella 2019