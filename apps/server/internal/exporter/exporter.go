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
	fmt.Fprintf(&output, "- Seed: `%d`\n- Generator: `%s`\n- Rules: `%s`\n- Schema: `%s`\n\n", doc.Seed, doc.GeneratorVersion, doc.RulesVersion, doc.SchemaVersion)
	for _, floor := range doc.Floors {
		fmt.Fprintf(&output, "## %s / %s\n\n", floor.Name.PTBR, floor.Name.ENUS)
		for _, room := range floor.Rooms {
			fmt.Fprintf(&output, "### %s / %s\n\n%s\n\n%s\n\n", room.Name.PTBR, room.Name.ENUS, room.Description.PTBR, room.Description.ENUS)
		}
		writeFloorContent(&output, doc, floor.ID)
	}
	output.WriteString("## Atribuições / Attributions\n\n")
	for _, attribution := range doc.Attributions {
		fmt.Fprintf(&output, "- **%s** — %s, %s. %s\n", attribution.Title, attribution.Creator, attribution.License, attribution.Notice)
	}
	return []byte(output.String())
}

func writeFloorContent(output *strings.Builder, doc domain.AdventureDocument, floorID string) {
	for _, encounter := range doc.Encounters {
		if encounter.FloorID != floorID {
			continue
		}
		fmt.Fprintf(output, "#### Encontro / Encounter — %s\n\n", encounter.Difficulty)
		for _, creature := range encounter.Creatures {
			fmt.Fprintf(output, "- %d× %s / %s (CR %.2g, %d XP cada / each)\n", creature.Count, creature.Name.PTBR, creature.Name.ENUS, creature.CR, creature.XP)
		}
		fmt.Fprintf(output, "\nBudget: %d XP · Total: %d XP\n\n", encounter.BudgetXP, encounter.TotalXP)
	}
	for _, puzzle := range doc.Puzzles {
		if puzzle.FloorID != floorID {
			continue
		}
		fmt.Fprintf(output, "#### %s / %s (DC %d)\n\n%s\n\n%s\n\n**Solução / Solution:** %s / %s\n\n", puzzle.Name.PTBR, puzzle.Name.ENUS, puzzle.CheckDC, puzzle.Prompt.PTBR, puzzle.Prompt.ENUS, puzzle.Solution.PTBR, puzzle.Solution.ENUS)
	}
	for _, trap := range doc.Traps {
		if trap.FloorID != floorID {
			continue
		}
		fmt.Fprintf(output, "#### %s / %s\n\n%s / %s\n\n- Tier: %s\n- Save DC: %d\n- Dano / Damage: %s / %s\n\n", trap.Name.PTBR, trap.Name.ENUS, trap.Trigger.PTBR, trap.Trigger.ENUS, trap.LevelTier, trap.SaveDC, trap.Damage.PTBR, trap.Damage.ENUS)
	}
	for _, treasure := range doc.Treasures {
		if treasure.FloorID != floorID {
			continue
		}
		fmt.Fprintf(output, "#### %s / %s\n\n%s\n\n%s\n\n**%d GP · %s**\n\n", treasure.Name.PTBR, treasure.Name.ENUS, treasure.Description.PTBR, treasure.Description.ENUS, treasure.ValueGP, treasure.Quality)
	}
	for _, rest := range doc.RestPoints {
		if rest.FloorID != floorID {
			continue
		}
		fmt.Fprintf(output, "#### %s / %s\n\n%s\n\n%s\n\n", rest.Name.PTBR, rest.Name.ENUS, rest.Description.PTBR, rest.Description.ENUS)
	}
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
.en{color:#5d5865}.content{break-inside:avoid;border:1px solid #ddd;padding:12px;margin:12px 0}.attribution{font-size:12px;border-top:1px solid #bbb;margin-top:40px;padding-top:16px}
@media print{body{padding:0}.floor:first-of-type{break-before:auto}}
</style>
</head>
<body>
<h1>{{.Name.PTBR}}<br><span class="en">{{.Name.ENUS}}</span></h1>
<p>{{.Summary.PTBR}}</p><p class="en">{{.Summary.ENUS}}</p>
<p class="meta">Seed {{.Seed}} · {{.GeneratorVersion}} · {{.RulesVersion}}</p>
{{range .Floors}}<section class="floor"><h2>{{.Name.PTBR}} / <span class="en">{{.Name.ENUS}}</span></h2>
{{range .Rooms}}<article class="room"><h3>{{.Name.PTBR}} / <span class="en">{{.Name.ENUS}}</span></h3>
<p>{{.Description.PTBR}}</p><p class="en">{{.Description.ENUS}}</p></article>{{end}}</section>{{end}}
<section><h2>Encontros / Encounters</h2>{{range .Encounters}}<article class="content"><h3>{{.Difficulty}} · {{.TotalXP}}/{{.BudgetXP}} XP</h3>{{range .Creatures}}<p>{{.Count}}× {{.Name.PTBR}} / <span class="en">{{.Name.ENUS}}</span> (CR {{.CR}})</p>{{end}}</article>{{end}}</section>
<section><h2>Puzzles</h2>{{range .Puzzles}}<article class="content"><h3>{{.Name.PTBR}} / <span class="en">{{.Name.ENUS}}</span> · DC {{.CheckDC}}</h3><p>{{.Prompt.PTBR}}</p><p class="en">{{.Prompt.ENUS}}</p><p><strong>Solução / Solution:</strong> {{.Solution.PTBR}} / <span class="en">{{.Solution.ENUS}}</span></p></article>{{end}}</section>
<section><h2>Armadilhas / Traps</h2>{{range .Traps}}<article class="content"><h3>{{.Name.PTBR}} / <span class="en">{{.Name.ENUS}}</span></h3><p>{{.Trigger.PTBR}}</p><p class="en">{{.Trigger.ENUS}}</p><p>{{.LevelTier}} · DC {{.SaveDC}} · {{.Damage.PTBR}} / <span class="en">{{.Damage.ENUS}}</span></p></article>{{end}}</section>
<section><h2>Tesouros / Treasures</h2>{{range .Treasures}}<article class="content"><h3>{{.Name.PTBR}} / <span class="en">{{.Name.ENUS}}</span></h3><p>{{.Description.PTBR}}</p><p class="en">{{.Description.ENUS}}</p><p>{{.ValueGP}} GP · {{.Quality}}</p></article>{{end}}</section>
<section><h2>Descansos / Rests</h2>{{range .RestPoints}}<article class="content"><h3>{{.Name.PTBR}} / <span class="en">{{.Name.ENUS}}</span></h3><p>{{.Description.PTBR}}</p><p class="en">{{.Description.ENUS}}</p></article>{{end}}</section>
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
