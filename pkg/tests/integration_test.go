package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/xKARASb/Calculator/internal/agent"
	"github.com/xKARASb/Calculator/internal/orchestrator/delivery/rest/routes"
	"github.com/xKARASb/Calculator/internal/orchestrator/repository"
	"github.com/xKARASb/Calculator/internal/orchestrator/service"
	"github.com/xKARASb/Calculator/pkg/db/cache"
	"github.com/xKARASb/Calculator/pkg/db/postgres"
	"github.com/xKARASb/Calculator/pkg/utils/statuses"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	integrationTestPort = "8081"
	integrationTimeout  = 10 * time.Second
	integrationBaseURL  = "http://localhost:" + integrationTestPort
)

// Сервер для интеграционных тестов
type integrationServer struct {
	echo   *echo.Echo
	client *http.Client
}

func setupIntegrationServer(t *testing.T) (*integrationServer, func()) {
	// Если установлена переменная окружения SKIP_INTEGRATION, пропускаем тест
	if os.Getenv("SKIP_INTEGRATION") == "true" {
		t.Skip("Пропускаем интеграционный тест, установлена переменная SKIP_INTEGRATION=true")
	}

	// Конфигурация для подключения к базе данных и Redis
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

	// Подключение к базе данных
	db, err := postgres.New(pgConfig)
	if err != nil {
		t.Skipf("Пропускаем интеграционный тест из-за ошибки подключения к БД: %v", err)
		return nil, func() {}
	}

	// Подключение к Redis
	redisClient := cache.New(redisConfig)
	_, err = redisClient.Ping(context.Background())
	if err != nil {
		t.Skipf("Пропускаем интеграционный тест из-за ошибки подключения к Redis: %v", err)
		return nil, func() {}
	}

	// Инициализация репозитория и сервиса
	ctx := context.Background()
	repo := repository.NewCalculatorRepository(ctx, db, redisClient)
	srv := service.NewCalculatorService(repo)

	// Настройка сервера Echo
	e := echo.New()
	routes.CalculatorRoutes(e, srv)

	// Добавляем эндпоинт для проверки здоровья
	e.GET("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	server := &integrationServer{
		echo: e,
		client: &http.Client{
			Timeout: integrationTimeout,
		},
	}

	// Запускаем сервер
	go func() {
		if err := e.Start(":" + integrationTestPort); err != nil && err != http.ErrServerClosed {
			t.Logf("Ошибка сервера: %v", err)
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
			t.Logf("Ошибка агента: %v", err)
		}
	}()

	// Ждем, пока сервер станет доступным
	waitForIntegrationServer(t, server.client)

	// Функция очистки ресурсов
	cleanup := func() {
		close(agentStop)
		server.echo.Close()
		for !agentStopped {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return server, cleanup
}

func waitForIntegrationServer(t *testing.T, client *http.Client) {
	deadline := time.Now().Add(integrationTimeout)
	for time.Now().Before(deadline) {
		_, err := client.Get(integrationBaseURL + "/health")
		if err == nil {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("Сервер не запустился в течение отведенного времени")
}

// Кейс для интеграционного теста
type integrationTestCase struct {
	name           string
	expression     string
	expectedStatus int
	expectedResult float64
	checkResult    bool
}

// Интеграционный тест для всей системы
func TestIntegrationCalculatorSystem(t *testing.T) {
	server, cleanup := setupIntegrationServer(t)
	if server == nil {
		return
	}
	defer cleanup()

	tests := []integrationTestCase{
		{
			name:           "Простое сложение",
			expression:     "2+2",
			expectedStatus: http.StatusOK,
			expectedResult: 4,
			checkResult:    true,
		},
		{
			name:           "Сложное выражение",
			expression:     "(10+2)*2",
			expectedStatus: http.StatusOK,
			expectedResult: 24,
			checkResult:    true,
		},
		{
			name:           "Числа с плавающей точкой",
			expression:     "3.14*2.0",
			expectedStatus: http.StatusOK,
			expectedResult: 6.28,
			checkResult:    true,
		},
		{
			name:           "Отрицательные числа",
			expression:     "-5+3",
			expectedStatus: http.StatusOK,
			expectedResult: -2,
			checkResult:    true,
		},
		{
			name:           "Несколько операций",
			expression:     "2+2*2/2-1",
			expectedStatus: http.StatusOK,
			expectedResult: 3,
			checkResult:    true,
		},
		{
			name:           "Недопустимое выражение - двойной оператор",
			expression:     "2++2",
			expectedStatus: http.StatusBadRequest,
			checkResult:    false,
		},
		{
			name:           "Деление на ноль",
			expression:     "1/0",
			expectedStatus: http.StatusBadRequest,
			checkResult:    false,
		},
		{
			name:           "Пустое выражение",
			expression:     "",
			expectedStatus: http.StatusBadRequest,
			checkResult:    false,
		},
		{
			name:           "Недопустимые символы",
			expression:     "2+a",
			expectedStatus: http.StatusBadRequest,
			checkResult:    false,
		},
		{
			name:           "Несбалансированные скобки",
			expression:     "(2+2",
			expectedStatus: http.StatusBadRequest,
			checkResult:    false,
		},
	}

	for _, tt := range tests {
		tt := tt // Создаем локальную копию для параллельного выполнения
		t.Run(tt.name, func(t *testing.T) {
			testIntegrationCalculation(t, server, tt)
		})
	}
}

func testIntegrationCalculation(t *testing.T, server *integrationServer, tt integrationTestCase) {
	url := fmt.Sprintf("%s/calculate?expression=%s", integrationBaseURL, tt.expression)

	resp, err := server.client.Get(url)
	require.NoError(t, err, "Не удалось выполнить запрос")
	defer resp.Body.Close()

	assert.Equal(t, tt.expectedStatus, resp.StatusCode, "Неожиданный статус код")

	if !tt.checkResult {
		return
	}

	var idResp struct {
		ID int `json:"id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&idResp), "Не удалось разобрать ответ")

	result := waitForIntegrationResult(t, server, idResp.ID)
	require.True(t, closeEnoughIntegration(result, tt.expectedResult),
		"Ожидаемый результат %v, получен %v", tt.expectedResult, result)
}

func waitForIntegrationResult(t *testing.T, server *integrationServer, id int) float64 {
	t.Helper()
	deadline := time.Now().Add(integrationTimeout)

	for time.Now().Before(deadline) {
		resultResp, err := server.client.Get(fmt.Sprintf("%s/expressions/%d", integrationBaseURL, id))
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

	t.Fatal("Не удалось получить результат вовремя")
	return 0
}

func closeEnoughIntegration(a, b float64) bool {
	const epsilon = 1e-9
	return math.Abs(a-b) < epsilon
}
