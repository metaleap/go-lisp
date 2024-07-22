;; read
(def READ (fn [strng]
  strng))

;; eval
(def EVAL (fn [ast]
  ast))

;; print
(def PRINT (fn [exp] exp))

;; repl
(def rep (fn [strng]
  (PRINT (EVAL (READ strng)))))

;; repl loop
(def repl-loop (fn [line]
  (if line
    (do
      (if (not (= "" line))
        (println (rep line)))
      (repl-loop (readLine "mal-user> "))))))

;; main
(repl-loop "")
