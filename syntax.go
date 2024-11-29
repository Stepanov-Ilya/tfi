package main

import (
	"fmt"
)

// Структура парсера
type Syntax struct {
	tokens []Token
	pos    int
	vars   map[string]bool
}

func (p *Syntax) currentToken() Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return Token{Type: TokenEOF, Lexeme: "", LineNum: 0, ColNum: 0}
}

func (p *Syntax) nextToken() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}

func (p *Syntax) matchToken(expectedType TokenType, expectedLexeme string) error {
	token := p.currentToken()
	if token.Type == expectedType && token.Lexeme == expectedLexeme {
		p.nextToken()
		return nil
	}
	return fmt.Errorf("Ожидалось %s '%s', получено %s '%s' на строке %d столбце %d",
		TokenTypeToString(expectedType), expectedLexeme,
		TokenTypeToString(token.Type), token.Lexeme,
		token.LineNum, token.ColNum)
}

// Функция синтаксического анализа
func (p *Syntax) ParseProgram() error {
	// program
	err := p.matchToken(TokenKeyword, "program")
	if err != nil {
		return err
	}

	// var
	err = p.matchToken(TokenKeyword, "var")
	if err != nil {
		return err
	}

	// parse declarations
	for {
		err = p.parseDeclaration()
		if err != nil {
			return err
		}
		// Проверяем, есть ли еще объявления
		token := p.currentToken()
		if token.Type == TokenKeyword && token.Lexeme == "begin" {
			break
		}
	}

	// begin
	err = p.matchToken(TokenKeyword, "begin")
	if err != nil {
		return err
	}

	// parse operations
	err = p.parseOperations()
	if err != nil {
		return err
	}

	// end
	err = p.matchToken(TokenKeyword, "end")
	if err != nil {
		return err
	}

	// '.'
	err = p.matchToken(TokenDelimiter, ".")
	if err != nil {
		return err
	}

	return nil
}

// Парсинг объявления переменных
func (p *Syntax) parseDeclaration() error {
	for {
		token := p.currentToken()
		if token.Type != TokenIdentifier {
			return fmt.Errorf("Ожидался идентификатор, получено %s '%s' на строке %d столбце %d",
				TokenTypeToString(token.Type), token.Lexeme,
				token.LineNum, token.ColNum)
		}
		if p.vars == nil {
			p.vars = make(map[string]bool)
		}
		if p.vars[token.Lexeme] {
			return fmt.Errorf("Переменная '%s' уже объявлена на строке %d столбце %d",
				token.Lexeme, token.LineNum, token.ColNum)
		}
		p.vars[token.Lexeme] = true
		p.nextToken()

		token = p.currentToken()
		if token.Type == TokenDelimiter && token.Lexeme == "," {
			p.nextToken()
			continue
		} else if token.Type == TokenDelimiter && token.Lexeme == ":" {
			p.nextToken()
			break
		} else {
			return fmt.Errorf("Ожидалось ',' или ':', получено %s '%s' на строке %d столбце %d",
				TokenTypeToString(token.Type), token.Lexeme,
				token.LineNum, token.ColNum)
		}
	}

	// Тип
	token := p.currentToken()
	if token.Type != TokenKeyword || (token.Lexeme != "int" && token.Lexeme != "float" && token.Lexeme != "bool") {
		return fmt.Errorf("Ожидался тип 'int', 'float' или 'bool', получено %s '%s' на строке %d столбце %d",
			TokenTypeToString(token.Type), token.Lexeme,
			token.LineNum, token.ColNum)
	}
	p.nextToken()

	// ';'
	token = p.currentToken()
	if token.Type != TokenDelimiter || token.Lexeme != ";" {
		return fmt.Errorf("Ожидалось ';', получено %s '%s' на строке %d столбце %d",
			TokenTypeToString(token.Type), token.Lexeme,
			token.LineNum, token.ColNum)
	}
	p.nextToken()

	return nil
}

// Парсинг списка операций
func (p *Syntax) parseOperations() error {
	for {
		token := p.currentToken()
		if token.Type == TokenKeyword && token.Lexeme == "end" {
			// Если мы встретили 'end', выходим из цикла операций
			break
		}

		err := p.parseOperation()
		if err != nil {
			return err
		}

		token = p.currentToken()
		if token.Type == TokenDelimiter && token.Lexeme == ";" {
			p.nextToken()
			continue
		} else if token.Type == TokenKeyword && token.Lexeme == "end" {
			// Если после операции нет ';', но есть 'end', завершаем парсинг операций
			break
		} else {
			return fmt.Errorf("Ожидалось ';' или 'end', получено %s '%s' на строке %d столбце %d",
				TokenTypeToString(token.Type), token.Lexeme,
				token.LineNum, token.ColNum)
		}
	}
	return nil
}

