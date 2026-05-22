package xf

import (
	"github.com/gofiber/fiber/v3"
	"github.com/go-playground/validator/v10"
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"bytes"
	"io"
	"fmt"
	"net/url"
)

type Context struct {
	*fiber.DefaultCtx
	q url.Values
}

func (c *Context) Param(name string) string {
	return c.DefaultCtx.Params(name)
}

// read values from path, query string, form, header or cookie, store them in a struct specified by vals.
// param name is specified by field tag. e.g.
//
// var vals struct {
//    V1 int    `path:"v1" optional`       // read path param "v1" optionally, any integer type (u)int8/16/32/64 is acceptable
//    V2 bool   `query:"v2" ignore-error`  // read query param "v2=xxx", ignore error occurring
//    V3 string `form:"v3"`                // read form param "v3=xxx", type string can be replace with []byte
//    V4 int    `header:"Content-Length"`  // read header "Content-Length"
//    V5 []byte `cookie:"v5" optional`     // read cookie "v4" opiontally
// }
// if status, err := c.ReadParams(&vals); err != nil {
//    c.Error(status, err.Error())
//    return
// }
func (c *Context) ReadParams(vals interface{}) (status int, err error) {
	return c.readParams(vals, false)
}

// read values from path, query string, form, header or cookie, store them in a struct specified by vals,
// then values are validated by using "github.com/go-playground/validator/v10".
// param name is specified by field tag. e.g.
//
// var vals struct {
//    V1 int    `path:"v1" validate:"gt=0"`     // read path param "v1", the value must be greater than 0
//    V2 bool   `query:"v2"`                    // read query param "v2=xxx"
//    V3 string `form:"v3" validate:"required"` // read form param "v3=xxx",
//    V4 int    `header:"Content-Length"`       // read header "Content-Length"
//    V5 []byte `cookie:"v5"`                   // read cookie "v4"
// }
// if status, err := c.ReadAndValidate(&vals); err != nil {
//    c.Error(status, err.Error())
//    return
// }
func (c *Context) ReadAndValidate(vals interface{}) (status int, err error) {
	if status, err = c.readParams(vals, true); err != nil {
		return
	}
	v := validator.New()
	if err = v.Struct(vals); err != nil {
		status = http.StatusBadRequest
	}
	return
}

func (c *Context) readParams(vals interface{}, onlyRead bool) (status int, err error) {
	type tagHandler struct {
		tagName string
		getVal  func(key string, defaultValue ...string) string
	}

	if vals == nil {
		return http.StatusOK, nil
	}
	p := reflect.ValueOf(vals)
	if p.Kind() != reflect.Ptr {
		return http.StatusInternalServerError, fmt.Errorf("vals must be pointer")
	}
	v := p.Elem() // struct Value
	if v.Kind() != reflect.Struct {
		return http.StatusInternalServerError, fmt.Errorf("vals must be pointer of struct")
	}
	t := v.Type() // struct Type
	n := t.NumField()
	tagHandlers := []tagHandler{
		tagHandler{"path",  c.DefaultCtx.Params},
		tagHandler{"query", c.DefaultCtx.Query},
		tagHandler{"form",  c.DefaultCtx.FormValue},
		tagHandler{"header",c.DefaultCtx.Get},
		tagHandler{"cookie",c.DefaultCtx.Cookies},
	}
	for i:=0; i<n; i++ {
		field := t.Field(i) // StructField
		val := ""
		for _, tagHandler := range tagHandlers {
			if tag, ok := field.Tag.Lookup(tagHandler.tagName); ok {
				val = tagHandler.getVal(tag)
				break
			}
		}
		if !onlyRead {
			if len(val) == 0 {
				if _, optional := field.Tag.Lookup("optional"); optional {
					continue
				}
				return http.StatusBadRequest, fmt.Errorf("no value specified for field %s", field.Name)
			}
		}
		_, ignoreError := field.Tag.Lookup("ignore-error")
		ignoreError = !onlyRead && ignoreError

		fv := v.Field(i) // field Value
		ft := field.Type // field Type
		switch ft.Kind() {
		case reflect.String:
			fv.SetString(val)
		case reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64:
			i, err := strconv.ParseInt(val, 10, ft.Bits())
			if err != nil {
				if !ignoreError { return http.StatusBadRequest, err }
				continue
			}
			fv.SetInt(i)
		case reflect.Uint,reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64:
			i, err := strconv.ParseUint(val, 10, ft.Bits())
			if err != nil {
				if !ignoreError { return http.StatusBadRequest, err }
				continue
			}
			fv.SetUint(i)
		case reflect.Float64,reflect.Float32:
			f, err := strconv.ParseFloat(val, ft.Bits())
			if err != nil {
				if !ignoreError { return http.StatusBadRequest, err }
				continue
			}
			fv.SetFloat(f)
		case reflect.Bool:
			b, err := strconv.ParseBool(val)
			if err != nil {
				if !ignoreError { return http.StatusBadRequest, err }
				continue
			}
			fv.SetBool(b)
		case reflect.Slice:
			if ft.Elem().Kind() == reflect.Uint8 {
				fv.SetBytes([]byte(val))
				break
			}
			fallthrough
		default:
			return http.StatusNotImplemented, fmt.Errorf("value of type %s not implemented", ft.Name())
		}
	}
	return http.StatusOK, nil
}

func (c *Context) QueryParam(name string) string {
	return c.DefaultCtx.Query(name)
}

