package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/xKARASb/Calculator/internal/agent"
	"github.com/xKARASb/Calculator/internal/orchestrator/repository"
	"github.com/xKARASb/Calculator/internal/orchestrator/service"
	"github.com/xKARASb/Calculator/pkg/db/cache"
	"github.com/xKARASb/Calculator/pkg/db/postgres"
	"github.com/xKARASb/Calculator/pkg/models"
	"github.com/xKARASb/Calculator/pkg/utils/statuses"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xKARASb/Calculator/internal/orchestrator/delivery/rest/routes"
)

const (
	testPort    = "8080"
	testTimeout = 5 * time.Second
	baseURL     = "http://localhost:" + testPort
)

// Мок для репозитория
type MockCalculatorRepository struct {
	repo repository.CalculatorRepository
}

func NewMockCalculatorRepository() *MockCalculatorRepository {
	return &MockCalculatorRepository{}
}

func (m *MockCalculatorRepository) Calculate(expression string) (int, error) {
	return 1, nil
}

func (m *MockCalculatorRepository) GetAllExpressions() ([]models.Expression, error) {
	return []models.Expression{}, nil
}

func (m *MockCalculatorRepository) GetExpressionByID(id int) (*models.Expression, error) {
	return &models.Expression{
		Expression: models.ExpressionData{
			ID:     id,
			Status: statuses.StatusComplete,
			Result: 4,
		},
	}, nil
}

func (m *MockCalculatorRepository) GetCurrentTask() (*models.Task, error) {
	return &models.Task{
		Task: models.TaskData{
			ID: 1,
		},
	}, nil
}

func (m *MockCalculatorRepository) GetResult() (*models.Result, error) {
	return &models.Result{
		ID:     1,
		Result: 4,
	}, nil
}

func (m *MockCalculatorRepository) SetExpression(expression models.Expression) error {
	return nil
}

func (m *MockCalculatorRepository) Register(login, password string) error {
	return nil
}

func (m *MockCalculatorRepository) Login(login, password string) error {
	return nil
}

func (m *MockCalculatorRepository) ValidateToken(token string) (int, error) {
	return 1, nil
}

// Unit тесты для сервиса калькулятора
func TestCalculatorService_Calculate(t *testing.T) {
	mockRepo := NewMockCalculatorRepository()

	svc := service.NewCalculatorService(mockRepo)

	id, err := svc.Calculate("2+2")

	assert.NoError(t, err)
	assert.Equal(t, 1, id)
}

func TestCalculatorService_GetExpressionByID(t *testing.T) {
	mockRepo := NewMockCalculatorRepository()
	expectedExpression := &models.Expression{
		Expression: models.ExpressionData{
			ID:     1,
			Status: statuses.StatusComplete,
			Result: 4,
		},
	}

	svc := service.NewCalculatorService(mockRepo)

	expr, err := svc.GetExpressionByID(1)

	assert.NoError(t, err)
	assert.Equal(t, expectedExpression.Expression.ID, expr.Expression.ID)
	assert.Equal(t, expectedExpression.Expression.Result, expr.Expression.Result)
}

// Вспомогательные функции для интеграционного тестирования
type testServer struct {
	echo   *echo.Echo
	client *http.Client
}

func setupMockServer(t *testing.T) (*testServer, *MockCalculatorRepository) {
	mockRepo := NewMockCalculatorRepository()
	srv := service.NewCalculatorService(mockRepo)

	e := echo.New()
	routes.CalculatorRoutes(e, srv)

	server := &testServer{
		echo: e,
		client: &http.Client{
			Timeout: testTimeout,
		},
	}

	return server, mockRepo
}

