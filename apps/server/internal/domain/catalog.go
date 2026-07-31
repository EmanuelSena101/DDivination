package domain

const (
	SRDTitle   = "System Reference Document 5.2.1"
	SRDSource  = "https://www.dndbeyond.com/srd"
	SRDLicense = "CC-BY-4.0"
)

type CatalogItem struct {
	Index    string        `json:"index"`
	Name     LocalizedText `json:"name"`
	Kind     string        `json:"kind" enum:"monster,trap"`
	CR       float64       `json:"cr,omitempty"`
	XP       int           `json:"xp,omitempty"`
	MinLevel int           `json:"minLevel,omitempty"`
	MaxLevel int           `json:"maxLevel,omitempty"`
	Source   string        `json:"source"`
	License  string        `json:"license"`
}

type Catalog struct {
	Version     string        `json:"version"`
	Ruleset     string        `json:"ruleset"`
	Items       []CatalogItem `json:"items"`
	Attribution Attribution   `json:"attribution"`
}

var catalogItems = []CatalogItem{
	monster("skeleton", "Esqueleto", "Skeleton", .25, 50),
	monster("goblin-warrior", "Guerreiro Goblin", "Goblin Warrior", .25, 50),
	monster("tough", "Valentão", "Tough", .5, 100),
	monster("bugbear-warrior", "Guerreiro Bugbear", "Bugbear Warrior", 1, 200),
	monster("ogre", "Ogro", "Ogre", 2, 450),
	monster("wight", "Inumano", "Wight", 3, 700),
	monster("succubus", "Súcubo", "Succubus", 4, 1100),
	monster("troll", "Troll", "Troll", 5, 1800),
	monster("mage", "Mago", "Mage", 6, 2300),
	monster("giant-ape", "Gorila Gigante", "Giant Ape", 7, 2900),
	monster("assassin", "Assassino", "Assassin", 8, 3900),
	monster("fire-giant", "Gigante do Fogo", "Fire Giant", 9, 5000),
	monster("stone-golem", "Golem de Pedra", "Stone Golem", 10, 5900),
	monster("remorhaz", "Remorhaz", "Remorhaz", 11, 7200),
	monster("archmage", "Arquimago", "Archmage", 12, 8400),
	monster("storm-giant", "Gigante da Tempestade", "Storm Giant", 13, 10000),
	monster("adult-black-dragon", "Dragão Negro Adulto", "Adult Black Dragon", 14, 11500),
	monster("mummy-lord", "Senhor das Múmias", "Mummy Lord", 15, 13000),
	monster("adult-blue-dragon", "Dragão Azul Adulto", "Adult Blue Dragon", 16, 15000),
	monster("adult-red-dragon", "Dragão Vermelho Adulto", "Adult Red Dragon", 17, 18000),
	monster("demilich", "Demilich", "Demilich", 18, 20000),
	monster("balor", "Balor", "Balor", 19, 22000),
	monster("pit-fiend", "Diabo do Fosso", "Pit Fiend", 20, 25000),
	monster("solar", "Solar", "Solar", 21, 33000),
	monster("ancient-green-dragon", "Dragão Verde Ancião", "Ancient Green Dragon", 22, 41000),
	monster("ancient-blue-dragon", "Dragão Azul Ancião", "Ancient Blue Dragon", 23, 50000),
	monster("ancient-red-dragon", "Dragão Vermelho Ancião", "Ancient Red Dragon", 24, 62000),
	monster("tarrasque", "Tarrasque", "Tarrasque", 30, 155000),
	trapCatalogItem("collapsing-roof", "Teto Desabando", "Collapsing Roof"),
	trapCatalogItem("poisoned-needle", "Agulha Envenenada", "Poisoned Needle"),
}

var xpBudgets = [20][3]int{
	{50, 75, 100}, {100, 150, 200}, {150, 225, 400}, {250, 375, 500},
	{500, 750, 1100}, {600, 1000, 1400}, {750, 1300, 1700}, {1000, 1700, 2100},
	{1300, 2000, 2600}, {1600, 2300, 3100}, {1900, 2900, 4100}, {2200, 3700, 4700},
	{2600, 4200, 5400}, {2900, 4900, 6200}, {3300, 5400, 7800}, {3800, 6100, 9800},
	{4500, 7200, 11700}, {5000, 8700, 14200}, {5500, 10700, 17200}, {6400, 13200, 22000},
}

func monster(index, ptBR, enUS string, cr float64, xp int) CatalogItem {
	return CatalogItem{Index: index, Name: LocalizedText{PTBR: ptBR, ENUS: enUS}, Kind: "monster", CR: cr, XP: xp, Source: SRDTitle, License: SRDLicense}
}

func trapCatalogItem(index, ptBR, enUS string) CatalogItem {
	return CatalogItem{Index: index, Name: LocalizedText{PTBR: ptBR, ENUS: enUS}, Kind: "trap", MinLevel: 1, MaxLevel: 20, Source: SRDTitle, License: SRDLicense}
}

func SRDAttribution() Attribution {
	return Attribution{
		Title:   SRDTitle,
		Creator: "Wizards of the Coast LLC",
		Source:  SRDSource,
		License: SRDLicense,
		Notice:  "This work includes material from the System Reference Document 5.2.1 (“SRD 5.2.1”) by Wizards of the Coast LLC, available at https://www.dndbeyond.com/srd. The SRD 5.2.1 is licensed under the Creative Commons Attribution 4.0 International License, available at https://creativecommons.org/licenses/by/4.0/legalcode.",
	}
}

func BundledCatalog() Catalog {
	items := append([]CatalogItem(nil), catalogItems...)
	return Catalog{Version: RulesVersion, Ruleset: "5E 2024 / SRD 5.2.1", Items: items, Attribution: SRDAttribution()}
}

func CatalogItemByIndex(index string) (CatalogItem, bool) {
	for _, item := range catalogItems {
		if item.Index == index {
			return item, true
		}
	}
	return CatalogItem{}, false
}

func MonsterCatalogItems() []CatalogItem {
	items := make([]CatalogItem, 0, len(catalogItems))
	for _, item := range catalogItems {
		if item.Kind == "monster" {
			items = append(items, item)
		}
	}
	return items
}

func EncounterBudget(level, partySize int, difficulty string) (int, string, bool) {
	if level < 1 || level > 20 || partySize < 1 || partySize > 8 {
		return 0, "", false
	}
	column := 0
	tier := "low"
	switch difficulty {
	case "easy":
	case "medium":
		column, tier = 1, "moderate"
	case "hard", "deadly":
		column, tier = 2, "high"
	default:
		return 0, "", false
	}
	return xpBudgets[level-1][column] * partySize, tier, true
}

func LevelTier(level int) string {
	switch {
	case level <= 4:
		return "1-4"
	case level <= 10:
		return "5-10"
	case level <= 16:
		return "11-16"
	default:
		return "17-20"
	}
}
