package main

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"
)

// Определение типов токенов
type TokenType int

const (
	TokenKeyword TokenType = iota
	TokenOperator
	TokenDelimiter
	TokenIdentifier
	TokenNumber
	TokenEOF
)

type Token struct {
	Type    TokenType
	Lexeme  string
	LineNum int
	ColNum  int
}

// Списки ключевых слов, операторов и разделителей
var keyWords = []string{
	"or", "and", "not", "program", "var", "begin", "end", "int", "float", "bool", "as", "if", "else", "then",
	"for", "to", "do", "while", "read", "write", "true", "false",
}

var operators = []string{
	"NE", "EQ", "LT", "LE", "GT", "GE",
	"plus", "min", "or", "mult", "div", "and", "~",
}

var delimiters = []string{
	";", ":", ",", "(", ")", ".", "=", "{", "}", "[", "]",
}

// Функции для проверки типов токенов
func isNumber(s string) bool {
	patterns := []string{
		`^\d+d?$`,                // Целое число
		`^\d*\.\d+(e[-+]?\d+)?$`, // Вещественное число с десятичной точкой
		`^\d+e[-+]?\d+$`,         // Число с экспонентой
		`^[01]+b$`,               // Двоичное число
		`^[0-7]+o$`,              // Восьмеричное число
		`^[0-9a-fA-F]+h$`,        // Шестнадцатеричное число
	}
	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, s)
		if matched {
			return true
		}
	}
	return false
}

func isKeyword(s string) bool {
	for _, kw := range keyWords {
		if s == kw {
			return true
		}
	}
	return false
}

func isOperator(s string) bool {
	for _, op := range operators {
		if s == op {
			return true
		}
	}
	return false
}

func isDelimiter(c rune) bool {
	for _, d := range delimiters {
		if string(c) == d {
			return true
		}
	}
	return false
}

