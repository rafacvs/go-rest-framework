package id

import (
	"encoding/json"
	"my-go-framework/example/app/services"
	"my-go-framework/pkg/goframework"
	"strconv"
)

func init() {
	goframework.Register(Put)
}

func Put(ctx *goframework.Context) error {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return goframework.NewHTTPError(400, "id inválido")
	}

	body, err := ctx.Body()
	if err != nil {
		return goframework.NewHTTPError(400, "erro ao ler corpo da requisição")
	}

	var user services.User
	if err := json.Unmarshal(body, &user); err != nil {
		return goframework.NewHTTPError(400, "dados inválidos")
	}

	service := services.NewUserService()
	updatedUser, err := service.UpdateUser(id, user)
	if err != nil {
		if err.Error() == "usuário não encontrado" {
			return goframework.NewHTTPError(404, err.Error())
		}
		return goframework.NewHTTPError(400, err.Error())
	}

	return ctx.JSON(200, updatedUser)
}

