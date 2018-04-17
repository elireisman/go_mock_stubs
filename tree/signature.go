package tree

import (
	"go/ast"
	"strings"
)

type Signature struct {
	Name     string
	Receiver Field
	Args     []Field
	Returns  []Field
}

// the declared arguments for this method
func (s Signature) ListArgs() string {
	out := []string{}
	for _, arg := range s.Args {
		out = append(out, arg.Render())
	}

	return strings.Join(out, ", ")
}

// the declared return types for this method
func (s Signature) ListReturns() string {
	switch len(s.Returns) {
	case 0:
		return ""

	case 1:
		if s.Returns[0].Name != "" {
			return " (" + s.Returns[0].Render() + ")"
		} else {
			return " " + s.Returns[0].Render()
		}

	default:
		return " (" + strings.Join(s.getDeclaredReturns(), ", ") + ")"
	}
}

func (s Signature) getDeclaredReturns() []string {
	out := []string{}
	for _, f := range s.Returns {
		out = append(out, f.Render())
	}

	return out
}

func (s *Signature) ProcessArgs(args *ast.FieldList) {
	s.Args = s.processFields(args)
}

func (s *Signature) ProcessReturns(rets *ast.FieldList) {
	s.Returns = s.processFields(rets)
}

func (s *Signature) processFields(args *ast.FieldList) []Field {
	fields := []Field{}
	if args != nil {
		for _, f := range args.List {
			field := NewField(f)
			fields = append(fields, field)
		}
	}

	return fields
}
