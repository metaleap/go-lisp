package main

import (
	"errors"
	"fmt"
)

func evalAndApply(env *Env, expr Expr) (Expr, error) {
	if it, is := expr.(ExprList); is && len(it) > 0 {
		var special_form Expr
		if isIdent(it[0]) {
			special_form = envSpecials.Map[it[0].(ExprIdent)]
		}

		var err error
		var list ExprList
		if special_form != nil {
			list = make(ExprList, len(it))
			copy(list, it)
			list[0], err = evalExpr(env, list[0])
			if err != nil {
				return nil, err
			}
		} else {
			expr, err = evalExpr(env, it)
			if err != nil {
				return nil, err
			}
			list = expr.(ExprList)
		}

		fn, err := reqType[ExprFunc](list[0])
		if err != nil {
			return nil, errors.New("uncallable: " + fmt.Sprintf("%#v", list[0]))
		}
		return fn(env, list[1:])
	}
	return evalExpr(env, expr)
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
