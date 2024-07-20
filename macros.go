package main

func macroCallCallee(env *Env, expr Expr) *ExprFn {
	if list, _ := expr.(ExprList); len(list) > 0 {
		if ident, _ := list[0].(ExprIdent); ident != "" {
			if maybe_fn := env.find(ident); maybe_fn != nil {
				if fn, _ := maybe_fn.(*ExprFn); (fn != nil) && fn.isMacro {
					return fn
				}
			}
		}
	}
	return nil
}

func macroExpand(env *Env, expr Expr, onOneExpansion func(Expr)) (Expr, error) {
	for callee := macroCallCallee(env, expr); callee != nil; callee = macroCallCallee(env, expr) {
		it, err := callee.Call(expr.(ExprList)[1:])
		if err != nil {
			return nil, err
		}
		expr = it
		if onOneExpansion != nil { // usually `nil`, but dev tooling could use it to collect and show multiple expansions as they unfold one by one
			onOneExpansion(expr)
		}
	}
	return expr, nil
}
