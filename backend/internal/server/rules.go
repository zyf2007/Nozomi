package server

import (
	"context"
	stderrors "errors"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja/ast"
	"github.com/dop251/goja/parser"
)

func runRule(script string, input MailInput) (RuleResult, bool, error) {
	vm := goja.New()
	_ = vm.Set("input", ruleInput(input))
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	var value goja.Value
	var err error
	done := make(chan struct{})
	go func() {
		for _, source := range ruleSourceVariants(script) {
			value, err = vm.RunString(source)
			if err == nil {
				break
			}
		}
		close(done)
	}()
	select {
	case <-ctx.Done():
		vm.Interrupt("timeout")
		return RuleResult{}, false, stderrors.New("规则执行超时")
	case <-done:
	}
	if err != nil {
		return RuleResult{}, false, err
	}
	if value == nil || goja.IsNull(value) || goja.IsUndefined(value) {
		return RuleResult{}, false, nil
	}
	result, err := exportRuleResult(vm, value)
	if err != nil {
		return RuleResult{}, false, err
	}
	if result.Variables == nil {
		result.Variables = map[string]string{}
	}
	return result, true, nil
}

func ruleInput(input MailInput) map[string]any {
	return map[string]any{
		"from":    input.From,
		"to":      input.To,
		"subject": input.Subject,
		"text":    input.Text,
		"html":    input.HTML,
		"raw":     input.Raw,
		"headers": input.Headers,
	}
}

func ruleSourceVariants(script string) []string {
	var out []string
	if transformed, ok := returnLastIfExpression(script); ok {
		out = append(out, transformed)
	}
	out = append(out, script)
	if strings.Contains(script, "return") {
		out = append(out, "(function(){\n"+script+"\n})()")
	}
	return out
}

func returnLastIfExpression(script string) (string, bool) {
	program, err := parser.ParseFile(nil, "", script, 0)
	if err != nil || len(program.Body) == 0 {
		return "", false
	}
	lastIf, ok := program.Body[len(program.Body)-1].(*ast.IfStatement)
	if !ok || lastIf.Alternate == nil {
		return "", false
	}
	consequent, ok := lastIf.Consequent.(*ast.ExpressionStatement)
	if !ok {
		return "", false
	}
	alternate, ok := lastIf.Alternate.(*ast.ExpressionStatement)
	if !ok {
		return "", false
	}
	prefixEnd := strings.LastIndex(script[:idx0(lastIf.Test)], "if")
	if prefixEnd < 0 {
		return "", false
	}
	prefix := script[:prefixEnd]
	test := nodeSource(script, lastIf.Test)
	whenTrue := nodeSource(script, consequent.Expression)
	whenFalse := nodeSource(script, alternate.Expression)
	return "(function(){\n" + prefix + "\nreturn ((" + test + ") ? (" + whenTrue + ") : (" + whenFalse + "));\n})()", true
}

func nodeSource(script string, node ast.Node) string {
	return strings.TrimSpace(script[idx0(node):idx1(node)])
}

func idx0(node ast.Node) int {
	idx := int(node.Idx0()) - 1
	if idx < 0 {
		return 0
	}
	return idx
}

func idx1(node ast.Node) int {
	idx := int(node.Idx1()) - 1
	if idx < 0 {
		return 0
	}
	return idx
}

func exportRuleResult(vm *goja.Runtime, value goja.Value) (RuleResult, error) {
	obj := value.ToObject(vm)
	if obj == nil {
		return RuleResult{}, nil
	}
	var result RuleResult
	result.TemplateID = uint64(valueInteger(obj.Get("templateId")))
	if result.TemplateID == 0 {
		result.TemplateID = uint64(valueInteger(obj.Get("TemplateID")))
	}
	if subject := obj.Get("subject"); subject != nil && !goja.IsUndefined(subject) && !goja.IsNull(subject) {
		result.Subject = subject.String()
	} else if subject := obj.Get("Subject"); subject != nil && !goja.IsUndefined(subject) && !goja.IsNull(subject) {
		result.Subject = subject.String()
	}
	var variables map[string]string
	rawVariables := obj.Get("variables")
	if goja.IsUndefined(rawVariables) {
		rawVariables = obj.Get("Variables")
	}
	if !goja.IsUndefined(rawVariables) && !goja.IsNull(rawVariables) {
		if err := vm.ExportTo(rawVariables, &variables); err != nil {
			return RuleResult{}, err
		}
	}
	result.Variables = variables
	return result, nil
}

func valueInteger(value goja.Value) int64 {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return 0
	}
	return value.ToInteger()
}
