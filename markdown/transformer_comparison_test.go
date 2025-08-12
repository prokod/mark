package mark_test

import (
	"testing"
	"text/template"

	mark "github.com/kovetskiy/mark/markdown" // Import the markdown package for the transformer
	"github.com/kovetskiy/mark/stdlib"
	"github.com/kovetskiy/mark/types"
	"github.com/stretchr/testify/assert"
)

func TestBasicTransformerFunctionality(t *testing.T) {
	testMarkdown := "> [!NOTE]\n> This is a test note."

	stdlib := &stdlib.Lib{
		Templates: template.New("test"),
	}

	cfg := types.MarkConfig{
		Features:      []string{},
		StripNewlines: false,
		DropFirstH1:   false,
	}

	result, attachments := mark.CompileMarkdownWithTransformer([]byte(testMarkdown), stdlib, "/test", cfg)

	// Basic checks
	assert.NotEmpty(t, result)
	assert.Empty(t, attachments)
	assert.Contains(t, result, "structured-macro")

	t.Logf("Transformer result: %s", result)
}
