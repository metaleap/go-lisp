package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	const src_stdlib = `
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

	if _, err := readAndEval("(do " + src_stdlib + "\n:nil)"); err != nil {
		panic(err)
	}

	readln := bufio.NewScanner(os.Stdin) // want line-editing? just run with `rlwrap`
	const prompt = "\n࿊  "
	fmt.Print(prompt)
	for readln.Scan() { // read-eval-print loop (REPL)
		input := strings.TrimSpace(readln.Text())
		expr, err := readAndEval(input)
		if err != nil {
			msg := err.Error()
			os.Stderr.WriteString(strings.Repeat("~", 2+len(msg)) + "\n " + msg + "\n" + strings.Repeat("~", 2+len(msg)) + "\n")
		} else if output := exprToString(expr, true); output != "" {
			fmt.Println(output)
		}
		fmt.Print(prompt)
	}
	if err := readln.Err(); err != nil {
		panic(err)
	}
}

func readAndEval(str string) (Expr, error) {
	expr, err := readExpr(str)
	if err != nil || expr == nil {
		return nil, err
	}
	return evalAndApply(&envMain, expr)
}
