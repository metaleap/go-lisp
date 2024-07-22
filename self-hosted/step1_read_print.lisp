;; read
(def READ readExpr)


;; eval
(def EVAL (fn [ast]
  ast))

;; print
(def PRINT str)

;; repl
(def rep (fn [strng]
  (PRINT (EVAL (READ strng)))))

;; repl loop
(def repl-loop (fn [line]
  (if line
    (do
      (if (not (= "" line))
        (try
          (println (rep line))
          (catch exc
            (println "Uncaught exception:" exc))))
      (repl-loop (readLine "mal-user> "))))))

;; main
(repl-loop "")
