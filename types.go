package main

import (
	"errors"
	"reflect"
)

type Expr interface {
	isExpr()
}

type ExprIdent string
type ExprKeyword string
type ExprStr string
type ExprNum int
type ExprFunc func(*Env, []Expr) (Expr, error)
type ExprList []Expr
type ExprVec []Expr
type ExprHashMap map[ExprStr]Expr

func (ExprIdent) isExpr()   {}
func (ExprKeyword) isExpr() {}
func (ExprStr) isExpr()     {}
func (ExprNum) isExpr()     {}
func (ExprFunc) isExpr()    {}
func (ExprList) isExpr()    {}
func (ExprVec) isExpr()     {}
func (ExprHashMap) isExpr() {}

func newHashMap(seq Expr) (Expr, error) {
	list, err := mustSeq(seq)
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

// General functions

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
		sl1, _ := mustSeq(arg1)
		sl2, _ := mustSeq(arg2)
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