func setupIntegrationTestServer(t *testing.T) (*testServer, func()) {
	// Проверяем наличие переменных окружения для базы данных и redis
	pgConfig := postgres.PostgresConfig{
		Username: os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		DbName:   os.Getenv("POSTGRES_DB"),
	}

	redisConfig := cache.RedisConfig{
		Host: os.Getenv("REDIS_HOST"),
		Port: os.Getenv("REDIS_PORT"),
	}

	db, err := postgres.New(pgConfig)
	if err != nil {
		t.Skipf("Skipping integration test due to database error: %v", err)
		return nil, func() {}
	}

	redisClient := cache.New(redisConfig)
	_, err = redisClient.Ping(context.Background())
	if err != nil {
		t.Skipf("Skipping integration test due to Redis error: %v", err)
		return nil, func() {}
	}

	ctx := context.Background()
	repo := repository.NewCalculatorRepository(ctx, db, redisClient)
	srv := service.NewCalculatorService(repo)

	e := echo.New()
	routes.CalculatorRoutes(e, srv)

	e.GET("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	server := &testServer{
		echo: e,
		client: &http.Client{
			Timeout: testTimeout,
		},
	}

	// Запускаем сервер на случайном порту
	go func() {
		if err := e.Start(":" + testPort); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
	}()

	// Запускаем агента
	var agentStopped bool
	agentStop := make(chan struct{})
	go func() {
		a := agent.NewAgent(1)
		go func() {
			<-agentStop
			agentStopped = true
		}()
		if err := a.CalculateExpression(); err != nil {
			t.Logf("Agent error: %v", err)
		}
	}()

	// Ждем, пока сервер станет доступным
	waitForServer(t, server.client)

	// Функция cleanup для закрытия ресурсов
	cleanup := func() {
		close(agentStop)
		server.echo.Close()
		for !agentStopped {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return server, cleanup
}

func waitForServer(t *testing.T, client *http.Client) {
	deadline := time.Now().Add(testTimeout)
	for time.Now().Before(deadline) {
		_, err := client.Get(baseURL + "/health")
		if err == nil {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("Server did not become ready within timeout")
}

// Тесты API с использованием моков
func TestCalculatorAPI_Calculate_Mock(t *testing.T) {
	server, _ := setupMockServer(t)

	// Создаем запрос
	req := httptest.NewRequest(http.MethodGet, "/calculate?expression=2+2", nil)
	rec := httptest.NewRecorder()

	// Выполняем запрос
	server.echo.ServeHTTP(rec, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		ID int `json:"id"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 1, resp.ID)
}

func TestCalculatorAPI_GetExpression_Mock(t *testing.T) {
	server, _ := setupMockServer(t)

	// Создаем запрос
	req := httptest.NewRequest(http.MethodGet, "/expressions/1", nil)
	rec := httptest.NewRecorder()

	// Выполняем запрос
	server.echo.ServeHTTP(rec, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp models.Expression
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 1, resp.Expression.ID)
	assert.Equal(t, statuses.StatusComplete, resp.Expression.Status)
	assert.Equal(t, float64(4), resp.Expression.Result)
}

// Интеграционные тесты
type testCase struct {
	name           string
	expression     string
	expectedStatus int
	expectedResult float64
	checkResult    bool
}

func TestCalculatorAPI_Integration(t *testing.T) {
	// Пропускаем тест, если запущен в режиме unit-тестов
	if os.Getenv("SKIP_INTEGRATION") == "true" {
		t.Skip("Skipping integration test")
	}

	server, cleanup := setupIntegrationTestServer(t)
	if server == nil {
		t.Skip("Failed to setup integration test server")
		return
	}
	defer cleanup()

	tests := []testCase{
		{
			name:           "Simple Addition",
			expression:     "2+2",
			expectedStatus: http.StatusOK,
			expectedResult: 4,
			checkResult:    true,
		},
		{
			name:           "Complex Expression",
			expression:     "(10+2)*2",
			expectedStatus: http.StatusOK,
			expectedResult: 24,
			checkResult:    true,
		},
		{
			name:           "Floating Point Numbers",
			expression:     "3.14*2.0",
			expectedStatus: http.StatusOK,
			expectedResult: 6.28,
			checkResult:    true,
		},
		{
			name:           "Negative Numbers",
			expression:     "-5+3",
			expectedStatus: http.StatusOK,
			expectedResult: -2,
			checkResult:    true,
		},
		{
			name:           "Multiple Operations",
			expression:     "2+2*2/2-1",
			expectedStatus: http.StatusOK,
			expectedResult: 3,
			checkResult:    true,
		},
		{
			name:           "Invalid Expression - Double Operator",
			expression:     "2++2",
			expectedStatus: http.StatusBadRequest,
			checkResult:    false,
		},
		{
			name:           "Division by Zero",
			expression:     "1/0",
			expectedStatus: http.StatusBadRequest,
			checkResult:    false,
		},
		{
			name:           "Empty Expression",
			expression:     "",
			expectedStatus: http.StatusBadRequest,
			checkResult:    false,
		},
		{
			name:           "Invalid Characters",
			expression:     "2+a",
			expectedStatus: http.StatusBadRequest,
			checkResult:    false,
		},
		{
			name:           "Unbalanced Parentheses",
			expression:     "(2+2",
			expectedStatus: http.StatusBadRequest,
			checkResult:    false,
		},
	}

	for _, tt := range tests {
		tt := tt // Создаем локальную копию для параллельного выполнения
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testCalculation(t, server, tt)
		})
	}
}

func testCalculation(t *testing.T, server *testServer, tt testCase) {
	url := fmt.Sprintf("%s/calculate?expression=%s", baseURL, tt.expression)

	resp, err := server.client.Get(url)
	require.NoError(t, err, "Failed to send request")
	defer resp.Body.Close()

	assert.Equal(t, tt.expectedStatus, resp.StatusCode, "Unexpected status code")

	if !tt.checkResult {
		return
	}

	var idResp struct {
		ID int `json:"id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&idResp), "Failed to parse response")

	result := waitForResult(t, server, idResp.ID)
	require.True(t, closeEnough(result, tt.expectedResult),
		"Expected result %v, got %v", tt.expectedResult, result)
}

func waitForResult(t *testing.T, server *testServer, id int) float64 {
	t.Helper()
	deadline := time.Now().Add(testTimeout)

	for time.Now().Before(deadline) {
		resultResp, err := server.client.Get(fmt.Sprintf("%s/expressions/%d", baseURL, id))
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		defer resultResp.Body.Close()

		if resultResp.StatusCode != http.StatusOK {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		var result struct {
			Expression struct {
				Result float64 `json:"result"`
				Status string  `json:"status"`
			} `json:"expression"`
		}

		if err := json.NewDecoder(resultResp.Body).Decode(&result); err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if result.Expression.Status == statuses.StatusComplete {
			return result.Expression.Result
		}

		time.Sleep(100 * time.Millisecond)
	}

	t.Fatal("Failed to get result in time")
	return 0
}

func closeEnough(a, b float64) bool {
	const epsilon = 1e-9
	return math.Abs(a-b) < epsilon
}

func stringReader(s string) io.Reader {
	return strings.NewReader(s)
}
