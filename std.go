package main

import (
	"cmp"
	"fmt"
	"os"
	"strings"
)

var (
	exprTrue          = ExprKeyword(":true")
	exprFalse         = ExprKeyword(":false")
	exprNil           = ExprKeyword(":nil")
	exprDo            = ExprIdent("do")
	exprQuote         = ExprIdent("quote")
	exprQuasiQuote    = ExprIdent("quasiQuote")
	exprUnquote       = ExprIdent("unquote")
	exprSpliceUnquote = ExprIdent("spliceUnquote")
	exprCons          = ExprIdent("cons")
	exprConcat        = ExprIdent("concat")

	envMain = Env{Map: map[ExprIdent]Expr{
		"osArgs":       ExprList{}, // populated by `main` when running a user-specified source file
		"print":        ExprFunc(stdPrint),
		"println":      ExprFunc(stdPrintln),
		"str":          ExprFunc(stdStr),
		"show":         ExprFunc(stdShow),
		"list":         ExprFunc(stdList),
		"is":           ExprFunc(stdIs),
		"isEmpty":      ExprFunc(stdIsEmpty),
		"count":        ExprFunc(stdCount),
		"cmp":          ExprFunc(stdCmp),
		"+":            ExprFunc(stdAdd),
		"-":            ExprFunc(stdSub),
		"*":            ExprFunc(stdMul),
		"/":            ExprFunc(stdDiv),
		"=":            ExprFunc(stdEq),
		"<":            ExprFunc(stdLt),
		">":            ExprFunc(stdGt),
		"<=":           ExprFunc(stdLe),
		">=":           ExprFunc(stdGe),
		"readExpr":     ExprFunc(stdReadExpr),
		"readTextFile": ExprFunc(stdReadTextFile),
		"atomFrom":     ExprFunc(stdAtomFrom),
		"atomGet":      ExprFunc(stdAtomGet),
		"atomSet":      ExprFunc(stdAtomSet),
		"atomSwap":     ExprFunc(stdAtomSwap),
		exprCons:       ExprFunc(stdCons),
		exprConcat:     ExprFunc(stdConcat),
		"vec":          ExprFunc(stdVec),
	}}
	specialForms map[ExprIdent]FnSpecial
)

func init() { // in here, rather than above, to avoid "initialization cycle" error:
	specialForms = map[ExprIdent]FnSpecial{
		"def":          stdDef,
		"set":          stdSet,
		"if":           stdIf,
		"let":          stdLet,
		"fn":           stdFn,
		exprDo:         stdDo,
		exprQuote:      stdQuote,
		exprQuasiQuote: stdQuasiQuote,
	}
	envMain.Map["eval"] = ExprFunc(stdEval)
}

func exprBool(b bool) ExprKeyword {
	if b {
		return exprTrue
	}
	return exprFalse
}

func checkArgsCountExactly(want int, have []Expr) error {
	if len(have) != want {
		return fmt.Errorf("expected %d arg(s), not %d", want, len(have))
	}
	return nil
}

func checkArgsCountAtLeast(want int, have []Expr) error {
	if len(have) < want {
		return fmt.Errorf("expected at least %d arg(s), not %d", want, len(have))
	}
	return nil
}

func checkIs[T Expr](have Expr) (T, error) {
	ret, ok := have.(T)
	if !ok {
		return ret, fmt.Errorf("expected %T, not %T", ret, have)
	}
	return ret, nil
}

func checkAre[T Expr](have ...Expr) error {
	for _, expr := range have {
		if _, err := checkIs[T](expr); err != nil {
			return err
		}
	}
	return nil
}

func checkAreBoth[T1 Expr, T2 Expr](have []Expr, exactArgsCount bool) (ret1 T1, ret2 T2, err error) {
	check_args_count := checkArgsCountExactly
	if !exactArgsCount {
		check_args_count = checkArgsCountAtLeast
	}
	if err = check_args_count(2, have); err != nil {
		return
	}
	if ret1, err = checkIs[T1](have[0]); err != nil {
		return
	}
	if ret2, err = checkIs[T2](have[1]); err != nil {
		return
	}
	return
}

