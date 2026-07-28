package exporter

import (
	"strings"
	"testing"
	"time"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/generator"
)

func TestPrintableExportsContainBothLanguagesAndAttribution(t *testing.T) {
	doc, err := generator.Generate(domain.DefaultAdventureSpec(), 31, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	markdown := string(Markdown(doc))
	html, err := HTML(doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, content := range []string{markdown, string(html)} {
		if !strings.Contains(content, doc.Name.PTBR) ||
			!strings.Contains(content, doc.Name.ENUS) ||
			!strings.Contains(content, "CC-BY-4.0") {
			t.Fatal("export is missing bilingual content or attribution")
		}
	}
}
