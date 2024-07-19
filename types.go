package main

import (
	"errors"
	"reflect"
)

// General types
type Expr interface {
	isExpr()
}

func (ExprIdent) isExpr()   {}
func (ExprKeyword) isExpr() {}
func (ExprStr) isExpr()     {}
func (ExprNum) isExpr()     {}
func (ExprFunc) isExpr()    {}
func (ExprList) isExpr()    {}
func (ExprVec) isExpr()     {}
func (ExprHashMap) isExpr() {}

// Scalars

type ExprIdent string

func isIdent(expr Expr) bool {
	_, ok := expr.(ExprIdent)
	return ok
}

func (me ExprIdent) isNil() bool {
	return string(me) == "nil"
}
func (me ExprIdent) isTrue() bool {
	return string(me) == "true"
}
func (me ExprIdent) isFalse() bool {
	return string(me) == "false"
}

type ExprNum int

func isNum(expr Expr) bool {
	_, ok := expr.(ExprNum)
	return ok
}

type ExprKeyword string

func isKeyword(expr Expr) bool {
	_, ok := expr.(ExprKeyword)
	return ok
}

type ExprStr string

func isString(expr Expr) bool {
	_, ok := expr.(ExprStr)
	return ok
}

// Functions
type ExprFunc func([]Expr) (Expr, error)

func isFunc(expr Expr) bool {
	_, ok := expr.(ExprFunc)
	return ok
}

// Lists
type ExprList []Expr

func newList(exprs ...Expr) Expr {
	return ExprList(exprs)
}

func isList(expr Expr) bool {
	_, ok := expr.(ExprList)
	return ok
}

// Vectors
type ExprVec []Expr

func isVec(expr Expr) bool {
	_, ok := expr.(ExprVec)
	return ok
}

func getSlice(expr Expr) ([]Expr, error) {
	switch expr := expr.(type) {
	case ExprList:
		return ([]Expr)(expr), nil
	case ExprVec:
		return ([]Expr)(expr), nil
	default:
		return nil, errors.New("getSlice called on non-sequence")
	}
}

// Hash Maps
type ExprHashMap map[ExprStr]Expr

func newHashMap(seq Expr) (Expr, error) {
	list, err := getSlice(seq)
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

func isHashMap(expr Expr) bool {
	_, ok := expr.(ExprHashMap)
	return ok
}

// General functions

func isListOrVec(seq Expr) bool {
	if seq == nil {
		return false
	}
	return (reflect.TypeOf(seq) == reflect.TypeOf(ExprList{})) ||
		(reflect.TypeOf(seq) == reflect.TypeOf(ExprVec{}))
}

func isEq(a Expr, b Expr) bool {
	ota, otb := reflect.TypeOf(a), reflect.TypeOf(b)
	if (ota != otb) && ((!isListOrVec(a)) || !isListOrVec(b)) {
		return false
	}
	switch a.(type) {
	case ExprVec, ExprList:
		sa, _ := getSlice(a)
		sb, _ := getSlice(b)
		if len(sa) != len(sb) {
			return false
		}
		for i := 0; i < len(sa); i += 1 {
			if !isEq(sa[i], sb[i]) {
				return false
			}
		}
		return true
	case ExprHashMap:
		ma, mb := a.(ExprHashMap), b.(ExprHashMap)
		if len(ma) != len(mb) {
			return false
		}
		for k, v := range ma {
			if !isEq(v, mb[k]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}
