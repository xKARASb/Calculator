package application

import "github.com/xKARASb/Calculator/pkg/rpn"

type Expression struct {
	Data   string  `json:"data"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
}

type ExpressionWithID struct {
	ID IDExpression `json:"id"`
	Expression
}

type IDExpression = uint32

type GetExpressionHandlerResult struct {
	Expression ExpressionWithID `json:"expression"`
}

type AddHandlerResult struct {
	ID uint32 `json:"id"`
}

type GetExpressionsHandlerResult struct {
	Expressions []ExpressionWithID `json:"expressions"`
}

type GetTaskHandlerResult struct {
	Task rpn.TaskID `json:"task"`
}

type AgentResult struct {
	ID     rpn.IDTask `json:"id"`
	Result float64    `json:"result"`
}
