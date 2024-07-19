package main

import (
	"errors"
	"regexp"
	"strconv"
)

type Reader interface {
	next() *string
	peek() *string
}

type TokenReader struct {
	tokens   []string
	position int
}

func (me *TokenReader) next() *string {
	if me.position >= len(me.tokens) {
		return nil
	}
	token := me.tokens[me.position]
	me.position = me.position + 1
	return &token
}

func (me *TokenReader) peek() *string {
	if me.position >= len(me.tokens) {
		return nil
	}
	return &me.tokens[me.position]
}

func tokenize(src string) []string {
	results := make([]string, 0, 1)
	// Work around lack of quoting in backtick
	regex := regexp.MustCompile(`[\s,]*(~@|[\[\]{}()'` + "`" + `~^@]|"(?:\\.|[^\\"])*"?|;.*|[^\s\[\]{}('"` + "`" + `,;)]*)`)
	for _, group := range regex.FindAllStringSubmatch(src, -1) {
		if (group[1] == "") || (group[1][0] == ';') {
			continue
		}
		results = append(results, group[1])
	}
	return results
}

func readAtom(r Reader) (Expr, error) {
	token := r.next()
	if token == nil {
		return nil, errors.New("readAtom underflow")
	}
	tok := *token
	if match, _ := regexp.MatchString(`^-?[0-9]+$`, tok); match {
		num, err := strconv.Atoi(tok)
		return Num(num), err
	} else if match, _ := regexp.MatchString(`^"(?:\\.|[^\\"])*"$`, tok); match {
		str, err := strconv.Unquote(tok)
		return Str(str), err
	} else if (tok)[0] == '"' {
		return nil, errors.New("expected '\"', got EOF")
	} else if (tok)[0] == ':' {
		return Keyword(tok), nil
	} else {
		return Ident(tok), nil
	}
}

func readList(r Reader, start string, end string) (Expr, error) {
	token := r.next()
	if token == nil {
		return nil, errors.New("readList underflow")
	} else if *token != start {
		return nil, errors.New("expected '" + start + "'")
	}

	var ast_list []Expr
	token = r.peek()
	for ; true; token = r.peek() {
		if token == nil {
			return nil, errors.New("exepected '" + end + "', got EOF")
		}
		if *token == end {
			break
		}
		form, err := readForm(r)
		if err != nil {
			return nil, err
		}
		ast_list = append(ast_list, form)
	}
	r.next()
	return List{List: ast_list}, nil
}

func readVec(r Reader) (Expr, error) {
	list, err := readList(r, "[", "]")
	if err != nil {
		return nil, err
	}
	vec := Vec{List: list.(List).List}
	return vec, nil
}

func readHashMap(r Reader) (Expr, error) {
	list, err := readList(r, "{", "}")
	if err != nil {
		return nil, err
	}
	return newHashMap(list)
}

func readForm(r Reader) (Expr, error) {
	token := r.peek()
	if token == nil {
		return nil, errors.New("readForm underflow")
	}
	switch *token {
	case `'`:
		r.next()
		form, err := readForm(r)
		if err != nil {
			return nil, err
		}
		return List{List: []Expr{Ident("quote"), form}}, nil
	case "`":
		r.next()
		form, e := readForm(r)
		if e != nil {
			return nil, e
		}
		return List{List: []Expr{Ident("quasiquote"), form}}, nil
	case `~`:
		r.next()
		form, e := readForm(r)
		if e != nil {
			return nil, e
		}
		return List{[]Expr{Ident("unquote"), form}, nil}, nil
	case `~@`:
		r.next()
		form, e := readForm(r)
		if e != nil {
			return nil, e
		}
		return List{[]Expr{Ident("splice-unquote"), form}, nil}, nil
	case `^`:
		r.next()
		meta, e := readForm(r)
		if e != nil {
			return nil, e
		}
		form, e := readForm(r)
		if e != nil {
			return nil, e
		}
		return List{[]Expr{Ident("with-meta"), form, meta}, nil}, nil
	case `@`:
		r.next()
		form, e := readForm(r)
		if e != nil {
			return nil, e
		}
		return List{[]Expr{Ident("deref"), form}, nil}, nil

	// list
	case ")":
		return nil, errors.New("unexpected ')'")
	case "(":
		return readList(r, "(", ")")

	// vector
	case "]":
		return nil, errors.New("unexpected ']'")
	case "[":
		return readVec(r)

	// hash-map
	case "}":
		return nil, errors.New("unexpected '}'")
	case "{":
		return readHashMap(r)
	default:
		return readAtom(r)
	}
}

func Read_str(str string) (Expr, error) {
	var tokens = tokenize(str)
	if len(tokens) == 0 {
		return nil, errors.New("<empty line>")
	}

	return readForm(&TokenReader{tokens: tokens, position: 0})
}
