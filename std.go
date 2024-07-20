package main

import (
	"fmt"
	"os"
)

var (
	envMain = Env{Map: map[ExprIdent]Expr{
		"osArgs":       ExprList{}, // populated by `main` when running a user-specified source file
		"print":        ExprFunc(stdPrint),
		"println":      ExprFunc(stdPrintln),
		"str":          ExprFunc(stdStr),
		"show":         ExprFunc(stdShow),
		"list":         ExprFunc(stdList),
		"is":           ExprFunc(stdIs),
		"isEmpty":      ExprFunc(stdIsEmpty),
		"count":        ExprFunc(stdCount),
		"cmp":          ExprFunc(stdCmp),
		"+":            ExprFunc(stdAdd),
		"-":            ExprFunc(stdSub),
		"*":            ExprFunc(stdMul),
		"/":            ExprFunc(stdDiv),
		"=":            ExprFunc(stdEq),
		"<":            ExprFunc(stdLt),
		">":            ExprFunc(stdGt),
		"<=":           ExprFunc(stdLe),
		">=":           ExprFunc(stdGe),
		"readExpr":     ExprFunc(stdReadExpr),
		"readTextFile": ExprFunc(stdReadTextFile),
		"atomFrom":     ExprFunc(stdAtomFrom),
		"atomGet":      ExprFunc(stdAtomGet),
		"atomSet":      ExprFunc(stdAtomSet),
		"atomSwap":     ExprFunc(stdAtomSwap),
		"cons":         ExprFunc(stdCons),
		"concat":       ExprFunc(stdConcat),
		"vec":          ExprFunc(stdVec),
	}}
)

func init() { // in here, instead of above, to avoid "initialization cycle" error:
	envMain.Map["eval"] = ExprFunc(stdEval)
}

func checkArgsCountExactly(want int, have []Expr) error {
	if (want >= 0) && (len(have) != want) {
		return fmt.Errorf("expected %d arg(s), not %d", want, len(have))
	}
	return nil
}

func checkArgsCountAtLeast(want int, have []Expr) error {
	if len(have) < want {
		return fmt.Errorf("expected at least %d arg(s), not %d", want, len(have))
	}
	return nil
}

func checkIs[T Expr](have Expr) (T, error) {
	ret, ok := have.(T)
	if !ok {
		return ret, fmt.Errorf("expected %T, not %T", ret, have)
	}
	return ret, nil
}

func checkAre[T Expr](have ...Expr) error {
	for _, expr := range have {
		if _, err := checkIs[T](expr); err != nil {
			return err
		}
	}
	return nil
}

func checkAreBoth[T1 Expr, T2 Expr](have []Expr, exactArgsCount bool) (ret1 T1, ret2 T2, err error) {
	check_args_count := checkArgsCountExactly
	if !exactArgsCount {
		check_args_count = checkArgsCountAtLeast
	}
	if err = check_args_count(2, have); err != nil {
		return
	}
	if ret1, err = checkIs[T1](have[0]); err != nil {
		return
	}
	if ret2, err = checkIs[T2](have[1]); err != nil {
		return
	}
	return
}

func checkIsSeq(expr Expr) ([]Expr, error) {
	switch expr := expr.(type) {
	case ExprList:
		return ([]Expr)(expr), nil
	case ExprVec:
		return ([]Expr)(expr), nil
	default:
		return nil, fmt.Errorf("expected list or vector, not %T", expr)
	}
}

func stdAdd(args []Expr) (Expr, error) {
	op1, op2, err := checkAreBoth[ExprNum, ExprNum](args, true)
	if err != nil {
		return nil, err
	}
	return op1 + op2, nil
}

func stdSub(args []Expr) (Expr, error) {
	op1, op2, err := checkAreBoth[ExprNum, ExprNum](args, true)
	if err != nil {
		return nil, err
	}
	return op1 - op2, nil
}

func stdMul(args []Expr) (Expr, error) {
	op1, op2, err := checkAreBoth[ExprNum, ExprNum](args, true)
	if err != nil {
		return nil, err
	}
	return op1 * op2, nil
}

func stdDiv(args []Expr) (Expr, error) {
	op1, op2, err := checkAreBoth[ExprNum, ExprNum](args, true)
	if err != nil {
		return nil, err
	}
	return op1 / op2, nil
}

func stdPrint(args []Expr) (Expr, error) {
	os.Stdout.WriteString(str(args, true))
	return exprNil, nil
}
func stdPrintln(args []Expr) (Expr, error) {
	os.Stdout.WriteString(str(args, false) + "\n")
	return exprNil, nil
}
func stdStr(args []Expr) (Expr, error) {
	return ExprStr(str(args, false)), nil
}
func stdShow(args []Expr) (Expr, error) {
	return ExprStr(str(args, true)), nil
}

func stdList(args []Expr) (Expr, error) {
	return ExprList(args), nil
}

