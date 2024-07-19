package main

import (
	"errors"
	"fmt"
)

var (
	repl_env = Env{Map: map[string]Expr{
		"+": stdAdd,
		"-": stdSub,
		"*": stdMul,
		"/": stdDiv,
	}}
)

type Env struct {
	Parent *Env
	Map    map[string]Expr
}

func (me *Env) Lookup(name string) (Expr, bool) {
	found, ok := me.Map[name]
	if (!ok) && (me.Parent != nil) {
		return me.Parent.Lookup(name)
	}
	return found, ok
}

func eval(expr Expr, env *Env) (Expr, error) {
	switch it := expr.(type) {
	case Symbol:
		expr, ok := env.Lookup(it.Val)
		if !ok {
			return nil, errors.New("undefined: " + it.Val)
		}
		return expr, nil
	case Vector:
		var err error
		vec := Vector{Val: make([]Expr, len(it.Val))}
		for i, item := range it.Val {
			vec.Val[i], err = eval(item, env)
			if err != nil {
				return nil, err
			}
		}
		return vec, nil
	case HashMap:
		var err error
		hash_map := HashMap{Val: make(map[string]Expr, len(it.Val))}
		for key, value := range it.Val {
			hash_map.Val[key], err = eval(value, env)
			if err != nil {
				return nil, err
			}
		}
		return hash_map, nil
	case List:
		var err error
		list := List{Val: make([]Expr, len(it.Val))}
		for i, item := range it.Val {
			list.Val[i], err = eval(item, env)
			if err != nil {
				return nil, err
			}
		}
		if len(list.Val) > 0 {
			if fn, ok := list.Val[0].(func([]Expr) (Expr, error)); !ok {
				return nil, errors.New("uncallable: " + fmt.Sprintf("%#v", list.Val[0]))
			} else {
				return fn(list.Val[1:])
			}
		}
		return list, nil
	}
	return expr, nil
}
