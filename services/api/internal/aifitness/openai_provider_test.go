package aifitness

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAIProviderFoodExtractionUsesForcedTool(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("authorization=%q", got)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload["model"] != "test-model" {
			t.Fatalf("model=%v", payload["model"])
		}
		choice, ok := payload["tool_choice"].(map[string]any)
		if !ok || choice["name"] != "submit_food_items" {
			t.Fatalf("tool_choice=%#v", payload["tool_choice"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"test-model","output":[{"type":"function_call","call_id":"fc1","name":"submit_food_items","arguments":"{\"items\":[{\"name\":\"Банан\",\"quantity_g\":120,\"confidence\":0.92,\"notes\":\"estimated\"}]}"}]}`))
	}))
	defer srv.Close()

	p := NewOpenAIProvider("test-key", "test-model")
	p.baseURL = srv.URL
	items, err := p.ExtractFoodText(context.Background(), "банан")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Name != "Банан" || items[0].QuantityG != 120 {
		t.Fatalf("items=%+v", items)
	}
}

func TestOpenAIProviderCoachExecutesToolLoop(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		if calls == 1 {
			_, _ = w.Write([]byte(`{"model":"test-model","output":[{"type":"function_call","call_id":"fc-nutrition","name":"get_today_nutrition","arguments":"{}"}]}`))
			return
		}
		raw, _ := json.Marshal(payload["input"])
		if !strings.Contains(string(raw), "function_call_output") || !strings.Contains(string(raw), "remaining_protein_g") {
			t.Fatalf("second input does not contain tool result: %s", raw)
		}
		_, _ = w.Write([]byte(`{"model":"test-model","output":[{"type":"message","content":[{"type":"output_text","text":"Осталось 42 г белка."}]}]}`))
	}))
	defer srv.Close()

	p := NewOpenAIProvider("test-key", "test-model")
	p.baseURL = srv.URL
	toolExecutions := 0
	res, err := p.Coach(context.Background(), CoachRequest{
		Message: "Сколько белка осталось?",
		Tools:   []ToolDefinition{{Name: "get_today_nutrition", Description: "nutrition", Parameters: emptyObjectSchema()}},
		ExecuteTool: func(_ context.Context, name string, _ json.RawMessage) (string, error) {
			toolExecutions++
			if name != "get_today_nutrition" {
				t.Fatalf("tool=%s", name)
			}
			return `{"remaining_protein_g":42}`, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Message != "Осталось 42 г белка." || toolExecutions != 1 || calls != 2 {
		t.Fatalf("res=%+v executions=%d calls=%d", res, toolExecutions, calls)
	}
}
