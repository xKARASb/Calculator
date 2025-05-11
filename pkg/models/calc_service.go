package models

import "github.com/volatiletech/null/v9"

type Request struct {
	Expression string `json:"expression"`
}

type Response struct {
	ID int `json:"id"`
}

type ExpressionData struct {
	ID     int     `json:"id"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
}

type Expression struct {
	Expression ExpressionData `json:"expression"`
}

type TaskData struct {
	ID            int     `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

type Task struct {
	Task TaskData `json:"task"`
}

type Result struct {
	ID     int     `json:"id"`
	Result float64 `json:"result"`
}

type Auth struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type User struct {
	ID       int         `json:"id" db:"id"`
	Login    string      `json:"login" db:"login"`
	Password string      `json:"password" db:"password"`
	Refresh  null.String `json:"refresh_token" db:"refresh_token"`
}
