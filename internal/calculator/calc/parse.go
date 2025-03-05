package calc

import (
	"errors"
	"strconv"
	"strings"

	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/calc/stackx"
)

var ErrInvalidExpr = errors.New("invalid expression")

type Token struct {
	IsNumber bool
	Number   float64
	Symbol   string
}

func (c *Calculator) Parse(s string) ([]Token, error) {
	tokens, err := c.tokenize(s)
	if err != nil {
		return nil, err
	}
	rpn, err := c.toRPN(tokens)
	if err != nil {
		return nil, err
	}
	return rpn, nil
}

func (c *Calculator) tokenize(s string) ([]Token, error) {
	tokens := make([]Token, 0, len(s))
	var numberBuf strings.Builder
	for _, ch := range strings.Split(s, "") {
		if ch >= "0" && ch <= "9" || ch == "." {
			numberBuf.WriteString(ch)
		} else if ch != " " {
			if numberBuf.Len() > 0 {
				num, err := strconv.ParseFloat(numberBuf.String(), 64)
				if err != nil {
					return nil, ErrInvalidExpr
				}
				tokens = append(tokens, Token{IsNumber: true, Number: num})
				numberBuf.Reset()
			}
			tokens = append(tokens, Token{IsNumber: false, Symbol: ch})
		}
	}
	if numberBuf.Len() > 0 {
		num, err := strconv.ParseFloat(numberBuf.String(), 64)
		if err != nil {
			return nil, ErrInvalidExpr
		}
		tokens = append(tokens, Token{IsNumber: true, Number: num})
	}
	return tokens, nil
}

func (c *Calculator) toRPN(tokens []Token) ([]Token, error) {
	rpn := make([]Token, 0, len(tokens))
	stack := stackx.New[Token]()
	for _, t := range tokens {
		switch {
		case t.IsNumber:
			rpn = append(rpn, t)
		case t.Symbol == "(":
			stack.Push(t)
		case t.Symbol == ")":
			for stack.Size() > 0 && stack.SafePeek().Symbol != "(" {
				rpn = append(rpn, stack.SafePop())
			}
			if stack.Size() > 0 {
				stack.SafePop()
			}
		default:
			for stack.Size() > 0 && c.precedence(stack.SafePeek().Symbol) >= c.precedence(t.Symbol) {
				rpn = append(rpn, stack.SafePop())
			}
			stack.Push(t)
		}
	}
	for stack.Size() > 0 {
		rpn = append(rpn, stack.SafePop())
	}

	if err := c.validateRPN(rpn); err != nil {
		return nil, err
	}
	return rpn, nil
}

func (c *Calculator) validateRPN(rpn []Token) error {
	stack := stackx.New[Token]()
	for _, token := range rpn {
		if !c.isOp(token.Symbol) {
			stack.Push(token)
			continue
		}

		if stack.Size() < 2 {
			return ErrInvalidExpr
		}

		_, _ = stack.SafePop(), stack.SafePop()
		stack.Push(Token{IsNumber: false, Symbol: "$"})
	}

	if stack.Size() != 1 {
		return ErrInvalidExpr
	}
	return nil
}

func (c *Calculator) precedence(op string) int {
	switch op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	default:
		return 0
	}
}

func (c *Calculator) isOp(s string) bool {
	switch s {
	case "+", "-", "*", "/":
		return true
	default:
		return false
	}
}
