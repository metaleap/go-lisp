package main

import (
	"errors"
	"reflect"
)

// General types
type Expr interface {
	isExpr()
}

func (Ident) isExpr()   {}
func (Keyword) isExpr() {}
func (Str) isExpr()     {}
func (Num) isExpr()     {}
func (Func) isExpr()    {}
func (List) isExpr()    {}
func (Vec) isExpr()     {}
func (HashMap) isExpr() {}

// Scalars

type Ident string

func isIdent(obj Expr) bool {
	_, ok := obj.(Ident)
	return ok
}

func (me Ident) isNil() bool {
	return string(me) == "nil"
}
func (me Ident) isTrue() bool {
	return string(me) == "true"
}
func (me Ident) isFalse() bool {
	return string(me) == "false"
}

type Num int

func isNum(obj Expr) bool {
	_, ok := obj.(Num)
	return ok
}

type Keyword string

func isKeyword(obj Expr) bool {
	_, ok := obj.(Keyword)
	return ok
}

type Str string

func isString(obj Expr) bool {
	_, ok := obj.(Str)
	return ok
}

// Functions
type Func struct {
	Fn   func([]Expr) (Expr, error)
	Meta Expr
}

func isFunc(obj Expr) bool {
	_, ok := obj.(Func)
	return ok
}

// Lists
type List struct {
	List []Expr
	Meta Expr
}

func newList(a ...Expr) Expr {
	return List{a, nil}
}

func isList(obj Expr) bool {
	_, ok := obj.(List)
	return ok
}

// Vectors
type Vec struct {
	List []Expr
	Meta Expr
}

func isVec(obj Expr) bool {
	_, ok := obj.(Vec)
	return ok
}

func getSlice(seq Expr) ([]Expr, error) {
	switch obj := seq.(type) {
	case List:
		return obj.List, nil
	case Vec:
		return obj.List, nil
	default:
		return nil, errors.New("GetSlice called on non-sequence")
	}
}

// Hash Maps
type HashMap struct {
	Map  map[Str]Expr
	Meta Expr
}

func newHashMap(seq Expr) (Expr, error) {
	list, err := getSlice(seq)
	if err != nil {
		return nil, err
	}
	if (len(list) % 2) != 0 {
		return nil, errors.New("odd number of arguments to NewHashMap")
	}
	hash_map := map[Str]Expr{}
	for i := 1; i < len(list); i += 2 {
		str, ok := list[i-1].(Str)
		if !ok {
			return nil, errors.New("expected hash-map key string")
		}
		hash_map[str] = list[i]
	}
	return HashMap{Map: hash_map}, nil
}

func isHashMap(obj Expr) bool {
	_, ok := obj.(HashMap)
	return ok
}

// General functions

func isListOrVec(seq Expr) bool {
	if seq == nil {
		return false
	}
	return (reflect.TypeOf(seq) == reflect.TypeOf(List{})) ||
		(reflect.TypeOf(seq) == reflect.TypeOf(Vec{}))
}

func isEq(a Expr, b Expr) bool {
	ota, otb := reflect.TypeOf(a), reflect.TypeOf(b)
	if (ota != otb) && ((!isListOrVec(a)) || !isListOrVec(b)) {
		return false
	}
	switch a.(type) {
	case Vec, List:
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
	case HashMap:
		ma, mb := a.(HashMap).Map, b.(HashMap).Map
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
