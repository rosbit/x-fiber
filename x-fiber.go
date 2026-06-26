package xf

import (
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3"
	"time"
)

func WithLogger(name string) fiber.Handler {
	return logger.New(logger.Config{
			TimeFormat: time.RFC3339,
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
