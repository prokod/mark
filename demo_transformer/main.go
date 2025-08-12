package main

import (
	"fmt"
	"os"
	"text/template"

	mark "github.com/kovetskiy/mark/markdown"
	"github.com/kovetskiy/mark/stdlib"
	"github.com/kovetskiy/mark/types"
)

func main() {
	// Test markdown with GitHub Alerts
	testMarkdown := `# Test GitHub Alerts

## Note Alert

> [!NOTE]
> This is a note alert with **markdown** formatting.
> 
> - Item 1
> - Item 2

## Tip Alert

> [!TIP]
> This is a tip alert.

## Warning Alert  

> [!WARNING]
> This is a warning alert.

## Regular Blockquote

> This is a regular blockquote without GitHub Alert syntax.
`

	stdlib := &stdlib.Lib{
		Templates: template.New("test"),
	}

	cfg := types.MarkConfig{
		Features:      []string{"ghalerts"},
		StripNewlines: false,
		DropFirstH1:   false,
	}

	fmt.Println("=== Testing Transformer Approach ===")
	transformerResult, _ := mark.CompileMarkdownWithTransformer([]byte(testMarkdown), stdlib, "/test", cfg)
	fmt.Println(transformerResult)

	fmt.Println("\n=== Testing Original Approach ===")
	originalResult, _ := mark.CompileMarkdown([]byte(testMarkdown), stdlib, "/test", cfg)
	fmt.Println(originalResult)

	fmt.Println("\n=== Comparison Summary ===")
	if len(transformerResult) > 0 && len(originalResult) > 0 {
		fmt.Println("✅ Both approaches successfully rendered the markdown")

		// Check for key differences
		fmt.Println("\nKey Differences:")
		fmt.Println("- Original approach preserves '[!NOTE]' syntax:", contains(originalResult, "[!NOTE]"))
		fmt.Println("- Transformer approach uses friendly text:", contains(transformerResult, " Note "))
		fmt.Println("- Both generate Confluence macros:",
			contains(originalResult, "ac:structured-macro") && contains(transformerResult, "ac:structured-macro"))
	} else {
		fmt.Println("❌ One or both approaches failed to render")
		os.Exit(1)
	}
}

func contains(text, substr string) bool {
	return len(text) > 0 && len(substr) > 0 &&
		(len(text) >= len(substr) && findSubstring(text, substr))
}

func findSubstring(text, substr string) bool {
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
