package id

import (
	"my-go-framework/example/app/services"
	"my-go-framework/pkg/goframework"
	"strconv"
)

func init() {
	goframework.Register(Delete)
}

func Delete(ctx *goframework.Context) error {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return goframework.NewHTTPError(400, "id inválido")
	}

	service := services.NewUserService()
	if err := service.DeleteUser(id); err != nil {
		if err.Error() == "usuário não encontrado" {
			return goframework.NewHTTPError(404, err.Error())
		}
		return goframework.NewHTTPError(500, "erro ao deletar usuário")
	}

	return ctx.JSON(200, map[string]string{"message": "usuário deletado com sucesso"})
}

