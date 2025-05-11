package controllers

import (
	"net/http"
	"strconv"

	"github.com/xKARASb/Calculator/pkg/models"
	"github.com/xKARASb/Calculator/pkg/utils/jwt"

	"github.com/labstack/echo/v4"
)

type CalculatorService interface {
	Calculate(expression string) (int, error)
	GetAllExpressions() ([]models.Expression, error)
	GetExpressionByID(id int) (*models.Expression, error)
	GetCurrentTask() (*models.Task, error)
	GetResult() (*models.Result, error)
	SetExpression(expression models.Expression) error
	Register(login string, password string) error
	Login(login string, password string) error
}

type CalculatorController struct {
	CalculatorService CalculatorService
}

func NewCalculatorController(calculatorService CalculatorService) *CalculatorController {
	return &CalculatorController{CalculatorService: calculatorService}
}

func (cc *CalculatorController) Calculate(c echo.Context) error {
	var request models.Request

	if err := c.Bind(&request); err != nil {
		err = c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
		return err
	}
	id, err := cc.CalculatorService.Calculate(request.Expression)
	if err != nil {
		err = c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		return err
	}
	response := models.Response{ID: id}
	err = c.JSON(http.StatusOK, response)
	if err != nil {
		return err
	}
	return nil
}

func (cc *CalculatorController) GetAllExpressions(c echo.Context) error {
	expressions, err := cc.CalculatorService.GetAllExpressions()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, expressions)
}

func (cc *CalculatorController) GetExpressionByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	expression, err := cc.CalculatorService.GetExpressionByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, expression)
}

func (cc *CalculatorController) GetCurrentTask(c echo.Context) error {
	task, err := cc.CalculatorService.GetCurrentTask()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, task)
}

func (cc *CalculatorController) GetResult(c echo.Context) error {
	result, err := cc.CalculatorService.GetResult()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}

func (cc *CalculatorController) SetExpression(c echo.Context) error {
	var request models.Expression

	if err := c.Bind(&request); err != nil {
		err = c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
		return err
	}
	err := cc.CalculatorService.SetExpression(request)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"status": "success"})
}

func (cc *CalculatorController) Register(c echo.Context) error {
	var request models.Auth

	if err := c.Bind(&request); err != nil {
		err = c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
		return err
	}
	err := cc.CalculatorService.Register(request.Login, request.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"status": "success"})
}

func (cc *CalculatorController) Login(c echo.Context) error {
	var request models.Auth

	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}

	err := cc.CalculatorService.Login(request.Login, request.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
	}

	token := jwt.NewAccessToken(1, "secret")

	return c.JSON(http.StatusOK, echo.Map{
		"status": "success",
		"token":  token,
	})
}
