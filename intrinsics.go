package main

import (
	"fmt"
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

func stdAdd(args []Expr) (Expr, error) {
	if err := mustArgCountExactly(2, args); err != nil {
		return nil, err
	}
	op1, err := mustType[Num](args[0])
	if err != nil {
		return nil, err
	}
	op2, err := mustType[Num](args[1])
	if err != nil {
		return nil, err
	}
	return op1 + op2, nil
}

func stdSub(args []Expr) (Expr, error) {
	if err := mustArgCountExactly(2, args); err != nil {
		return nil, err
	}
	op1, err := mustType[Num](args[0])
	if err != nil {
		return nil, err
	}
	op2, err := mustType[Num](args[1])
	if err != nil {
		return nil, err
	}
	return op1 - op2, nil
}

func stdMul(args []Expr) (Expr, error) {
	if err := mustArgCountExactly(2, args); err != nil {
		return nil, err
	}
	op1, err := mustType[Num](args[0])
	if err != nil {
		return nil, err
	}
	op2, err := mustType[Num](args[1])
	if err != nil {
		return nil, err
	}
	return op1 * op2, nil
}

func stdDiv(args []Expr) (Expr, error) {
	if err := mustArgCountExactly(2, args); err != nil {
		return nil, err
	}
	op1, err := mustType[Num](args[0])
	if err != nil {
		return nil, err
	}
	op2, err := mustType[Num](args[1])
	if err != nil {
		return nil, err
	}
	return op1 / op2, nil
}
