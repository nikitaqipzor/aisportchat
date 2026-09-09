package aifitness

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type OpenAIProvider struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	if strings.TrimSpace(model) == "" {
		model = "gpt-5.6-terra"
	}
	return &OpenAIProvider{apiKey: strings.TrimSpace(apiKey), model: model, baseURL: "https://api.openai.com/v1", client: &http.Client{Timeout: 45 * time.Second}}
}
func (p *OpenAIProvider) Name() string  { return "openai" }
func (p *OpenAIProvider) Model() string { return p.model }

type responseEnvelope struct {
	Output []json.RawMessage `json:"output"`
	Model  string            `json:"model"`
}
type responseItem struct {
	Type      string `json:"type"`
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
	CallID    string `json:"call_id,omitempty"`
	Content   []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content,omitempty"`
}

func (p *OpenAIProvider) postResponses(ctx context.Context, payload map[string]any) (responseEnvelope, error) {
	if p.apiKey == "" {
		return responseEnvelope{}, errors.New("OPENAI_API_KEY is not configured")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return responseEnvelope{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/responses", bytes.NewReader(body))
	if err != nil {
		return responseEnvelope{}, err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return responseEnvelope{}, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return responseEnvelope{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return responseEnvelope{}, fmt.Errorf("openai responses API: status=%d body=%s", resp.StatusCode, truncate(string(raw), 1000))
	}
	var out responseEnvelope
	if err := json.Unmarshal(raw, &out); err != nil {
		return responseEnvelope{}, fmt.Errorf("decode openai response: %w", err)
	}
	return out, nil
}

func foodTool() map[string]any {
	return map[string]any{
		"type": "function", "name": "submit_food_items", "description": "Return every food or drink item detected in the user's input. Quantity must be edible grams; estimate only when needed.", "strict": true,
		"parameters": map[string]any{"type": "object", "properties": map[string]any{
			"items": map[string]any{"type": "array", "items": map[string]any{"type": "object", "properties": map[string]any{
				"name": map[string]any{"type": "string"}, "quantity_g": map[string]any{"type": "number"}, "confidence": map[string]any{"type": "number"}, "notes": map[string]any{"type": "string"},
			}, "required": []string{"name", "quantity_g", "confidence", "notes"}, "additionalProperties": false}},
		}, "required": []string{"items"}, "additionalProperties": false},
	}
}

func (p *OpenAIProvider) extract(ctx context.Context, content any) ([]ExtractedFood, error) {
	payload := map[string]any{"model": p.model, "store": false, "instructions": "You extract foods for a fitness diary. Return foods in the user's language when possible. Never invent nutrition values. Estimate portion grams conservatively and say so in notes. Use the submit_food_items tool exactly once.", "input": content, "tools": []any{foodTool()}, "tool_choice": map[string]any{"type": "function", "name": "submit_food_items"}}
	resp, err := p.postResponses(ctx, payload)
	if err != nil {
		return nil, err
	}
	for _, raw := range resp.Output {
		var item responseItem
		if json.Unmarshal(raw, &item) == nil && item.Type == "function_call" && item.Name == "submit_food_items" {
			var args struct {
				Items []ExtractedFood `json:"items"`
			}
			if err := json.Unmarshal([]byte(item.Arguments), &args); err != nil {
				return nil, err
			}
			if len(args.Items) == 0 {
				return nil, errors.New("no foods detected")
			}
			return args.Items, nil
		}
	}
	return nil, errors.New("OpenAI did not return food extraction tool output")
}

func (p *OpenAIProvider) ExtractFoodText(ctx context.Context, text string) ([]ExtractedFood, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("text is required")
	}
	return p.extract(ctx, []any{map[string]any{"role": "user", "content": text}})
}
func (p *OpenAIProvider) ExtractFoodImage(ctx context.Context, imageDataURL string) ([]ExtractedFood, error) {
	imageDataURL = strings.TrimSpace(imageDataURL)
	if !strings.HasPrefix(imageDataURL, "data:image/") {
		return nil, errors.New("image_data_url must be a data:image URL")
	}
	if len(imageDataURL) > 7_000_000 {
		return nil, errors.New("image is too large")
	}
	input := []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "input_text", "text": "Identify the foods and estimate edible portion grams. Return every visible food item."}, map[string]any{"type": "input_image", "image_url": imageDataURL, "detail": "auto"}}}}
	return p.extract(ctx, input)
}

