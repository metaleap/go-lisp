package main

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	envMain = Env{Map: map[ExprIdent]Expr{
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
	}}
	specialForms = map[ExprIdent]FnSpecial{}
)

func init() { // in here, rather than above, to avoid "initialization cycle" error
	envMain.Map["eval"] = ExprFunc(stdEval)
	specialForms = map[ExprIdent]FnSpecial{
		"def": stdDef,
		"set": stdSet,
		"let": stdLet,
		"do":  stdDo,
		"if":  stdIf,
		"fn":  stdFn,
	}
}

var (
	exprTrue  = ExprKeyword(":true")
	exprFalse = ExprKeyword(":false")
	exprNil   = ExprKeyword(":nil")
)

func exprBool(b bool) ExprKeyword {
	if b {
		return exprTrue
	}
	return exprFalse
}

func checkArgCountExactly(want int, have []Expr) error {
	if len(have) != want {
		return fmt.Errorf("expected %d arg(s), not %d", want, len(have))
	}
	return nil
}

func checkArgCountAtLeast(want int, have []Expr) error {
	if len(have) < want {
		return fmt.Errorf("expected at least %d arg(s), not %d", want, len(have))
	}
	return nil
}

func checkIs[T any](have Expr) (T, error) {
	ret, ok := have.(T)
	if !ok {
		return ret, fmt.Errorf("expected %T, not %T", ret, have)
	}
	return ret, nil
}

func checkAre[T any](have ...Expr) error {
	for _, expr := range have {
		if _, err := checkIs[T](expr); err != nil {
			return err
		}
	}
	return nil
}

