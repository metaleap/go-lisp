package main

import (
	"fmt"
	"io"
	"reflect"
)

func isNilOrFalse(expr Expr) bool {
	return isEq(exprNil, expr) || isEq(exprFalse, expr)
}

func readUntil(r io.Reader, until byte, initialBufCapacity int) (string, error) {
	buf := make([]byte, 0, initialBufCapacity)
	var b [1]byte
	for {
		_, err := r.Read(b[0:1])
		if err != nil {
			return "", err
		} else if b[0] == until {
			break
		} else {
			buf = append(buf, b[0])
		}
	}
	line := string(buf)
	return line, nil
}

func checkArgsCount(wantAtLeast int, wantAtMost int, have []Expr) error {
	if wantAtLeast < 0 {
		return nil
	} else if want_exactly := wantAtLeast; (want_exactly == wantAtMost) && (want_exactly != len(have)) {
		return fmt.Errorf("expected %d arg(s), not %d", want_exactly, len(have))
	} else if len(have) < wantAtLeast {
		return fmt.Errorf("expected at least %d arg(s), not %d", wantAtLeast, len(have))
	} else if (wantAtMost > wantAtLeast) && (len(have) > wantAtMost) {
		return fmt.Errorf("expected %d to %d arg(s), not %d", wantAtLeast, wantAtMost, len(have))
	}
	return nil
}

func checkIs[T Expr](have Expr) (T, error) {
	ret, ok := have.(T)
	if !ok {
		if reflect.TypeOf(ret) == reflect.TypeFor[*ExprFn]() {
			panic("WHOIS?")
		}
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
	max_args_count := -1
	if exactArgsCount {
		max_args_count = 2
	}
	if err = checkArgsCount(2, max_args_count, have); err != nil {
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

func newErrNotCallable(expr Expr) error {
	return fmt.Errorf("not callable: `%s`", str(true, expr))
}