// Функция лексического анализа
func Lexer(reader io.Reader) ([]Token, error) {
	var tokens []Token
	var lineNum, colNum int = 1, 0

	bufReader := bufio.NewReader(reader)
	var sb strings.Builder

	state := "H"
	for {
		ch, size, err := bufReader.ReadRune()
		if err != nil {
			if err == io.EOF {
				if state == "ID" {
					lexeme := sb.String()
					if isKeyword(lexeme) {
						tokens = append(tokens, Token{Type: TokenKeyword, Lexeme: lexeme, LineNum: lineNum, ColNum: colNum - len([]rune(lexeme)) + 1})
					} else if isOperator(lexeme) {
						tokens = append(tokens, Token{Type: TokenOperator, Lexeme: lexeme, LineNum: lineNum, ColNum: colNum - len([]rune(lexeme)) + 1})
					} else {
						tokens = append(tokens, Token{Type: TokenIdentifier, Lexeme: lexeme, LineNum: lineNum, ColNum: colNum - len([]rune(lexeme)) + 1})
					}
				} else if state == "NUM" {
					lexeme := sb.String()
					if isNumber(lexeme) {
						tokens = append(tokens, Token{Type: TokenNumber, Lexeme: lexeme, LineNum: lineNum, ColNum: colNum - len([]rune(lexeme)) + 1})
					} else {
						return nil, fmt.Errorf("Лексическая ошибка в строке %d, столбец %d: некорректное число '%s'", lineNum, colNum-len([]rune(lexeme))+1, lexeme)
					}
				} else if state == "OP" {
					lexeme := sb.String()
					if isOperator(lexeme) {
						tokens = append(tokens, Token{Type: TokenOperator, Lexeme: lexeme, LineNum: lineNum, ColNum: colNum})
					} else {
						return nil, fmt.Errorf("Неизвестный оператор '%s' в строке %d, столбец %d", lexeme, lineNum, colNum)
					}
				}
				break
			}
			return nil, err
		}

		colNum += size

		switch state {
		case "H":
			if unicode.IsSpace(ch) {
				if ch == '\n' {
					lineNum++
					colNum = 0
				}
				continue
			} else if unicode.IsLetter(ch) {
				sb.WriteRune(ch)
				state = "ID"
			} else if unicode.IsDigit(ch) {
				sb.WriteRune(ch)
				state = "NUM"
			} else if isDelimiter(ch) {
				// Обработка комментариев
				if ch == '{' {
					for {
						ch, size, err = bufReader.ReadRune()
						if err != nil {
							return nil, fmt.Errorf("Некорректный комментарий: ожидался '}'")
						}
						if ch == '}' {
							break
						}
						if ch == '\n' {
							lineNum++
							colNum = 0
						} else {
							colNum += size
						}
					}
				} else {
					tokens = append(tokens, Token{Type: TokenDelimiter, Lexeme: string(ch), LineNum: lineNum, ColNum: colNum})
				}
			} else if isOperator(string(ch)) {
				sb.WriteRune(ch)
				state = "OP"
			} else {
				return nil, fmt.Errorf("Неизвестный символ '%c' в строке %d, столбец %d", ch, lineNum, colNum)
			}
		case "ID":
			if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
				sb.WriteRune(ch)
			} else {
				lexeme := sb.String()
				if isKeyword(lexeme) {
					tokens = append(tokens, Token{Type: TokenKeyword, Lexeme: lexeme, LineNum: lineNum, ColNum: colNum - len([]rune(lexeme))})
				} else if isOperator(lexeme) {
					tokens = append(tokens, Token{Type: TokenOperator, Lexeme: lexeme, LineNum: lineNum, ColNum: colNum - len([]rune(lexeme))})
				} else {
					tokens = append(tokens, Token{Type: TokenIdentifier, Lexeme: lexeme, LineNum: lineNum, ColNum: colNum - len([]rune(lexeme))})
				}
				sb.Reset()
				state = "H"
				bufReader.UnreadRune()
				colNum -= size
			}
		case "NUM":
			if unicode.IsDigit(ch) || ch == '.' || ch == 'e' || ch == 'E' || ch == '+' || ch == '-' ||
				ch == 'b' || ch == 'o' || ch == 'h' || ch == 'd' ||
				(ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F') {
				sb.WriteRune(ch)
			} else {
				lexeme := sb.String()
				if isNumber(lexeme) {
					// Проверяем, что следующий символ не является буквой или цифрой
					if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
						return nil, fmt.Errorf("Лексическая ошибка в строке %d, столбец %d: некорректное число '%s'", lineNum, colNum-len([]rune(lexeme)), lexeme+string(ch))
					}
					tokens = append(tokens, Token{Type: TokenNumber, Lexeme: lexeme, LineNum: lineNum, ColNum: colNum - len([]rune(lexeme))})
				} else {
					return nil, fmt.Errorf("Лексическая ошибка в строке %d, столбец %d: некорректное число '%s'", lineNum, colNum-len([]rune(lexeme)), lexeme)
				}
				sb.Reset()
				state = "H"
				bufReader.UnreadRune()
				colNum -= size
			}

		case "OP":
			lexeme := sb.String()
			if isOperator(lexeme) {
				tokens = append(tokens, Token{Type: TokenOperator, Lexeme: lexeme, LineNum: lineNum, ColNum: colNum - size + 1})
				sb.Reset()
				state = "H"
				bufReader.UnreadRune()
				colNum -= size
			} else {
				chNext, sizeNext, err := bufReader.ReadRune()
				if err != nil && err != io.EOF {
					return nil, err
				}
				if err == nil {
					sb.WriteRune(chNext)
					lexeme = sb.String()
					if isOperator(lexeme) {
						tokens = append(tokens, Token{Type: TokenOperator, Lexeme: lexeme, LineNum: lineNum, ColNum: colNum - len([]rune(lexeme)) + 1})
						sb.Reset()
						state = "H"
					} else {
						// Если это не оператор, возвращаем последний прочитанный символ
						sb.Reset()
						sb.WriteRune(ch)
						bufReader.UnreadRune()
						colNum -= sizeNext
						tokens = append(tokens, Token{Type: TokenOperator, Lexeme: string(ch), LineNum: lineNum, ColNum: colNum - size + 1})
						state = "H"
					}
				} else {
					// EOF после оператора
					if isOperator(lexeme) {
						tokens = append(tokens, Token{Type: TokenOperator, Lexeme: lexeme, LineNum: lineNum, ColNum: colNum - size + 1})
						sb.Reset()
						state = "H"
					} else {
						return nil, fmt.Errorf("Неизвестный оператор '%s' в строке %d, столбец %d", lexeme, lineNum, colNum-len([]rune(lexeme))+1)
					}
				}
			}
		}
	}

	return tokens, nil
}

// Функция для преобразования типа токена в строку
func TokenTypeToString(t TokenType) string {
	switch t {
	case TokenKeyword:
		return "Keyword"
	case TokenOperator:
		return "Operator"
	case TokenDelimiter:
		return "Delimiter"
	case TokenIdentifier:
		return "Identifier"
	case TokenNumber:
		return "Number"
	case TokenEOF:
		return "EOF"
	default:
		return "Unknown"
	}
}
