package users

import (
	"my-go-framework/example/app/services"
	"my-go-framework/pkg/goframework"
)

func init() {
	goframework.Register(Get)
}

func Get(ctx *goframework.Context) error {
	service := services.NewUserService()
	users, err := service.GetAllUsers()
	if err != nil {
		return goframework.NewHTTPError(500, "erro ao buscar usuários")
	}

	return ctx.JSON(200, users)
}

