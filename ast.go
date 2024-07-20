package main

import (
	"cmp"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type Expr interface {
	isExpr()
}

func (ExprIdent) isExpr()   {}
func (ExprKeyword) isExpr() {}
func (ExprStr) isExpr()     {}
func (ExprNum) isExpr()     {}
func (ExprList) isExpr()    {}
func (ExprVec) isExpr()     {}
func (ExprHashMap) isExpr() {}
func (*ExprAtom) isExpr()   {}
func (ExprFunc) isExpr()    {}
func (*ExprFn) isExpr()     {}

type ExprIdent string
type ExprKeyword string
type ExprStr string
type ExprNum int
type ExprList []Expr
type ExprVec []Expr
type ExprHashMap map[ExprStr]Expr
type ExprAtom struct{ Ref Expr }
type ExprFunc func([]Expr) (Expr, error)
type ExprFn struct { // if it weren't for TCO, just the above `ExprFunc` would suffice.
	params []Expr // all are guaranteed to be `ExprIdent` before constructing an `ExprFn`
	body   Expr
	env    *Env
}

func exprBool(b bool) ExprKeyword {
	if b {
		return exprTrue
	}
	return exprFalse
}

func compare(args []Expr) (int, error) {
	if err := checkArgsCountExactly(2, args); err != nil {
		return 0, err
	}
	switch it := args[0].(type) {
	case ExprNum:
		if other, ok := args[1].(ExprNum); ok {
			return cmp.Compare(it, other), nil
		}
	case ExprStr:
		if other, ok := args[1].(ExprStr); ok {
			return cmp.Compare(it, other), nil
		}
	}
	return 0, fmt.Errorf("specified operands `%#v` and `%#v` are not comparable", args[0], args[1])
}

func isListOrVec(seq Expr) bool {
	ty := reflect.TypeOf(seq)
	return (ty == reflect.TypeFor[ExprList]()) || (ty == reflect.TypeFor[ExprVec]())
}

func isEq(arg1 Expr, arg2 Expr) bool {
	ty1, ty2 := reflect.TypeOf(arg1), reflect.TypeOf(arg2)
	if (ty1 != ty2) && ((!isListOrVec(arg1)) || !isListOrVec(arg2)) {
		return false
	}
	switch arg1.(type) {
	case ExprVec, ExprList:
		sl1, _ := checkIsSeq(arg1)
		sl2, _ := checkIsSeq(arg2)
		if len(sl1) != len(sl2) {
			return false
		}
		for i := 0; i < len(sl1); i += 1 {
			if !isEq(sl1[i], sl2[i]) {
				return false
			}
		}
		return true
	case ExprHashMap:
		hm1, hm2 := arg1.(ExprHashMap), arg2.(ExprHashMap)
		if len(hm1) != len(hm2) {
			return false
		}
		for k, v := range hm1 {
			if !isEq(v, hm2[k]) {
				return false
			}
		}
		return true
	default:
		return arg1 == arg2
	}
}

func newHashMap(seq Expr) (Expr, error) {
	list, err := checkIsSeq(seq)
	if err != nil {
		return nil, err
	}
	if (len(list) % 2) != 0 {
		return nil, errors.New("odd number of arguments to NewHashMap")
	}
	hash_map := ExprHashMap{}
	for i := 1; i < len(list); i += 2 {
		str, ok := list[i-1].(ExprStr)
		if !ok {
			return nil, errors.New("expected hash-map key string")
		}
		hash_map[str] = list[i]
	}
	return hash_map, nil
}

func str(args []Expr, printReadably bool) string {
	var buf strings.Builder
	for i, arg := range args {
		if i > 0 && printReadably {
			buf.WriteByte(' ')
		}
		exprWriteTo(&buf, arg, printReadably)
	}
	return buf.String()
}