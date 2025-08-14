package renderer

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

type ConfluenceGHAlertsTextRenderer struct {
	html.Config
	softBreak byte
}

// NewConfluenceGHAlertsTextRenderer creates a new instance of the renderer for GitHub Alerts text
func NewConfluenceGHAlertsTextRenderer(stripNewlines bool, opts ...html.Option) renderer.NodeRenderer {
	sb := '\n'
	if stripNewlines {
		sb = ' '
	}
	return &ConfluenceGHAlertsTextRenderer{
		Config:    html.NewConfig(),
		softBreak: byte(sb),
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

	// Default text rendering behavior (same as original ConfluenceTextRenderer)
	segment := n.Segment
	if n.IsRaw() {
		r.Writer.RawWrite(writer, segment.Value(source))
	} else {
		value := segment.Value(source)
		r.Writer.Write(writer, value)
		if n.HardLineBreak() || (n.SoftLineBreak() && r.HardWraps) {
			if r.XHTML {
				_, _ = writer.WriteString("<br />\n")
			} else {
				_, _ = writer.WriteString("<br>\n")
			}
		} else if n.SoftLineBreak() {
			_ = writer.WriteByte(r.softBreak)
		}
	}

	return ast.WalkContinue, nil
}
