package main

import (
	"fmt"
	"strconv"
	"strings"
)

func printExpr(obj Expr, printReadably bool) string {
	print_list := func(lst []Expr, pr bool, opening string, closing string, sep string) string {
		ret := make([]string, 0, len(lst))
		for _, e := range lst {
			ret = append(ret, printExpr(e, pr))
		}
		return opening + strings.Join(ret, sep) + closing
	}

	switch tobj := obj.(type) {
	case ExprList:
		return print_list(tobj, printReadably, "(", ")", " ")
	case ExprVec:
		return print_list(tobj, printReadably, "[", "]", " ")
	case ExprHashMap:
		str_list := make([]string, 0, len(tobj)*2)
		for k, v := range tobj {
			str_list = append(str_list, printExpr(k, printReadably), printExpr(v, printReadably))
		}
		return "{" + strings.Join(str_list, " ") + "}"
	case ExprKeyword:
		return string(tobj)
	case ExprStr:
		if printReadably {
			return strconv.Quote(string(tobj))
		} else {
			return string(tobj)
		}
	case ExprIdent:
		return string(tobj)
	case ExprFunc:
		return fmt.Sprintf("<function %#v>", tobj)
	case *ExprFn:
		return fmt.Sprintf("<function %#v>", tobj)
	case ExprNum:
		return strconv.Itoa(int(tobj))
	default:
		return fmt.Sprintf("UNKNOWN %T %#v", tobj, tobj)
	}
}
