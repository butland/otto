package parser

import (
	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/file"
	"github.com/robertkrimen/otto/token"
)

// A Mode value is a set of flags (or 0). They control optional parser functionality.
type Mode uint

type _parser struct {
	filename string
	str      string
	length   int
	base     int

	chr       rune // The current character
	chrOffset int  // The offset of current character
	offset    int  // The offset after current character (may be greater than 1)

	idx     file.Idx    // The index of token
	token   token.Token // The token
	literal string      // The literal of the token, if any

	scope             *_scope
	insertSemicolon   bool // If we see a newline, then insert an implicit semicolon
	implicitSemicolon bool // An implicit semicolon exists

	errors []error

	recover struct {
		// Scratch when trying to seek to the next statement, etc.
		idx   file.Idx
		count int
	}
}

func _newParser(filename, src string, base int) *_parser {
	return &_parser{
		chr:    ' ', // This is set so we can start scanning by skipping whitespace
		str:    src,
		length: len(src),
		base:   base,
	}
}

func newParser(filename, src string) *_parser {
	return _newParser(filename, src, 1)
}

// ParseFile parses the source code of a single JavaScript/ECMAScript source file and returns
// the corresponding ast.Program node.
//
// If fileSet == nil, ParseFile parses source without a FileSet.
// If fileSet != nil, ParseFile first adds filename and src to fileSet.
//
// The filename argument is optional and is used for labelling errors, etc.
//
// The mode parameter mode will control optional parser functionality, but does nothing right now (pass 0).
func ParseFile(fileSet *file.FileSet, filename, src string, mode Mode) (*ast.Program, error) {
	base := 1
	if fileSet != nil {
		base = fileSet.AddFile(filename, src)
	}
	parser := _newParser(filename, src, base)
	return parser.parse()
}

func (self *_parser) slice(idx0, idx1 file.Idx) string {
	from := int(idx0) - self.base
	to := int(idx1) - self.base
	if from >= 0 && to <= len(self.str) {
		return self.str[from:to]
	}

	return ""
}

func (self *_parser) parse() (*ast.Program, error) {
	self.next()
	program := self.parseProgram()
	if len(self.errors) > 0 {
		return program, self.errors[0]
	}
	return program, nil
}

func (self *_parser) next() {
	self.token, self.literal, self.idx = self.scan()
}

func (self *_parser) optionalSemicolon() {
	if self.token == token.SEMICOLON {
		self.next()
		return
	}

	if self.implicitSemicolon {
		self.implicitSemicolon = false
		return
	}

	if self.token != token.EOF && self.token != token.RIGHT_BRACE {
		self.expect(token.SEMICOLON)
	}
}

func (self *_parser) semicolon() {
	if self.token != token.RIGHT_PARENTHESIS && self.token != token.RIGHT_BRACE {
		if self.implicitSemicolon {
			self.implicitSemicolon = false
			return
		}

		self.expect(token.SEMICOLON)
	}
}

func (self *_parser) idxOf(offset int) file.Idx {
	return file.Idx(self.base + offset)
}

func (self *_parser) expect(value token.Token) file.Idx {
	idx := self.idx
	if self.token != value {
		self.errorUnexpectedToken(self.token)
	}
	self.next()
	return idx
}

func lineCount(str string) (int, int) {
	line, last := 0, -1
	pair := false
	for index, chr := range str {
		switch chr {
		case '\r':
			line += 1
			last = index
			pair = true
			continue
		case '\n':
			if !pair {
				line += 1
			}
			last = index
		case '\u2028', '\u2029':
			line += 1
			last = index + 2
		}
		pair = false
	}
	return line, last
}

func (self *_parser) position(idx file.Idx) file.Position {
	position := file.Position{}
	offset := int(idx) - self.base
	str := self.str[:offset]
	position.Filename = self.filename
	line, last := lineCount(str)
	position.Line = 1 + line
	if last >= 0 {
		position.Column = offset - last
	} else {
		position.Column = 1 + len(str)
	}

	return position
}
