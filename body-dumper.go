package xf

import (
	logr "github.com/rosbit/reader-logger"
	"github.com/gofiber/fiber/v3"
	"fmt"
	"io"
	"os"
	"strconv"
)

func bodyDumper(body io.Reader, dumper io.Writer, prompts ...string) (reader io.Reader, deferFunc func()) {
	var prompt string
	if len(prompts) > 0 {
		prompt = prompts[0]
	}

	if dumper != nil {
		if f, ok := dumper.(*os.File); ok {
			if f == os.Stderr || f == os.Stdout {
				if len(prompt) == 0 {
					prompt = "fiber adapter dumping body"
				}
			}
		}
	}
	return logr.ReaderLogger(body, dumper, prompt)
}

func CreateBodyDumpingMiddleware(dumper io.Writer, prompt ...string) fiber.Handler {
	var reqPrompt string

	if len(prompt) > 0 && len(prompt[0]) > 0 {
		reqPrompt = prompt[0]
	}
	return createBodyDumppingFiberMiddleware(dumper, reqPrompt, "")
}

func CreateBodyDumpingMiddleware2(dumper io.Writer, reqPrompt, respPrompt string, querySwitchParam ...string) fiber.Handler {
	return createBodyDumppingFiberMiddleware(dumper, reqPrompt, respPrompt, querySwitchParam...)
}

func createBodyDumppingFiberMiddleware(dumper io.Writer, reqPrompt, respPrompt string, querySwitchParam ...string) fiber.Handler {
	var prompt string
	if len(reqPrompt) > 0 {
		prompt = reqPrompt
	}
	var querySwitchName string
	if len(querySwitchParam) > 0 && len(querySwitchParam[0]) > 0 {
		querySwitchName = querySwitchParam[0]
	}

	if dumper != nil {
		if f, ok := dumper.(*os.File); ok {
			if f == os.Stderr || f == os.Stdout {
				if len(prompt) == 0 {
					prompt = "dumping body"
				}
			}
		}
	}

	return func(c fiber.Ctx) error {
		if len(querySwitchName) > 0 {
			dumpSwitch := c.Query(querySwitchName)
			if len(dumpSwitch) == 0 {
				return c.Next()
			} else if b, err := strconv.ParseBool(dumpSwitch); err != nil || !b {
				return c.Next()
			}
		}

		// 1. 捕获请求体
		reqBody := c.Body()
		if len(reqBody) > 0 {
			if len(prompt) > 0 {
				fmt.Fprintf(dumper, "--- %s begin ---\n", prompt)
			}
			fmt.Fprintf(dumper, "%s\n", reqBody)
			if len(prompt) > 0 {
				fmt.Fprintf(dumper, "--- %s end ---\n", prompt)
			}
		}

		// 2. 执行后续中间件与业务逻辑
		err := c.Next()

		if len(respPrompt) > 0 {
			// 3. 捕获响应体（必须在 c.Next() 之后）
			resBody := c.Response().Body()

			// 4. 输出响应体
			if len(resBody) > 0 {
				fmt.Fprintf(dumper, "--- %s begin ---\n", respPrompt)
				fmt.Fprintf(dumper, "%s\n", resBody)
				fmt.Fprintf(dumper, "--- %s end ---\n", respPrompt)
			}
		}

		return err
	}
}
