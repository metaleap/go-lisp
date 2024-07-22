; run with: `rlwrap go run *.go hello.lisp` or `rlwrap go run *.go hello.lisp World`

(def greet
    (fn (name)
        (def repeat (not (bool name)))
        (set name (or name (readLine "Name: ")))
        (if (= "" name)
            (set repeat :false)
            (println "Hello, " name "!"))
        (if repeat (greet :nil) :nil)))

(if (isEmpty osArgs)
    (greet :nil)
    (map greet osArgs))
