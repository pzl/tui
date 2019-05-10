package ansi

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

/* ----- Top-level exported funcs ------ */

func TestEffect_Color(t *testing.T) {
	assert.Equal(t, "\x1b[34m", Effect(Blue))
}

func TestEffect_Style(t *testing.T) {
	assert.Equal(t, "\x1b[4m", Effect(Underline))
}

func TestEffect_Multiple(t *testing.T) {
	assert.Equal(t, "\x1b[31;1m", Effect(Red, Bold))
}

func TestEffect_MultipleStyles(t *testing.T) {
	assert.Equal(t, "\x1b[4;5m", Effect(Underline, Blink))
}

func TestEffect_EmptyResets(t *testing.T) {
	assert.Equal(t, "\x1b[m", Effect())
}

func TestStyles(t *testing.T) {
	assert.Equal(t, "\x1b[3m", Style(It))
}

func TestStyles_Multiple(t *testing.T) {
	assert.Equal(t, "\x1b[2;9m", Style(Dim, Strikethrough))
}

func TestColor_Basic(t *testing.T) {
	assert.Equal(t, "\x1b[30m", Color(Black))
}

func TestColor_8Bit(t *testing.T) {
	assert.Equal(t, "\x1b[38;5;234m", Color(EBColor(234)))
}

func TestColor_24bit(t *testing.T) {
	assert.Equal(t, "\x1b[38;2;254;171;39m", Color(TrueColor(0xfeab27)))
}

// test that T can be used as an RGB shorthand
func Test24bitColor_T(t *testing.T) {
	assert.Equal(t, TrueColor(0xab12ce), T(0xab, 0x12, 0xce))
}

// Fmt Strings

func TestTextStyle_FmtString(t *testing.T) {
	assert.Equal(t, "\x1b[8mGONE", fmt.Sprintf("%sGONE", Hidden))
}

func TestBasicColor_FmtString(t *testing.T) {
	assert.Equal(t, "\x1b[32mOK", fmt.Sprintf("%sOK", Green))
}

func Test8BitColor_FmtString(t *testing.T) {
	assert.Equal(t, "\x1b[38;5;122mOK", fmt.Sprintf("%sOK", EBColor(122)))
}

func Test24BitColor_FmtString(t *testing.T) {
	assert.Equal(t, "\x1b[38;2;18;52;86mOK", fmt.Sprintf("%sOK", TrueColor(0x123456)))
}
