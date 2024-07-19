package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func replRead(str string) (Expr, error) {
	return readExpr(str)
}

func replEval(expr Expr, env *Env) (Expr, error) {
	return eval(expr, env)
}

func replPrint(expr Expr) string {
	return printExpr(expr, true)
}

func repl(str string) (string, error) {
	expr, err := replRead(str)
	if err != nil || expr == nil {
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
			os.Stderr.WriteString(err.Error() + "\n")
		} else if output != "" {
			fmt.Println(output)
		}
		fmt.Print("\nrepl> ")
	}
	if err := readln.Err(); err != nil {
		panic(err)
	}
}
