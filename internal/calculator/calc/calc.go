package calc

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/model"
	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/repository"
	calculatorv1 "github.com/belo4ya/edu-dist-calculate-api/pkg/calculator/v1"

	"github.com/rs/xid"
)

// ExpressionParser парсит арифметические выражения и преобразует их в дерево задач
type ExpressionParser struct {
	repo               *repository.Repository
	operationDurations map[calculatorv1.Operation]time.Duration
}

func NewExpressionParser(repo *repository.Repository, opDurations map[calculatorv1.Operation]time.Duration) *ExpressionParser {
	return &ExpressionParser{
		repo:               repo,
		operationDurations: opDurations,
	}
}

// ParseExpression разбирает арифметическое выражение и создает соответствующие задачи
func (p *ExpressionParser) ParseExpression(expressionText string) (*model.Expression, error) {
	// Удаляем пробелы из выражения
	expressionText = strings.ReplaceAll(expressionText, " ", "")
	if expressionText == "" {
		return nil, errors.New("пустое выражение")
	}

	// Создаем новое выражение
	expr := &model.Expression{
		ID:        xid.New().String(),
		Text:      expressionText,
		Status:    calculatorv1.ExpressionStatus_EXPRESSION_STATUS_PENDING,
		Nodes:     make(map[string]*model.ExpressionNode),
		CreatedAt: time.Now(),
	}

	// Разбираем выражение в дерево
	tokens, err := tokenizeExpression(expressionText)
	if err != nil {
		return nil, fmt.Errorf("ошибка токенизации: %w", err)
	}

	rootID, err := p.buildExpressionTree(tokens, expr)
	if err != nil {
		return nil, fmt.Errorf("ошибка построения дерева: %w", err)
	}
	expr.RootNodeID = rootID

	// Создаем задачи для всех операций
	if err := p.createTasksForExpression(expr); err != nil {
		return nil, fmt.Errorf("ошибка создания задач: %w", err)
	}

	return expr, nil
}

// Токенизация выражения
func tokenizeExpression(expr string) ([]string, error) {
	var tokens []string
	var currentNum strings.Builder

	flushNumber := func() {
		if currentNum.Len() > 0 {
			tokens = append(tokens, currentNum.String())
			currentNum.Reset()
		}
	}

	for i := 0; i < len(expr); i++ {
		ch := expr[i]
		switch ch {
		case '+', '-', '*', '/', '(', ')':
			flushNumber()
			tokens = append(tokens, string(ch))
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
			currentNum.WriteByte(ch)
		default:
			return nil, fmt.Errorf("недопустимый символ в выражении: %c", ch)
		}
	}

	flushNumber()
	return tokens, nil
}

