package main

import (
	"fmt"
	"strconv"
	"strings"
)

func exprToString(expr Expr, srcLike bool) string {
	print_list := func(lst []Expr, pr bool, opening string, closing string, sep string) string {
		ret := make([]string, 0, len(lst))
		for _, e := range lst {
			ret = append(ret, exprToString(e, pr))
		}
		return opening + strings.Join(ret, sep) + closing
	}

	switch it := expr.(type) {
	case ExprList:
		return print_list(it, srcLike, "(", ")", " ")
	case ExprVec:
		return print_list(it, srcLike, "[", "]", " ")
	case ExprHashMap:
		str_list := make([]string, 0, len(it)*2)
		for k, v := range it {
			str_list = append(str_list, exprToString(k, srcLike), exprToString(v, srcLike))
		}
		return "{" + strings.Join(str_list, " ") + "}"
	case ExprIdent:
		return string(it)
	case ExprKeyword:
		return string(it)
	case ExprStr:
		if srcLike {
			return strconv.Quote(string(it))
		} else {
			return string(it)
		}
	case ExprNum:
		return strconv.Itoa(int(it))
	case ExprAtom:
		if srcLike {
			return fmt.Sprintf("(atom %s)", exprToString(it.Ref, true))
		}
		return exprToString(it.Ref, false)
	default:
		return fmt.Sprintf("%#v", it)
	}
}
