package repository

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"sync"
	"time"
	"unicode"

	"github.com/xKARASb/Calculator/pkg/db/cache"
	"github.com/xKARASb/Calculator/pkg/db/postgres"
	"github.com/xKARASb/Calculator/pkg/models"
	"github.com/xKARASb/Calculator/pkg/utils/errors"
	"github.com/xKARASb/Calculator/pkg/utils/hash"
	"github.com/xKARASb/Calculator/pkg/utils/jwt"
	"github.com/xKARASb/Calculator/pkg/utils/statuses"
	"github.com/xKARASb/Calculator/pkg/utils/timings"

	"github.com/lib/pq"
)

type CalculatorRepository struct {
	ctx   context.Context
	id    int
	task  chan models.Task
	db    *postgres.DB
	redis *cache.RedisClient
	mu    sync.Mutex
}

func NewCalculatorRepository(ctx context.Context, db *postgres.DB, redis *cache.RedisClient) *CalculatorRepository {
	repo := &CalculatorRepository{
		ctx:   ctx,
		id:    0,
		task:  make(chan models.Task, 1),
		db:    db,
		redis: redis,
	}

	lastID, err := repo.getLastID()
	if err == nil && lastID > 0 {
		repo.id = lastID
		log.Printf("Last ID was restored from Redis: %d", lastID)
	}

	return repo
}

func (r *CalculatorRepository) getLastID() (int, error) {
	ids, err := r.redis.SMembers(r.ctx, "expressions:all")
	if err != nil || len(ids) == 0 {
		return 0, errors.ErrNotFound
	}

	numIDs := make([]int, 0, len(ids))
	for _, idStr := range ids {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		numIDs = append(numIDs, id)
	}

	if len(numIDs) == 0 {
		return 0, errors.ErrNotFound
	}

	maxID := numIDs[0]
	for _, id := range numIDs {
		if id > maxID {
			maxID = id
		}
	}

	return maxID, nil
}

func (r *CalculatorRepository) Calculate(expression string) (int, error) {
	r.mu.Lock()

	if len(r.task) > 0 {
		<-r.task
	}

	r.id++
	id := r.id

	expr := models.Expression{Expression: models.ExpressionData{ID: id, Status: statuses.StatusPending}}
	err := r.SetExpression(expr)
	if err != nil {
		r.mu.Unlock()
		return 0, err
	}

	r.mu.Unlock()

	expr = models.Expression{Expression: models.ExpressionData{ID: id, Status: statuses.StatusProgress}}

	err = r.SetExpression(expr)
	if err != nil {
		return 0, err
	}

	task := models.Task{Task: models.TaskData{
		ID: r.id,
	}}

	r.task <- task

	err = r.prioritize(expression)
	if err == nil {
		log.Println("Passed id:", id)
	} else {
		log.Println("Failed id:", id)
		return 0, err
	}

	return id, err
}

func (r *CalculatorRepository) prioritize(expression string) (err error) {
	var values []float64
	var ops []byte

	task := models.Task{}
	task.Task.ID = r.id

	applyOperation := func(op byte, a, b float64) (float64, error) {
		task.Task.Arg1 = a
		task.Task.Arg2 = b
		task.Task.Operation = string(op)
		switch op {
		case '+':
			task.Task.OperationTime = timings.TimeAdditionMS
		case '-':
			task.Task.OperationTime = timings.TimeSubtractionMS
		case '*':
			task.Task.OperationTime = timings.TimeMultiplicationMS
		case '/':
			if b == 0 {
				return 0, errors.ErrDivisionByZero
			}
			task.Task.OperationTime = timings.TimeDivisionMS
		default:
			return 0, errors.ErrUnknownOperation
		}
		select {
		case <-r.task:
			r.task <- task
		default:
			return 0, errors.ErrTaskChanIsFull
		}

		for {
			expr, err := r.GetExpressionByID(r.id)
			if err != nil {
				return 0, err
			}
			if expr.Expression.Status == statuses.StatusComplete {
				r.mu.Lock()

				expr.Expression.Status = statuses.StatusProgress
				err = r.SetExpression(*expr)

				r.mu.Unlock()

				if err != nil {
					return 0, err
				}

				result, err := r.GetResult()
				if err != nil {
					return 0, err
				}
				return result.Result, nil
			}
			time.Sleep(1 * time.Millisecond)
		}
	}

	precedence := func(op byte) int {
		switch op {
		case '+', '-':
			return 1
		case '*', '/':
			return 2
		default:
			return 0
		}
	}

	applyTopOperation := func() error {
		if len(ops) == 0 || len(values) < 2 {
			return errors.ErrInvalidOperation
		}

		sec := values[len(values)-1]
		fst := values[len(values)-2]
		values = values[:len(values)-2]

		op := ops[len(ops)-1]
		ops = ops[:len(ops)-1]

		result, err := applyOperation(op, fst, sec)
		if err != nil {
			expr := models.Expression{Expression: models.ExpressionData{ID: r.id, Status: statuses.StatusError}}
			err = r.SetExpression(expr)
			if err != nil {
				return err
			}
			return err
		}
		values = append(values, result)
		return nil
	}

	for i := 0; i < len(expression); i++ {
		ch := expression[i]

		if ch == ' ' {
			continue
		}
		if unicode.IsDigit(rune(ch)) {
			start := i
			for i < len(expression) && (unicode.IsDigit(rune(expression[i])) || expression[i] == '.') {
				i++
			}
			value, err := strconv.ParseFloat(expression[start:i], 64)
			if err != nil {
				return errors.ErrInvalidExpression
			}
			values = append(values, value)
			i--
		} else if ch == '(' {
			ops = append(ops, ch)
		} else if ch == ')' {
			for len(ops) > 0 && ops[len(ops)-1] != '(' {
				if err = applyTopOperation(); err != nil {
					return err
				}
			}
			if len(ops) == 0 {
				return errors.ErrMismatchedParentheses
			}
			ops = ops[:len(ops)-1]
		} else if ch == '+' || ch == '-' || ch == '*' || ch == '/' {
			for len(ops) > 0 && precedence(ops[len(ops)-1]) >= precedence(ch) {
				if err = applyTopOperation(); err != nil {
					return err
				}
			}
			ops = append(ops, ch)
		} else {
			return errors.ErrUnknownOperation
		}
	}
	for len(ops) > 0 {
		if err = applyTopOperation(); err != nil {
			return err
		}
	}
	if len(values) != 1 {
		return errors.ErrInvalidExpression
	}

	return nil
}

