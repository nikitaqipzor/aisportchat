package aifitness

import (
	"context"
	"strings"
	"testing"
)

func TestLocalProviderKeepsQuantitiesAttachedToNearestFood(t *testing.T) {
	p := NewLocalProvider()
	items, err := p.ExtractFoodText(context.Background(), "Съел 200 г творога, банан и 30 г протеина")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("items=%d want=3: %+v", len(items), items)
	}
	want := []float64{200, 120, 30}
	for i := range want {
		if items[i].QuantityG != want[i] {
			t.Fatalf("item %d %q quantity=%v want=%v", i, items[i].Name, items[i].QuantityG, want[i])
		}
	}
}

func TestLocalCoachCanExplainReadinessContext(t *testing.T) {
	p := NewLocalProvider()
	out, err := p.Coach(context.Background(), CoachRequest{Message: "Какая у меня готовность сегодня?", BaseContext: FitnessContext{Readiness: map[string]any{"score": 72}}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.Message, "72/100") {
		t.Fatalf("unexpected answer: %s", out.Message)
	}
}
