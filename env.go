package main

import "errors"

var (
	envUnEvals = newEnv(nil)
	envMain    = newEnv(envUnEvals)
)

type Env struct {
	Parent *Env
	Map    map[ExprIdent]Expr
}

func newEnv(parent *Env, bindsExprs ...Expr) *Env {
	ret := Env{Parent: parent, Map: make(map[ExprIdent]Expr, len(bindsExprs)/2)}
	for i := 1; i < len(bindsExprs); i += 2 {
		ret.Map[bindsExprs[i-1].(ExprIdent)] = bindsExprs[i]
	}
	return &ret
}

func (me *Env) hasOwn(name ExprIdent) (ret bool) {
	_, ret = me.Map[name]
	return
}

func (me *Env) set(name ExprIdent, value Expr) {
	me.Map[name] = value
}

func (me *Env) find(name ExprIdent) Expr {
	found, ok := me.Map[name]
	if (!ok) && (me.Parent != nil) {
		return me.Parent.find(name)
	}
	return found
}

func (me *Env) get(name ExprIdent) (Expr, error) {
	expr := me.find(name)
	if expr == nil {
		return nil, errors.New("undefined: " + string(name))
	}
	return expr, nil
}
