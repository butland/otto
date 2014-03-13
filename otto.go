/*
Package otto is a JavaScript parser and interpreter written natively in Go.

http://godoc.org/github.com/robertkrimen/otto

	// Create a new runtime
	ctx := otto.New()

	ctx.Run(`
        abc = 2 + 2;
		console.log("The value of abc is " + abc);
		// The value of abc is 4
	`)

	value, err := ctx.Get("abc")
	{
		// value is an int64 with a value of 4
		value, _ := value.ToInteger()
	}

	ctx.Set("def", 11)
	ctx.Run(`
		console.log("The value of def is " + def);
		// The value of def is 11
	`)

	ctx.Set("xyzzy", "Nothing happens.")
	ctx.Run(`
		console.log(xyzzy.length); // 16
	`)

	value, _ = ctx.Run("xyzzy.length")
	{
		// value is an int64 with a value of 16
		value, _ := value.ToInteger()
	}

	value, err = ctx.Run("abcdefghijlmnopqrstuvwxyz.length")
	if err != nil {
		// err = ReferenceError: abcdefghijlmnopqrstuvwxyz is not defined
		// If there is an error, then value.IsUndefined() is true
		...
	}

Embedding a Go function in JavaScript:

	ctx.Set("sayHello", func(call otto.FunctionCall) otto.Value {
		fmt.Printf("Hello, %s.\n", call.Argument(0).String())
		return otto.UndefinedValue()
	})

	ctx.Set("twoPlus", func(call otto.FunctionCall) otto.Value {
		right, _ := call.Argument(0).ToInteger()
		result, _ := ctx.ToValue(2 + right)
		return result
	})

	result, _ = ctx.Run(`
		// First, say a greeting
		sayHello("Xyzzy"); // Hello, Xyzzy.
		sayHello(); // Hello, undefined

		result = twoPlus(2.0); // 4
	`)

You can run (Go) JavaScript from the commandline with: http://github.com/robertkrimen/otto/tree/master/otto

	$ go get -v github.com/robertkrimen/otto/otto

Run JavaScript by entering some source on stdin or by giving otto a filename:

	$ otto example.js

Optionally include the JavaScript utility-belt library, underscore, with this import:

	import (
		"github.com/robertkrimen/otto"
		_ "github.com/robertkrimen/otto/underscore"
	)

	// Now every otto runtime will come loaded with underscore

For more information: http://github.com/robertkrimen/otto/tree/master/underscore

Caveat Emptor

    * For now, otto is a hybrid ECMA3/ECMA5 interpreter. Parts of the specification are still works in progress.
    * For example, "use strict" will parse, but does nothing.
    * Error reporting needs to be improved.
    * Does not support the (?!) or (?=) regular expression syntax (because Go does not)
    * JavaScript considers a vertical tab (\000B <VT>) to be part of the whitespace class (\s), while RE2 does not.
    * Really, error reporting could use some improvement.

Regular Expression Syntax

Go translates JavaScript-style regular expressions into something that is "regexp" package compatible.

Unfortunately, JavaScript has positive lookahead, negative lookahead, and backreferencing,
all of which are not supported by Go's RE2-like engine: https://code.google.com/p/re2/wiki/Syntax

A brief discussion of these limitations: "Regexp (?!re)" https://groups.google.com/forum/?fromgroups=#%21topic/golang-nuts/7qgSDWPIh_E

More information about RE2: https://code.google.com/p/re2/

JavaScript considers a vertical tab (\000B <VT>) to be part of the whitespace class (\s), while RE2 does not.

Halting Problem

If you want to stop long running executions (like third-party code), you can use the interrupt channel to do this:

    package main

    import (
        "errors"
        "fmt"
        "os"
        "time"

        "github.com/robertkrimen/otto"
    )

    var Halt = errors.New("Halt")

    func main() {
        runUnsafe(`var abc = [];`)
        runUnsafe(`
        while (true) {
            // Loop forever
        }`)
    }

    func runUnsafe(unsafe string) {
        start := time.Now()
        defer func() {
            duration := time.Since(start)
            if caught := recover(); caught != nil {
                if caught == Halt {
                    fmt.Fprintf(os.Stderr, "Some code took to long! Stopping after: %v\n", duration)
                    return
                }
                panic(caught) // Something else happened, repanic!
            }
            fmt.Fprintf(os.Stderr, "Ran code successfully: %v\n", duration)
        }()
        ctx := otto.New()
        ctx.Interrupt = make(chan func())
        go func() {
            time.Sleep(2 * time.Second) // Stop after two seconds
            ctx.Interrupt <- func() {
                panic(Halt)
            }
        }()
        ctx.Run(unsafe) // Here be dragons (risky code)
        ctx.Interrupt = nil
    }

Where is setTimeout/setInterval?

These timing functions are not actually part of the ECMA-262 specification. Typically, they belong to the `windows` object (in the browser).
It would not be difficult to provide something like these via Go, but you probably want to wrap otto in an event loop in that case.

Here is some discussion of the problem:

* http://book.mixu.net/node/ch2.html

* http://en.wikipedia.org/wiki/Reentrancy_%28computing%29

* http://aaroncrane.co.uk/2009/02/perl_safe_signals/

*/
package otto

