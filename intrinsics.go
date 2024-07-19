package main

import (
	"errors"
	"fmt"
)

func init() {
	for k, v := range map[ExprIdent]Expr{
		"def": ExprFunc(stdDef),
		"set": ExprFunc(stdSet),
		"let": ExprFunc(stdLet),
		"do":  ExprFunc(stdDo),
		"if":  ExprFunc(stdIf),
		"fn":  ExprFunc(stdFn),
	} {
		envUnEvals.Map[k] = v
	}
	for k, v := range map[ExprIdent]Expr{
		"+": ExprFunc(stdAdd),
		"-": ExprFunc(stdSub),
		"*": ExprFunc(stdMul),
		"/": ExprFunc(stdDiv),
		"=": ExprFunc(stdEq),
	} {
		envMain.Map[k] = v
	}
}

var (
	exprTrue  = ExprKeyword(":true")
	exprFalse = ExprKeyword(":false")
	exprNil   = ExprKeyword(":nil")
)

func mustArgCountExactly(want int, have []Expr) error {
	if len(have) != want {
		return fmt.Errorf("expected %d args, not %d", want, len(have))
	}
	return nil
}

func mustArgCountAtLeast(want int, have []Expr) error {
	if len(have) < want {
		return fmt.Errorf("expected at least %d args, not %d", want, len(have))
	}
	return nil
}

func mustType[T any](have Expr) (T, error) {
	ret, ok := have.(T)
	if !ok {
		return ret, fmt.Errorf("expected %T, not %T", ret, have)
	}
	return ret, nil
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

func stdAdd(env *Env, args []Expr) (Expr, error) {
	if err := mustArgCountExactly(2, args); err != nil {
		return nil, err
	}
	op1, err := mustType[ExprNum](args[0])
	if err != nil {
		return nil, err
	}
	op2, err := mustType[ExprNum](args[1])
	if err != nil {
		return nil, err
	}
	return op1 + op2, nil
}

func stdSub(env *Env, args []Expr) (Expr, error) {
	if err := mustArgCountExactly(2, args); err != nil {
		return nil, err
	}
	op1, err := mustType[ExprNum](args[0])
	if err != nil {
		return nil, err
	}
	op2, err := mustType[ExprNum](args[1])
	if err != nil {
		return nil, err
	}
	return op1 - op2, nil
}

func stdMul(env *Env, args []Expr) (Expr, error) {
	if err := mustArgCountExactly(2, args); err != nil {
		return nil, err
	}
	op1, err := mustType[ExprNum](args[0])
	if err != nil {
		return nil, err
	}
	op2, err := mustType[ExprNum](args[1])
	if err != nil {
		return nil, err
	}
	return op1 * op2, nil
}

func stdDiv(env *Env, args []Expr) (Expr, error) {
	if err := mustArgCountExactly(2, args); err != nil {
		return nil, err
	}
	op1, err := mustType[ExprNum](args[0])
	if err != nil {
		return nil, err
	}
	op2, err := mustType[ExprNum](args[1])
	if err != nil {
		return nil, err
	}
	return op1 / op2, nil
}

func stdDef(env *Env, args []Expr) (Expr, error) {
	return defOrSet(true, env, args)
}
func stdSet(env *Env, args []Expr) (Expr, error) {
	return defOrSet(false, env, args)
}
func defOrSet(isDef bool, env *Env, args []Expr) (Expr, error) {
	if err := mustArgCountExactly(2, args); err != nil {
		return nil, err
	}
	name, err := mustType[ExprIdent](args[0])
	if err != nil {
		return nil, err
	}
	if isDef && env.hasOwn(name) {
		return nil, errors.New("already defined: " + string(name))
	} else if _, err := env.get(name); (!isDef) && err != nil {
		return nil, err
	}

	expr, err := eval(env, args[1])
	if err != nil {
		return nil, err
	}
	env.set(name, expr)
	return expr, nil
}

func stdDo(env *Env, args []Expr) (expr Expr, err error) {
	if err = mustArgCountAtLeast(1, args); err != nil {
		return
	}
	for _, arg := range args {
		if expr, err = eval(env, arg); err != nil {
			return
		}
	}
	return
}

func stdLet(env *Env, args []Expr) (Expr, error) {
	if err := mustArgCountAtLeast(2, args); err != nil {
		return nil, err
	}
	bindings, err := mustSeq(args[0])
	if err != nil {
		return nil, err
	}
	let_env := newEnv(env, nil, nil)
	for _, binding := range bindings {
		pair, err := mustSeq(binding)
		if err != nil {
			return nil, err
		}
		if err = mustArgCountExactly(2, pair); err != nil {
			return nil, err
		}
		name, err := mustType[ExprIdent](pair[0])
		if err != nil {
			return nil, err
		}
		expr, err := eval(let_env, pair[1])
		if err != nil {
			return nil, err
		}
		let_env.set(name, expr)
	}
	return stdDo(let_env, args[1:])
}

func stdEq(env *Env, args []Expr) (Expr, error) {
	if err := mustArgCountExactly(2, args); err != nil {
		return nil, err
	}
	expr1, err := eval(env, args[0])
	if err != nil {
		return nil, err
	}
	expr2, err := eval(env, args[1])
	if err != nil {
		return nil, err
	}
	if isEq(expr1, expr2) {
		return exprTrue, nil
	}
	return exprFalse, nil
}

func stdIf(env *Env, args []Expr) (Expr, error) {
	if err := mustArgCountExactly(3, args); err != nil {
		return nil, err
	}
	expr, err := eval(env, args[0])
	if err != nil {
		return nil, err
	}
	idx := 1
	if isEq(expr, exprFalse) || isEq(expr, exprNil) {
		idx = 2
	}
	return eval(env, args[idx])
}

func stdFn(env *Env, args []Expr) (Expr, error) {
	if err := mustArgCountAtLeast(2, args); err != nil {
		return nil, err
	}
	params, err := mustSeq(args[0])
	if err != nil {
		return nil, err
	}
	for _, param := range params {
		if _, err = mustType[ExprIdent](param); err != nil {
			return nil, err
		}
	}
	return ExprFunc(func(_callerEnv *Env, callerArgs []Expr) (Expr, error) {
		if err := mustArgCountExactly(len(params), callerArgs); err != nil {
			return nil, err
		}
		env_closure := newEnv(env, params, callerArgs)
		return stdDo(env_closure, args[1:])
	}), nil
}
