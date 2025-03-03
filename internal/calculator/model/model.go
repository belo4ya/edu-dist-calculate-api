package model

import (
	"time"

	calculatorv1 "github.com/belo4ya/edu-dist-calculate-api/pkg/calculator/v1"
)

// ExpressionNode представляет узел в дереве выражения
type ExpressionNode struct {
	ID         string                        // Уникальный идентификатор узла
	ParentID   string                        // Идентификатор родительского узла (если есть)
	Type       string                        // Тип узла: "operation" или "value"
	Operation  calculatorv1.Operation        // Операция (если тип - "operation")
	Value      float64                       // Значение (если тип - "value")
	LeftChild  string                        // ID левого потомка (для операций)
	RightChild string                        // ID правого потомка (для операций)
	TaskID     string                        // ID связанной задачи (для операций)
	Status     calculatorv1.ExpressionStatus // Статус этого узла
}

// Expression представляет арифметическое выражение
type Expression struct {
	ID          string                        // Уникальный идентификатор выражения
	Text        string                        // Исходный текст выражения
	RootNodeID  string                        // ID корневого узла дерева выражения
	Status      calculatorv1.ExpressionStatus // Общий статус выражения
	Result      *float64                      // Результат вычисления (если есть)
	Nodes       map[string]*ExpressionNode    // Все узлы дерева выражения
	CreatedAt   time.Time                     // Время создания выражения
	CompletedAt *time.Time                    // Время завершения вычисления (если есть)
}

// Task представляет элементарную задачу для вычисления
type Task struct {
	ID            string                 // Уникальный идентификатор задачи
	ExpressionID  string                 // ID выражения, к которому относится задача
	NodeID        string                 // ID узла выражения, к которому относится задача
	Arg1          float64                // Первый аргумент
	Arg2          float64                // Второй аргумент
	Operation     calculatorv1.Operation // Операция
	OperationTime time.Duration          // Время выполнения операции
	Result        *float64               // Результат выполнения (если есть)
	Status        string                 // Статус задачи: "pending", "in_progress", "completed", "failed"
	DependsOn     []string               // Список ID задач, от которых зависит эта задача
	CreatedAt     time.Time              // Время создания задачи
	AssignedAt    *time.Time             // Время назначения задачи агенту
	CompletedAt   *time.Time             // Время завершения выполнения задачи
	AgentID       string                 // ID агента, который выполняет задачу (если назначена)
}
