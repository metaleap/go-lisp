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
		"print":   ExprFunc(stdPrint),
		"println": ExprFunc(stdPrintln),
		"str":     ExprFunc(stdStr),
		"show":    ExprFunc(stdShow),
		"list":    ExprFunc(stdList),
		"is":      ExprFunc(stdIs),
		"isEmpty": ExprFunc(stdIsEmpty),
		"count":   ExprFunc(stdCount),
		"cmp":     ExprFunc(stdCmp),
		"+":       ExprFunc(stdAdd),
		"-":       ExprFunc(stdSub),
		"*":       ExprFunc(stdMul),
		"/":       ExprFunc(stdDiv),
		"=":       ExprFunc(stdEq),
		"<":       ExprFunc(stdLt),
		">":       ExprFunc(stdGt),
		"<=":      ExprFunc(stdLe),
		">=":      ExprFunc(stdGe),
	}}
	specialForms = map[ExprIdent]FnSpecial{}
)

func init() {
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

func reqArgCountExactly(want int, have []Expr) error {
	if len(have) != want {
		return fmt.Errorf("expected %d arg(s), not %d", want, len(have))
	}
	return nil
}

func reqArgCountAtLeast(want int, have []Expr) error {
	if len(have) < want {
		return fmt.Errorf("expected at least %d arg(s), not %d", want, len(have))
	}
	return nil
}

func reqType[T any](have Expr) (T, error) {
	ret, ok := have.(T)
	if !ok {
		return ret, fmt.Errorf("expected %T, not %T", ret, have)
	}
	return ret, nil
}

func reqTypes[T any](have ...Expr) error {
	for _, expr := range have {
		if _, err := reqType[T](expr); err != nil {
			return err
		}
	}
	return nil
}

func reqTypes2[T1 any, T2 any](have ...Expr) (ret1 T1, ret2 T2, err error) {
	if err = reqArgCountExactly(2, have); err != nil {
		return
	}
	if ret1, err = reqType[T1](have[0]); err != nil {
		return
	}
	if ret2, err = reqType[T2](have[1]); err != nil {
		return
	}
	return
}

func mustSeq(expr Expr) ([]Expr, error) {
	switch expr := expr.(type) {
	case ExprList:
		return ([]Expr)(expr), nil
	case ExprVec:
		return ([]Expr)(expr), nil
	default:
		return nil, fmt.Errorf("expected list or vector, not %T", expr)
	}
}

func stdAdd(_ *Env, args []Expr) (Expr, error) {
	op1, op2, err := reqTypes2[ExprNum, ExprNum](args...)
	if err != nil {
		return nil, err
	}
	return op1 + op2, nil
}

func stdSub(_ *Env, args []Expr) (Expr, error) {
	op1, op2, err := reqTypes2[ExprNum, ExprNum](args...)
	if err != nil {
		return nil, err
	}
	return op1 - op2, nil
}

func stdMul(_ *Env, args []Expr) (Expr, error) {
	op1, op2, err := reqTypes2[ExprNum, ExprNum](args...)
	if err != nil {
		return nil, err
	}
	return op1 * op2, nil
}

func stdDiv(_ *Env, args []Expr) (Expr, error) {
	op1, op2, err := reqTypes2[ExprNum, ExprNum](args...)
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
	if err := reqArgCountExactly(2, args); err != nil {
		return nil, nil, err
	}
	name, err := reqType[ExprIdent](args[0])
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
	if err = reqArgCountAtLeast(1, args); err != nil {
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

func tmpDo(env *Env, args []Expr) (tailEnv *Env, expr Expr, err error) {
	if err = reqArgCountAtLeast(1, args); err != nil {
		return
	}
	for _, arg := range args {
		if expr, err = evalAndApply(env, arg); err != nil {
			return
		}
	}
	return
}

func stdLet(env *Env, args []Expr) (*Env, Expr, error) {
	if err := reqArgCountAtLeast(2, args); err != nil {
		return nil, nil, err
	}
	bindings, err := mustSeq(args[0])
	if err != nil {
		return nil, nil, err
	}
	let_env := newEnv(env, nil, nil)
	for _, binding := range bindings {
		pair, err := mustSeq(binding)
		if err != nil {
			return nil, nil, err
		}
		if err = reqArgCountExactly(2, pair); err != nil {
			return nil, nil, err
		}
		name, err := reqType[ExprIdent](pair[0])
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
	if err := reqArgCountExactly(3, args); err != nil {
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
	if err := reqArgCountAtLeast(2, args); err != nil {
		return nil, nil, err
	}
	params, err := mustSeq(args[0])
	if err != nil {
		return nil, nil, err
	}
	if err := reqTypes[ExprIdent](params...); err != nil {
		return nil, nil, err
	}
	body := args[1]
	if len(args) > 2 {
		body = append(ExprList{ExprIdent("do")}, args[1:]...)
	}
	return nil, &ExprFn{params: params, body: body, env: env}, nil
}

func (me *ExprFn) newEnv(callerArgs []Expr) (*Env, error) {
	if err := reqArgCountExactly(len(me.params), callerArgs); err != nil {
		return nil, err
	}
	return newEnv(me.env, me.params, callerArgs), nil
}

func (me *ExprFn) Call(_ *Env, callerArgs []Expr) (Expr, error) {
	env, err := me.newEnv(callerArgs)
	if err != nil {
		return nil, err
	}
	return evalAndApply(env, me.body)
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

func stdPrint(_ *Env, args []Expr) (Expr, error) {
	os.Stdout.WriteString(str(args, true))
	return exprNil, nil
}
func stdPrintln(_ *Env, args []Expr) (Expr, error) {
	os.Stdout.WriteString(str(args, true) + "\n")
	return exprNil, nil
}
func stdStr(_ *Env, args []Expr) (Expr, error) {
	return ExprStr(str(args, false)), nil
}
func stdShow(_ *Env, args []Expr) (Expr, error) {
	return ExprStr(str(args, true)), nil
}

func stdList(_ *Env, args []Expr) (Expr, error) {
	return ExprList(args), nil
}

func stdIs(_ *Env, args []Expr) (Expr, error) {
	if err := reqArgCountExactly(2, args); err != nil {
		return nil, err
	}
	kind, err := reqType[ExprKeyword](args[0])
	if err != nil {
		return nil, err
	}
	var ok bool
	switch kind {
	case ":list":
		_, ok = args[1].(ExprList)
	case ":ident":
		_, ok = args[1].(ExprIdent)
	case ":str":
		_, ok = args[1].(ExprStr)
	case ":num":
		_, ok = args[1].(ExprNum)
	case ":vec":
		_, ok = args[1].(ExprVec)
	case ":hashmap":
		_, ok = args[1].(ExprHashMap)
	case ":fn":
		_, ok = args[1].(ExprFunc)
	case ":keyword":
		_, ok = args[1].(ExprKeyword)
	default:
		return nil, fmt.Errorf("expected not `%s` but one of: `:list`, `:ident`, `:str`, `:num`, `:vec`, `:hashmap`, `:fn`, `:keyword`", kind)
	}
	return exprBool(ok), nil
}

func stdIsEmpty(_ *Env, args []Expr) (Expr, error) {
	if err := reqArgCountExactly(1, args); err != nil {
		return nil, err
	}
	list, err := reqType[ExprList](args[0])
	if err != nil {
		return nil, err
	}
	return exprBool(len(list) == 0), nil
}

func stdCount(_ *Env, args []Expr) (Expr, error) {
	if err := reqArgCountExactly(1, args); err != nil {
		return nil, err
	}
	list, err := reqType[ExprList](args[0])
	if err != nil {
		return nil, err
	}
	return ExprNum(len(list)), nil
}

func stdEq(_ *Env, args []Expr) (Expr, error) {
	if err := reqArgCountExactly(2, args); err != nil {
		return nil, err
	}
	return exprBool(isEq(args[0], args[1])), nil
}

func compare(args []Expr) (int, error) {
	if err := reqArgCountExactly(2, args); err != nil {
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

func stdCmp(_ *Env, args []Expr) (Expr, error) {
	order, err := compare(args)
	if err != nil {
		return nil, err
	}
	return ExprNum(order), nil
}
func stdLt(_ *Env, args []Expr) (Expr, error) {
	order, err := compare(args)
	if err != nil {
		return nil, err
	}
	return exprBool(order == -1), nil
}
func stdLe(_ *Env, args []Expr) (Expr, error) {
	order, err := compare(args)
	if err != nil {
		return nil, err
	}
	return exprBool(order <= 0), nil
}
func stdGt(_ *Env, args []Expr) (Expr, error) {
	order, err := compare(args)
	if err != nil {
		return nil, err
	}
	return exprBool(order == 1), nil
}
func stdGe(_ *Env, args []Expr) (Expr, error) {
	order, err := compare(args)
	if err != nil {
		return nil, err
	}
	return exprBool(order >= 0), nil
}
