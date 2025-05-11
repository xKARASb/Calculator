package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/xKARASb/Calculator/pkg/models"
	"github.com/xKARASb/Calculator/pkg/utils/errors"
	"github.com/xKARASb/Calculator/pkg/utils/statuses"
)

type Agent struct {
	ID    int
	mu    sync.Mutex
	Host  string
	token string
}

func NewAgent(id int, host string) *Agent {
	agent := &Agent{ID: id, Host: host}
	token, err := agent.login("agent", "agent_password")
	if err != nil {
		fmt.Printf("Agent authentication error: %v\n", err)
	} else {
		agent.token = token
	}
	return agent
}

func (a *Agent) login(login, password string) (string, error) {
	auth := models.Auth{
		Login:    login,
		Password: password,
	}

	authBody, err := json.Marshal(auth)
	if err != nil {
		return "", err
	}

	resp, err := http.Post("http://"+a.Host+":8080/api/v1/login", "application/json", bytes.NewBuffer(authBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login error: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response struct {
		Status string `json:"status"`
		Token  string `json:"token"`
	}

	if err = json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	return response.Token, nil
}

func (a *Agent) CalculateExpression() error {
	localRand := rand.New(rand.NewSource(time.Now().UnixNano() + int64(a.ID)))

	time.Sleep(time.Duration(localRand.Intn(1000)) * time.Millisecond)

	var (
		task *models.Task
		err  error
		id   int
	)

	for {
		a.mu.Lock()
		task, err = a.getTask()
		if err != nil || task == nil {
			a.mu.Unlock()
			time.Sleep(1 * time.Second)
			continue
		}

		id = task.Task.ID

		expression, err := a.getExpression(id)
		if err != nil {
			a.mu.Unlock()
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if id == 0 || task.Task.OperationTime == 0 || expression.Expression.Status == statuses.StatusComplete {
			a.mu.Unlock()
			time.Sleep(500 * time.Millisecond)
			continue
		}

		a.mu.Unlock()

		var result float64
		switch task.Task.Operation {
		case "+":
			result = task.Task.Arg1 + task.Task.Arg2
		case "-":
			result = task.Task.Arg1 - task.Task.Arg2
		case "*":
			result = task.Task.Arg1 * task.Task.Arg2
		case "/":
			if task.Task.Arg2 == 0 {
				return errors.ErrDivisionByZero
			}
			result = task.Task.Arg1 / task.Task.Arg2
		}
		{
			a.mu.Lock()
			err := a.setExpression(id, result)
			if err != nil {
				a.mu.Unlock()
				time.Sleep(1 * time.Second)
				continue
			}
			a.mu.Unlock()
		}

		time.Sleep(time.Duration(localRand.Intn(1000)+500) * time.Millisecond)

		if a.token == "" {
			token, err := a.login("agent", "agent_password")
			if err == nil {
				a.token = token
			}
		}
	}
}

func (a *Agent) setExpression(id int, result float64) error {
	expr := models.Expression{Expression: models.ExpressionData{ID: id, Status: statuses.StatusComplete, Result: result}}
	exprBody, err := json.Marshal(expr)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "http://"+a.Host+":8080/data/setExpression", bytes.NewBuffer(exprBody))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+a.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (a *Agent) getExpression(id int) (*models.Expression, error) {
	req, err := http.NewRequest("GET", "http://"+a.Host+":8080/api/v1/expressions/"+fmt.Sprint(id), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+a.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	exprBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	var expression models.Expression
	if err = json.Unmarshal(exprBody, &expression); err != nil {
		return nil, err
	}
	return &expression, nil
}

func (a *Agent) getTask() (task *models.Task, err error) {
	req, err := http.NewRequest("GET", "http://"+a.Host+":8080/internal/task", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+a.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(body, &task); err != nil {
		return nil, err
	}
	return task, nil
}