import (
	"fmt"
	"strings"

	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/registry"
)

// Otto is the representation of the JavaScript runtime. Each instance of Otto has a self-contained namespace.
type Otto struct {
	// Interrupt is a channel for interrupting the runtime. You can use this to halt a long running execution, for example.
	// See "Halting Problem" for more information.
	Interrupt chan func()
	runtime   *_runtime
}

// New will allocate a new JavaScript runtime
func New() *Otto {
	self := &Otto{
		runtime: newContext(),
	}
	self.runtime.Otto = self
	self.Set("console", self.runtime.newConsole())

	registry.Apply(func(entry registry.Entry) {
		self.Run(entry.Source())
	})

	return self
}

func (otto *Otto) clone() *Otto {
	self := &Otto{
		runtime: otto.runtime.clone(),
	}
	self.runtime.Otto = self
	return self
}

// Run will allocate a new JavaScript runtime, run the given source
// on the allocated runtime, and return the runtime, resulting value, and
// error (if any).
func Run(source string) (*Otto, Value, error) {
	otto := New()
	value, err := otto.Run(source)
	return otto, value, err
}

// Run will run the given source (parsing it first), returning the resulting value and error (if any)
//
// If the runtime is unable to parse source, then this function will return undefined and the parse error (nothing
// will be evaluated in this case).
func (self Otto) Run(source string) (Value, error) {
	return self.runtime.run(source)
}

// Get the value of the top-level binding of the given name.
//
// If there is an error (like the binding does not exist), then the value
// will be undefined.
func (self Otto) Get(name string) (Value, error) {
	value := UndefinedValue()
	err := catchPanic(func() {
		value = self.getValue(name)
	})
	return value, err
}

func (self Otto) getValue(name string) Value {
	return self.runtime.GlobalEnvironment.GetValue(name, false)
}

// Set the top-level binding of the given name to the given value.
//
// Set will automatically apply ToValue to the given value in order
// to convert it to a JavaScript value (type Value).
//
// If there is an error (like the binding is read-only, or the ToValue conversion
// fails), then an error is returned.
//
// If the top-level binding does not exist, it will be created.
func (self Otto) Set(name string, value interface{}) error {
	{
		value, err := self.ToValue(value)
		if err != nil {
			return err
		}
		err = catchPanic(func() {
			self.setValue(name, value)
		})
		return err
	}
}

func (self Otto) setValue(name string, value Value) {
	self.runtime.GlobalEnvironment.SetValue(name, value, false)
}

// Call the given JavaScript with a given this and arguments.
//
// If this is nil, then some special handling takes place to determine the proper
// this value, falling back to a "standard" invocation if necessary (where this is
// undefined).
//
// If source begins with "new " (A lowercase new followed by a space), then
// Call will invoke the function constructor rather than performing a function call.
// In this case, the this argument has no effect.
//
//      // value is a String object                                                       
//      value, _ := ctx.Call("Object", nil, "Hello, World.")                             
//                                                                                        
//      // Likewise...                                                                    
//      value, _ := ctx.Call("new Object", nil, "Hello, World.")                         
//                                                                                        
//      // This will perform a concat on the given array and return the result            
//      // value is [ 1, 2, 3, undefined, 4, 5, 6, 7, "abc" ]                             
//      value, _ := ctx.Call(`[ 1, 2, 3, undefined, 4 ].concat`, nil, 5, 6, 7, "abc")    
//
func (self Otto) Call(source string, this interface{}, argumentList ...interface{}) (Value, error) {

	thisValue := UndefinedValue()

	new_ := false
	switch {
	case strings.HasPrefix(source, "new "):
		source = source[4:]
		new_ = true
	}

	if !new_ && this == nil {
		value := UndefinedValue()
		fallback := false
		err := catchPanic(func() {
			// FIXME
			programNode := mustParse(self.runtime.parse(source + "()"))
			if callNode, valid := programNode.Body[0].(*ast.ExpressionStatement).Expression.(*ast.CallExpression); valid {
				value = self.runtime.evaluateCall(callNode, argumentList)
			} else {
				fallback = true
			}
		})
		if !fallback && err == nil {
			return value, nil
		}
	} else {
		value, err := self.ToValue(this)
		if err != nil {
			return UndefinedValue(), err
		}
		thisValue = value
	}

	fnValue, err := self.Run(source)
	if err != nil {
		return UndefinedValue(), err
	}

	value := UndefinedValue()
	if new_ {
		value, err = fnValue.constructSafe(thisValue, argumentList...)
		if err != nil {
			return UndefinedValue(), err
		}
	} else {
		value, err = fnValue.Call(thisValue, argumentList...)
		if err != nil {
			return UndefinedValue(), err
		}
	}

	return value, nil
}

