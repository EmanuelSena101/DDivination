package exporter

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
)

func Markdown(doc domain.AdventureDocument) []byte {
	var output strings.Builder
	fmt.Fprintf(&output, "# %s / %s\n\n", doc.Name.PTBR, doc.Name.ENUS)
	fmt.Fprintf(&output, "> %s\n>\n> %s\n\n", doc.Summary.PTBR, doc.Summary.ENUS)
	fmt.Fprintf(&output, "- Seed: `%d`\n- Generator: `%s`\n- Schema: `%s`\n\n", doc.Seed, doc.GeneratorVersion, doc.SchemaVersion)
	for _, floor := range doc.Floors {
		fmt.Fprintf(&output, "## %s / %s\n\n", floor.Name.PTBR, floor.Name.ENUS)
		for _, room := range floor.Rooms {
			fmt.Fprintf(&output, "### %s / %s\n\n%s\n\n%s\n\n", room.Name.PTBR, room.Name.ENUS, room.Description.PTBR, room.Description.ENUS)
		}
	}
	output.WriteString("## Atribuições / Attributions\n\n")
	for _, attribution := range doc.Attributions {
		fmt.Fprintf(&output, "- **%s** — %s, %s. %s\n", attribution.Title, attribution.Creator, attribution.License, attribution.Notice)
	}
	return []byte(output.String())
}

var printable = template.Must(template.New("adventure").Parse(`<!doctype html>
<html lang="pt-BR">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>{{.Name.PTBR}} — DDivination</title>
<style>
body{max-width:850px;margin:0 auto;padding:40px;font:16px/1.55 Georgia,serif;color:#24222a}
h1,h2,h3{line-height:1.2;color:#2d2350}h1{font-size:34px;border-bottom:3px solid #7359d9;padding-bottom:12px}
.meta{color:#6c6774}.floor{break-before:page}.room{break-inside:avoid;border-left:3px solid #d8c57b;padding-left:16px;margin:22px 0}
.en{color:#5d5865}.attribution{font-size:12px;border-top:1px solid #bbb;margin-top:40px;padding-top:16px}
@media print{body{padding:0}.floor:first-of-type{break-before:auto}}
</style>
</head>
<body>
<h1>{{.Name.PTBR}}<br><span class="en">{{.Name.ENUS}}</span></h1>
<p>{{.Summary.PTBR}}</p><p class="en">{{.Summary.ENUS}}</p>
<p class="meta">Seed {{.Seed}} · {{.GeneratorVersion}}</p>
{{range .Floors}}<section class="floor"><h2>{{.Name.PTBR}} / <span class="en">{{.Name.ENUS}}</span></h2>
{{range .Rooms}}<article class="room"><h3>{{.Name.PTBR}} / <span class="en">{{.Name.ENUS}}</span></h3>
<p>{{.Description.PTBR}}</p><p class="en">{{.Description.ENUS}}</p></article>{{end}}</section>{{end}}
<section class="attribution"><h2>Atribuições / Attributions</h2>
{{range .Attributions}}<p><strong>{{.Title}}</strong> — {{.Creator}}, {{.License}}.<br>{{.Notice}}</p>{{end}}</section>
</body></html>`))

func HTML(doc domain.AdventureDocument) ([]byte, error) {
	var output bytes.Buffer
	if err := printable.Execute(&output, doc); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
