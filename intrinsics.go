package main

import (
	"errors"
	"fmt"
)

func init() {
	for k, v := range map[ExprIdent]Expr{
		"def":   ExprFunc(stdDef),
		"set":   ExprFunc(stdSet),
		"let":   ExprFunc(stdLet),
		"begin": ExprFunc(stdBegin),
	} {
		envUnEvals.Map[k] = v
	}
	for k, v := range map[ExprIdent]Expr{
		"+": ExprFunc(stdAdd),
		"-": ExprFunc(stdSub),
		"*": ExprFunc(stdMul),
		"/": ExprFunc(stdDiv),
	} {
		envMain.Map[k] = v
	}
}

var (
	exprTrue  = ExprIdent("true")
	exprFalse = ExprIdent("false")
	exprNil   = ExprIdent("nil")
)

func mustType[T any](have Expr) (T, error) {
	ret, ok := have.(T)
	if !ok {
		return ret, fmt.Errorf("expected %T, not %T", ret, have)
	}
	return ret, nil
}

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
	} else if _, err := env.Get(name); (!isDef) && err != nil {
		return nil, err
	}

	expr, err := eval(env, args[1])
	if err != nil {
		return nil, err
	}
	env.Set(name, expr)
	return expr, nil
}

func stdBegin(env *Env, args []Expr) (expr Expr, err error) {
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
	let_env := Env{Parent: env, Map: make(map[ExprIdent]Expr, len(bindings))}
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
		expr, err := eval(&let_env, pair[1])
		if err != nil {
			return nil, err
		}
		let_env.Set(name, expr)
	}
	return stdBegin(&let_env, args[1:])
}
