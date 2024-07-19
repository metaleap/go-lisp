package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func replRead(str string) (Expr, error) {
	return Read_str(str)
}

func replEval(expr Expr, env *Env) (Expr, error) {
	return eval(expr, env)
}

func replPrint(expr Expr) string {
	return Pr_str(expr, true)
}

func repl(str string) (string, error) {
	expr, err := replRead(str)
	if err != nil {
		return "", err
	}
	expr, err = replEval(expr, &repl_env)
	if err != nil {
		return "", err
	}
	return replPrint(expr), nil
}

func main() {
	readln := bufio.NewScanner(os.Stdin)
	// repl loop
	fmt.Print("\nrepl> ")
	for readln.Scan() {
		input := strings.TrimSpace(readln.Text())
		output, err := repl(input)
		if err != nil {
			os.Stderr.WriteString(err.Error())
		} else {
			fmt.Println(output)
		}
		fmt.Print("\nrepl> ")
	}
	if err := readln.Err(); err != nil {
		panic(err)
	}
}
