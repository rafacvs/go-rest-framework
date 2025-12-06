package index

import (
	"my-go-framework/pkg/goframework"
)

func init() {
	goframework.Register(Get)
}

func Get(ctx *goframework.Context) error {
	return ctx.JSON(200, map[string]string{
		"message": "hello from example route",
	})
}
