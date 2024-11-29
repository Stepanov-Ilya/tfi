package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Использование: go run main.go <имя файла>")
		return
	}

	fileName := os.Args[1]
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Ошибка при открытии файла: %v\n", err)
		return
	}
	defer file.Close()

	// Лексический анализ
	tokens, err := Lexer(file)
	if err != nil {
		fmt.Printf("Ошибка лексического анализа: %v\n", err)
		return
	}

	for _, token := range tokens {
		fmt.Printf("Token: %-15s Lexeme: %-10s Line: %d Col: %d\n", TokenTypeToString(token.Type), token.Lexeme, token.LineNum, token.ColNum)
	}

	// Синтаксический анализ
	parser := Syntax{tokens: tokens, pos: 0}
	err = parser.ParseProgram()
	if err != nil {
		fmt.Printf("Ошибка синтаксического анализа: %v\n", err)
	} else {
		fmt.Println("Синтаксический анализ успешно завершен.")
	}
}
