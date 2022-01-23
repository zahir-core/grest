package grest

import (
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"grest.dev/grest/swagger"
)

const Version = "0.0.1"

type App struct {
	IsUseTLS              bool
	CertFile              string
	KeyFile               string
	DisableStartupMessage bool
	Config                fiber.Config
	Recover               fiber.Handler
	ErrorHandler          fiber.ErrorHandler
	NotFoundHandler       fiber.Handler // to make sure it added at the very bottom of the stack (below all other functions) to handle a 404 response
	Fiber                 *fiber.App
	OpenAPI               func() swagger.OpenAPI
}

func New(a ...App) *App {
	app := checkConfig(a...)
	app.Fiber = fiber.New(app.Config)
	app.Fiber.Use(app.Recover)

	return &app
}

func checkConfig(a ...App) App {
	app := App{}
	if len(a) > 0 {
		app = a[0]
	}
	app.Config.DisableStartupMessage = true
	if app.ErrorHandler == nil && app.Config.ErrorHandler == nil {
		app.Config.ErrorHandler = NewErrorHandler()
	}
	if app.Recover == nil {
		app.Recover = recover.New()
	}
	if app.NotFoundHandler == nil {
		app.NotFoundHandler = NewNotFoundHandler()
	}
	return app
}

// use grest to add route so it can generate swagger api documentation automatically
func (app *App) AddRoute(path, method string, handler fiber.Handler, model swagger.Component) {
	app.Fiber.Add(method, path, handler)
	if len(os.Args) == 3 && os.Args[1] == "update" && os.Args[2] == "doc" {
		swagger.AddComponent(path, method, model)
	}
}

func (app *App) AddStaticRoute(root, path string, config ...fiber.Static) {
	app.Fiber.Static(path, root, config...)
}

func (app *App) AddSwagger(path string) {
	app.Fiber.Static(path, "./docs")
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
	fmt.Println(`
        __________________________________________________
       /         ____  __    ____  ___   _____ v` + Version + `    /
      /      -- / __/  _ \  / __/ / __/ _   _/          /
     /    ---- / / /  / _/ / _/  _\ \   / /            /
    /     --- /___/ _/\ \ /___/ /___/  /_/            /
   /      An instant, full-featured and scalable     /
  /            REST APIs framework for Go           /
 /                 https://grest.dev               /
/_________________________________________________/

http server listening on http://` + addr)
}
