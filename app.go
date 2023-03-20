package grest

import (
	"fmt"
	"strings"
)

const Version = "v0.0.1"

func StartupMessage(addr string) {
	addrPart := strings.Split(addr, ":")
	addr = "127.0.0.1"
	if addrPart[0] != "" && addrPart[0] != "0.0.0.0" {
		addr = addrPart[0]
	}
	if len(addrPart) > 1 {
		addr = addr + ":" + addrPart[1]
	}

	msg := strings.Builder{}
	msg.WriteString(Fmt(`        __________________________________________`, FmtHiMagenta, FmtBold, FmtItalic))
	msg.WriteString("\n")

	msg.WriteString(Fmt(`       /        `, FmtHiMagenta, FmtBold, FmtItalic))
	msg.WriteString(Fmt(`____`, FmtHiCyan, FmtBold, FmtBlinkRapid))
	msg.WriteString(Fmt(`___  `, FmtHiBlue, FmtBold))
	msg.WriteString(Fmt(`____`, FmtHiBlue, FmtBold))
	msg.WriteString(Fmt(`____`, FmtHiBlue, FmtBold))
	msg.WriteString(Fmt(`_____ `, FmtHiBlue, FmtBold))
	msg.WriteString(Fmt(`         /`, FmtHiMagenta, FmtBold, FmtItalic))
	msg.WriteString("\n")

	msg.WriteString(Fmt(`      /    `, FmtHiMagenta, FmtBold, FmtItalic))
	msg.WriteString(Fmt(`--- / __/`, FmtHiCyan, FmtBold, FmtBlinkRapid))
	msg.WriteString(Fmt(` _ \`, FmtHiBlue, FmtBold))
	msg.WriteString(Fmt(`/ __/`, FmtHiBlue, FmtBold))
	msg.WriteString(Fmt(` __/`, FmtHiBlue, FmtBold))
	msg.WriteString(Fmt(`_  _/`, FmtHiBlue, FmtBold))
	msg.WriteString(Fmt(`         /`, FmtHiMagenta, FmtBold, FmtItalic))
	msg.WriteString("\n")

	msg.WriteString(Fmt(`     /   `, FmtHiMagenta, FmtBold, FmtItalic))
	msg.WriteString(Fmt(`---- / / /`, FmtHiCyan, FmtBold, FmtBlinkRapid))
	msg.WriteString(Fmt(` / _/`, FmtHiBlue, FmtBold))
	msg.WriteString(Fmt(` _/`, FmtHiBlue, FmtBold))
	msg.WriteString(Fmt(`_\ \`, FmtHiBlue, FmtBold))
	msg.WriteString(Fmt(`  / /`, FmtHiBlue, FmtBold))
	msg.WriteString(Fmt(`          /`, FmtHiMagenta, FmtBold, FmtItalic))
	msg.WriteString("\n")

	msg.WriteString(Fmt(`    /     `, FmtHiMagenta, FmtBold, FmtItalic))
	msg.WriteString(Fmt(`-- /___/`, FmtHiCyan, FmtBold, FmtBlinkRapid))
	msg.WriteString(Fmt(`_/\ \`, FmtHiBlue, FmtBold))
	msg.WriteString(Fmt(`___/`, FmtHiBlue, FmtBold))
	msg.WriteString(Fmt(`___/`, FmtHiBlue, FmtBold))
	msg.WriteString(Fmt(` /_/ `, FmtHiBlue, FmtBold))
	msg.WriteString(Fmt(" ", FmtBgRed))
	msg.WriteString(Fmt(Version, FmtBgRed, FmtBold))
	msg.WriteString(Fmt(" ", FmtBgRed))
	msg.WriteString(Fmt(` /`, FmtHiMagenta, FmtBold, FmtItalic))
	msg.WriteString("\n")

	msg.WriteString(Fmt(`   /`, FmtHiMagenta, FmtBold, FmtItalic))
	msg.WriteString(` An instant, full-featured and scalable `)
	msg.WriteString(Fmt(`/`, FmtHiMagenta, FmtBold, FmtItalic))
	msg.WriteString("\n")

	msg.WriteString(Fmt(`  /`, FmtHiMagenta, FmtBold, FmtItalic))
	msg.WriteString(`       REST APIs framework for `)
	msg.WriteString(Fmt(`Go`, FmtHiCyan, FmtBold, FmtItalic))
	msg.WriteString(Fmt(`       /`, FmtHiMagenta, FmtBold, FmtItalic))
	msg.WriteString("\n")

	msg.WriteString(Fmt(` /             `, FmtHiMagenta, FmtBold, FmtItalic))
	msg.WriteString(Fmt("https://grest.dev", FmtBlue))
	msg.WriteString(Fmt(`          /`, FmtHiMagenta, FmtBold, FmtItalic))
	msg.WriteString("\n")

	msg.WriteString(Fmt(`/________________________________________/`, FmtHiMagenta, FmtBold, FmtItalic))
	msg.WriteString("\n")

	msg.WriteString("\n")
	msg.WriteString(`http server listening on `)
	msg.WriteString(Fmt("http://"+addr, FmtHiGreen))
	msg.WriteString("\n")

	fmt.Fprintln(FmtStdout, msg.String())
}
