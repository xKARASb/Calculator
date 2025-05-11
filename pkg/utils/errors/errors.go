package errors

import "errors"

var (
	ErrInvalidExpression     = errors.New("Invalid expression")
	ErrInvalidOperation      = errors.New("Invalid operation")
	ErrDivisionByZero        = errors.New("Division by zero")
	ErrUnknownOperation      = errors.New("Unknown operation")
	ErrTaskChanIsFull        = errors.New("Task channel is full")
	ErrNotAvailable          = errors.New("No available")
	ErrMismatchedParentheses = errors.New("Mismatched parentheses")
	ErrNotFound              = errors.New("Not found")
	ErrUserAlreadyExists     = errors.New("User already exists")
	ErrInvalidCredentials    = errors.New("Invalid login or password")
	ErrInvalidToken          = errors.New("Invalid token")
)
