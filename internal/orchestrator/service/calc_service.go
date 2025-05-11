package service

import (
	"github.com/xKARASb/Calculator/pkg/models"
)

type CalculatorRepository interface {
	Calculate(expression string) (int, error)
	GetAllExpressions() ([]models.Expression, error)
	GetExpressionByID(id int) (*models.Expression, error)
	GetCurrentTask() (*models.Task, error)
	GetResult() (*models.Result, error)
	SetExpression(expression models.Expression) error
	Register(login, password string) error
	Login(login, password string) error
}

type CalculatorService struct {
	repository CalculatorRepository
}

func NewCalculatorService(repo CalculatorRepository) CalculatorService {
	return CalculatorService{repository: repo}
}

func (s CalculatorService) Calculate(expression string) (int, error) {
	return s.repository.Calculate(expression)
}

func (s CalculatorService) GetAllExpressions() ([]models.Expression, error) {
	return s.repository.GetAllExpressions()
}

func (s CalculatorService) GetExpressionByID(id int) (*models.Expression, error) {
	return s.repository.GetExpressionByID(id)
}

func (s CalculatorService) GetCurrentTask() (*models.Task, error) {
	return s.repository.GetCurrentTask()
}

func (s CalculatorService) GetResult() (*models.Result, error) {
	return s.repository.GetResult()
}

func (s CalculatorService) SetExpression(expression models.Expression) error {
	return s.repository.SetExpression(expression)
}
func (s CalculatorService) Register(login, password string) error {
	return s.repository.Register(login, password)
}

func (s CalculatorService) Login(login, password string) error {
	return s.repository.Login(login, password)
}