// Парсинг одной операции
func (p *Syntax) parseOperation() error {
	token := p.currentToken()
	if token.Type == TokenKeyword {
		switch token.Lexeme {
		case "if":
			return p.parseIf()
		case "for":
			return p.parseFor()
		case "while":
			return p.parseWhile()
		case "read":
			return p.Syntaxead()
		case "write":
			return p.parseWrite()
		case "begin":
			return p.parseCompositeOperation()
		default:
			return fmt.Errorf("Неизвестный оператор '%s' на строке %d столбце %d",
				token.Lexeme, token.LineNum, token.ColNum)
		}
	} else if token.Type == TokenIdentifier {
		// Присваивание
		return p.parseAssignment()
	} else if token.Type == TokenDelimiter && token.Lexeme == "[" {
		// Составной оператор
		return p.parseCompositeOperation()
	} else {
		return fmt.Errorf("Ожидался оператор, получено %s '%s' на строке %d столбце %d",
			TokenTypeToString(token.Type), token.Lexeme,
			token.LineNum, token.ColNum)
	}
}

// Парсинг составного оператора
func (p *Syntax) parseCompositeOperation() error {
	// '[' <оператор> { (: | ';') <оператор> } ']'
	err := p.matchToken(TokenDelimiter, "[")
	if err != nil {
		return err
	}

	for {
		err = p.parseOperation()
		if err != nil {
			return err
		}
		token := p.currentToken()
		if token.Type == TokenDelimiter && (token.Lexeme == ":" || token.Lexeme == ";") {
			p.nextToken()
			continue
		} else if token.Type == TokenDelimiter && token.Lexeme == "]" {
			p.nextToken()
			break
		} else {
			return fmt.Errorf("Ожидалось ':' или ']' в составном операторе, получено %s '%s' на строке %d столбце %d",
				TokenTypeToString(token.Type), token.Lexeme,
				token.LineNum, token.ColNum)
		}
	}

	return nil
}

// Парсинг операции присваивания
func (p *Syntax) parseAssignment() error {
	// <идентификатор> as <выражение>
	token := p.currentToken()
	if token.Type != TokenIdentifier {
		return fmt.Errorf("Ожидался идентификатор в присваивании, получено %s '%s' на строке %d столбце %d",
			TokenTypeToString(token.Type), token.Lexeme,
			token.LineNum, token.ColNum)
	}
	if !p.vars[token.Lexeme] {
		return fmt.Errorf("Необъявленная переменная '%s' на строке %d столбце %d",
			token.Lexeme, token.LineNum, token.ColNum)
	}
	p.nextToken()

	err := p.matchToken(TokenKeyword, "as")
	if err != nil {
		return err
	}

	err = p.parseExpression()
	if err != nil {
		return err
	}

	return nil
}

// Парсинг конструкции if
func (p *Syntax) parseIf() error {
	// if <выражение> then <оператор> [ else <оператор> ]
	err := p.matchToken(TokenKeyword, "if")
	if err != nil {
		return err
	}

	err = p.parseExpression()
	if err != nil {
		return err
	}

	err = p.matchToken(TokenKeyword, "then")
	if err != nil {
		return err
	}

	err = p.parseOperation()
	if err != nil {
		return err
	}

	token := p.currentToken()
	if token.Type == TokenKeyword && token.Lexeme == "else" {
		p.nextToken()
		err = p.parseOperation()
		if err != nil {
			return err
		}
	}

	return nil
}

// Парсинг цикла for
func (p *Syntax) parseFor() error {
	// for <присваивание> to <выражение> do <оператор>
	err := p.matchToken(TokenKeyword, "for")
	if err != nil {
		return err
	}

	err = p.parseAssignment()
	if err != nil {
		return err
	}

	err = p.matchToken(TokenKeyword, "to")
	if err != nil {
		return err
	}

	err = p.parseExpression()
	if err != nil {
		return err
	}

	err = p.matchToken(TokenKeyword, "do")
	if err != nil {
		return err
	}

	err = p.parseOperation()
	if err != nil {
		return err
	}

	return nil
}

// Парсинг цикла while
func (p *Syntax) parseWhile() error {
	// while <выражение> do <оператор>
	err := p.matchToken(TokenKeyword, "while")
	if err != nil {
		return err
	}

	err = p.parseExpression()
	if err != nil {
		return err
	}

	err = p.matchToken(TokenKeyword, "do")
	if err != nil {
		return err
	}

	err = p.parseOperation()
	if err != nil {
		return err
	}

	return nil
}

