# token
--
    import "github.com/robertkrimen/otto/token"

Package token defines constants representing the lexical tokens of JavaScript
(ECMA5).

## Usage

#### type Token

```go
type Token int
```

Token is the set of lexical tokens in JavaScript (ECMA5).

```go
const (
	ILLEGAL Token = iota
	EOF
	COMMENT

	STRING
	BOOLEAN
	NULL
	NUMBER
	IDENTIFIER

	PLUS      // +
	MINUS     // -
	MULTIPLY  // *
	SLASH     // /
	REMAINDER // %

	AND                  // &
	OR                   // |
	EXCLUSIVE_OR         // ^
	SHIFT_LEFT           // <<
	SHIFT_RIGHT          // >>
	UNSIGNED_SHIFT_RIGHT // >>>
	AND_NOT              // &^

	ADD_ASSIGN       // +=
	SUBTRACT_ASSIGN  // -=
	MULTIPLY_ASSIGN  // *=
	QUOTIENT_ASSIGN  // /=
	REMAINDER_ASSIGN // %=

	AND_ASSIGN                  // &=
	OR_ASSIGN                   // |=
	EXCLUSIVE_OR_ASSIGN         // ^=
	SHIFT_LEFT_ASSIGN           // <<=
	SHIFT_RIGHT_ASSIGN          // >>=
	UNSIGNED_SHIFT_RIGHT_ASSIGN // >>>=
	AND_NOT_ASSIGN              // &^=

	LOGICAL_AND // &&
	LOGICAL_OR  // ||
	INCREMENT   // ++
	DECREMENT   // --

	EQUAL        // ==
	STRICT_EQUAL // ===
	LESS         // <
	GREATER      // >
	ASSIGN       // =
	NOT          // !

	BITWISE_NOT // ~

	NOT_EQUAL        // !=
	STRICT_NOT_EQUAL // !==
	LESS_OR_EQUAL    // <=
	GREATER_OR_EQUAL // <=

	LEFT_PARENTHESIS // (
	LEFT_BRACKET     // [
	LEFT_BRACE       // {
	COMMA            // ,
	PERIOD           // .

	RIGHT_PARENTHESIS // )
	RIGHT_BRACKET     // ]
	RIGHT_BRACE       // }
	SEMICOLON         // ;
	COLON             // :
	QUESTION_MARK     // ?

	IF
	IN
	DO

	VAR
	FOR
	NEW
	TRY
	LET

	THIS
	ELSE
	CASE
	VOID
	WITH
	ENUM

	WHILE
	BREAK
	CATCH
	THROW
	CONST
	YIELD
	CLASS
	SUPER

	RETURN
	TYPEOF
	DELETE
	SWITCH
	EXPORT
	IMPORT

	DEFAULT
	FINALLY
	EXTENDS

	FUNCTION
	CONTINUE
	DEBUGGER

	INSTANCEOF
)
```

#### func  IsKeyword

```go
func IsKeyword(literal string) Token
```
IsKeyword returns true if the literal represents a keyword; false otherwise.

#### func (Token) String

```go
func (tkn Token) String() string
```
String returns the string corresponding to the token. For operators, delimiters,
and keywords the string is the actual token string (e.g., for the token PLUS,
the String() is "+"). For all other tokens the string corresponds to the token
name (e.g. for the token IDENTIFIER, the string is "IDENTIFIER").

--
**godocdown** http://github.com/robertkrimen/godocdown
