package parser

import (
	"encoding/json"
	"reflect"

	"github.com/robertkrimen/otto/ast"
)

func marshal(name string, children ...interface{}) map[string]interface{} {
	if len(children) == 1 {
		return map[string]interface{}{
			name: children[0],
		}
	}
	map_ := map[string]interface{}{}
	length := len(children) / 2
	for i := 0; i < length; i++ {
		name := children[i*2].(string)
		value := children[i*2+1]
		map_[name] = value
	}
	if name == "" {
		return map_
	}
	return map[string]interface{}{
		name: map_,
	}
}

func testMarshalNode(node interface{}) interface{} {
	switch node := node.(type) {

	// Expression

	case *ast.ArrayLiteral:
		return marshal("Array", testMarshalNode(node.Value))

	case *ast.AssignExpression:
		return marshal("Assign",
			"Left", testMarshalNode(node.Left),
			"Right", testMarshalNode(node.Right),
		)

	case *ast.BinaryExpression:
		return marshal("BinaryExpression",
			"Operator", node.Operator.String(),
			"Left", testMarshalNode(node.Left),
			"Right", testMarshalNode(node.Right),
		)

	case *ast.BooleanLiteral:
		return marshal("Literal", node.Value)

	case *ast.CallExpression:
		return marshal("Call",
			"Callee", testMarshalNode(node.Callee),
			"ArgumentList", testMarshalNode(node.ArgumentList),
		)

	case *ast.ConditionalExpression:
		return marshal("Conditional",
			"Test", testMarshalNode(node.Test),
			"Consequent", testMarshalNode(node.Consequent),
			"Alternate", testMarshalNode(node.Alternate),
		)

	case *ast.DotExpression:
		return marshal("Dot",
			"Left", testMarshalNode(node.Left),
			"Member", node.Identifier.Name,
		)

	case *ast.NewExpression:
		return marshal("New",
			"Callee", testMarshalNode(node.Callee),
			"ArgumentList", testMarshalNode(node.ArgumentList),
		)

	case *ast.NullLiteral:
		return marshal("Literal", nil)

	case *ast.NumberLiteral:
		return marshal("Literal", node.Value)

	case *ast.ObjectLiteral:
		return marshal("Object", testMarshalNode(node.Value))

	case *ast.RegExpLiteral:
		return marshal("Literal", node.Literal)

	case *ast.StringLiteral:
		return marshal("Literal", node.Literal)

	// Statement

	case *ast.Program:
		return testMarshalNode(node.Body)

	case *ast.BlockStatement:
		return marshal("BlockStatement", testMarshalNode(node.List))

	case *ast.EmptyStatement:
		return "EmptyStatement"

	case *ast.ExpressionStatement:
		return testMarshalNode(node.Expression)

	case *ast.FunctionExpression:
		return marshal("Function", testMarshalNode(node.Body))

	case *ast.Identifier:
		return marshal("Identifier", node.Name)

	case *ast.IfStatement:
		if_ := marshal("",
			"Test", testMarshalNode(node.Test),
			"Consequent", testMarshalNode(node.Consequent),
		)
		if node.Alternate != nil {
			if_["Alternate"] = testMarshalNode(node.Alternate)
		}
		return marshal("If", if_)

	case *ast.LabelledStatement:
		return marshal("Label",
			"Name", node.Label.Name,
			"Statement", testMarshalNode(node.Statement),
		)
	case ast.Property:
		return marshal("",
			"Key", node.Key,
			"Value", testMarshalNode(node.Value),
		)

	case *ast.ReturnStatement:
		return marshal("Return", testMarshalNode(node.Argument))

	case *ast.SequenceExpression:
		return marshal("Sequence", testMarshalNode(node.Sequence))
	}

	{
		value := reflect.ValueOf(node)
		if value.Kind() == reflect.Slice {
			tmp0 := []interface{}{}
			for index := 0; index < value.Len(); index++ {
				tmp0 = append(tmp0, testMarshalNode(value.Index(index).Interface()))
			}
			return tmp0
		}
	}

	return nil
}

func testMarshal(node interface{}) string {
	value, err := json.Marshal(testMarshalNode(node))
	if err != nil {
		panic(err)
	}
	return string(value)
}
