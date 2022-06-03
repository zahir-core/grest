package grest

import (
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"

	"grest.dev/grest/log"
	"grest.dev/grest/swagger"
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
	OpenAPI               func() swagger.OpenAPI
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
func (app *App) AddRoute(path, method string, handler fiber.Handler, model swagger.Component) {
	if method == "ALL" {
		for _, m := range []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE"} {
			app.AddRoute(path, m, handler, model)
		}
	} else {
		app.Fiber.Add(method, path, handler)
		if len(os.Args) == 3 && os.Args[1] == "update" && os.Args[2] == "doc" {
			swagger.AddComponent(path, method, model)
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
		swagger.Generate(app.OpenAPI)
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
	msg.WriteString(log.Fmt(`        __________________________________________`, log.HiMagenta, log.Bold, log.Italic))
	msg.WriteString("\n")

	msg.WriteString(log.Fmt(`       /        `, log.HiMagenta, log.Bold, log.Italic))
	msg.WriteString(log.Fmt(`____`, log.HiCyan, log.Bold, log.BlinkRapid))
	msg.WriteString(log.Fmt(`___  `, log.Red, log.Bold))
	msg.WriteString(log.Fmt(`____`, log.Red, log.Bold))
	msg.WriteString(log.Fmt(`____`, log.Red, log.Bold))
	msg.WriteString(log.Fmt(`_____ `, log.Red, log.Bold))
	msg.WriteString(log.Fmt(`         /`, log.HiMagenta, log.Bold, log.Italic))
	msg.WriteString("\n")

	msg.WriteString(log.Fmt(`      /    `, log.HiMagenta, log.Bold, log.Italic))
	msg.WriteString(log.Fmt(`--- / __/`, log.HiCyan, log.Bold, log.BlinkRapid))
	msg.WriteString(log.Fmt(` _ \`, log.Red, log.Bold))
	msg.WriteString(log.Fmt(`/ __/`, log.Red, log.Bold))
	msg.WriteString(log.Fmt(` __/`, log.Red, log.Bold))
	msg.WriteString(log.Fmt(`_  _/`, log.Red, log.Bold))
	msg.WriteString(log.Fmt(`         /`, log.HiMagenta, log.Bold, log.Italic))
	msg.WriteString("\n")

	msg.WriteString(log.Fmt(`     /   `, log.HiMagenta, log.Bold, log.Italic))
	msg.WriteString(log.Fmt(`---- / / /`, log.HiCyan, log.Bold, log.BlinkRapid))
	msg.WriteString(log.Fmt(` / _/`, log.Red, log.Bold))
	msg.WriteString(log.Fmt(` _/`, log.Red, log.Bold))
	msg.WriteString(log.Fmt(`_\ \`, log.Red, log.Bold))
	msg.WriteString(log.Fmt(`  / /`, log.Red, log.Bold))
	msg.WriteString(log.Fmt(`          /`, log.HiMagenta, log.Bold, log.Italic))
	msg.WriteString("\n")

	msg.WriteString(log.Fmt(`    /     `, log.HiMagenta, log.Bold, log.Italic))
	msg.WriteString(log.Fmt(`-- /___/`, log.HiCyan, log.Bold, log.BlinkRapid))
	msg.WriteString(log.Fmt(`_/\ \`, log.White, log.Bold))
	msg.WriteString(log.Fmt(`___/`, log.White, log.Bold))
	msg.WriteString(log.Fmt(`___/`, log.White, log.Bold))
	msg.WriteString(log.Fmt(` /_/ `, log.White, log.Bold))
	msg.WriteString(log.Fmt(" ", log.BgRed))
	msg.WriteString(log.Fmt(Version, log.BgRed, log.Bold))
	msg.WriteString(log.Fmt(" ", log.BgRed))
	msg.WriteString(log.Fmt(` /`, log.HiMagenta, log.Bold, log.Italic))
	msg.WriteString("\n")

	msg.WriteString(log.Fmt(`   /`, log.HiMagenta, log.Bold, log.Italic))
	msg.WriteString(` An instant, full-featured and scalable `)
	msg.WriteString(log.Fmt(`/`, log.HiMagenta, log.Bold, log.Italic))
	msg.WriteString("\n")

	msg.WriteString(log.Fmt(`  /`, log.HiMagenta, log.Bold, log.Italic))
	msg.WriteString(`       REST APIs framework for `)
	msg.WriteString(log.Fmt(`Go`, log.HiCyan, log.Bold, log.Italic))
	msg.WriteString(log.Fmt(`       /`, log.HiMagenta, log.Bold, log.Italic))
	msg.WriteString("\n")

	msg.WriteString(log.Fmt(` /             `, log.HiMagenta, log.Bold, log.Italic))
	msg.WriteString(log.Fmt("https://grest.dev", log.Blue))
	msg.WriteString(log.Fmt(`          /`, log.HiMagenta, log.Bold, log.Italic))
	msg.WriteString("\n")

	msg.WriteString(log.Fmt(`/________________________________________/`, log.HiMagenta, log.Bold, log.Italic))
	msg.WriteString("\n")

	msg.WriteString("\n")
	msg.WriteString(`http server listening on `)
	msg.WriteString(log.Fmt("http://"+addr, log.HiGreen))
	msg.WriteString("\n")

	fmt.Fprintln(log.Stdout, msg.String())
}