func checkIsSeq(expr Expr) ([]Expr, error) {
	switch expr := expr.(type) {
	case ExprList:
		return ([]Expr)(expr), nil
	case ExprVec:
		return ([]Expr)(expr), nil
	default:
		return nil, fmt.Errorf("expected list or vector, not %T", expr)
	}
}

func stdAdd(args []Expr) (Expr, error) {
	op1, op2, err := checkAreBoth[ExprNum, ExprNum](args, true)
	if err != nil {
		return nil, err
	}
	return op1 + op2, nil
}

func stdSub(args []Expr) (Expr, error) {
	op1, op2, err := checkAreBoth[ExprNum, ExprNum](args, true)
	if err != nil {
		return nil, err
	}
	return op1 - op2, nil
}

func stdMul(args []Expr) (Expr, error) {
	op1, op2, err := checkAreBoth[ExprNum, ExprNum](args, true)
	if err != nil {
		return nil, err
	}
	return op1 * op2, nil
}

func stdDiv(args []Expr) (Expr, error) {
	op1, op2, err := checkAreBoth[ExprNum, ExprNum](args, true)
	if err != nil {
		return nil, err
	}
	return op1 / op2, nil
}

func stdDef(env *Env, args []Expr) (*Env, Expr, error) {
	return defOrSet(true, env, args)
}
func stdSet(env *Env, args []Expr) (*Env, Expr, error) {
	return defOrSet(false, env, args)
}
func defOrSet(isDef bool, env *Env, args []Expr) (*Env, Expr, error) {
	if err := checkArgsCountExactly(2, args); err != nil {
		return nil, nil, err
	}
	name, err := checkIs[ExprIdent](args[0])
	if err != nil {
		return nil, nil, err
	}
	if _, is_reserved := specialForms[name]; is_reserved {
		return nil, nil, fmt.Errorf("cannot redefine `%s`", name)
	}
	if !isDef { // check if this `set` refers to something that's been `def`d
		if _, err := env.get(name); err != nil {
			return nil, nil, err
		}
	} else if env.hasOwn(name) { // cannot re`def` locals
		return nil, nil, fmt.Errorf("cannot redefine `%s` (use `set` here instead of `def`)", name)
	}

	expr, err := evalAndApply(env, args[1])
	if err != nil {
		return nil, nil, err
	}
	env.set(name, expr)
	return nil, expr, nil
}

func stdDo(env *Env, args []Expr) (tailEnv *Env, expr Expr, err error) {
	if err = checkArgsCountAtLeast(1, args); err != nil {
		return
	}
	for _, arg := range args[:len(args)-1] {
		if expr, err = evalAndApply(env, arg); err != nil {
			return
		}
	}
	tailEnv, expr = env, args[len(args)-1]
	return
}

func stdLet(env *Env, args []Expr) (*Env, Expr, error) {
	if err := checkArgsCountAtLeast(2, args); err != nil {
		return nil, nil, err
	}
	bindings, err := checkIsSeq(args[0])
	if err != nil {
		return nil, nil, err
	}
	let_env := newEnv(env, nil, nil)
	for _, binding := range bindings {
		pair, err := checkIsSeq(binding)
		if err != nil {
			return nil, nil, err
		}
		if err = checkArgsCountExactly(2, pair); err != nil {
			return nil, nil, err
		}
		name, err := checkIs[ExprIdent](pair[0])
		if err != nil {
			return nil, nil, err
		}
		if _, is_special := specialForms[name]; is_special {
			return nil, nil, fmt.Errorf("cannot redefine `%s`", name)
		}
		expr, err := evalAndApply(let_env, pair[1])
		if err != nil {
			return nil, nil, err
		}
		let_env.set(name, expr)
	}
	return stdDo(let_env, args[1:])
}

