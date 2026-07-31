package domain

import "testing"

func TestEncounterBudgetUsesSRD521Table(t *testing.T) {
	tests := []struct {
		level, party int
		difficulty   string
		budget       int
		tier         string
	}{
		{1, 4, "easy", 200, "low"},
		{3, 5, "medium", 1125, "moderate"},
		{5, 5, "medium", 3750, "moderate"},
		{15, 6, "hard", 46800, "high"},
		{20, 8, "deadly", 176000, "high"},
	}
	for _, test := range tests {
		budget, tier, ok := EncounterBudget(test.level, test.party, test.difficulty)
		if !ok || budget != test.budget || tier != test.tier {
			t.Fatalf("level %d %s: got %d %s", test.level, test.difficulty, budget, tier)
		}
	}
}

func TestBundledCatalogIsBilingualAndAttributed(t *testing.T) {
	catalog := BundledCatalog()
	if catalog.Version != RulesVersion || len(catalog.Items) < 30 || catalog.Attribution.Notice == "" {
		t.Fatal("catalog metadata is incomplete")
	}
	for _, item := range catalog.Items {
		if item.Index == "" || item.Name.PTBR == "" || item.Name.ENUS == "" || item.Source != SRDTitle || item.License != SRDLicense {
			t.Fatalf("invalid catalog item: %+v", item)
		}
	}
}
