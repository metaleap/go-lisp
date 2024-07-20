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
	return evalAndApply(env, expr)
}

func replPrint(expr Expr) string {
	return printExpr(expr, true)
}

func repl(str string) (string, error) {
	expr, err := replRead(str)
	if err != nil || expr == nil {
		return "", err
	}
	expr, err = replEval(expr, envMain)
	if err != nil {
		return "", err
	}
	return replPrint(expr), nil
}

func main() {
	readln := bufio.NewScanner(os.Stdin) // for line-editing, just run with `rlwrap`

	// define `not` in source:
	input := "(def not (fn (b) (if b :false :true)))"
	if _, err := repl(input); err != nil {
		panic(err)
	}

	// repl loop
	fmt.Print("\nrepl> ")
	for readln.Scan() {
		input := strings.TrimSpace(readln.Text())
		output, err := repl(input)
		if err != nil {
			msg := err.Error()
			os.Stderr.WriteString(strings.Repeat("~", 2+len(msg)) + "\n " + msg + "\n" + strings.Repeat("~", 2+len(msg)) + "\n")
		} else if output != "" {
			fmt.Println(output)
		}
		fmt.Print("\nrepl> ")
	}
	if err := readln.Err(); err != nil {
		panic(err)
	}
}
