package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xKARASb/Calculator/internal/orchestrator/service"
	"github.com/xKARASb/Calculator/pkg/models"
	"github.com/xKARASb/Calculator/pkg/utils/statuses"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/xKARASb/Calculator/internal/orchestrator/delivery/rest/routes"
)

// Тесты для сервиса калькулятора
func TestUnitCalculatorService_Calculate(t *testing.T) {
	mockRepo := NewMockCalculatorRepository()
	svc := service.NewCalculatorService(mockRepo)

	id, err := svc.Calculate("2+2")

	assert.NoError(t, err)
	assert.Equal(t, 1, id)
}

func TestUnitCalculatorService_GetExpressionByID(t *testing.T) {
	mockRepo := NewMockCalculatorRepository()
	svc := service.NewCalculatorService(mockRepo)

	expr, err := svc.GetExpressionByID(1)

	assert.NoError(t, err)
	assert.Equal(t, 1, expr.Expression.ID)
	assert.Equal(t, float64(4), expr.Expression.Result)
	assert.Equal(t, statuses.StatusComplete, expr.Expression.Status)
}

func TestUnitCalculatorService_GetAllExpressions(t *testing.T) {
	mockRepo := NewMockCalculatorRepository()
	svc := service.NewCalculatorService(mockRepo)

	expressions, err := svc.GetAllExpressions()

	assert.NoError(t, err)
	assert.Empty(t, expressions)
}

func TestUnitCalculatorService_GetCurrentTask(t *testing.T) {
	mockRepo := NewMockCalculatorRepository()
	svc := service.NewCalculatorService(mockRepo)

	task, err := svc.GetCurrentTask()

	assert.NoError(t, err)
	assert.Equal(t, 1, task.Task.ID)
}

func TestUnitCalculatorService_GetResult(t *testing.T) {
	mockRepo := NewMockCalculatorRepository()
	svc := service.NewCalculatorService(mockRepo)

	result, err := svc.GetResult()

	assert.NoError(t, err)
	assert.Equal(t, 1, result.ID)
	assert.Equal(t, float64(4), result.Result)
}

// Тесты для API калькулятора
func TestUnitCalculatorAPI_Calculate(t *testing.T) {
	// Настройка
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/calculate?expression=2+2", nil)
	rec := httptest.NewRecorder()

	// Создаем сервис с мок-репозиторием
	mockRepo := NewMockCalculatorRepository()
	svc := service.NewCalculatorService(mockRepo)

	// Настраиваем ручки API
	routes.CalculatorRoutes(e, svc)

	// Выполняем запрос
	e.ServeHTTP(rec, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUnitCalculatorAPI_GetExpression(t *testing.T) {
	// Настройка
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/expressions/1", nil)
	rec := httptest.NewRecorder()

	// Создаем сервис с мок-репозиторием
	mockRepo := NewMockCalculatorRepository()
	svc := service.NewCalculatorService(mockRepo)

	// Настраиваем ручки API
	routes.CalculatorRoutes(e, svc)

	// Выполняем запрос
	e.ServeHTTP(rec, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, rec.Code)

	// Десериализуем ответ
	var response models.Expression
	err := json.Unmarshal(rec.Body.Bytes(), &response)

	assert.NoError(t, err)
	assert.Equal(t, 1, response.Expression.ID)
	assert.Equal(t, float64(4), response.Expression.Result)
	assert.Equal(t, statuses.StatusComplete, response.Expression.Status)
}

func TestUnitCalculatorAPI_GetAllExpressions(t *testing.T) {
	// Настройка
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/expressions", nil)
	rec := httptest.NewRecorder()

	// Создаем сервис с мок-репозиторием
	mockRepo := NewMockCalculatorRepository()
	svc := service.NewCalculatorService(mockRepo)

	// Настраиваем ручки API
	routes.CalculatorRoutes(e, svc)

	// Выполняем запрос
	e.ServeHTTP(rec, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUnitCalculatorAPI_InvalidExpression(t *testing.T) {
	tests := []struct {
		name       string
		expression string
	}{
		{
			name:       "Пустое выражение",
			expression: "",
		},
		{
			name:       "Некорректный оператор",
			expression: "2++2",
		},
		{
			name:       "Неподдерживаемые символы",
			expression: "2+a",
		},
		{
			name:       "Несбалансированные скобки",
			expression: "(2+2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настройка
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/calculate?expression="+tt.expression, nil)
			rec := httptest.NewRecorder()

			// Создаем сервис с мок-репозиторием, который будет возвращать ошибку
			mockRepo := NewErrorMockCalculatorRepository()
			svc := service.NewCalculatorService(mockRepo)

			// Настраиваем ручки API
			routes.CalculatorRoutes(e, svc)

			// Выполняем запрос
			e.ServeHTTP(rec, req)

			// Проверяем, что вернулся статус BadRequest
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
	}
}

// Мок репозитория, который всегда возвращает ошибку
type ErrorMockCalculatorRepository struct {
	*MockCalculatorRepository
}

func NewErrorMockCalculatorRepository() *ErrorMockCalculatorRepository {
	return &ErrorMockCalculatorRepository{
		MockCalculatorRepository: NewMockCalculatorRepository(),
	}
}

func (m *ErrorMockCalculatorRepository) Calculate(expression string) (int, error) {
	return 0, echo.NewHTTPError(http.StatusBadRequest, "Invalid expression")
}
