# parser
--
    import "github.com/robertkrimen/otto/parser"


## Usage

#### func  ParseFile

```go
func ParseFile(fileSet *file.FileSet, filename, src string, mode Mode) (*ast.Program, error)
```
ParseFile parses the source code of a single JavaScript/ECMAScript source file
and returns the corresponding ast.Program node.

If fileSet == nil, ParseFile parses source without a FileSet. If fileSet != nil,
ParseFile first adds filename and src to fileSet.

The filename argument is optional and is used for labelling errors, etc.

The mode parameter mode will control optional parser functionality, but does
nothing right now (pass 0).

#### func  TransformRegExp

```go
func TransformRegExp(pattern string) (string, error)
```
TransformRegExp transforms a Go "regexp" pattern from a JavaScript/ECMAScript
pattern.

re2 (Go) cannot do backtracking, so the presence of a lookahead (?=) (?!) or
backreference (\1, \2, ...) will cause an error.

re2 (Go) has a different definition for \s: [\t\n\f\r ] The JavaScript
definition includes \v, Unicode "Separator, Space", etc.

If the pattern is invalid (not valid even in JavaScript), then this function
returns the empty string and an error.

If the pattern is valid, but incompatible (contains a lookahead or
backreference), then this function returns a non-empty string and an error.

#### type Error

```go
type Error struct {
	Message  string
	Position file.Position
}
```

An Error represents a parsing error. It includes the position where the error
occurred and a message/description.

#### func (Error) Error

```go
func (self Error) Error() string
```

#### type Mode

```go
type Mode uint
```

A Mode value is a set of flags (or 0). They control optional parser
functionality.

--
**godocdown** http://github.com/robertkrimen/godocdown
