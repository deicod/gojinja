package nodes

import (
	"testing"
)

func TestBaseNode(t *testing.T) {
	pos := NewPosition(10, 5)
	node := &BaseNode{}
	node.SetPosition(pos)

	if node.GetPosition().Line != 10 {
		t.Errorf("Expected line 10, got %d", node.GetPosition().Line)
	}

	if node.GetPosition().Column != 5 {
		t.Errorf("Expected column 5, got %d", node.GetPosition().Column)
	}

	if node.Type() != "BaseNode" {
		t.Errorf("Expected type BaseNode, got %s", node.Type())
	}

	if len(node.GetChildren()) != 0 {
		t.Errorf("Expected 0 children, got %d", len(node.GetChildren()))
	}
}

func TestConst(t *testing.T) {
	node := NewConst("hello", 1, 1)

	if node.Value != "hello" {
		t.Errorf("Expected value 'hello', got %v", node.Value)
	}

	if node.GetPosition().Line != 1 {
		t.Errorf("Expected line 1, got %d", node.GetPosition().Line)
	}

	if node.Type() != "Const" {
		t.Errorf("Expected type Const, got %s", node.Type())
	}

	ctx := NewEvalContext(nil, "")
	value, err := node.AsConst(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if value != "hello" {
		t.Errorf("Expected 'hello', got %v", value)
	}
}

func TestName(t *testing.T) {
	node := NewName("variable", CtxLoad, 2, 10)

	if node.Name != "variable" {
		t.Errorf("Expected name 'variable', got %s", node.Name)
	}

	if node.Ctx != CtxLoad {
		t.Errorf("Expected ctx '%s', got %s", CtxLoad, node.Ctx)
	}

	if !node.CanAssign() {
		t.Error("Expected name to be assignable")
	}

	// Test reserved names
	reservedNames := []string{"true", "false", "none", "True", "False", "None"}
	for _, name := range reservedNames {
		n := NewName(name, CtxStore, 1, 1)
		if n.CanAssign() {
			t.Errorf("Expected reserved name '%s' to not be assignable", name)
		}
	}
}

func TestBinExpr(t *testing.T) {
	left := NewConst(1, 1, 1)
	right := NewConst(2, 1, 10)

	node := NewAdd(left, right)

	if node.Operator != "+" {
		t.Errorf("Expected operator '+', got %s", node.Operator)
	}

	if node.Type() != "Add" {
		t.Errorf("Expected type Add, got %s", node.Type())
	}

	if len(node.GetChildren()) != 2 {
		t.Errorf("Expected 2 children, got %d", len(node.GetChildren()))
	}
}

func TestUnaryExpr(t *testing.T) {
	operand := NewConst(5, 1, 5)
	node := NewNeg(operand)

	if node.Operator != "-" {
		t.Errorf("Expected operator '-', got %s", node.Operator)
	}

	if node.Type() != "Neg" {
		t.Errorf("Expected type Neg, got %s", node.Type())
	}

	ctx := NewEvalContext(nil, "")
	value, err := node.AsConst(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if value != -5 {
		t.Errorf("Expected -5, got %v", value)
	}
}

func TestList(t *testing.T) {
	items := []Expr{
		NewConst(1, 1, 1),
		NewConst(2, 1, 5),
		NewConst(3, 1, 9),
	}

	node := &List{Items: items}

	if len(node.GetChildren()) != 3 {
		t.Errorf("Expected 3 children, got %d", len(node.GetChildren()))
	}

	ctx := NewEvalContext(nil, "")
	value, err := node.AsConst(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if list, ok := value.([]interface{}); ok {
		if len(list) != 3 {
			t.Errorf("Expected list length 3, got %d", len(list))
		}
		if list[0] != 1 || list[1] != 2 || list[2] != 3 {
			t.Errorf("Expected [1, 2, 3], got %v", list)
		}
	} else {
		t.Errorf("Expected []interface{}, got %T", value)
	}
}

func TestDict(t *testing.T) {
	pairs := []*Pair{
		{Key: NewConst("a", 1, 1), Value: NewConst(1, 1, 5)},
		{Key: NewConst("b", 1, 10), Value: NewConst(2, 1, 15)},
	}

	node := &Dict{Items: pairs}

	if len(node.GetChildren()) != 2 {
		t.Errorf("Expected 2 children, got %d", len(node.GetChildren()))
	}

	ctx := NewEvalContext(nil, "")
	value, err := node.AsConst(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if dict, ok := value.(map[interface{}]interface{}); ok {
		if len(dict) != 2 {
			t.Errorf("Expected dict length 2, got %d", len(dict))
		}
		if dict["a"] != 1 || dict["b"] != 2 {
			t.Errorf("Expected {a: 1, b: 2}, got %v", dict)
		}
	} else {
		t.Errorf("Expected map[interface{}]interface{}, got %T", value)
	}
}

func TestCompare(t *testing.T) {
	left := NewConst(5, 1, 1)
	right := NewConst(10, 1, 5)
	operand := &Operand{Op: "lt", Expr: right}

	node := &Compare{Expr: left, Ops: []*Operand{operand}}

	if len(node.GetChildren()) != 2 {
		t.Errorf("Expected 2 children, got %d", len(node.GetChildren()))
	}

	ctx := NewEvalContext(nil, "")
	value, err := node.AsConst(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result, ok := value.(bool); ok {
		if !result {
			t.Error("Expected 5 < 10 to be true")
		}
	} else {
		t.Errorf("Expected bool, got %T", value)
	}
}

func TestAnd(t *testing.T) {
	left := NewConst(true, 1, 1)
	right := NewConst(false, 1, 10)

	node := NewAnd(left, right)

	ctx := NewEvalContext(nil, "")
	value, err := node.AsConst(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result, ok := value.(bool); ok {
		if result {
			t.Error("Expected true AND false to be false")
		}
	} else {
		t.Errorf("Expected bool, got %T", value)
	}
}

func TestOr(t *testing.T) {
	left := NewConst(true, 1, 1)
	right := NewConst(false, 1, 10)

	node := NewOr(left, right)

	ctx := NewEvalContext(nil, "")
	value, err := node.AsConst(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result, ok := value.(bool); ok {
		if !result {
			t.Error("Expected true OR false to be true")
		}
	} else {
		t.Errorf("Expected bool, got %T", value)
	}
}

func TestNot(t *testing.T) {
	operand := NewConst(true, 1, 1)
	node := NewNot(operand)

	ctx := NewEvalContext(nil, "")
	value, err := node.AsConst(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result, ok := value.(bool); ok {
		if result {
			t.Error("Expected NOT true to be false")
		}
	} else {
		t.Errorf("Expected bool, got %T", value)
	}
}

func TestVisitorPattern(t *testing.T) {
	// Build a simple AST: (1 + 2) * 3
	left := NewAdd(NewConst(1, 1, 1), NewConst(2, 1, 5))
	right := NewConst(3, 1, 10)
	root := NewMul(left, right)

	// Count the number of Const nodes
	constCount := 0
	visitor := NodeVisitorFunc(func(node Node) interface{} {
		if _, ok := node.(*Const); ok {
			constCount++
		}
		return nil
	})

	Walk(visitor, root)

	if constCount != 3 {
		t.Errorf("Expected 3 Const nodes, got %d", constCount)
	}
}

func TestFind(t *testing.T) {
	// Build a simple AST: 1 + 2 * 3
	left := NewConst(1, 1, 1)
	right := NewMul(NewConst(2, 1, 5), NewConst(3, 1, 9))
	root := NewAdd(left, right)

	// Find first Const node
	found := Find(root, "Const")
	if found == nil {
		t.Error("Expected to find a Const node")
	}

	if found.Type() != "Const" {
		t.Errorf("Expected Const node, got %s", found.Type())
	}

	// Find all Const nodes
	allConsts := FindAll(root, "Const")
	if len(allConsts) != 3 {
		t.Errorf("Expected 3 Const nodes, got %d", len(allConsts))
	}

	// Find all Add nodes
	allAdds := FindAll(root, "Add")
	if len(allAdds) != 1 {
		t.Errorf("Expected 1 Add node, got %d", len(allAdds))
	}
}

func TestSetCtx(t *testing.T) {
	// Build AST with Name nodes
	name1 := NewName("var1", CtxLoad, 1, 1)
	name2 := NewName("var2", CtxLoad, 1, 5)
	root := NewAdd(name1, name2)

	// Set context to store
	SetCtx(root, CtxStore)

	if name1.Ctx != CtxStore {
		t.Errorf("Expected name1 ctx to be %s, got %s", CtxStore, name1.Ctx)
	}

	if name2.Ctx != CtxStore {
		t.Errorf("Expected name2 ctx to be %s, got %s", CtxStore, name2.Ctx)
	}
}

func TestSetLineNo(t *testing.T) {
	// Build AST
	left := NewConst(1, 0, 0) // No position set
	right := NewConst(2, 5, 10) // Position already set
	root := NewAdd(left, right)

	// Set line number
	SetLineNo(root, 10, false)

	if left.GetPosition().Line != 10 {
		t.Errorf("Expected left node line to be 10, got %d", left.GetPosition().Line)
	}

	if right.GetPosition().Line != 5 {
		t.Errorf("Expected right node line to remain 5, got %d", right.GetPosition().Line)
	}

	// Override existing line numbers
	SetLineNo(root, 20, true)

	if right.GetPosition().Line != 20 {
		t.Errorf("Expected right node line to be overridden to 20, got %d", right.GetPosition().Line)
	}
}

func TestDump(t *testing.T) {
	// Build a simple AST
	left := NewConst(1, 1, 1)
	right := NewConst(2, 1, 5)
	root := NewAdd(left, right)

	// Dump the AST
	dump := Dump(root)

	if dump == "" {
		t.Error("Expected non-empty dump string")
	}

	// Check that it contains expected node types
	if !contains(dump, "Add") {
		t.Error("Expected dump to contain 'Add'")
	}

	if !contains(dump, "Const") {
		t.Error("Expected dump to contain 'Const'")
	}
}

func TestTypeChecking(t *testing.T) {
	// Test various node types
	nodes := []Node{
		NewConst(1, 1, 1),
		NewName("test", CtxLoad, 1, 1),
		&Template{},
		&Pair{},
	}

	for _, node := range nodes {
		if IsStmt(node) {
			stmt, ok := AsStmt(node)
			if !ok {
				t.Errorf("Node %s reported as Stmt but AsStmt failed", node.Type())
			}
			if stmt == nil {
				t.Error("AsStmt returned nil for Stmt node")
			}
		}

		if IsExpr(node) {
			expr, ok := AsExpr(node)
			if !ok {
				t.Errorf("Node %s reported as Expr but AsExpr failed", node.Type())
			}
			if expr == nil {
				t.Error("AsExpr returned nil for Expr node")
			}
		}

		if IsHelper(node) {
			helper, ok := AsHelper(node)
			if !ok {
				t.Errorf("Node %s reported as Helper but AsHelper failed", node.Type())
			}
			if helper == nil {
				t.Error("AsHelper returned nil for Helper node")
			}
		}
	}
}

func TestEvalContext(t *testing.T) {
	ctx := NewEvalContext(nil, "test.html")

	if ctx.AutoEscape != true {
		t.Error("Expected default AutoEscape to be true")
	}

	if ctx.Volatile != false {
		t.Error("Expected default Volatile to be false")
	}

	// Test save/restore
	ctx.AutoEscape = false
	ctx.Volatile = true

	saved := ctx.Save()

	ctx.AutoEscape = true
	ctx.Volatile = false

	ctx.Revert(saved)

	if ctx.AutoEscape != false {
		t.Error("Expected AutoEscape to be restored to false")
	}

	if ctx.Volatile != true {
		t.Error("Expected Volatile to be restored to true")
	}
}

func TestTruthiness(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected bool
	}{
		{nil, false},
		{false, false},
		{true, true},
		{0, false},
		{1, true},
		{0.0, false},
		{1.5, true},
		{"", false},
		{"hello", true},
		{[]interface{}{}, false},
		{[]interface{}{1}, true},
		{map[interface{}]interface{}{}, false},
		{map[interface{}]interface{}{"a": 1}, true},
	}

	for _, test := range tests {
		result := isTruthy(test.value)
		if result != test.expected {
			t.Errorf("isTruthy(%v) = %v, expected %v", test.value, result, test.expected)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		func() bool {
			for i := 1; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())))
}

func TestStringRepresentations(t *testing.T) {
	tests := []struct {
		node     Node
		expected string
	}{
		{NewConst("hello", 1, 1), "Const"},
		{NewName("test", CtxLoad, 1, 1), "Name"},
		{NewAdd(NewConst(1, 1, 1), NewConst(2, 1, 5)), "BinExpr"},
		{NewNot(NewConst(true, 1, 1)), "UnaryExpr"},
		{&List{Items: []Expr{NewConst(1, 1, 1)}}, "List"},
		{&Template{Body: []Node{}}, "Template"},
		{&Output{Nodes: []Expr{}}, "Output"},
	}

	for _, test := range tests {
		str := test.node.String()
		if !contains(str, test.expected) {
			t.Errorf("String representation of %s should contain '%s', got: %s",
				test.node.Type(), test.expected, str)
		}
	}
}