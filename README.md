# Let's Write an Interpreter

## Idea

Book on writing scheme like interpreter in Go.

### Learning Goals
Always compare our lisp to Go and talk on design decisions

- program structure (syntax)
- basic types in the language
    - typing
- basic operations (+, - ...)
    - operator precedence
- functions
- eager/lazy evaluation (if/or/and vs function call)
- scope & closure
- type composition (cons, car, cdr ...)
- syntactic sugar (defn, defstruct ...)
- error handling
- concurrency
- byte/machine code
- OO ?

## Resources
- [lis.py](http://norvig.com/lispy.html)
- [mal](https://github.com/kanaka/mal)
- [Bytecode compilers and interpreters](https://bernsteinbear.com/blog/bytecode-interpreters/)


## Chapters

- Introduction to Go
- Our toy language (scheme-ish)
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
