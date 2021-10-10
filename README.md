# Let's Write an Interpreter

## Idea

Book on writing scheme/clojure like interpreter in Go.

### Learning Goals
Always compare our lisp to Go and talk on design decisions

- Overview - How programs are written 
- A look at Scheme/Clojure
    - syntax
    - types
- Lexing & Parsing
- Evaluation of simple math expressions
    - operators
    - precedence
- REPL
    - 
- Variables & environments
    - lisp1 vs lisp2
    - what do we expose to the user (types)
- Logic (booleans, if, or ...)
    - what is true/false?
    - lazy evaluation
- Functions & callables
    - scope
    - closure
- Adding strings
    - strings & bytes
    - string syntax (quote, raw ...)
    - type system discussion
- Error handling
    - types of error handling
    - exceptions
    - multiple return values
- Optimizations
    - Types of optimization
    - Example: constant folding
- Generating bytecode & creating a VM
    - web assembly? llvm?
- A standard library & package managers
    - library design
    - lisp vs go in stdlib
    - package management problem
    - writing a package manager
- Tooling
    - debugger
    - profiler

### Optional Subjects

- type composition (cons, car, cdr ...)
- syntactic sugar (defn, defstruct ...)
- error handling
- concurrency
- OO
- macros

## Resources
- [lis.py](http://norvig.com/lispy.html)
- [mal](https://github.com/kanaka/mal)
- [Bytecode compilers and interpreters](https://bernsteinbear.com/blog/bytecode-interpreters/)


## Chapters

- Introduction to Go
- Our toy language (scheme-ish/clojure)
- Interpreter overview
- Lexer
    - Strings in Go
    - slices
    - struct & methods
	- NewLexer
	- pointer receivers
    - commenting methods
    - Test
- Parser
- Eval
- Parallel eval


## License
[Attribution 4.0 International (CC BY 4.0)](https://creativecommons.org/licenses/by/4.0/)

Copyright &copy; 353solutions & Miki Tebeka