// Построение дерева выражения по токенам
func (p *ExpressionParser) buildExpressionTree(tokens []string, expr *model.Expression) (string, error) {
	// Используем алгоритм перевода выражения в постфиксную нотацию (алгоритм сортировочной станции)
	output := make([]string, 0)    // Выходная очередь
	operators := make([]string, 0) // Стек операторов

	precedence := map[string]int{
		"+": 1,
		"-": 1,
		"*": 2,
		"/": 2,
	}

	for _, token := range tokens {
		// Если токен - число
		if _, err := strconv.ParseFloat(token, 64); err == nil {
			output = append(output, token)
			continue
		}

		// Если токен - открывающая скобка
		if token == "(" {
			operators = append(operators, token)
			continue
		}

		// Если токен - закрывающая скобка
		if token == ")" {
			for len(operators) > 0 && operators[len(operators)-1] != "(" {
				output = append(output, operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}
			if len(operators) > 0 && operators[len(operators)-1] == "(" {
				operators = operators[:len(operators)-1] // Удаляем открывающую скобку
			} else {
				return "", errors.New("несбалансированные скобки")
			}
			continue
		}

		// Если токен - оператор
		for len(operators) > 0 && operators[len(operators)-1] != "(" &&
			precedence[operators[len(operators)-1]] >= precedence[token] {
			output = append(output, operators[len(operators)-1])
			operators = operators[:len(operators)-1]
		}
		operators = append(operators, token)
	}

	// Переносим оставшиеся операторы в выходную очередь
	for len(operators) > 0 {
		if operators[len(operators)-1] == "(" {
			return "", errors.New("несбалансированные скобки")
		}
		output = append(output, operators[len(operators)-1])
		operators = operators[:len(operators)-1]
	}

	// Строим дерево из постфиксной нотации
	var stack []string
	nodeMap := make(map[string]bool)

	for _, token := range output {
		if _, err := strconv.ParseFloat(token, 64); err == nil {
			// Это число - создаем узел-значение
			nodeID := xid.New().String()
			value, _ := strconv.ParseFloat(token, 64)

			expr.Nodes[nodeID] = &model.ExpressionNode{
				ID:     nodeID,
				Type:   "value",
				Value:  value,
				Status: calculatorv1.ExpressionStatus_EXPRESSION_STATUS_COMPLETED,
			}
			stack = append(stack, nodeID)
			nodeMap[nodeID] = true
		} else {
			// Это оператор - создаем узел-операцию
			if len(stack) < 2 {
				return "", errors.New("неправильное выражение")
			}

			nodeID := xid.New().String()
			rightID := stack[len(stack)-1]
			leftID := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			var op calculatorv1.Operation
			switch token {
			case "+":
				op = calculatorv1.Operation_OPERATION_ADDITION
			case "-":
				op = calculatorv1.Operation_OPERATION_SUBTRACTION
			case "*":
				op = calculatorv1.Operation_OPERATION_MULTIPLICATION
			case "/":
				op = calculatorv1.Operation_OPERATION_DIVISION
			default:
				return "", fmt.Errorf("неизвестный оператор: %s", token)
			}

			expr.Nodes[nodeID] = &model.ExpressionNode{
				ID:         nodeID,
				Type:       "operation",
				Operation:  op,
				LeftChild:  leftID,
				RightChild: rightID,
				Status:     calculatorv1.ExpressionStatus_EXPRESSION_STATUS_PENDING,
			}

			// Устанавливаем родительские связи
			expr.Nodes[leftID].ParentID = nodeID
			expr.Nodes[rightID].ParentID = nodeID

			stack = append(stack, nodeID)
			nodeMap[nodeID] = true
		}
	}

	if len(stack) != 1 {
		return "", errors.New("неправильное выражение")
	}

	return stack[0], nil
}

// Создание задач для всех операций в дереве выражения
func (p *ExpressionParser) createTasksForExpression(expr *model.Expression) error {
	tasks := make([]*model.Task, 0)

	// Рекурсивно обходим дерево и создаем задачи для операций
	var createTasks func(string) error
	createTasks = func(nodeID string) error {
		node := expr.Nodes[nodeID]

		// Если это узел со значением, ничего делать не нужно
		if node.Type == "value" {
			return nil
		}

		// Рекурсивно обрабатываем потомков
		if err := createTasks(node.LeftChild); err != nil {
			return err
		}
		if err := createTasks(node.RightChild); err != nil {
			return err
		}

		// Создаем задачу для этого узла
		taskID := xid.New().String()
		node.TaskID = taskID

		// Задача зависит от результатов вычисления потомков
		dependsOn := make([]string, 0)
		leftNode := expr.Nodes[node.LeftChild]
		rightNode := expr.Nodes[node.RightChild]

		if leftNode.Type == "operation" {
			dependsOn = append(dependsOn, leftNode.TaskID)
		}
		if rightNode.Type == "operation" {
			dependsOn = append(dependsOn, rightNode.TaskID)
		}

		task := &model.Task{
			ID:            taskID,
			ExpressionID:  expr.ID,
			NodeID:        nodeID,
			Operation:     node.Operation,
			OperationTime: p.operationDurations[node.Operation],
			Arg1:          leftNode.Value,
			Arg2:          rightNode.Value,
			Status:        "pending",
			DependsOn:     dependsOn,
			CreatedAt:     time.Now(),
		}

		tasks = append(tasks, task)
		return nil
	}

	if err := createTasks(expr.RootNodeID); err != nil {
		return err
	}

	for _, node := range expr.Nodes {
		fmt.Printf("node: %+v\n", node)
	}

	for _, task := range tasks {
		fmt.Printf("task: %+v\n", task)
	}

	// Сохраняем задачи в хранилище
	// Примечание: здесь должна быть логика сохранения в репозиторий
	return nil
}
