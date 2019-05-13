package ansi_test

import (
	"bufio"
	"fmt"
	"os"

	"github.com/pzl/tui/ansi"
)

func ExampleWriter() {
	w := ansi.NewWriter(os.Stdout)

	w.Screen(ansi.Alt)
	w.ClearAll()
	w.MoveTo(10, 10)
	fmt.Print("example!")

	// Output:[?1049h[2J[10;10Hexample!
}

func ExampleWriter_buffered() {
	buf := bufio.NewWriter(os.Stdout) // buffer the commands going to stdout
	w := ansi.NewWriter(buf)          // create the writer connected to that buffer

	w.Up(2)
	w.ClearLineRight()
	buf.Write([]byte("replacement"))
	w.Down(2)
	buf.Flush() // commands and all sent all at once here

	// Output:[2A[Kreplacement[2B
}

func ExampleEffect() {
	fmt.Printf("%sThis Will be Blue and Underlined%s", ansi.Effect(ansi.Blue, ansi.Underline), ansi.Effect(ansi.Reset))
	// Output:[34;4mThis Will be Blue and Underlined[0m
}

func ExampleEffect_direct() {
	fmt.Printf("%s%sThis Will be Blue and Underlined%s", ansi.Blue, ansi.Underline, ansi.Reset)
	// Output: [34m[4mThis Will be Blue and Underlined[0m
}

func ExampleStyle() {
	fmt.Printf("%sThis Text Will Be Underlined%s", ansi.Style(ansi.Underline), ansi.Style(ansi.Reset))
	// Output: [4mThis Text Will Be Underlined[0m
}

//An effect can be used directly with the %s formatter, which calls effect.String() which will return the needed ansi sequence
func ExampleTextStyle() {
	fmt.Printf("%sThis Text Will Be Underlined%s", ansi.Underline, ansi.Reset)
	// Output: [4mThis Text Will Be Underlined[0m
}

func ExampleColor_basic() {
	fmt.Printf("%sRed", ansi.Color(ansi.Red))
	// Output: [31mRed
}

func ExampleColor_eightBit() {
	fmt.Printf("%sReddish", ansi.Color(ansi.EBColor(216)))
	// Output: [38;5;216mReddish
}

func ExampleColor_trueColor() {
	fmt.Printf("%sAlso Red", ansi.Color(ansi.TrueColor(0xc45679)))
	// Output: [38;2;196;86;121mAlso Red
}

func ExampleBasicColor_String_printf() {
	fmt.Printf("%sThis is Red%sNow Blue%s", ansi.Red, ansi.Blue, ansi.Reset)
	// Output: [31mThis is Red[34mNow Blue[0m
}

func ExampleEBColor_String_printf() {
	fmt.Printf("%sThis is Red%sNow Blue%s", ansi.EBColor(196), ansi.EBColor(81), ansi.Reset)
	// Output: [38;5;196mThis is Red[38;5;81mNow Blue[0m
}

func ExampleT() {
	fmt.Printf("%sAny Color You want", ansi.Color(ansi.T(12, 91, 202)))
	// Output: [38;2;12;91;202mAny Color You want
}

func ExampleTrueColor_String_printf() {
	fmt.Printf("%sAny Color You want", ansi.TrueColor(0x0C5BCA))
	// Output: [38;2;12;91;202mAny Color You want
}
