package routes

import (
	"github.com/xKARASb/Calculator/internal/orchestrator/delivery/rest/controllers"
	"github.com/xKARASb/Calculator/internal/orchestrator/service"
	"github.com/xKARASb/Calculator/pkg/utils/jwt"

	"github.com/labstack/echo/v4"
)

func CalculatorRoutes(e *echo.Echo, calculatorService service.CalculatorService) {
	CalculatorController := controllers.NewCalculatorController(calculatorService)
	e.Static("/", "./bin/web") // ./bin/ когда уже в докере

	e.POST("/api/v1/register", CalculatorController.Register)
	e.POST("/api/v1/login", CalculatorController.Login)

	api := e.Group("/api/v1")
	api.Use(jwt.JWTAuth)

	api.POST("/calculate", CalculatorController.Calculate)
	api.GET("/expressions", CalculatorController.GetAllExpressions)
	api.GET("/expressions/:id", CalculatorController.GetExpressionByID)

	internal := e.Group("/internal")
	internal.GET("/task", CalculatorController.GetCurrentTask)
	internal.POST("/task", CalculatorController.GetResult)

	data := e.Group("/data")
	data.POST("/setExpression", CalculatorController.SetExpression)
}