func stdIf(env *Env, args []Expr) (*Env, Expr, error) {
	if err := checkArgsCountExactly(3, args); err != nil {
		return nil, nil, err
	}
	expr, err := evalAndApply(env, args[0])
	if err != nil {
		return nil, nil, err
	}
	idx := 1
	if isEq(expr, exprFalse) || isEq(expr, exprNil) {
		idx = 2
	}
	return env, args[idx], nil
}

func stdFn(env *Env, args []Expr) (*Env, Expr, error) {
	if err := checkArgsCountAtLeast(2, args); err != nil {
		return nil, nil, err
	}
	params, err := checkIsSeq(args[0])
	if err != nil {
		return nil, nil, err
	}
	if err := checkAre[ExprIdent](params...); err != nil {
		return nil, nil, err
	}
	body := args[1]
	if len(args) > 2 {
		body = append(ExprList{exprDo}, args[1:]...)
	}
	var expr Expr = &ExprFn{params: params, body: body, env: env}
	if disableTcoFuncs {
		expr = expr.(*ExprFn).ToFunc()
	}
	return nil, expr, nil
}

func (me *ExprFn) envWith(args []Expr) (*Env, error) {
	if err := checkArgsCountExactly(len(me.params), args); err != nil {
		return nil, err
	}
	return newEnv(me.env, me.params, args), nil
}

func (me *ExprFn) ToFunc() ExprFunc {
	return ExprFunc(func(args []Expr) (Expr, error) {
		env, err := me.envWith(args)
		if err != nil {
			return nil, err
		}
		return evalAndApply(env, me.body)
	})
}

func str(args []Expr, printReadably bool) string {
	var buf strings.Builder
	for i, arg := range args {
		if i > 0 && printReadably {
			buf.WriteByte(' ')
		}
		buf.WriteString(exprToString(arg, printReadably))
	}
	return buf.String()
}

func stdPrint(args []Expr) (Expr, error) {
	os.Stdout.WriteString(str(args, true))
	return exprNil, nil
}
func stdPrintln(args []Expr) (Expr, error) {
	os.Stdout.WriteString(str(args, false) + "\n")
	return exprNil, nil
}
func stdStr(args []Expr) (Expr, error) {
	return ExprStr(str(args, false)), nil
}
func stdShow(args []Expr) (Expr, error) {
	return ExprStr(str(args, true)), nil
}

func stdList(args []Expr) (Expr, error) {
	return ExprList(args), nil
}

func stdIs(args []Expr) (Expr, error) {
	if err := checkArgsCountExactly(2, args); err != nil {
		return nil, err
	}
	kind, err := checkIs[ExprKeyword](args[0])
	if err != nil {
		return nil, err
	}
	var ok bool
	switch kind {
	case ":ident":
		_, ok = args[1].(ExprIdent)
	case ":keyword":
		_, ok = args[1].(ExprKeyword)
	case ":str":
		_, ok = args[1].(ExprStr)
	case ":num":
		_, ok = args[1].(ExprNum)
	case ":list":
		_, ok = args[1].(ExprList)
	case ":vec":
		_, ok = args[1].(ExprVec)
	case ":hashmap":
		_, ok = args[1].(ExprHashMap)
	case ":fn":
		if _, ok = args[1].(*ExprFn); !ok {
			_, ok = args[1].(ExprFunc)
		}
	case ":atom":
		_, ok = args[1].(*ExprAtom)
	default:
		return nil, fmt.Errorf("expected not `%s` but one of: `:list`, `:ident`, `:str`, `:num`, `:vec`, `:hashmap`, `:fn`, `:keyword`, `:atom`", kind)
	}
	return exprBool(ok), nil
}

