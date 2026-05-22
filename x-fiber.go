package xf

import (
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3"
	"fmt"
)

func NewFiberApp(handlers ...any) *fiber.App {
	app := fiber.NewWithCustomCtx(func(app *fiber.App) fiber.CustomCtx {
		return &Context {
			DefaultCtx: fiber.NewDefaultCtx(app),
		}
	})
	app.Use(recoverer.New())
	app.Use(handlers...)
	return app
}

func ToXFiberCtx(ctx fiber.Ctx) (*Context, error) {
	c, ok := ctx.(*Context)
	if !ok {
		return nil, fmt.Errorf("bad context")
	}
	c.Init()
	return c, nil
}

func WithLogger(name string) fiber.Handler {
	return logger.New(logger.Config{
			CustomTags: map[string]logger.LogFunc{
				"logger_name": func(output logger.Buffer, c fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
					return output.WriteString(name)
				},
				"host_name": func(output logger.Buffer, c fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
					return output.WriteString(c.Host())
				},
			},
			// For more options, see the Config section
			Format: "[${logger_name}] ${time} | ${status} | \t ${latency} | ${host_name} | ${method} ${url}\n",
			ForceColors: true,
	})
}

func CreateCorsFreeHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Headers", "*")
		c.Set("Access-Control-Allow-Method", "POST, GET, PUT, OPTIONS, DELETE")
		if c.Method() == fiber.MethodOptions {
			return c.End()
		}
		return c.Next()
	}
}
