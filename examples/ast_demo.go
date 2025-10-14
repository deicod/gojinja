package examples

import (
	"fmt"

	"github.com/deicod/gojinja/nodes"
)

func RunASTDemo() {
	fmt.Println("Go Jinja2 AST Node System Example")
	fmt.Println("=================================")

	// Create a simple AST representing: {{ (x + 5) * y if condition else "default" }}

	// Create leaf nodes
	x := nodes.NewName("x", nodes.CtxLoad, 1, 5)
	five := nodes.NewConst(5, 1, 9)
	y := nodes.NewName("y", nodes.CtxLoad, 1, 14)
	condition := nodes.NewName("condition", nodes.CtxLoad, 1, 19)
	defaultValue := nodes.NewConst("default", 1, 33)

	// Build the expression: (x + 5) * y
	add := nodes.NewAdd(x, five)
	mul := nodes.NewMul(add, y)

	// Create the conditional expression
	condExpr := &nodes.CondExpr{
		Test:  condition,
		Expr1: mul,
		Expr2: defaultValue,
	}

	// Create an output node containing the expression
	output := nodes.NewOutput([]nodes.Expr{condExpr}, 1, 3)

	// Create the template
	template := nodes.NewTemplate([]nodes.Node{output})

	fmt.Println("\n1. AST Structure:")
	fmt.Println(DumpAST(template))

	fmt.Println("\n2. Visitor Pattern Example:")
	// Count different node types using visitor pattern
	var constCount, nameCount, binExprCount int

	visitor := nodes.NodeVisitorFunc(func(node nodes.Node) interface{} {
		switch node.(type) {
		case *nodes.Const:
			constCount++
		case *nodes.Name:
			nameCount++
		case *nodes.BinExpr:
			binExprCount++
		}
		return nil
	})

	nodes.Walk(visitor, template)

	fmt.Printf("   Const nodes: %d\n", constCount)
	fmt.Printf("   Name nodes: %d\n", nameCount)
	fmt.Printf("   BinExpr nodes: %d\n", binExprCount)

	fmt.Println("\n3. Constant Evaluation Example:")
	ctx := nodes.NewEvalContext(nil, "example.html")

	fmt.Println("   Evaluating leaf nodes:")
	fmt.Printf("   Five as const: %v\n", evalConst(five, ctx))

	fmt.Println("\n4. Context Setting Example:")
	fmt.Printf("   Original context for 'x': %s\n", x.Ctx)
	nodes.SetCtx(template, nodes.CtxStore)
	fmt.Printf("   After setting to store: %s\n", x.Ctx)

	fmt.Println("\n5. Node Finding Example:")
	addNode := nodes.Find(template, "Add")
	if addNode != nil {
		fmt.Printf("   Found Add node: %s\n", addNode.String())
	}

	allNames := nodes.FindAll(template, "Name")
	fmt.Printf("   All Name nodes (%d):\n", len(allNames))
	for i, name := range allNames {
		if nameNode, ok := name.(*nodes.Name); ok {
			fmt.Printf("     %d. %s (ctx: %s)\n", i+1, nameNode.Name, nameNode.Ctx)
		}
	}

	fmt.Println("\n6. Type Checking Example:")
	fmt.Printf("   Template is Stmt: %t\n", nodes.IsStmt(template))
	fmt.Printf("   Output is Stmt: %t\n", nodes.IsStmt(output))
	fmt.Printf("   Conditional expression is Expr: %t\n", nodes.IsExpr(condExpr))
	fmt.Printf("   Const is Expr: %t\n", nodes.IsExpr(five))

	fmt.Println("\n7. Position Information Example:")
	fmt.Printf("   Five position: Line %d, Column %d\n",
		five.GetPosition().Line, five.GetPosition().Column)

	// Set line numbers for all nodes
	nodes.SetLineNo(template, 10, true)
	fmt.Printf("   After SetLineNo(10, true): Five position: Line %d, Column %d\n",
		five.GetPosition().Line, five.GetPosition().Column)
}

// Helper function to safely evaluate constants
func evalConst(expr nodes.Expr, ctx *nodes.EvalContext) interface{} {
	value, err := expr.AsConst(ctx)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return value
}

// Helper function to dump AST structure
func DumpAST(node nodes.Node) string {
	return nodes.Dump(node)
}