func stdIsEmpty(args []Expr) (Expr, error) {
	if err := checkArgsCountExactly(1, args); err != nil {
		return nil, err
	}
	list, err := checkIs[ExprList](args[0])
	if err != nil {
		return nil, err
	}
	return exprBool(len(list) == 0), nil
}

func stdCount(args []Expr) (Expr, error) {
	if err := checkArgsCountExactly(1, args); err != nil {
		return nil, err
	}
	list, err := checkIs[ExprList](args[0])
	if err != nil {
		return nil, err
	}
	return ExprNum(len(list)), nil
}

func stdEq(args []Expr) (Expr, error) {
	if err := checkArgsCountExactly(2, args); err != nil {
		return nil, err
	}
	return exprBool(isEq(args[0], args[1])), nil
}

func compare(args []Expr) (int, error) {
	if err := checkArgsCountExactly(2, args); err != nil {
		return 0, err
	}
	switch it := args[0].(type) {
	case ExprNum:
		if other, ok := args[1].(ExprNum); ok {
			return cmp.Compare(it, other), nil
		}
	case ExprStr:
		if other, ok := args[1].(ExprStr); ok {
			return cmp.Compare(it, other), nil
		}
	}
	return 0, fmt.Errorf("specified operands `%#v` and `%#v` are not comparable", args[0], args[1])
}

func stdCmp(args []Expr) (Expr, error) {
	order, err := compare(args)
	if err != nil {
		return nil, err
	}
	return ExprNum(order), nil
}
func stdLt(args []Expr) (Expr, error) {
	order, err := compare(args)
	if err != nil {
		return nil, err
	}
	return exprBool(order == -1), nil
}
func stdLe(args []Expr) (Expr, error) {
	order, err := compare(args)
	if err != nil {
		return nil, err
	}
	return exprBool(order <= 0), nil
}
func stdGt(args []Expr) (Expr, error) {
	order, err := compare(args)
	if err != nil {
		return nil, err
	}
	return exprBool(order == 1), nil
}
func stdGe(args []Expr) (Expr, error) {
	order, err := compare(args)
	if err != nil {
		return nil, err
	}
	return exprBool(order >= 0), nil
}

func stdReadExpr(args []Expr) (Expr, error) {
	if err := checkArgsCountExactly(1, args); err != nil {
		return nil, err
	}
	src, err := checkIs[ExprStr](args[0])
	if err != nil {
		return nil, err
	}
	return readExpr(string(src))
}

func stdReadTextFile(args []Expr) (Expr, error) {
	if err := checkArgsCountExactly(1, args); err != nil {
		return nil, err
	}
	file_path, err := checkIs[ExprStr](args[0])
	if err != nil {
		return nil, err
	}
	file_bytes, err := os.ReadFile(string(file_path))
	if err != nil {
		return nil, err
	}
	return ExprStr(file_bytes), nil
}

func stdEval(args []Expr) (Expr, error) {
	if err := checkArgsCountExactly(1, args); err != nil {
		return nil, err
	}
	return evalAndApply(&envMain, args[0])
}

func stdAtomFrom(args []Expr) (Expr, error) {
	if err := checkArgsCountExactly(1, args); err != nil {
		return nil, err
	}
	return &ExprAtom{Ref: args[0]}, nil
}
func stdAtomGet(args []Expr) (Expr, error) {
	if err := checkArgsCountExactly(1, args); err != nil {
		return nil, err
	}
	atom, err := checkIs[*ExprAtom](args[0])
	if err != nil {
		return nil, err
	}
	return atom.Ref, nil
}
func stdAtomSet(args []Expr) (Expr, error) {
	if err := checkArgsCountExactly(2, args); err != nil {
		return nil, err
	}
	atom, err := checkIs[*ExprAtom](args[0])
	if err != nil {
		return nil, err
	}
	atom.Ref = args[1]
	return atom.Ref, nil
}
func stdAtomSwap(args []Expr) (Expr, error) {
	if err := checkArgsCountAtLeast(2, args); err != nil {
		return nil, err
	}
	atom, fn, err := checkAreBoth[*ExprAtom, *ExprFn](args, false)
	if err != nil {
		return nil, err
	}
	atom.Ref, err = fn.ToFunc()(append([]Expr{atom.Ref}, args[2:]...))
	if err != nil {
		return nil, err
	}
	return atom.Ref, nil
}

