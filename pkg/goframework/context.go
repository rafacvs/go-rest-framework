package goframework

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sync"
)

type Context struct {
	Req    *http.Request
	Res    http.ResponseWriter
	Params map[string]string

	writer   *responseWriter
	bodyOnce sync.Once
	body     []byte
	bodyErr  error
}

func newContext(w *responseWriter, r *http.Request, params map[string]string) *Context {
	copiedParams := make(map[string]string, len(params))
	for k, v := range params {
		copiedParams[k] = v
	}

	return &Context{
		Req:    r,
		Res:    w,
		Params: copiedParams,
		writer: w,
	}
}

func (c *Context) JSON(status int, data any) error {
	if c.writer.wrote {
		return nil
	}

	c.writer.Header().Set("Content-Type", "application/json")
	c.writer.WriteHeader(status)
	return json.NewEncoder(c.writer).Encode(data)
}

func (c *Context) Param(name string) string {
	return c.Params[name]
}

func (c *Context) Query() url.Values {
	return c.Req.URL.Query()
}

func (c *Context) QueryParam(name string) string {
	return c.Query().Get(name)
}

func (c *Context) SetHeader(name, value string) {
	if c.writer.wrote {
		return
	}
	c.writer.Header().Set(name, value)
}

func (c *Context) Body() ([]byte, error) {
	c.bodyOnce.Do(func() {
		defer c.Req.Body.Close()
		c.body, c.bodyErr = io.ReadAll(c.Req.Body)
	})
	return c.body, c.bodyErr
}

type responseWriter struct {
	http.ResponseWriter
	wrote bool
}

func (w *responseWriter) WriteHeader(statusCode int) {
	if w.wrote {
		return
	}
	w.wrote = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.wrote {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}