func checkAreBoth[T1 any, T2 any](have ...Expr) (ret1 T1, ret2 T2, err error) {
	if err = checkArgCountExactly(2, have); err != nil {
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
	op1, op2, err := checkAreBoth[ExprNum, ExprNum](args...)
	if err != nil {
		return nil, err
	}
	return op1 + op2, nil
}

func stdSub(args []Expr) (Expr, error) {
	op1, op2, err := checkAreBoth[ExprNum, ExprNum](args...)
	if err != nil {
		return nil, err
	}
	return op1 - op2, nil
}

func stdMul(args []Expr) (Expr, error) {
	op1, op2, err := checkAreBoth[ExprNum, ExprNum](args...)
	if err != nil {
		return nil, err
	}
	return op1 * op2, nil
}

func stdDiv(args []Expr) (Expr, error) {
	op1, op2, err := checkAreBoth[ExprNum, ExprNum](args...)
	if err != nil {
		return nil, err
	}
	return op1 / op2, nil
}

func stdDef(env *Env, args []Expr) (*Env, Expr, error) {
	return defOrSet(true, env, args)
}
func stdSet(env *Env, args []Expr) (*Env, Expr, error) {
	return defOrSet(false, env, args)
}
func defOrSet(isDef bool, env *Env, args []Expr) (*Env, Expr, error) {
	if err := checkArgCountExactly(2, args); err != nil {
		return nil, nil, err
	}
	name, err := checkIs[ExprIdent](args[0])
	if err != nil {
		return nil, nil, err
	}
	if _, is_special := specialForms[name]; is_special {
		return nil, nil, fmt.Errorf("cannot redefine `%s`", name)
	}
	if isDef && env.hasOwn(name) {
		return nil, nil, errors.New("already defined: " + string(name))
	} else if _, err := env.get(name); (!isDef) && err != nil {
		return nil, nil, err
	}

	expr, err := evalAndApply(env, args[1])
	if err != nil {
		return nil, nil, err
	}
	env.set(name, expr)
	return nil, expr, nil
}

func stdDo(env *Env, args []Expr) (tailEnv *Env, expr Expr, err error) {
	if err = checkArgCountAtLeast(1, args); err != nil {
		return
	}
	for _, arg := range args[:len(args)-1] {
		if expr, err = evalAndApply(env, arg); err != nil {
			return
		}
	}
	tailEnv, expr = env, args[len(args)-1]
	return
}

func stdLet(env *Env, args []Expr) (*Env, Expr, error) {
	if err := checkArgCountAtLeast(2, args); err != nil {
		return nil, nil, err
	}
	bindings, err := checkIsSeq(args[0])
	if err != nil {
		return nil, nil, err
	}
	let_env := newEnv(env, nil, nil)
	for _, binding := range bindings {
		pair, err := checkIsSeq(binding)
		if err != nil {
			return nil, nil, err
		}
		if err = checkArgCountExactly(2, pair); err != nil {
			return nil, nil, err
		}
		name, err := checkIs[ExprIdent](pair[0])
		if err != nil {
			return nil, nil, err
		}
		if _, is_special := specialForms[name]; is_special {
			return nil, nil, fmt.Errorf("cannot redefine `%s`", name)
		}
		expr, err := evalAndApply(let_env, pair[1])
		if err != nil {
			return nil, nil, err
		}
		let_env.set(name, expr)
	}
	return stdDo(let_env, args[1:])
}

func stdIf(env *Env, args []Expr) (*Env, Expr, error) {
	if err := checkArgCountExactly(3, args); err != nil {
		return nil, nil, err
	}
	expr, err := evalAndApply(env, args[0])
	if err != nil {
		return nil, nil, err
	}
	idx := 1
	if isEq(expr, exprFalse) || isEq(expr, exprNil) {
		idx = 2
	}
	return env, args[idx], nil
}

func stdFn(env *Env, args []Expr) (*Env, Expr, error) {
	if err := checkArgCountAtLeast(2, args); err != nil {
		return nil, nil, err
	}
	params, err := checkIsSeq(args[0])
	if err != nil {
		return nil, nil, err
	}
	if err := checkAre[ExprIdent](params...); err != nil {
		return nil, nil, err
	}
	body := args[1]
	if len(args) > 2 {
		body = append(ExprList{ExprIdent("do")}, args[1:]...)
	}
	var expr Expr = &ExprFn{params: params, body: body, env: env}
	if disableTco {
		expr = expr.(*ExprFn).ToFunc()
	}
	return nil, expr, nil
}

func (me *ExprFn) newEnv(callerArgs []Expr) (*Env, error) {
	if err := checkArgCountExactly(len(me.params), callerArgs); err != nil {
		return nil, err
	}
	return newEnv(me.env, me.params, callerArgs), nil
}

func (me *ExprFn) ToFunc() ExprFunc {
	return ExprFunc(func(callerArgs []Expr) (Expr, error) {
		env, err := me.newEnv(callerArgs)
		if err != nil {
			return nil, err
		}
		return evalAndApply(env, me.body)
	})
}

func str(args []Expr, printReadably bool) string {
	var buf strings.Builder
	for i, arg := range args {
		if i > 0 && printReadably {
			buf.WriteByte(' ')
		}
		buf.WriteString(printExpr(arg, printReadably))
	}
	return buf.String()
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
	if err := checkArgCountExactly(2, args); err != nil {
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
	default:
		return nil, fmt.Errorf("expected not `%s` but one of: `:list`, `:ident`, `:str`, `:num`, `:vec`, `:hashmap`, `:fn`, `:keyword`", kind)
	}
	return exprBool(ok), nil
}

func stdIsEmpty(args []Expr) (Expr, error) {
	if err := checkArgCountExactly(1, args); err != nil {
		return nil, err
	}
	list, err := checkIs[ExprList](args[0])
	if err != nil {
		return nil, err
	}
	return exprBool(len(list) == 0), nil
}

func stdCount(args []Expr) (Expr, error) {
	if err := checkArgCountExactly(1, args); err != nil {
		return nil, err
	}
	list, err := checkIs[ExprList](args[0])
	if err != nil {
		return nil, err
	}
	return ExprNum(len(list)), nil
}

func stdEq(args []Expr) (Expr, error) {
	if err := checkArgCountExactly(2, args); err != nil {
		return nil, err
	}
	return exprBool(isEq(args[0], args[1])), nil
}

func compare(args []Expr) (int, error) {
	if err := checkArgCountExactly(2, args); err != nil {
		return 0, err
	}
	switch it := args[0].(type) {
	case ExprNum:
		if other, ok := args[1].(ExprNum); ok {
			return cmp.Compare(it, other), nil
		}
	case ExprStr:
		if other, ok := args[1].(ExprStr); ok {
			return cmp.Compare(it, other), nil
		}
	}
	return 0, fmt.Errorf("specified operands are not comparable")
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
	if err := checkArgCountExactly(1, args); err != nil {
		return nil, err
	}
	src, err := checkIs[ExprStr](args[0])
	if err != nil {
		return nil, err
	}
	return readExpr(string(src))
}

func stdReadTextFile(args []Expr) (Expr, error) {
	if err := checkArgCountExactly(1, args); err != nil {
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
	if err := checkArgCountExactly(1, args); err != nil {
		return nil, err
	}
	return evalAndApply(&envMain, args[0])
}
