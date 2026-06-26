package xf

import (
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3"
)

type XFiberApp struct {
	*fiber.App
}

func NewFiberApp(handlers ...any) *XFiberApp {
	app := fiber.NewWithCustomCtx(func(app *fiber.App) fiber.CustomCtx {
		return &Context {
			DefaultCtx: fiber.NewDefaultCtx(app),
		}
	})
	app.Use(recoverer.New())
	app.Use(handlers...)
	return &XFiberApp{App: app}
}

// ---- fiber.HandlerFunc ----
func (app *XFiberApp) Get(path string, handler any, handlers ...any) fiber.Router {
	return app.App.Get(path, handler, handlers...)
}

func (app *XFiberApp) Post(path string, handler any, handlers ...any) fiber.Router {
	return app.App.Post(path, handler, handlers...)
}

func (app *XFiberApp) Put(path string, handler any, handlers ...any) fiber.Router {
	return app.App.Put(path, handler, handlers...)
}

func (app *XFiberApp) Delete(path string, handler any, handlers ...any) fiber.Router {
	return app.App.Delete(path, handler, handlers...)
}

func (app *XFiberApp) Options(path string, handler any, handlers ...any) fiber.Router {
	return app.App.Options(path, handler, handlers...)
}

func (app *XFiberApp) Patch(path string, handler any, handlers ...any) fiber.Router {
	return app.App.Patch(path, handler, handlers...)
}

func (app *XFiberApp) All(path string, handler any, handlers ...any) fiber.Router {
	return app.App.All(path, handler, handlers...)
}

// ---- XFiberHandlerFunc ----
func (app *XFiberApp) GET(path string, h XFiberHandlerFunc) fiber.Router {
	return app.Get(path, unwrap(h))
}

func (app *XFiberApp) POST(path string, h XFiberHandlerFunc) fiber.Router {
	return app.Post(path, unwrap(h))
}

func (app *XFiberApp) PUT(path string, h XFiberHandlerFunc) fiber.Router {
	return app.Put(path, unwrap(h))
}

func (app *XFiberApp) DELETE(path string, h XFiberHandlerFunc) fiber.Router {
	return app.Delete(path, unwrap(h))
}

func (app *XFiberApp) OPTIONS(path string, h XFiberHandlerFunc) fiber.Router {
	return app.Options(path, unwrap(h))
}

func (app *XFiberApp) PATCH(path string, h XFiberHandlerFunc) fiber.Router {
	return app.Patch(path, unwrap(h))
}

func (app *XFiberApp) ALL(path string, h XFiberHandlerFunc) fiber.Router {
	return app.All(path, unwrap(h))
}
