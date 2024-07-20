package main

import (
	"errors"
	"fmt"
)

func evalAndApply(env *Env, expr Expr) (Expr, error) {
	var err error
	for env != nil {
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
				var fn ExprFunc
				if fn, err = reqType[ExprFunc](list[0]); err != nil {
					return nil, errors.New("not callable: " + fmt.Sprintf("%#v", list[0]))
				}
				expr, err = fn(env, list[1:])
				env = nil
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