func coachTools(defs []ToolDefinition) []any {
	out := make([]any, 0, len(defs))
	for _, d := range defs {
		out = append(out, map[string]any{"type": "function", "name": d.Name, "description": d.Description, "parameters": d.Parameters, "strict": true})
	}
	return out
}

func (p *OpenAIProvider) Coach(ctx context.Context, req CoachRequest) (CoachResponse, error) {
	input := make([]any, 0, len(req.History)+2)
	for _, m := range req.History {
		role := m.Role
		if role != "assistant" {
			role = "user"
		}
		input = append(input, map[string]any{"role": role, "content": m.Content})
	}
	input = append(input, map[string]any{"role": "user", "content": req.Message})
	tools := coachTools(req.Tools)
	called := []string{}
	instructions := "You are AI Coach inside a fitness application. Use tools whenever user-specific facts are needed. Do not diagnose disease or present estimates as medical facts. Training and nutrition calculations from tools are authoritative; do not recalculate or silently change targets. Be concise, practical, and answer in the user's language. If pain, alarming symptoms, or a medical question is raised, recommend appropriate professional evaluation rather than diagnosis."
	for round := 0; round < 4; round++ {
		payload := map[string]any{"model": p.model, "store": false, "instructions": instructions, "input": input, "tools": tools, "tool_choice": "auto"}
		resp, err := p.postResponses(ctx, payload)
		if err != nil {
			return CoachResponse{}, err
		}
		toolCount := 0
		for _, raw := range resp.Output {
			input = append(input, raw)
			var item responseItem
			if json.Unmarshal(raw, &item) != nil {
				continue
			}
			if item.Type == "function_call" {
				toolCount++
				called = append(called, item.Name)
				result, execErr := req.ExecuteTool(ctx, item.Name, json.RawMessage(item.Arguments))
				if execErr != nil {
					result = `{"error":"tool failed"}`
				}
				input = append(input, map[string]any{"type": "function_call_output", "call_id": item.CallID, "output": result})
			}
		}
		if toolCount == 0 {
			text := outputText(resp.Output)
			if strings.TrimSpace(text) == "" {
				return CoachResponse{}, errors.New("empty coach response")
			}
			return CoachResponse{Message: text, Model: p.model, Provider: p.Name(), ToolCalls: unique(called)}, nil
		}
	}
	return CoachResponse{}, errors.New("AI coach exceeded tool-call rounds")
}

func weeklyReportTool() map[string]any {
	return map[string]any{"type": "function", "name": "submit_weekly_report", "description": "Return a concise weekly fitness report based only on supplied stats.", "strict": true, "parameters": map[string]any{"type": "object", "properties": map[string]any{
		"summary": map[string]any{"type": "string"}, "wins": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}, "focus": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}, "next_actions": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
	}, "required": []string{"summary", "wins", "focus", "next_actions"}, "additionalProperties": false}}
}
func (p *OpenAIProvider) WeeklyReport(ctx context.Context, stats WeeklyStats) (WeeklyReport, error) {
	rawStats, _ := json.Marshal(stats)
	payload := map[string]any{"model": p.model, "store": false, "instructions": "You create a weekly fitness report. Use only the supplied numeric facts. Do not diagnose health conditions. Distinguish missing data from poor performance. Keep next actions realistic and short.", "input": []any{map[string]any{"role": "user", "content": "Weekly stats JSON: " + string(rawStats)}}, "tools": []any{weeklyReportTool()}, "tool_choice": map[string]any{"type": "function", "name": "submit_weekly_report"}}
	resp, err := p.postResponses(ctx, payload)
	if err != nil {
		return WeeklyReport{}, err
	}
	for _, raw := range resp.Output {
		var item responseItem
		if json.Unmarshal(raw, &item) == nil && item.Type == "function_call" && item.Name == "submit_weekly_report" {
			var r WeeklyReport
			if err := json.Unmarshal([]byte(item.Arguments), &r); err != nil {
				return WeeklyReport{}, err
			}
			r.Stats = stats
			r.Model = p.model
			r.Provider = p.Name()
			return r, nil
		}
	}
	return WeeklyReport{}, errors.New("OpenAI did not return weekly report tool output")
}

func outputText(items []json.RawMessage) string {
	var b strings.Builder
	for _, raw := range items {
		var item responseItem
		if json.Unmarshal(raw, &item) != nil || item.Type != "message" {
			continue
		}
		for _, c := range item.Content {
			if c.Type == "output_text" {
				if b.Len() > 0 {
					b.WriteString("\n")
				}
				b.WriteString(c.Text)
			}
		}
	}
	return b.String()
}
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
func unique(in []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, v := range in {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}
