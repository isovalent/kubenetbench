package utils

import (
	"strings"
	"testing"
	"text/template"
)

var mainTmpl = `title: {{.title}}
section:
    {{.section}}`

var pizzaTmpl = `title: Pizza is the best
body: I love pizza!
`

var expected = `title: Best food
section:
    title: Pizza is the best
    body: I love pizza!
`

func TestPizza(t *testing.T) {
	values := map[string]interface{}{
		"title":   "Best food",
		"section": "{{template \"pizza\"}}",
	}

	mainT := template.Must(template.New("main").Parse(mainTmpl))
	pizzaT := template.Must(template.New("pizza").Parse(pizzaTmpl))

	templates := map[string]PrefixRenderer{
		"pizza": func(pw *PrefixWriter, params map[string]interface{}) {
			err := pizzaT.Execute(pw, params)
			if err != nil {
				panic(err)
			}
		},
	}

	var bld strings.Builder
	err := RenderTemplate(mainT, values, templates, &bld)
	if err != nil {
		t.Errorf("RenderTemplate failed with %v", err)
	}

	if bld.String() != expected {
		t.Errorf("RenderTemplate produced unexpected result:\n-->%s<--\nvs:\n-->%s<--", bld.String(), expected)
	}
}
