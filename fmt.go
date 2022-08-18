package grest

import (
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
)

// Base attributes
const (
	FmtReset uint8 = iota
	FmtBold
	FmtFaint
	FmtItalic
	FmtUnderline
	FmtBlinkSlow
	FmtBlinkRapid
	FmtReverseVideo
	FmtConcealed
	FmtCrossedOut
)

// Foreground text colors
const (
	FmtBlack uint8 = iota + 30
	FmtRed
	FmtGreen
	FmtYellow
	FmtBlue
	FmtMagenta
	FmtCyan
	FmtWhite
)

// Foreground Hi-Intensity text colors
const (
	FmtHiBlack uint8 = iota + 90
	FmtHiRed
	FmtHiGreen
	FmtHiYellow
	FmtHiBlue
	FmtHiMagenta
	FmtHiCyan
	FmtHiWhite
)

// Background text colors
const (
	FmtBgBlack uint8 = iota + 40
	FmtBgRed
	FmtBgGreen
	FmtBgYellow
	FmtBgBlue
	FmtBgMagenta
	FmtBgCyan
	FmtBgWhite
)

// Background Hi-Intensity text colors
const (
	FmtBgHiBlack uint8 = iota + 100
	FmtBgHiRed
	FmtBgHiGreen
	FmtBgHiYellow
	FmtBgHiBlue
	FmtBgHiMagenta
	FmtBgHiCyan
	FmtBgHiWhite
)

var (
	// DisableFmt defines if the output is colorized or not.
	DisableFmt = (!isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()))

	// Output defines the standard output of the print functions. By default os.Stdout is used.
	FmtStdout = colorable.NewColorableStdout()
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
