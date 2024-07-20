package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

func exprToString(expr Expr, srcLike bool) string {
	var buf strings.Builder
	exprWriteTo(&buf, expr, srcLike)
	return buf.String()
}

type Writer interface {
	io.StringWriter
	io.ByteWriter
}

func exprWriteTo(w Writer, expr Expr, srcLike bool) {
	print_list := func(lst []Expr, opening byte, closing byte) {
		w.WriteByte(opening)
		for i, it := range lst {
			if i > 0 {
				w.WriteByte(' ')
			}
			exprWriteTo(w, it, srcLike)
		}
		w.WriteByte(closing)
	}

	switch it := expr.(type) {
	case ExprList:
		print_list(it, '(', ')')
	case ExprVec:
		print_list(it, '[', ']')
	case ExprHashMap:
		w.WriteByte('{')
		for k, v := range it {
			w.WriteByte(' ')
			w.WriteString(string(k))
			w.WriteByte(' ')
			exprWriteTo(w, v, srcLike)
		}
		w.WriteString(" }")
	case ExprIdent:
		w.WriteString(string(it))
	case ExprKeyword:
		w.WriteString(string(it))
	case ExprStr:
		if srcLike {
			w.WriteString(strconv.Quote(string(it)))
		} else {
			w.WriteString(string(it))
		}
	case ExprNum:
		w.WriteString(strconv.Itoa(int(it)))
	case *ExprAtom:
		if srcLike {
			w.WriteString("(atom ")
			exprWriteTo(w, it.Ref, true)
			w.WriteByte(')')
		}
		exprWriteTo(w, it.Ref, false)
	default:
		w.WriteString(fmt.Sprintf("%#v", it))
	}
}
