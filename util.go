package main

import (
	"io"
)

func isNilOrFalse(expr Expr) bool {
	return isEq(exprNil, expr) || isEq(exprFalse, expr)
}

func readUntil(r io.Reader, until byte, initialBufCapacity int) (string, error) {
	buf := make([]byte, 0, initialBufCapacity)
	var b [1]byte
	for {
		_, err := r.Read(b[0:1])
		if err != nil {
			return "", err
		} else if b[0] == until {
			break
		} else {
			buf = append(buf, b[0])
		}
	}
	line := string(buf)
	return line, nil
}

func makeCompatibleWithMAL() {
	for mals, ours := range map[ExprIdent]ExprIdent{
		"def!": "def",
		"fn*":  "fn",
		"try*": "try",
	} {
		it := specialForms[ours]
		if specialForms[mals] = it; it == nil {
			panic("mixed sth up huh?")
		}
	}

	for mals, ours := range map[ExprIdent]ExprIdent{
		"pr-str":      "str",
		"read-string": "readExpr",
		"readline":    "readLine",
	} {
		it := envMain.Map[ours]
		if envMain.Map[mals] = it; it == nil {
			panic("mixed sth up huh?")
		}
	}
}