// Object will run the given source and return the result as an object.
//
// For example, accessing an existing object:
//
//		object, _ := ctx.Object(`Number`)
//
// Or, creating a new object:
//
//		object, _ := ctx.Object(`({ xyzzy: "Nothing happens." })`)
//
// Or, creating and assigning an object:
//
//		object, _ := ctx.Object(`xyzzy = {}`)
//		object.Set("volume", 11)
//
// If there is an error (like the source does not result in an object), then
// nil and an error is returned.
func (self Otto) Object(source string) (*Object, error) {
	value, err := self.runtime.run(source)
	if err != nil {
		return nil, err
	}
	if value.IsObject() {
		return value.Object(), nil
	}
	return nil, fmt.Errorf("value is not an object")
}

// ToValue will convert an interface{} value to a value digestible by otto/JavaScript.
func (self Otto) ToValue(value interface{}) (Value, error) {
	return self.runtime.ToValue(value)
}

// Copy will create a copy/clone of the runtime.
//
// Copy is useful for saving some processing time when creating many similar
// runtimes.
//
// This implementation is alpha-ish, and works by introspecting every part of the runtime
// and reallocating and then relinking everything back together. Please report if you
// notice any inadvertent sharing of data between copies.
func (self *Otto) Copy() *Otto {
	otto := &Otto{
		runtime: self.runtime.clone(),
	}
	otto.runtime.Otto = otto
	return otto
}

// Object{}

// Object is the representation of a JavaScript object.
type Object struct {
	object *_object
	value  Value
}

func _newObject(object *_object, value Value) *Object {
	// value MUST contain object!
	return &Object{
		object: object,
		value:  value,
	}
}

// Call a method on the object.
//
// It is essentially equivalent to:
//
//		var method, _ := object.Get(name)
//		method.Call(object, argumentList...)
//
// An undefined value and an error will result if:
//
//		1. There is an error during conversion of the argument list
//		2. The property is not actually a function
//		3. An (uncaught) exception is thrown
//
func (self Object) Call(name string, argumentList ...interface{}) (Value, error) {
	// TODO: Insert an example using JavaScript below...
	// e.g., Object("JSON").Call("stringify", ...)

	function, err := self.Get(name)
	if err != nil {
		return UndefinedValue(), err
	}
	return function.Call(self.Value(), argumentList...)
}

// Value will return self as a value.
func (self Object) Value() Value {
	return self.value
}

// Get the value of the property with the given name.
func (self Object) Get(name string) (Value, error) {
	value := UndefinedValue()
	err := catchPanic(func() {
		value = self.object.get(name)
	})
	return value, err
}

// Set the property of the given name to the given value.
//
// An error will result if the setting the property triggers an exception (i.e. read-only),
// or there is an error during conversion of the given value.
func (self Object) Set(name string, value interface{}) error {
	{
		value, err := self.object.runtime.ToValue(value)
		if err != nil {
			return err
		}
		err = catchPanic(func() {
			self.object.put(name, value, true)
		})
		return err
	}
}

// Get the keys for the object
//
// Equivalent to calling Object.keys on the object
func (self Object) Keys() []string {
	var keys []string
	self.object.enumerate(false, func(name string) bool {
		keys = append(keys, name)
		return true
	})
	return keys
}

// Class will return the class string of the object.
//
// The return value will (generally) be one of:
//
//		Object
//		Function
//		Array
//		String
//		Number
//		Boolean
//		Date
//		RegExp
//
func (self Object) Class() string {
	return self.object.class
}
