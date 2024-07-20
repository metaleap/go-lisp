package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	src_stdlib := ` ; separate each top-level def by an empty line (ie. 2 newlines) due to the below hacky loading
(def not
  (fn (b)
    (if b :false :true)))

(def loadFile
  (fn (srcFilePath)
    (def src (readTextFile srcFilePath))
    (set src (str "(do " src "\n:nil)"))
    (def expr (readExpr src))
    (eval expr)))
`

	for _, stdlib_def := range strings.Split(src_stdlib, "\n\n") {
		if _, err := replNext(stdlib_def); err != nil {
			panic(err)
		}
	}

	readln := bufio.NewScanner(os.Stdin) // want line-editing? just run with `rlwrap`
	const prompt = "\n࿊  "
	fmt.Print(prompt)
	for readln.Scan() {
		input := strings.TrimSpace(readln.Text())
		output, err := replNext(input)
		if err != nil {
			msg := err.Error()
			os.Stderr.WriteString(strings.Repeat("~", 2+len(msg)) + "\n " + msg + "\n" + strings.Repeat("~", 2+len(msg)) + "\n")
		} else if output != "" {
			fmt.Println(output)
		}
		fmt.Print(prompt)
	}
	if err := readln.Err(); err != nil {
		panic(err)
	}
}

func replNext(str string) (string, error) {
	expr, err := replRead(str)
	if err != nil || expr == nil {
		return "", err
	}
	expr, err = replEval(expr, &envMain)
	if err != nil {
		return "", err
	}
	return replPrint(expr), nil
}

func replRead(str string) (Expr, error) {
	return readExpr(str)
}

func replEval(expr Expr, env *Env) (Expr, error) {
	return evalAndApply(env, expr)
}

func replPrint(expr Expr) string {
	return printExpr(expr, true)
}
