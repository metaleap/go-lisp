package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// read
func READ(str string) string {
	return str
}

// eval
func EVAL(ast string, env string) string {
	return ast
}

// print
func PRINT(exp string) string {
	return exp
}

// repl
func rep(str string) string {
	return PRINT(EVAL(READ(str), ""))
}

func main() {
	readln := bufio.NewScanner(os.Stdin)
	// repl loop
	fmt.Print("\nrepl> ")
	for readln.Scan() {
		text := strings.TrimSpace(readln.Text())
		fmt.Println(rep(text))
		fmt.Print("\nrepl> ")
	}
	if err := readln.Err(); err != nil {
		panic(err)
	}
}
