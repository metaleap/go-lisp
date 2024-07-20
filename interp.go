package main

import (
	"errors"
	"fmt"
)

const disableTco = false

// to confirm TCO still works, uncomment the 2 commented lines in here below.
// another way, run `(sum2 10000000 0)` with TCO disabled (stack overflow) and then re-enabled (no stack overflow), where `sum2` is in github.com/kanaka/mal/blob/master/impls/tests/step5_tco.mal
func evalAndApply(env *Env, expr Expr) (Expr, error) {
	// id := time.Now().UnixNano()
	var err error
	for env != nil {
		// println("ITER", id, printExpr(expr, true))
		if it, is := expr.(ExprList); (!is) || len(it) == 0 {
			expr, err = evalExpr(env, expr)
			env = nil
		} else {
			var special_form FnSpecial
			if ident, _ := it[0].(ExprIdent); ident != "" {
				special_form = specialForms[ident]
			}
			if special_form != nil {
				if env, expr, err = special_form(env, it[1:]); err != nil {
					return nil, err
				}
			} else {
				expr, err = evalExpr(env, it)
				if err != nil {
					return nil, err
				}
				list := expr.(ExprList)
				args := list[1:]
				switch fn := list[0].(type) {
				default:
					return nil, errors.New("not callable: " + fmt.Sprintf("%#v", list[0]))
				case ExprFunc:
					expr, err = fn(env, args)
					env = nil
				case *ExprFn:
					expr = fn.body
					env, err = fn.newEnv(args)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}
	return expr, err
}

func evalExpr(env *Env, expr Expr) (Expr, error) {
	switch it := expr.(type) {
	case ExprIdent:
		return env.get(it)
	case ExprHashMap:
		var err error
		hash_map := make(ExprHashMap, len(it))
		for key, value := range it {
			hash_map[key], err = evalAndApply(env, value)
			if err != nil {
				return nil, err
			}
		}
		return hash_map, nil
	case ExprVec:
		var err error
		vec := make(ExprVec, len(it))
		for i, item := range it {
			vec[i], err = evalAndApply(env, item)
			if err != nil {
				return nil, err
			}
		}
		return vec, nil
	case ExprList:
		var err error
		list := make(ExprList, len(it))
		for i, item := range it {
			list[i], err = evalAndApply(env, item)
			if err != nil {
				return nil, err
			}
		}
		return list, nil
	}
	return expr, nil
}
