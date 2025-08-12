package renderer

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

type ConfluenceGHAlertsTextRenderer struct {
	html.Config
	StripNewlines bool
}

// NewConfluenceGHAlertsTextRenderer creates a new instance of the renderer for GitHub Alerts text
func NewConfluenceGHAlertsTextRenderer(stripNewlines bool, opts ...html.Option) renderer.NodeRenderer {
	return &ConfluenceGHAlertsTextRenderer{
		Config:        html.NewConfig(),
		StripNewlines: stripNewlines,
	}
}

// RegisterFuncs implements NodeRenderer.RegisterFuncs
func (r *ConfluenceGHAlertsTextRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindText, r.renderText)
}

func (r *ConfluenceGHAlertsTextRenderer) renderText(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*ast.Text)

	// Check if this text node has replacement content from the GHAlerts transformer
	if replacementContent, hasAttribute := node.Attribute([]byte("replacement-content")); hasAttribute && replacementContent != nil {
		if contentBytes, ok := replacementContent.([]byte); ok {
			_, err := writer.Write(contentBytes)
			if err != nil {
				return ast.WalkStop, err
			}
			return ast.WalkContinue, nil
		}
	}

	// Default text rendering behavior
	segment := n.Segment
	value := segment.Value(source)

	if r.StripNewlines {
		value = []byte(stripNewlines(value))
	}

	if n.IsRaw() {
		r.Writer.RawWrite(writer, value)
	} else {
		r.Writer.Write(writer, value)
	}

	if n.SoftLineBreak() {
		if r.HardWraps {
			r.Writer.RawWrite(writer, []byte("<br />\n"))
		} else {
			r.Writer.RawWrite(writer, []byte("\n"))
		}
	}

	return ast.WalkContinue, nil
}

// stripNewlines removes newline characters from the text
func stripNewlines(value []byte) string {
	result := make([]byte, 0, len(value))
	for _, b := range value {
		if b != '\n' && b != '\r' {
			result = append(result, b)
		}
	}
	return string(result)
}
