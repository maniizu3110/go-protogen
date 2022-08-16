package generator

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"io/ioutil"
	"log"
	"strings"

	"github.com/iancoleman/strcase"
)

//go:embed template/template.txt
var templateTxt string

func GenerateProtoFile(inputFilePath string, outputPath string) error {
	fileName := strings.Split(inputFilePath, "/")[len(strings.Split(inputFilePath, "/"))-1]
	model := strings.Split(fileName, ".")[0]
	structFields := paramCreatorForProto(inputFilePath)

	params := map[string]any{
		"Model":  strcase.ToCamel(model),
		"model":  strcase.ToLowerCamel(model),
		"struct": structFields,
	}

	var buf bytes.Buffer
	t := template.Must(template.New("meta-txt").Parse(templateTxt))
	t.Execute(&buf, params)

	var out bytes.Buffer
	out.Write(buf.Bytes())

	if err := ioutil.WriteFile(outputPath+"/"+strcase.ToLowerCamel(model)+".proto", out.Bytes(), 0644); err != nil {
		log.Fatalf("writing output: %s", err)
	}

	return nil
}

func paramCreatorForProto(filePath string) string {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		log.Fatal(err)
	}
	typeInfo := make([]string, 0)
	jsonTag := make([]string, 0)
	// Make sure that the json tag is attached.
	ast.Inspect(f, func(n ast.Node) bool {
		ident, ok := n.(*ast.Field)
		if !ok {
			return true
		}
		if ident.Tag != nil {
			t := fmt.Sprintf("%s", ident.Type)
			tag := ident.Tag.Value
			_, right, ok := strings.Cut(tag, "json:")
			if ok {
				typeInfo = append(typeInfo, t)
				jsonTag = append(jsonTag, right[1:len(right)-2])
			}
		}
		return true
	})
	var ans string
	cnt := 1
	for i := range typeInfo {
		typ := typeInfo[i]
		tag := jsonTag[i]
		if strings.Contains(typ, "time") {
			typ = "google.protobuf.Timestamp"
		}
		ans += fmt.Sprintf("	%s %s = %d;\n", typ, tag, cnt)
		cnt++
	}
	return ans
}
