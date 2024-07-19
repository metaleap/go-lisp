package main

import (
	"errors"
	"fmt"
)

var (
	repl_env = Env{Map: map[Ident]Expr{
		"+": Func{Fn: stdAdd},
		"-": Func{Fn: stdSub},
		"*": Func{Fn: stdMul},
		"/": Func{Fn: stdDiv},
	}}
)

type Env struct {
	Parent *Env
	Map    map[Ident]Expr
}

func (me *Env) Lookup(name Ident) (Expr, bool) {
	found, ok := me.Map[name]
	if (!ok) && (me.Parent != nil) {
		return me.Parent.Lookup(name)
	}
	return found, ok
}

func eval(expr Expr, env *Env) (Expr, error) {
	switch it := expr.(type) {
	case Ident:
		expr, ok := env.Lookup(it)
		if !ok {
			return nil, errors.New("undefined: " + string(it))
		}
		return expr, nil
	case Vec:
		var err error
		vec := Vec{List: make([]Expr, len(it.List))}
		for i, item := range it.List {
			vec.List[i], err = eval(item, env)
			if err != nil {
				return nil, err
			}
		}
		return vec, nil
	case HashMap:
		var err error
		hash_map := HashMap{Map: make(map[Str]Expr, len(it.Map))}
		for key, value := range it.Map {
			hash_map.Map[key], err = eval(value, env)
			if err != nil {
				return nil, err
			}
		}
		return hash_map, nil
	case List:
		var err error
		list := List{List: make([]Expr, len(it.List))}
		for i, item := range it.List {
			list.List[i], err = eval(item, env)
			if err != nil {
				return nil, err
			}
		}
		if len(list.List) > 0 {
			fn, err := mustType[Func](list.List[0])
			if err != nil {
				return nil, errors.New("uncallable: " + fmt.Sprintf("%#v", list.List[0]))
			} else {
				return fn.Fn(list.List[1:])
			}
		}
		return list, nil
	}
	return expr, nil
}
