package utils

import (
	"fmt"
	"go/ast"
	"text/template"
)

const MockTemplate = `{{$unit := .}}

package {{.Pkg.Name}}

{{.FormatImports}}

{{range $rcvr, $sigs := .Pkg.Methods}}
{{$isLocal := $rcvr | $unit.IsDeclaredHere}}{{if $isLocal}}
type {{$rcvr}}Iface interface {
{{range $sig := $sigs}}  {{$sig.Name}}({{$sig.ListArgs}}){{$sig.ListReturns}}
{{end}}}
type {{$firstSig := index $sigs 0}}{{$firstSig.Receiver.ToMock}} struct { }
{{end}}{{end}}

{{range $rcvr, $sigs := .Pkg.Methods}}{{$isLocal := $rcvr | $unit.IsDeclaredHere}}{{if $isLocal}}
{{range $sig := $sigs}}func ({{$sig.Receiver.Name}} {{$sig.Receiver.Ptr}}{{$sig.Receiver.ToMock}}) {{$sig.Name}}({{$sig.ListArgs}}){{$sig.ListReturns}} {
  panic("mock: stub method not implemented")
}

{{end}}{{end}}

{{end}}
`

var Compiled *template.Template

func init() {
	var err error
	Compiled, err = template.New("mock").Parse(MockTemplate)
	if err != nil {
		panic(fmt.Sprintf("failed to parse template: %s", err))
	}
}

func IsPublicMethod(fn *ast.FuncDecl) bool {
	return len(fn.Name.Name) > 0 &&
		fn.Name.IsExported() &&
		fn.Recv != nil &&
		len(fn.Recv.List) > 0 &&
		len(fn.Recv.List[0].Names[0].Name) > 0
}
