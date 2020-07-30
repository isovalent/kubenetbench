package utils

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"
	"text/template/parse"
)

// PrefixRenderer writes data to pw
type PrefixRenderer = func(pw *PrefixWriter, params map[string]interface{})

// RenderTemplate renders templates respecting indentation
// Its intended use is for YAML templates
// check template_test for an example.
func RenderTemplate(
	main0 *template.Template,
	vmap map[string]interface{},
	tmap map[string]PrefixRenderer,
	wr io.Writer,
) error {
	var buff bytes.Buffer

	main0.Execute(&buff, vmap)
	main := template.Must(template.New("main").Parse(buff.String()))

	mainTree := main.Tree
	lastIndent := -1
	for _, node := range mainTree.Root.Nodes {
		switch ty := node.Type(); ty {
		case parse.NodeText:
			nodeTxt := node.(*parse.TextNode)
			_, err := wr.Write(nodeTxt.Text)
			if err != nil {
				return err
			}

			lastIndent = -1
			lastIdx := len(nodeTxt.Text) - 1
			cnt := 0
			for {
				if lastIdx == 0 {
					break
				}
				lastIdx--
				cnt++
				if nodeTxt.Text[lastIdx] == '\n' {
					lastIndent = cnt
					break
				}
			}
		case parse.NodeTemplate:
			nodeTmpl := node.(*parse.TemplateNode)
			renderer, ok := tmap[nodeTmpl.Name]
			if !ok {
				return fmt.Errorf("template %s does not exist", nodeTmpl.Name)
			}
			if lastIndent == -1 {
				panic("NYI")
			}
			pw := NewPrefixWriter(wr, true)
			pw.PushPrefix(strings.Repeat(" ", lastIndent))
			renderer(pw, vmap)
			pw.PopPrefix()
			err := pw.Done()
			if err != nil {
				return fmt.Errorf("error terminating prefix writer: %w", err)
			}
		default:
			panic("Unexpected node type")
		}
	}
	return nil
}
