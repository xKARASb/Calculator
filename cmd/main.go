package main

import (
	app "github.com/xKARASb/Calculator/internal/application"
)

func main() {
	a := app.New()
	a.RunServer() // Запускаем приложение
}
