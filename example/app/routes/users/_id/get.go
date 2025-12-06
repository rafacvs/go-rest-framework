package id

import (
	"my-go-framework/example/app/services"
	"my-go-framework/pkg/goframework"
	"strconv"
)

func init() {
	goframework.Register(Get)
}

func Get(ctx *goframework.Context) error {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return goframework.NewHTTPError(400, "id inválido")
	}

	service := services.NewUserService()
	user, err := service.GetUserByID(id)
	if err != nil {
		return goframework.NewHTTPError(404, "usuário não encontrado")
	}

	return ctx.JSON(200, user)
}