func stdQuote(_ *Env, args []Expr) (*Env, Expr, error) {
	if err := checkArgsCountExactly(1, args); err != nil {
		return nil, nil, err
	}
	return nil, args[0], nil
}

func stdCons(args []Expr) (Expr, error) {
	if err := checkArgsCountExactly(2, args); err != nil {
		return nil, err
	}
	list, err := checkIsSeq(args[1])
	if err != nil {
		return nil, err
	}
	return append(ExprList{args[0]}, list...), nil
}

func stdConcat(args []Expr) (Expr, error) {
	var list ExprList
	for _, arg := range args {
		seq, err := checkIsSeq(arg)
		if err != nil {
			return nil, err
		}
		list = append(list, seq...)
	}
	return list, nil
}

func stdQuasiQuote(env *Env, args []Expr) (*Env, Expr, error) {
	is_list_starting_with_ident := func(maybeList Expr, ident ExprIdent, mustHaveLen int) (_ []Expr, _ bool, err error) {
		if list, _ := maybeList.(ExprList); len(list) > 0 {
			if maybe_ident, _ := list[0].(ExprIdent); maybe_ident == ident {
				if err := checkArgsCountExactly(mustHaveLen, list); err == nil {
					return list, true, nil
				}
			}
		}
		return
	}

	if err := checkArgsCountExactly(1, args); err != nil {
		return nil, nil, err
	}
	_, is_vec := args[0].(ExprVec)
	list, not_a_seq := checkIsSeq(args[0])
	if not_a_seq != nil {
		return stdQuote(env, args)
	}

	if unquote, ok, err := is_list_starting_with_ident(args[0], exprUnquote, 2); err != nil {
		return nil, nil, err
	} else if ok {
		if unquoted, err := evalAndApply(env, unquote[1]); err != nil {
			return nil, nil, err
		} else {
			return nil, unquoted, nil
		}
	}

	expr := make(ExprList, 0, len(list))
	for _, item := range list {
		if unquote, ok, err := is_list_starting_with_ident(item, exprUnquote, 2); err != nil {
			return nil, nil, err
		} else if ok {
			if unquoted, err := evalAndApply(env, unquote[1]); err != nil {
				return nil, nil, err
			} else {
				expr = append(expr, unquoted)
			}
		} else if splice_unquote, ok, err := is_list_starting_with_ident(item, exprSpliceUnquote, 2); err != nil {
			return nil, nil, err
		} else if ok {
			evaled, err := evalAndApply(env, splice_unquote[1])
			if err != nil {
				return nil, nil, err
			}
			splicees, err := checkIs[ExprList](evaled)
			if err != nil {
				return nil, nil, err
			}
			for _, splicee := range splicees {
				if unquoted, err := evalAndApply(env, splicee); err != nil {
					return nil, nil, err
				} else {
					expr = append(expr, unquoted)
				}
			}
		} else {
			_, evaled, err := stdQuasiQuote(env, []Expr{item})
			if err != nil {
				return nil, nil, err
			}
			expr = append(expr, evaled)
		}
	}
	if is_vec {
		return nil, (ExprVec)(expr), nil
	}
	return nil, expr, nil
}

func stdVec(args []Expr) (Expr, error) {
	if err := checkArgsCountExactly(1, args); err != nil {
		return nil, err
	}
	if vec, is_vec := args[0].(ExprVec); is_vec {
		return vec, nil
	}
	list, err := checkIs[ExprList](args[0])
	if err != nil {
		return nil, err
	}
	return (ExprVec)(list), nil
}
