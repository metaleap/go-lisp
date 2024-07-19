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
		envSpecials.Map[k] = v
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

func reqArgCountExactly(want int, have []Expr) error {
	if len(have) != want {
		return fmt.Errorf("expected %d args, not %d", want, len(have))
	}
	return nil
}

func reqArgCountAtLeast(want int, have []Expr) error {
	if len(have) < want {
		return fmt.Errorf("expected at least %d args, not %d", want, len(have))
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

func stdAdd(env *Env, args []Expr) (Expr, error) {
	op1, op2, err := reqTypes2[ExprNum, ExprNum](args...)
	if err != nil {
		return nil, err
	}
	return op1 + op2, nil
}

func stdSub(env *Env, args []Expr) (Expr, error) {
	op1, op2, err := reqTypes2[ExprNum, ExprNum](args...)
	if err != nil {
		return nil, err
	}
	return op1 - op2, nil
}

func stdMul(env *Env, args []Expr) (Expr, error) {
	op1, op2, err := reqTypes2[ExprNum, ExprNum](args...)
	if err != nil {
		return nil, err
	}
	return op1 * op2, nil
}

func stdDiv(env *Env, args []Expr) (Expr, error) {
	op1, op2, err := reqTypes2[ExprNum, ExprNum](args...)
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
	if err := reqArgCountExactly(2, args); err != nil {
		return nil, err
	}
	name, err := reqType[ExprIdent](args[0])
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
	if err = reqArgCountAtLeast(1, args); err != nil {
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
	if err := reqArgCountAtLeast(2, args); err != nil {
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
		if err = reqArgCountExactly(2, pair); err != nil {
			return nil, err
		}
		name, err := reqType[ExprIdent](pair[0])
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
	if err := reqArgCountExactly(2, args); err != nil {
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
	if err := reqArgCountExactly(3, args); err != nil {
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
	if err := reqArgCountAtLeast(2, args); err != nil {
		return nil, err
	}
	params, err := mustSeq(args[0])
	if err != nil {
		return nil, err
	}
	if err := reqTypes[ExprIdent](params...); err != nil {
		return nil, err
	}
	return ExprFunc(func(_callerEnv *Env, callerArgs []Expr) (Expr, error) {
		if err := reqArgCountExactly(len(params), callerArgs); err != nil {
			return nil, err
		}
		env_closure := newEnv(env, params, callerArgs)
		return stdDo(env_closure, args[1:])
	}), nil
}
