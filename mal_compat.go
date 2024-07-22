package main

import (
	"fmt"
	"os"
)

var malCompat = (os.Getenv("MAL_COMPAT") != "")

func makeCompatibleWithMAL() {
	// simple aliases: special-forms
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

	// simple aliases: std funcs
	for mals, ours := range map[ExprIdent]ExprIdent{
		"read-string": "readExpr",
		"readline":    "readLine",
		"atom":        "atomFrom",
		"deref":       "atomGet",
		"reset!":      "atomSet",
		"swap!":       "atomSwap",
		"empty?":      "isEmpty",
		"get":         "hashmapGet",
	} {
		it := envMain.Map[ours]
		if envMain.Map[mals] = it; it == nil {
			panic("mixed sth up huh?")
		}
	}

	// non-alias-able funcs
	for name, expr := range map[ExprIdent]Expr{
		"pr-str": ExprFunc(func(args []Expr) (Expr, error) { return ExprStr(str(true, args...)), nil }),
	} {
		envMain.Map[name] = expr
	}

	// non-alias-able special-forms
	for name, sf := range map[ExprIdent]SpecialForm{

		"let*": SpecialForm(func(env *Env, args []Expr) (*Env, Expr, error) {
			bindings, err := checkIsSeq(args[0])
			if err != nil {
				return nil, nil, err
			}
			rewritten := make([]Expr, 0, len(bindings)/2)
			for i := 1; i < len(bindings); i += 2 {
				rewritten = append(rewritten, ExprList{bindings[i-1], bindings[i]})
			}
			args[0] = (ExprList)(rewritten)
			return stdLet(env, args)
		}),

		"cond": SpecialForm(func(env *Env, args []Expr) (*Env, Expr, error) { // dont have that as a macro due to github.com/kanaka/mal/issues/655
			if len(args) == 0 {
				return nil, exprNil, nil
			}
			if (len(args) % 2) != 0 {
				return nil, nil, fmt.Errorf("expected even number of args to `cond`, not %d", len(args))
			}
			the_bool, the_then, the_rest := args[0], args[1], args[2:]
			call_form := ExprList{ExprIdent("if"), the_bool, the_then, append(ExprList{ExprIdent("cond")}, the_rest...)}
			return env, call_form, nil
		}),
	} {
		specialForms[name] = sf
	}

	// the below helper defs
	if _, err := readAndEval("(" + string(exprIdentDo) + " " + srcMiniStdlibMalCompat + "\n" + string(exprNil) + ")"); err != nil {
		panic(err)
	}
}

const srcMiniStdlibMalCompat = `


(def checker
	(fn (tag)
		(fn (arg) (is tag arg))))

(def list? (checker :list))
(def symbol? (checker :ident))
(def vector? (checker :vec))
(def map? (checker :hashmap))




`
