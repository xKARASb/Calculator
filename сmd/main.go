package main

import (
	app "https://github.com/xKARASb/calculator/internal/application"
)

func main() {
	a := app.New()
	a.RunServer() // Запускаем приложение
}
