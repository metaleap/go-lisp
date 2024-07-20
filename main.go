package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	// load in the mini-stdlib
	if _, err := readAndEval("(" + string(exprIdentDo) + " " + srcMiniStdlib + "\n" + string(exprNil) + ")"); err != nil {
		panic(err)
	}

	// check if we are to run the REPL or run a specified source file
	if len(os.Args) > 1 { // run the specified source file and exit
		addOsArgsToEnv()
		if _, err := readAndEval(fmt.Sprintf("(loadFile %q)", os.Args[1])); err != nil {
			panic(err)
		}
		return
	}

	// read-eval-print loop (REPL)
	readln := bufio.NewScanner(os.Stdin) // want line-editing? just run with `rlwrap`
	const prompt = "\n࿊  "
	for fmt.Print(prompt); readln.Scan(); fmt.Print(prompt) {
		input := strings.TrimSpace(readln.Text())
		expr, err := readAndEval(input)
		if err != nil {
			msg := err.Error()
			os.Stderr.WriteString(strings.Repeat("~", 2+len(msg)) + "\n " + msg + "\n" + strings.Repeat("~", 2+len(msg)) + "\n")
		} else if output := exprToString(expr, true); output != "" {
			fmt.Println(output)
		}
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

func addOsArgsToEnv() {
	args := make(ExprList, 0, len(os.Args)-2)
	for _, arg := range os.Args[2:] {
		args = append(args, ExprStr(arg))
	}
	envMain.set("osArgs", args)
}

const srcMiniStdlib = `
(def not
  (fn (b)
    (if b :false :true)))
(def and
  (macro (b1 b2)
    ´(if ~b1 ~b2 :false)))
(def or
  (macro (b1 b2)
    ´(if ~b1 :true ~b2)))

(def loadFile
  (fn (srcFilePath)
    (def src (readTextFile srcFilePath))
    (set src (str "(do " src "\n:nil)"))
    (def expr (readExpr src))
    (eval expr)))

(def postfix ;;; turns (1 2 +) into (+ 1 2)
  (macro (call)
    (if (and (is :list call) (> (count call) 1))
      (cons (listAt call -1) (listAt call 0 -2))
      call)))
`
