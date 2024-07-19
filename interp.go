package main

import (
	"errors"
	"fmt"
)

func eval(env *Env, expr Expr) (Expr, error) {
	switch it := expr.(type) {
	case ExprIdent:
		return env.Get(it)
	case ExprVec:
		var err error
		vec := make(ExprVec, len(it))
		for i, item := range it {
			vec[i], err = eval(env, item)
			if err != nil {
				return nil, err
			}
		}
		return vec, nil
	case ExprHashMap:
		var err error
		hash_map := make(ExprHashMap, len(it))
		for key, value := range it {
			hash_map[key], err = eval(env, value)
			if err != nil {
				return nil, err
			}
		}
		return hash_map, nil
	case ExprList:
		if len(it) == 0 {
			return nil, errors.New("we have `nil` or `[]` for that")
		}

		var intrinsic_uneval Expr
		if isIdent(it[0]) {
			intrinsic_uneval = envUnEvals.Map[it[0].(ExprIdent)]
		}

		var err error
		list := make(ExprList, len(it))
		if intrinsic_uneval != nil {
			copy(list, it)
			list[0], err = eval(env, list[0])
			if err != nil {
				return nil, err
			}
		} else {
			for i, item := range it {
				list[i], err = eval(env, item)
				if err != nil {
					return nil, err
				}
			}
		}
		if len(list) > 0 {
			fn, err := mustType[ExprFunc](list[0])
			if err != nil {
				return nil, errors.New("uncallable: " + fmt.Sprintf("%#v", list[0]))
			} else {
				return fn(env, list[1:])
			}
		}
		return list, nil
	}
	return expr, nil
}