func stdIs(args []Expr) (Expr, error) {
	if err := checkArgsCountExactly(2, args); err != nil {
		return nil, err
	}
	kind, err := checkIs[ExprKeyword](args[0])
	if err != nil {
		return nil, err
	}
	var ok bool
	switch kind {
	case ":ident":
		_, ok = args[1].(ExprIdent)
	case ":keyword":
		_, ok = args[1].(ExprKeyword)
	case ":str":
		_, ok = args[1].(ExprStr)
	case ":num":
		_, ok = args[1].(ExprNum)
	case ":list":
		_, ok = args[1].(ExprList)
	case ":vec":
		_, ok = args[1].(ExprVec)
	case ":hashmap":
		_, ok = args[1].(ExprHashMap)
	case ":fn":
		if _, ok = args[1].(*ExprFn); !ok {
			_, ok = args[1].(ExprFunc)
		}
	case ":atom":
		_, ok = args[1].(*ExprAtom)
	default:
		return nil, fmt.Errorf("expected not `%s` but one of: `:list`, `:ident`, `:str`, `:num`, `:vec`, `:hashmap`, `:fn`, `:keyword`, `:atom`", kind)
	}
	return exprBool(ok), nil
}

func stdIsEmpty(args []Expr) (Expr, error) {
	expr, err := stdCount(args)
	if err != nil {
		return nil, err
	}
	return exprBool((expr.(ExprNum) == 0)), nil
}

func stdCount(args []Expr) (Expr, error) {
	if err := checkArgsCountExactly(1, args); err != nil {
		return nil, err
	}
	list, err := checkIs[ExprList](args[0])
	if err != nil {
		return nil, err
	}
	return ExprNum(len(list)), nil
}

func stdEq(args []Expr) (Expr, error) {
	if err := checkArgsCountExactly(2, args); err != nil {
		return nil, err
	}
	return exprBool(isEq(args[0], args[1])), nil
}

func stdCmp(args []Expr) (Expr, error) {
	order, err := compare(args)
	if err != nil {
		return nil, err
	}
	return ExprNum(order), nil
}
func stdLt(args []Expr) (Expr, error) {
	order, err := compare(args)
	if err != nil {
		return nil, err
	}
	return exprBool(order == -1), nil
}
func stdLe(args []Expr) (Expr, error) {
	order, err := compare(args)
	if err != nil {
		return nil, err
	}
	return exprBool(order <= 0), nil
}
func stdGt(args []Expr) (Expr, error) {
	order, err := compare(args)
	if err != nil {
		return nil, err
	}
	return exprBool(order == 1), nil
}
func stdGe(args []Expr) (Expr, error) {
	order, err := compare(args)
	if err != nil {
		return nil, err
	}
	return exprBool(order >= 0), nil
}

func stdReadExpr(args []Expr) (Expr, error) {
	if err := checkArgsCountExactly(1, args); err != nil {
		return nil, err
	}
	src, err := checkIs[ExprStr](args[0])
	if err != nil {
		return nil, err
	}
	return readExpr(string(src))
}

func stdReadTextFile(args []Expr) (Expr, error) {
	if err := checkArgsCountExactly(1, args); err != nil {
		return nil, err
	}
	file_path, err := checkIs[ExprStr](args[0])
	if err != nil {
		return nil, err
	}
	file_bytes, err := os.ReadFile(string(file_path))
	if err != nil {
		return nil, err
	}
	return ExprStr(file_bytes), nil
}

func stdEval(args []Expr) (Expr, error) {
	if err := checkArgsCountExactly(1, args); err != nil {
		return nil, err
	}
	return evalAndApply(&envMain, args[0])
}

func stdAtomFrom(args []Expr) (Expr, error) {
	if err := checkArgsCountExactly(1, args); err != nil {
		return nil, err
	}
	return &ExprAtom{Ref: args[0]}, nil
}
func stdAtomGet(args []Expr) (Expr, error) {
	if err := checkArgsCountExactly(1, args); err != nil {
		return nil, err
	}
	atom, err := checkIs[*ExprAtom](args[0])
	if err != nil {
		return nil, err
	}
	return atom.Ref, nil
}
func stdAtomSet(args []Expr) (Expr, error) {
	if err := checkArgsCountExactly(2, args); err != nil {
		return nil, err
	}
	atom, err := checkIs[*ExprAtom](args[0])
	if err != nil {
		return nil, err
	}
	atom.Ref = args[1]
	return atom.Ref, nil
}
func stdAtomSwap(args []Expr) (Expr, error) {
	if err := checkArgsCountAtLeast(2, args); err != nil {
		return nil, err
	}
	atom, fn, err := checkAreBoth[*ExprAtom, *ExprFn](args, false)
	if err != nil {
		return nil, err
	}
	atom.Ref, err = fn.Call(append([]Expr{atom.Ref}, args[2:]...))
	if err != nil {
		return nil, err
	}
	return atom.Ref, nil
}

func stdCons(args []Expr) (Expr, error) {
	if err := checkArgsCountExactly(2, args); err != nil {
		return nil, err
	}
	list, err := checkIsSeq(args[1])
	if err != nil {
		return nil, err
	}
	return append(ExprList{args[0]}, list...), nil
}

func stdConcat(args []Expr) (Expr, error) {
	var list ExprList
	for _, arg := range args {
		seq, err := checkIsSeq(arg)
		if err != nil {
			return nil, err
		}
		list = append(list, seq...)
	}
	return list, nil
}

func stdVec(args []Expr) (Expr, error) {
	if err := checkArgsCountExactly(1, args); err != nil {
		return nil, err
	}
	if vec, is_vec := args[0].(ExprVec); is_vec {
		return vec, nil
	}
	list, err := checkIs[ExprList](args[0])
	if err != nil {
		return nil, err
	}
	return (ExprVec)(list), nil
}
