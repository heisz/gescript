# gescript - An Embedded ECMA Scripting Engine for Go

A self-contained, 'lightweight' and embeddable script interpreter that is based
on the ECMA-262 specification.  It does *not* attempt to be a full reference
implementation of the specification - it is not intended to be a replacement
for node or goja or similar.  Instead, it selectively provides support for the
most common elements of ECMAScript to support a Go application providing a
level of configurable/extensible business logic through a script in a language
that is familiar to many.

## History

The original form of this engine was written in C (in the early 2000's), used
Yacc to parse and 'compile' the script in the style still used today - an array
of opcode or native function references executing against a variable stack.  It
was originally written as the 'scripting' element of the distributed Mousetrap
engine and has made other appearances in products over the years.

The early versions of the scripting library were extremely simple, only var
support, no hoisting, no first-class functions, etc. etc. etc.  In part this
was because ECMA script was evolving and in part because the intent as stated
was providing basic script support, not a full JavaScript implementation.

In more recent times, Go has become a system of choice for multi-platform
server application implementations (Mousetrap has moved to it).  And ECMAScript
has introduced a lot of elements that would be useful in this kind of script
environment.  So a complete porting of the original engine to Go was undertaken,
with mixed success.  The Yacc model was tried and abandoned in favour of the
Pratt-based parser (which was basically borrowed from that other project).  The
'array-of-functions' model was kept but somewhat mutated with the types system
to take advantage of Go features.  And as part of it, a much more complete
implementation of ECMAScript elements (either skipped originally or added to
the specification later) were included.

## Features (High Level)

- **Pure Go** - no CGO dependencies, completely cross-platform
- **Embeddable** - specifically intended to be used by Go applications for
                   business logic scripting
- **Extensible** - supports external registration of 'native' Go elements for
                   use in scripts
- **Expressions** - supports standard expression elements and operators,
                    including spread and rest operators/declarations
- **Statements** - supports most of the standard statement forms
- **Variables** - var/let/const support, hoisting and reference capture
                  (closures), plus 'this' and 'arguments' support
- **Functions** - first-class function support, arrow functions, closures
- **Standard Type/Libraries ** - 'native' implementations of array, object,
                                 boolean, number, string, promise, etc.

## Not Supported (High Level)

As described, this is not intended to be a fully compliant, standalone script
engine like node or a browser page.  Functions (with closures) are the highest
level supported capability.  The following ECMAScript features are not
supported:

- **Modules** - import and export, this should be managed by the external
                application in the parsing and execution of scripts
- **Classes/Prototypes** - maybe someday, but for non-persistent business logic
                           it's an unnecessary complexity.  Object-oriented
                           elements like inheritance, get/set, new, etc. are
                           not supported
- **Async** - there are elements to support this from the application (see
              promises) but in general no async/await/yield/generators.  Scripts
              support execution in goroutines to allow for parallelism
- **Quirks** - oddities of ECMASCript, like automatic semicolon insertion,
               no...

## Installation

```
go get github.com/heisz/gescript
```

Requires Go 1.21 or later.

## Usage

This is big enough to warrant its own page, see [USAGE](USAGE) for details.

## License  
        
MIT License - see [LICENSE](LICENSE) for details.
