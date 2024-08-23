package main

import (
	"bytes"
	"fmt"
	"html/template"
)

type Item struct {
	Name        string
	Enabled     bool
	ShouldBreak bool
}

type Data struct {
	Items []Item
}

func main() {
	// 定义测试数据
	data := Data{
		Items: []Item{
			{Name: "Item 1", Enabled: true, ShouldBreak: false},
			{Name: "Item 2", Enabled: false, ShouldBreak: true},
			{Name: "Item 3", Enabled: true, ShouldBreak: false},
			{Name: "Item 4", Enabled: false, ShouldBreak: false},
		},
	}

	// 渲染模板
	tmpl, err := template.New("template").Parse(templateSource)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		panic(err)
	}

	fmt.Println(buf.String())
}

const templateSource = `{{define "template"}}
{{range $index, $item := .Items}}
  {{if $item.Enabled}}
    {{$item.Name}}
  {{else if $item.ShouldBreak}}
    {{break}}
  {{else}}
    {{continue}}
  {{end}}
{{end}}
{{end}}`
