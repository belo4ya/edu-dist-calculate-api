package calc

import (
	"unicode"

	"github.com/belo4ya/edu-dist-calculate-api/internal/calculator/model"
)

func newToken(value string, isNumber bool) *model.Token {
	return &model.Token{
		Value:    value,
		IsNumber: isNumber,
	}
}

func tokenize(expression string) ([]model.Token, error) {
	var (
		tokens []model.Token
		number string
		err    error
	)

	for i, symbol := range expression {
		if unicode.IsSpace(symbol) {
			if number != "" {
				tokens = append(tokens, *newToken(number, true))
			}
			continue
		}

		if unicode.IsDigit(symbol) {
			number += string(symbol)
			if i+1 == len(expression) || !unicode.IsDigit(rune(expression[i+1])) {
				tokens = append(tokens, *newToken(number, true))
				number = ""
			}
			continue
		}

		switch string(symbol) {
		case "+", "-", "/", "*", "(", ")":
			tokens = append(tokens, *newToken(string(symbol), false))
		default:
			err = model.ErrorInvalidCharacter
			return nil, err
		}

	}

	if number != "" {
		tokens = append(tokens, *newToken(number, true))
		number = ""
	}

	if !checkEmptyBrackets(tokens) {
		return nil, model.ErrorEmptyBrackets
	}

	if !checkMissingOperand(tokens) {
		return nil, model.ErrorMissingOperand
	}

	if !checkMissingNumber(tokens) {
		return nil, model.ErrorInvalidInput
	}

	result := addMissingOperand(tokens)
	return result, nil
}

func checkEmptyBrackets(tokens []model.Token) bool {
	for i, token := range tokens {
		if i == len(tokens)-1 {
			break
		}
		if token.Value == "(" && tokens[i+1].Value == ")" {
			return false
		}
	}
	return true
}

func checkMissingOperand(tokens []model.Token) bool {
	for i, token := range tokens {
		if i == len(tokens)-1 {
			break
		}
		if token.IsNumber && tokens[i+1].IsNumber {
			return false
		}
	}
	return true
}

func checkMissingNumber(tokens []model.Token) bool {
	for i, token := range tokens {
		if i == len(tokens)-1 {
			break
		}
		if !token.IsNumber && !tokens[i+1].IsNumber && token.Value != ")" && tokens[i+1].Value != "(" {
			return false
		}
	}
	return true
}

func addMissingOperand(expression []model.Token) []model.Token {
	var result []model.Token

	for i, token := range expression {
		result = append(result, token)

		if i+1 < len(expression) {
			if (token.IsNumber || token.Value == ")") && expression[i+1].Value == "(" {
				result = append(result, model.Token{Value: "*", IsNumber: false})
			}
			if token.Value == ")" && expression[i+1].IsNumber {
				result = append(result, model.Token{Value: "*", IsNumber: false})
			}
		}
	}

	return result
}