func (c *Context) GetQueryParam(name string) (string, bool) {
	if c.q == nil {
		c.q = c.QueryParams()
	}
	if val, ok := c.q[name]; ok {
		return val[0], true
	} else {
		return "", false
	}
}

func (c *Context) QueryParams() url.Values {
	if c.q == nil {
		qs := c.QueryString()
		if len(qs) == 0 {
			c.q = url.Values{}
		} else {
			if u, err := url.ParseQuery(qs); err != nil {
				c.q = url.Values{}
			} else {
				c.q = u
			}
		}
	}
	return c.q
}

func (c *Context) QueryString() string {
	return string(c.DefaultCtx.Request().URI().QueryString())
}

func (c *Context) CookieValue(n string) string {
	return c.DefaultCtx.Cookies(n)
}

func (c *Context) SetCookie(cookie *http.Cookie) {
	c.DefaultCtx.Cookie(&fiber.Cookie{
		Name: cookie.Name,
		Value: cookie.Value,
		Path: cookie.Path,
		Domain: cookie.Domain,
		MaxAge: cookie.MaxAge,
		Expires: cookie.Expires,
		Secure: cookie.Secure,
		HTTPOnly: cookie.HttpOnly,
		SameSite: func()string{
			switch cookie.SameSite {
			case http.SameSiteLaxMode:     return "Lax"
			case http.SameSiteStrictMode:  return "Strict"
			case http.SameSiteNoneMode:    return "None"
			case http.SameSiteDefaultMode: return "Default"
			default: return "Default"
			}
		}(),
		Partitioned: cookie.Partitioned,
	})
}

func (c *Context) Header(name string) string {
	return c.DefaultCtx.Get(name)
}

func (c *Context) SetHeader(key, value string) {
	c.DefaultCtx.Set(key, value)
}

func (c *Context) AddHeader(key, value string) {
	c.DefaultCtx.Append(key, value)
}

func (c *Context) WriteHeader(statusCode int, contentType ...string) {
	if len(contentType) > 0 && len(contentType[0]) > 0 {
		c.writeContentType(contentType[0])
	}
	c.DefaultCtx.Status(statusCode)
}

func (c *Context) writeContentType(contentType string) {
	c.DefaultCtx.Set(fiber.HeaderContentType, contentType)
}

func (c *Context) Blob(code int, contentType string, b []byte) (err error) {
	c.WriteHeader(code, contentType)
	return c.DefaultCtx.Send(b)
}

func (c *Context) json(code int, i interface{}, indent string) error {
	return c.DefaultCtx.Status(code).JSON(i)
	/*
	enc := json.NewEncoder(c.DefaultCtx.DefaultRes)
	enc.SetEscapeHTML(false)
	if indent != "" {
		enc.SetIndent("", indent)
	}
	c.writeContentType(fiber.MIMEApplicationJSONCharsetUTF8)
	c.Status(code)
	return enc.Encode(i)
	*/
}

func (c *Context) JSONBlob(code int, b []byte) (err error) {
	return c.Blob(code, fiber.MIMEApplicationJSONCharsetUTF8, b)
}

func (c *Context) JSONPretty(code int, i interface{}, indent string) (err error) {
	return c.json(code, i, indent)
}

func (c *Context) Stream(code int, contentType string, r io.Reader) (err error) {
	c.writeContentType(contentType)
	c.DefaultCtx.Status(code)
	return c.DefaultCtx.SendStream(r)
}

func NotFoundHandler(c *Context) (err error) {
	return c.Error(http.StatusNotFound, "File not found")
}

func (c *Context) Error(code int, msg string) (err error) {
	return c.json(code, map[string]interface{}{"code":code, "msg":msg}, "")
}

func (c *Context) File(file string) (err error) {
	return c.DefaultCtx.SendFile(file)
}

func (c *Context) contentDisposition(file, name, dispositionType string) error {
	c.DefaultCtx.Set(fiber.HeaderContentDisposition, fmt.Sprintf("%s;filename=%s", dispositionType, name))
	return c.DefaultCtx.SendFile(file)
}

func (c *Context) Inline(file, name string) error {
	return c.contentDisposition(file, name, "inline")
}

func (c *Context) ReadJSON(res interface{}, dumper ...io.Writer) (code int, err error) {
	b := c.DefaultCtx.Body()
	if len(b) == 0 {
		return http.StatusBadRequest, fmt.Errorf("bad request")
	}

	body := bytes.NewReader(b)
	var realDumper io.Writer
	if len(dumper) > 0 && dumper[0] != nil {
		realDumper = dumper[0]
	}
	reqBody, deferFunc := bodyDumper(body, realDumper)
	defer deferFunc()

	jr := json.NewDecoder(reqBody)
	jr.UseNumber()
	if err = jr.Decode(res); err != nil {
		return http.StatusBadRequest, err
	}

	return http.StatusOK, nil
}

func (c *Context) ReadAndValidateJSON(res interface{}, dumper ...io.Writer) (code int, err error) {
	if code, err = c.ReadJSON(res, dumper...); err != nil {
		return
	}
	v := validator.New()
	if err = v.Struct(res); err != nil {
		code = http.StatusBadRequest
	}
	return
}

/*
func (c *Context) SSEvent(name string, message interface{}) {
	sse.Event{Event: name, Data: message}.Render(c.w)
}*/

func (c *Context) RemoteAddr() string {
	if addr := c.DefaultCtx.Get("X-Forwarded-For"); len(addr) > 0 {
		return addr
	}
	return c.DefaultCtx.IP()
}