func (r *CalculatorRepository) GetAllExpressions() ([]models.Expression, error) {
	ids, err := r.redis.SMembers(r.ctx, "expressions:all")
	if err != nil || len(ids) == 0 {
		return nil, errors.ErrNotFound
	}

	expressions := make([]models.Expression, 0, len(ids))

	for _, idStr := range ids {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}

		expr, err := r.GetExpressionByID(id)
		if err == nil {
			expressions = append(expressions, *expr)
		}
	}

	sort.Slice(expressions, func(i, j int) bool {
		return expressions[i].Expression.ID < expressions[j].Expression.ID
	})

	if len(expressions) == 0 {
		return nil, errors.ErrNotFound
	}

	return expressions, nil
}

func (r *CalculatorRepository) GetExpressionByID(id int) (*models.Expression, error) {
	key := fmt.Sprintf("expression:%d", id)

	data, err := r.redis.Get(r.ctx, key)
	if err != nil {
		return nil, errors.ErrNotFound
	}

	var expression models.Expression
	err = json.Unmarshal([]byte(data), &expression)
	if err != nil {
		return nil, err
	}

	return &expression, nil
}

func (r *CalculatorRepository) GetCurrentTask() (*models.Task, error) {

	select {
	case task := <-r.task:
		if task.Task.ID == 0 && task.Task.OperationTime == 0 {
			return nil, errors.ErrNotFound
		}

		r.task <- task

		return &task, nil
	default:
		return nil, errors.ErrNotAvailable
	}
}

func (r *CalculatorRepository) GetResult() (*models.Result, error) {
	select {
	case task := <-r.task:
		id := task.Task.ID
		if id == 0 {
			return nil, errors.ErrNotFound
		}

		result := &models.Result{
			ID: id,
		}

		expression, err := r.GetExpressionByID(id)
		if err != nil {
			return nil, err
		}

		result.Result = expression.Expression.Result

		r.task <- task

		return result, nil
	default:

		return nil, errors.ErrNotAvailable
	}
}

func (r *CalculatorRepository) SetExpression(expression models.Expression) error {
	jsonData, err := json.Marshal(expression)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("expression:%d", expression.Expression.ID)

	err = r.redis.Set(r.ctx, key, string(jsonData), 24*time.Hour)
	if err != nil {
		return err
	}

	err = r.redis.SAdd(r.ctx, "expressions:all", strconv.Itoa(expression.Expression.ID))
	return err
}

func (r *CalculatorRepository) Register(login, password string) error {
	password, err := hash.HashString(password)
	if err != nil {
		return err
	}

	query := `INSERT INTO public.users (login, password) VALUES ($1, $2);`
	_, err = r.db.Db.Exec(query, login, password)
	if err != nil {
		pgErr, ok := err.(*pq.Error)
		if ok && pgErr.Code == "23505" {
			return errors.ErrUserAlreadyExists
		}
		return err
	}
	return nil
}

func (r *CalculatorRepository) Login(login, password string) error {
	var user models.User

	query := `SELECT id, login, password FROM public.users WHERE login = $1;`
	err := r.db.Db.QueryRow(query, login).Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		return errors.ErrInvalidCredentials
	}

	success, err := hash.CheckStirngHash(password, user.Password)
	if err != nil {
		return err
	}

	if !success {
		return errors.ErrInvalidCredentials
	}

	refreshToken, err := jwt.NewRefreshToken()
	if err != nil {
		return err
	}

	encodedToken := base64.StdEncoding.EncodeToString([]byte(refreshToken))

	query = `UPDATE public.users SET refresh_token = $1 WHERE id = $2;`
	_, err = r.db.Db.Exec(query, encodedToken, user.ID)

	return err
}

func (r *CalculatorRepository) ValidateToken(token string) (int, error) {
	claims, err := jwt.ValidateToken(token, "secret")
	if err != nil {
		return 0, err
	}

	userID, ok := (*claims)["sub"].(float64)
	if !ok {
		return 0, errors.ErrInvalidToken
	}

	return int(userID), nil
}
