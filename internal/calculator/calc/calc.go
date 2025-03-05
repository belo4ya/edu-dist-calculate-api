package calc

import (
	"fmt"
	"strconv"

	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/model"
)

type Repository interface {
	InsertTask(*model.Task) (int, error)
}

// Calc вызывает токенизацию выражения, записывает его в RPN. а затем в параллельных горутинах подсчитывает значения выражений в скобках
func Calc(stringExpression string, id int, repo Repository) error {
	expression, err := tokenize(stringExpression)
	if err != nil {
		return err
	}

	if len(expression) == 0 {
		return model.ErrorEmptyExpression
	}

	reversePolishNotation, err := toReversePolishNotation(expression)
	if err != nil {
		return err
	}

	err = parseRPN(reversePolishNotation, id, repo)
	if err != nil {
		return err
	}

	return nil
}

// NewTask создает экземпляр структуры Task
func NewTask(id int, arg1, arg2 float64, operation string) *model.Task {
	newTask := model.Task{
		ExpressionID: id,
		Arg1:         arg1,
		Arg2:         arg2,
		Operation:    operation,
	}
	return &newTask
}

func toReversePolishNotation(expression []model.Token) ([]model.Token, error) {
	priority := map[string]int{
		"(": 0,
		")": 1,
		"+": 2,
		"-": 2,
		"*": 3,
		"/": 3,
	}
	var stack []model.Token
	var reversePolishNotation []model.Token

	for _, token := range expression {
		if _, ok := priority[token.Value]; ok {
			if token.Value == ")" {
				for i := len(stack) - 1; i >= 0 && stack[i].Value != "("; i-- {
					reversePolishNotation = append(reversePolishNotation, lastToken(stack))
					stack = stack[:len(stack)-1]
				}

				if len(stack) > 0 && lastToken(stack).Value == "(" {
					stack = stack[:len(stack)-1]
				} else {
					return nil, model.ErrorUnclosedBracket
				}
				continue
			}

			for len(stack) > 0 && priority[lastToken(stack).Value] >= priority[token.Value] && token.Value != "(" {
				reversePolishNotation = append(reversePolishNotation, lastToken(stack))
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token)

		} else if token.IsNumber {
			reversePolishNotation = append(reversePolishNotation, token)
		} else {
			return nil, model.ErrorInvalidInput
		}
	}

	for len(stack) > 0 {

		reversePolishNotation = append(reversePolishNotation, lastToken(stack))
		stack = stack[:len(stack)-1]
	}
	return reversePolishNotation, nil
}

func parseRPN(expression []model.Token, exprID int, repo Repository) error {
	type StackElement struct {
		Value  float64
		TaskID int
		IsTask bool
	}

	var stack []StackElement

	for _, token := range expression {
		if token.IsNumber {
			value, err := strconv.ParseFloat(token.Value, 64)
			if err != nil {
				return fmt.Errorf("failed to parse number: %v", err)
			}
			stack = append(stack, StackElement{Value: value})
		} else {
			if len(stack) < 2 {
				return fmt.Errorf("not enough operands for operation %s", token.Value)
			}

			right := stack[len(stack)-1]
			left := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			task := NewTask(exprID, left.Value, right.Value, token.Value)
			task.Status = model.StatusWait

			if left.IsTask {
				task.PrevTaskID1 = left.TaskID
			} else {
				task.Arg1 = left.Value
			}

			if right.IsTask {
				task.PrevTaskID2 = right.TaskID
			} else {
				task.Arg2 = right.Value
			}

			taskID, err := repo.InsertTask(task)
			if err != nil {
				return fmt.Errorf("failed to insert task: %v", err)
			}

			stack = append(stack, StackElement{TaskID: taskID, IsTask: true})
		}
	}

	if len(stack) != 1 {
		return fmt.Errorf("invalid expression")
	}

	return nil
}

func lastToken(tokens []model.Token) model.Token {
	return tokens[len(tokens)-1]
}
