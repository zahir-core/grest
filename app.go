package grest

import (
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const Version = "v0.0.0"

type App struct {
	IsUseTLS              bool
	CertFile              string
	KeyFile               string
	DisableStartupMessage bool
	Config                fiber.Config
	ErrorHandler          fiber.ErrorHandler
	NotFoundHandler       fiber.Handler // to make sure it added at the very bottom of the stack (below all other functions) to handle a 404 response
	Fiber                 *fiber.App
	OpenAPI               OpenAPIInterface
}

func New(a ...App) *App {
	app := checkConfig(a...)
	app.Fiber = fiber.New(app.Config)

	return &app
}

func checkConfig(a ...App) App {
	app := App{}
	if len(a) > 0 {
		app = a[0]
	}
	app.Config.DisableStartupMessage = true
	if app.ErrorHandler == nil && app.Config.ErrorHandler == nil {
		app.ErrorHandler = NewErrorHandler()
	}
	if app.Config.ErrorHandler == nil {
		app.Config.ErrorHandler = app.ErrorHandler
	}
	if app.NotFoundHandler == nil {
		app.NotFoundHandler = NewNotFoundHandler()
	}
	return app
}

// use grest to add route so it can generate swagger api documentation automatically
func (app *App) AddRoute(path, method string, handler fiber.Handler, operation OpenAPIOperationInterface) {
	if method == "ALL" {
		for _, m := range []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE"} {
			app.AddRoute(path, m, handler, operation)
		}
	} else {
		app.Fiber.Add(method, path, handler)
		if len(os.Args) == 3 && os.Args[1] == "update" && os.Args[2] == "doc" {
			app.OpenAPI.AddRoute(path, method, operation)
		}
	}
}

func (app *App) AddStaticRoute(root, path string, config ...fiber.Static) {
	app.Fiber.Static(path, root, config...)
}

func (app *App) AddSwagger(path string) {
	app.Fiber.Static(path, "./docs", fiber.Static{CacheDuration: -1})
}

func (app *App) AddMiddleware(handler fiber.Handler) {
	app.Fiber.Use(handler)
}

func (app *App) Start(addr string) error {
	if len(os.Args) == 3 && os.Args[1] == "update" && os.Args[2] == "doc" {
		app.OpenAPI.Generate()
	}
	app.Fiber.Use(app.NotFoundHandler)
	if !app.DisableStartupMessage {
		app.startupMessage(addr)
	}
	if app.IsUseTLS {
		return app.Fiber.ListenTLS(addr, app.CertFile, app.KeyFile)
	}
	return app.Fiber.Listen(addr)
}

func (app *App) startupMessage(addr string) {
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
