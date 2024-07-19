package main

import (
	"errors"
	"fmt"
)

var (
	repl_env = Env{Map: map[ExprIdent]Expr{
		"+": ExprFunc(stdAdd),
		"-": ExprFunc(stdSub),
		"*": ExprFunc(stdMul),
		"/": ExprFunc(stdDiv),
	}}
)

type Env struct {
	Parent *Env
	Map    map[ExprIdent]Expr
}

func (me *Env) Lookup(name ExprIdent) (Expr, bool) {
	found, ok := me.Map[name]
	if (!ok) && (me.Parent != nil) {
		return me.Parent.Lookup(name)
	}
	return found, ok
}

func eval(expr Expr, env *Env) (Expr, error) {
	switch it := expr.(type) {
	case ExprIdent:
		expr, ok := env.Lookup(it)
		if !ok {
			return nil, errors.New("undefined: " + string(it))
		}
		return expr, nil
	case ExprVec:
		var err error
		vec := make(ExprVec, len(it))
		for i, item := range it {
			vec[i], err = eval(item, env)
			if err != nil {
				return nil, err
			}
		}
		return vec, nil
	case ExprHashMap:
		var err error
		hash_map := make(ExprHashMap, len(it))
		for key, value := range it {
			hash_map[key], err = eval(value, env)
			if err != nil {
				return nil, err
			}
		}
		return hash_map, nil
	case ExprList:
		var err error
		list := make(ExprList, len(it))
		for i, item := range it {
			list[i], err = eval(item, env)
			if err != nil {
				return nil, err
			}
		}
		if len(list) > 0 {
			fn, err := mustType[ExprFunc](list[0])
			if err != nil {
				return nil, errors.New("uncallable: " + fmt.Sprintf("%#v", list[0]))
			} else {
				return fn(list[1:])
			}
		}
		return list, nil
	}
	return expr, nil
}
