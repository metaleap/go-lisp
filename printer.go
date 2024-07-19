package main

import (
	"fmt"
	"strings"
)

func Pr_list(lst []MalType, pr bool, start string, end string, join string) string {
	str_list := make([]string, 0, len(lst))
	for _, e := range lst {
		str_list = append(str_list, Pr_str(e, pr))
	}
	return start + strings.Join(str_list, join) + end
}

func Pr_str(obj MalType, printReadably bool) string {
	switch tobj := obj.(type) {
	case List:
		return Pr_list(tobj.Val, printReadably, "(", ")", " ")
	case Vector:
		return Pr_list(tobj.Val, printReadably, "[", "]", " ")
	case HashMap:
		str_list := make([]string, 0, len(tobj.Val)*2)
		for k, v := range tobj.Val {
			str_list = append(str_list, Pr_str(k, printReadably))
			str_list = append(str_list, Pr_str(v, printReadably))
		}
		return "{" + strings.Join(str_list, " ") + "}"
	case string:
		if strings.HasPrefix(tobj, "\u029e") {
			return ":" + tobj[2:]
		} else if printReadably {
			return `"` + strings.ReplaceAll(
				strings.ReplaceAll(
					strings.ReplaceAll(tobj, `\`, `\\`),
					`"`, `\"`),
				"\n", `\n`) + `"`
		} else {
			return tobj
		}
	case Symbol:
		return tobj.Val
	case nil:
		return "nil"
	case MalFunc:
		return "(fn* " +
			Pr_str(tobj.Params, true) + " " +
			Pr_str(tobj.Exp, true) + ")"
	case func([]MalType) (MalType, error):
		return fmt.Sprintf("<function %v>", obj)
	case *Atom:
		return "(atom " +
			Pr_str(tobj.Val, true) + ")"
	default:
		return fmt.Sprintf("%v", obj)
	}
}
