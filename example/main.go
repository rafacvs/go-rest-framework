package main

import (
	"log"

	"my-go-framework/pkg/goframework"

	_ "my-go-framework/example/app/routes/index"
	_ "my-go-framework/example/app/routes/users"
	_ "my-go-framework/example/app/routes/users/_id"
)

func main() {
	app := goframework.New()

	err := app.LoadRoutes("example/app/routes")
	if err != nil {
		log.Fatalf("erro carregando rotas: %v", err)
	}

	log.Println("servidor rodando em http://localhost:8080")
	err = app.Listen(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
