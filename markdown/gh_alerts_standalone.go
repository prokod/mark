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

// StandaloneGHAlertsExtension is a self-contained GitHub Alerts extension that doesn't depend on feature flags
type StandaloneGHAlertsExtension struct {
	html.Config
	Stdlib      *stdlib.Lib
	Path        string
	MarkConfig  types.MarkConfig
	Attachments []attachment.Attachment
}

// NewStandaloneGHAlertsExtension creates a new instance of the standalone GitHub Alerts extension
func NewStandaloneGHAlertsExtension(stdlib *stdlib.Lib, path string, cfg types.MarkConfig) *StandaloneGHAlertsExtension {
	return &StandaloneGHAlertsExtension{
		Config:      html.NewConfig(),
		Stdlib:      stdlib,
		Path:        path,
		MarkConfig:  cfg,
		Attachments: []attachment.Attachment{},
	}
}

func (c *StandaloneGHAlertsExtension) Attach(a attachment.Attachment) {
	c.Attachments = append(c.Attachments, a)
}

func (c *StandaloneGHAlertsExtension) Extend(m goldmark.Markdown) {
	// Register core renderers (excluding blockquote and text which we'll replace)
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(crenderer.NewConfluenceCodeBlockRenderer(c.Stdlib, c.Path), 100),
		util.Prioritized(crenderer.NewConfluenceFencedCodeBlockRenderer(c.Stdlib, c, c.MarkConfig), 100),
		util.Prioritized(crenderer.NewConfluenceHTMLBlockRenderer(c.Stdlib), 100),
		util.Prioritized(crenderer.NewConfluenceHeadingRenderer(c.MarkConfig.DropFirstH1), 100),
		util.Prioritized(crenderer.NewConfluenceImageRenderer(c.Stdlib, c, c.Path), 100),
		util.Prioritized(crenderer.NewConfluenceParagraphRenderer(), 100),
		util.Prioritized(crenderer.NewConfluenceLinkRenderer(), 100),
	))

	// Add GitHub Alerts specific renderers with higher priority
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(crenderer.NewConfluenceGHAlertsBlockQuoteRenderer(), 200),
		util.Prioritized(crenderer.NewConfluenceTextRenderer(c.MarkConfig.StripNewlines), 200),
	)) // Add the GitHub Alerts transformer
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(ctransformer.NewGHAlertsTransformer(), 100),
	))

	// Add mkdocsadmonitions if requested
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

	// Add confluence tag parser
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(cparser.NewConfluenceTagParser(), 199),
	))
}

// CompileMarkdownWithGHAlertsTransformer compiles markdown using the standalone GitHub Alerts transformer
func CompileMarkdownWithGHAlertsTransformer(markdown []byte, stdlib *stdlib.Lib, path string, cfg types.MarkConfig) (string, []attachment.Attachment) {
	log.Tracef(nil, "rendering markdown with standalone GH Alerts transformer:\n%s", string(markdown))

	// Create the standalone GitHub Alerts extension
	ghAlertsExtension := NewStandaloneGHAlertsExtension(stdlib, path, cfg)

	converter := goldmark.New(
		goldmark.WithExtensions(
			extension.Footnote,
			extension.DefinitionList,
			extension.NewTable(
				extension.WithTableCellAlignMethod(extension.TableCellAlignStyle),
			),
			ghAlertsExtension,
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

	log.Tracef(nil, "rendered markdown to html with standalone GH Alerts transformer:\n%s", string(html))

	return string(html), ghAlertsExtension.Attachments
}

// CompileMarkdownWithLegacyRenderer compiles markdown using the legacy blockquote renderer approach
func CompileMarkdownWithLegacyRenderer(markdown []byte, stdlib *stdlib.Lib, path string, cfg types.MarkConfig) (string, []attachment.Attachment) {
	log.Tracef(nil, "rendering markdown with legacy renderer:\n%s", string(markdown))

	// Create the original extension (this uses the legacy renderer in blockquote.go)
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

	log.Tracef(nil, "rendered markdown to html with legacy renderer:\n%s", string(html))

	return string(html), confluenceExtension.Attachments
}
