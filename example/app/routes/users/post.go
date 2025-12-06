package users

import (
	"encoding/json"
	"my-go-framework/example/app/services"
	"my-go-framework/pkg/goframework"
)

func init() {
	goframework.Register(Post)
}

func Post(ctx *goframework.Context) error {
	body, err := ctx.Body()
	if err != nil {
		return goframework.NewHTTPError(400, "erro ao ler corpo da requisição")
	}

	var user services.User
	if err := json.Unmarshal(body, &user); err != nil {
		return goframework.NewHTTPError(400, "dados inválidos")
	}

	service := services.NewUserService()
	createdUser, err := service.CreateUser(user)
	if err != nil {
		return goframework.NewHTTPError(400, err.Error())
	}

	return ctx.JSON(201, createdUser)
}

