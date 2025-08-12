package mark

import (
	"bytes"
	"slices"

	"github.com/kovetskiy/mark/attachment"
	cparser "github.com/kovetskiy/mark/parser"
	crenderer "github.com/kovetskiy/mark/renderer"
	"github.com/kovetskiy/mark/stdlib"
	ctransformer "github.com/kovetskiy/mark/transformer"
	"github.com/kovetskiy/mark/types"
	"github.com/reconquest/pkg/log"
	mkDocsParser "github.com/stefanfritsch/goldmark-admonitions"
	"github.com/yuin/goldmark"

	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// Renderer renders anchor [Node]s.
type ConfluenceExtension struct {
	html.Config
	Stdlib      *stdlib.Lib
	Path        string
	MarkConfig  types.MarkConfig
	Attachments []attachment.Attachment
}

// NewConfluenceRenderer creates a new instance of the ConfluenceRenderer
func NewConfluenceExtension(stdlib *stdlib.Lib, path string, cfg types.MarkConfig) *ConfluenceExtension {
	return &ConfluenceExtension{
		Config:      html.NewConfig(),
		Stdlib:      stdlib,
		Path:        path,
		MarkConfig:  cfg,
		Attachments: []attachment.Attachment{},
	}
}

func (c *ConfluenceExtension) Attach(a attachment.Attachment) {
	c.Attachments = append(c.Attachments, a)
}

func (c *ConfluenceExtension) Extend(m goldmark.Markdown) {

	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(crenderer.NewConfluenceTextRenderer(c.MarkConfig.StripNewlines), 100),
		util.Prioritized(crenderer.NewConfluenceBlockQuoteRenderer(), 100),
		util.Prioritized(crenderer.NewConfluenceCodeBlockRenderer(c.Stdlib, c.Path), 100),
		util.Prioritized(crenderer.NewConfluenceFencedCodeBlockRenderer(c.Stdlib, c, c.MarkConfig), 100),
		util.Prioritized(crenderer.NewConfluenceHTMLBlockRenderer(c.Stdlib), 100),
		util.Prioritized(crenderer.NewConfluenceHeadingRenderer(c.MarkConfig.DropFirstH1), 100),
		util.Prioritized(crenderer.NewConfluenceImageRenderer(c.Stdlib, c, c.Path), 100),
		util.Prioritized(crenderer.NewConfluenceParagraphRenderer(), 100),
		util.Prioritized(crenderer.NewConfluenceLinkRenderer(), 100),
	))

	if slices.Contains(c.MarkConfig.Features, "mkdocsadmonitions") {
		m.Parser().AddOptions(
			parser.WithBlockParsers(
				util.Prioritized(mkDocsParser.NewAdmonitionParser(), 100),
			),
		)

		m.Renderer().AddOptions(renderer.WithNodeRenderers(
			util.Prioritized(crenderer.NewConfluenceMkDocsAdmonitionRenderer(), 100),
		))
	}

	m.Parser().AddOptions(parser.WithInlineParsers(
		// Must be registered with a higher priority than goldmark's linkParser to make sure goldmark doesn't parse
		// the <ac:*/> tags.
		util.Prioritized(cparser.NewConfluenceTagParser(), 199),
	))
}

func CompileMarkdown(markdown []byte, stdlib *stdlib.Lib, path string, cfg types.MarkConfig) (string, []attachment.Attachment) {
	log.Tracef(nil, "rendering markdown:\n%s", string(markdown))

	confluenceExtension := NewConfluenceExtension(stdlib, path, cfg)

	converter := goldmark.New(
		goldmark.WithExtensions(
			extension.Footnote,
			extension.DefinitionList,
			extension.NewTable(
				extension.WithTableCellAlignMethod(extension.TableCellAlignStyle),
			),
			confluenceExtension,
			extension.GFM,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
			html.WithXHTML(),
		))

	ctx := parser.NewContext(parser.WithIDs(&cparser.ConfluenceIDs{Values: map[string]bool{}}))

	var buf bytes.Buffer
	err := converter.Convert(markdown, &buf, parser.WithContext(ctx))

	if err != nil {
		panic(err)
	}

	html := buf.Bytes()

	log.Tracef(nil, "rendered markdown to html:\n%s", string(html))

	return string(html), confluenceExtension.Attachments
}

// ConfluenceGHAlertsExtension is a goldmark extension for GitHub Alerts with Transformer approach
type ConfluenceGHAlertsExtension struct {
	*ConfluenceExtension
}

// NewConfluenceGHAlertsExtension creates a new instance of the GitHub Alerts extension
func NewConfluenceGHAlertsExtension(baseExtension *ConfluenceExtension) *ConfluenceGHAlertsExtension {
	return &ConfluenceGHAlertsExtension{
		ConfluenceExtension: baseExtension,
	}
}

// Extend extends the Goldmark processor with GitHub Alerts transformer and renderers
func (c *ConfluenceGHAlertsExtension) Extend(m goldmark.Markdown) {
	// First add the base extension functionality
	c.ConfluenceExtension.Extend(m)

	// Add the GitHub Alerts transformer
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(ctransformer.NewGHAlertsTransformer(), 100),
	))

	// Replace the blockquote renderer with the GitHub Alerts aware version
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(crenderer.NewConfluenceGHAlertsBlockQuoteRenderer(), 200), // Higher priority than base
	))

	// Add the text renderer that handles replacement content for GitHub Alerts
	if slices.Contains(c.ConfluenceExtension.MarkConfig.Features, "ghalerts") || slices.Contains(c.ConfluenceExtension.MarkConfig.Features, "gh-alerts") {
		m.Renderer().AddOptions(renderer.WithNodeRenderers(
			util.Prioritized(crenderer.NewConfluenceGHAlertsTextRenderer(c.ConfluenceExtension.MarkConfig.StripNewlines), 200), // Higher priority than base
		))
	}
}

// CompileMarkdownWithTransformer compiles markdown using the transformer approach for GitHub Alerts
func CompileMarkdownWithTransformer(markdown []byte, stdlib *stdlib.Lib, path string, cfg types.MarkConfig) (string, []attachment.Attachment) {
	log.Tracef(nil, "rendering markdown with transformer:\n%s", string(markdown))

	// Create base extension
	confluenceExtension := NewConfluenceExtension(stdlib, path, cfg)

	// Create GitHub Alerts extension that wraps the base extension
	ghAlertsExtension := NewConfluenceGHAlertsExtension(confluenceExtension)

	converter := goldmark.New(
		goldmark.WithExtensions(
			extension.Footnote,
			extension.DefinitionList,
			extension.NewTable(
				extension.WithTableCellAlignMethod(extension.TableCellAlignStyle),
			),
			ghAlertsExtension, // Use the GitHub Alerts extension instead of base
			extension.GFM,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
			html.WithXHTML(),
		))

	ctx := parser.NewContext(parser.WithIDs(&cparser.ConfluenceIDs{Values: map[string]bool{}}))

	var buf bytes.Buffer
	err := converter.Convert(markdown, &buf, parser.WithContext(ctx))

	if err != nil {
		panic(err)
	}

	html := buf.Bytes()

	log.Tracef(nil, "rendered markdown to html with transformer:\n%s", string(html))

	return string(html), confluenceExtension.Attachments
}
