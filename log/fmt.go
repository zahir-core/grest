package log

import (
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
)

// Base attributes
const (
	Reset uint8 = iota
	Bold
	Faint
	Italic
	Underline
	BlinkSlow
	BlinkRapid
	ReverseVideo
	Concealed
	CrossedOut
)

// Foreground text colors
const (
	Black uint8 = iota + 30
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

// Foreground Hi-Intensity text colors
const (
	HiBlack uint8 = iota + 90
	HiRed
	HiGreen
	HiYellow
	HiBlue
	HiMagenta
	HiCyan
	HiWhite
)

// Background text colors
const (
	BgBlack uint8 = iota + 40
	BgRed
	BgGreen
	BgYellow
	BgBlue
	BgMagenta
	BgCyan
	BgWhite
)

// Background Hi-Intensity text colors
const (
	BgHiBlack uint8 = iota + 100
	BgHiRed
	BgHiGreen
	BgHiYellow
	BgHiBlue
	BgHiMagenta
	BgHiCyan
	BgHiWhite
)

var (
	// DisableFmt defines if the output is colorized or not.
	DisableFmt = (!isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()))

	// Output defines the standard output of the print functions. By default os.Stdout is used.
	Stdout = colorable.NewColorableStdout()
)

// Fmt format log with attribute
// an example log.Fmt("text", log.Bold, log.Red) output might be: "\x1b[1;31mtext\x1b[0m" -> text with bold red foreground
func Fmt(s string, attribute ...uint8) string {
	if DisableFmt {
		return s
	}
	format := make([]string, len(attribute))
	for i, v := range attribute {
		format[i] = strconv.Itoa(int(v))
	}
	return "\x1b[" + strings.Join(format, ";") + "m" + s + "\x1b[0m"
}
