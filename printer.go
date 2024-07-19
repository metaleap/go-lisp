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
	case List:
		return print_list(tobj.List, printReadably, "(", ")", " ")
	case Vec:
		return print_list(tobj.List, printReadably, "[", "]", " ")
	case HashMap:
		str_list := make([]string, 0, len(tobj.Map)*2)
		for k, v := range tobj.Map {
			str_list = append(str_list, printExpr(k, printReadably))
			str_list = append(str_list, printExpr(v, printReadably))
		}
		return "{" + strings.Join(str_list, " ") + "}"
	case Keyword:
		return string(tobj)
	case Str:
		if printReadably {
			return strconv.Quote(string(tobj))
		} else {
			return string(tobj)
		}
	case Ident:
		return string(tobj)
	case Func:
		return fmt.Sprintf("<function %#v>", tobj)
	default:
		return fmt.Sprintf("UNKNOWN:%#v", tobj)
	}
}