// Парсинг оператора read
func (p *Syntax) Syntaxead() error {
	// read ( <идентификатор> { , <идентификатор> } )
	err := p.matchToken(TokenKeyword, "read")
	if err != nil {
		return err
	}

	err = p.matchToken(TokenDelimiter, "(")
	if err != nil {
		return err
	}

	for {
		token := p.currentToken()
		if token.Type != TokenIdentifier {
			return fmt.Errorf("Ожидался идентификатор в read, получено %s '%s' на строке %d столбце %d",
				TokenTypeToString(token.Type), token.Lexeme,
				token.LineNum, token.ColNum)
		}
		if !p.vars[token.Lexeme] {
			return fmt.Errorf("Необъявленная переменная '%s' на строке %d столбце %d",
				token.Lexeme, token.LineNum, token.ColNum)
		}
		p.nextToken()

		token = p.currentToken()
		if token.Type == TokenDelimiter && token.Lexeme == "," {
			p.nextToken()
			continue
		} else if token.Type == TokenDelimiter && token.Lexeme == ")" {
			p.nextToken()
			break
		} else {
			return fmt.Errorf("Ожидалось ',' или ')', получено %s '%s' на строке %d столбце %d",
				TokenTypeToString(token.Type), token.Lexeme,
				token.LineNum, token.ColNum)
		}
	}

	return nil
}

// Парсинг оператора write
func (p *Syntax) parseWrite() error {
	// write ( <выражение> { , <выражение> } )
	err := p.matchToken(TokenKeyword, "write")
	if err != nil {
		return err
	}

	err = p.matchToken(TokenDelimiter, "(")
	if err != nil {
		return err
	}

	for {
		err = p.parseExpression()
		if err != nil {
			return err
		}

		token := p.currentToken()
		if token.Type == TokenDelimiter && token.Lexeme == "," {
			p.nextToken()
			continue
		} else if token.Type == TokenDelimiter && token.Lexeme == ")" {
			p.nextToken()
			break
		} else {
			return fmt.Errorf("Ожидалось ',' или ')', получено %s '%s' на строке %d столбце %d",
				TokenTypeToString(token.Type), token.Lexeme,
				token.LineNum, token.ColNum)
		}
	}

	return nil
}

// Парсинг выражения
func (p *Syntax) parseExpression() error {
	// Реализуем разбор выражений с учетом приоритетов операций

	err := p.parseOperand()
	if err != nil {
		return err
	}

	token := p.currentToken()
	for token.Type == TokenOperator && isRelationOperator(token.Lexeme) {
		p.nextToken()
		err = p.parseOperand()
		if err != nil {
			return err
		}
		token = p.currentToken()
	}

	return nil
}

// Проверка, является ли оператор оператором отношения
func isRelationOperator(op string) bool {
	switch op {
	case "EQ", "NE", "LT", "LE", "GT", "GE":
		return true
	}
	return false
}

// Парсинг операнда
func (p *Syntax) parseOperand() error {
	err := p.parseTerm()
	if err != nil {
		return err
	}

	token := p.currentToken()
	for token.Type == TokenOperator && isAdditionOperator(token.Lexeme) {
		p.nextToken()
		err = p.parseTerm()
		if err != nil {
			return err
		}
		token = p.currentToken()
	}

	return nil
}

// Проверка, является ли оператор оператором сложения
func isAdditionOperator(op string) bool {
	switch op {
	case "plus", "min", "or":
		return true
	}
	return false
}

// Парсинг терма
func (p *Syntax) parseTerm() error {
	err := p.parseFactor()
	if err != nil {
		return err
	}

	token := p.currentToken()
	for token.Type == TokenOperator && isMultiplicationOperator(token.Lexeme) {
		p.nextToken()
		err = p.parseFactor()
		if err != nil {
			return err
		}
		token = p.currentToken()
	}

	return nil
}

// Проверка, является ли оператор оператором умножения
func isMultiplicationOperator(op string) bool {
	switch op {
	case "mult", "div", "and":
		return true
	}
	return false
}

// Парсинг фактора
func (p *Syntax) parseFactor() error {
	token := p.currentToken()

	if token.Type == TokenOperator && token.Lexeme == "~" {
		p.nextToken()
		return p.parseFactor()
	} else if token.Type == TokenDelimiter && token.Lexeme == "(" {
		p.nextToken()
		err := p.parseExpression()
		if err != nil {
			return err
		}
		err = p.matchToken(TokenDelimiter, ")")
		if err != nil {
			return err
		}
		return nil
	} else if token.Type == TokenIdentifier {
		if !p.vars[token.Lexeme] {
			return fmt.Errorf("Необъявленная переменная '%s' на строке %d столбце %d",
				token.Lexeme, token.LineNum, token.ColNum)
		}
		p.nextToken()
		return nil
	} else if token.Type == TokenNumber {
		p.nextToken()
		return nil
	} else if token.Type == TokenKeyword && (token.Lexeme == "true" || token.Lexeme == "false") {
		p.nextToken()
		return nil
	} else {
		return fmt.Errorf("Ожидался фактор, получено %s '%s' на строке %d столбце %d",
			TokenTypeToString(token.Type), token.Lexeme,
			token.LineNum, token.ColNum)
	}
}
