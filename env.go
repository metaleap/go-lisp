package main

import "errors"

var (
	envUnEvals = Env{Map: map[ExprIdent]Expr{}}
	envMain    = Env{Parent: &envUnEvals, Map: map[ExprIdent]Expr{}}
)

func init() {
	for k, v := range map[ExprIdent]Expr{
		"def!": ExprFunc(stdDef),
		"set!": ExprFunc(stdSet),
	} {
		envUnEvals.Map[k] = v
	}
	for k, v := range map[ExprIdent]Expr{
		"+": ExprFunc(stdAdd),
		"-": ExprFunc(stdSub),
		"*": ExprFunc(stdMul),
		"/": ExprFunc(stdDiv),
	} {
		envMain.Map[k] = v
	}
}

type Env struct {
	Parent *Env
	Map    map[ExprIdent]Expr
}

func (me *Env) hasOwn(name ExprIdent) (ret bool) {
	_, ret = me.Map[name]
	return
}

func (me *Env) Set(name ExprIdent, value Expr) {
	me.Map[name] = value
}

func (me *Env) Find(name ExprIdent) Expr {
	found, ok := me.Map[name]
	if (!ok) && (me.Parent != nil) {
		return me.Parent.Find(name)
	}
	return found
}

func (me *Env) Get(name ExprIdent) (Expr, error) {
	expr := me.Find(name)
	if expr == nil {
		return nil, errors.New("undefined: " + string(name))
	}
	return expr, nil
}
