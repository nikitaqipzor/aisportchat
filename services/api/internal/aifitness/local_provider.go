package aifitness

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type LocalProvider struct{}

func NewLocalProvider() *LocalProvider { return &LocalProvider{} }
func (p *LocalProvider) Name() string  { return "local" }
func (p *LocalProvider) Model() string { return "deterministic-local-v1" }

var quantityPattern = regexp.MustCompile(`(?i)(\d+(?:[\.,]\d+)?)\s*(?:г|гр|грамм|граммов|g)`)
var wordOrNumberPattern = regexp.MustCompile(`[\p{L}\p{N}]`)

func (p *LocalProvider) ExtractFoodText(_ context.Context, text string) ([]ExtractedFood, error) {
	lower := strings.ToLower(text)
	aliases := []struct {
		Name    string
		Keys    []string
		Default float64
	}{
		{"Творог 5%", []string{"творог"}, 200},
		{"Банан", []string{"банан"}, 120},
		{"Сырники", []string{"сырник"}, 200},
		{"Куриная грудка", []string{"куриц", "грудк"}, 150},
		{"Рис белый варёный", []string{"рис"}, 200},
		{"Овсяные хлопья сухие", []string{"овсян", "овсянк"}, 80},
		{"Яйцо куриное", []string{"яйц"}, 110},
		{"Сывороточный протеин", []string{"протеин"}, 30},
		{"Макароны варёные", []string{"макарон", "паст"}, 200},
		{"Лосось", []string{"лосос"}, 150},
	}
	results := make([]ExtractedFood, 0)
	for _, a := range aliases {
		pos := -1
		for _, key := range a.Keys {
			if idx := strings.Index(lower, key); idx >= 0 {
				pos = idx
				break
			}
		}
		if pos < 0 {
			continue
		}
		q := a.Default
		start := pos - 28
		if start < 0 {
			start = 0
		}
		segment := lower[start:pos]
		matches := quantityPattern.FindAllStringSubmatchIndex(segment, -1)
		for i := len(matches) - 1; i >= 0; i-- {
			m := matches[i]
			// Accept a quantity only when no other word sits between the unit and
			// the detected food. This prevents "200 г творога, банан" from
			// assigning 200 g to the banana.
			if wordOrNumberPattern.MatchString(segment[m[1]:]) {
				continue
			}
			raw := strings.ReplaceAll(segment[m[2]:m[3]], ",", ".")
			if v, err := strconv.ParseFloat(raw, 64); err == nil && v > 0 {
				q = v
			}
			break
		}
		results = append(results, ExtractedFood{Name: a.Name, QuantityG: q, Confidence: 0.72, Notes: "Локальный разбор без AI — проверь граммовку."})
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("не удалось распознать продукты локально")
	}
	return results, nil
}
func (p *LocalProvider) ExtractFoodImage(_ context.Context, _ string) ([]ExtractedFood, error) {
	return nil, fmt.Errorf("анализ фото требует OPENAI_API_KEY")
}
func (p *LocalProvider) Coach(_ context.Context, req CoachRequest) (CoachResponse, error) {
	msg := strings.ToLower(req.Message)
	answer := "Я вижу твой фитнес-контекст. Подключи OPENAI_API_KEY, чтобы включить полноценного AI Coach с tool calling."
	if strings.Contains(msg, "разгруз") || strings.Contains(msg, "deload") || strings.Contains(msg, "адаптиров") || strings.Contains(msg, "автоперенос") {
		answer = "Изменение сделано по правилам программы, а не случайно: разгрузочная неделя уменьшает объём и интенсивность, чтобы снизить накопленную тренировочную нагрузку, а перенос сохраняет регулярность после пропуска. Это тренировочная адаптация, а не медицинская оценка восстановления."
	}
	if strings.Contains(msg, "восстанов") || strings.Contains(msg, "readiness") || strings.Contains(msg, "готов") || strings.Contains(msg, "устал") {
		if raw, err := json.Marshal(req.BaseContext.Readiness); err == nil {
			var r map[string]any
			if json.Unmarshal(raw, &r) == nil {
				if score, ok := r["score"].(float64); ok {
					answer = fmt.Sprintf("Сегодняшний readiness по тренировочной модели: %.0f/100. Это не медицинская оценка; score собирается из сна, энергии, стресса, soreness и недавней нагрузки.", score)
				}
			}
		}
	}
	if (strings.Contains(msg, "норм") || strings.Contains(msg, "обычно")) && strings.Contains(msg, "сон") {
		if raw, err := json.Marshal(req.BaseContext.Readiness); err == nil {
			var r map[string]any
			if json.Unmarshal(raw, &r) == nil {
				if hi, ok := r["health_insights"].(map[string]any); ok {
					if b, ok := hi["baseline_28d"].(map[string]any); ok {
						if sleep, ok := b["sleep_minutes"].(map[string]any); ok {
							if avg, ok := sleep["average"].(float64); ok {
								answer = fmt.Sprintf("Твоя текущая 28-дневная норма сна по синхронизированным данным — примерно %.0f ч %.0f мин. Это персональный fitness-тренд, а не медицинская норма.", math.Floor(avg/60), math.Mod(math.Round(avg), 60))
							}
						}
					}
				}
			}
		}
	}
	if strings.Contains(msg, "бел") || strings.Contains(msg, "protein") {
		if raw, err := json.Marshal(req.BaseContext.NutritionToday); err == nil {
			var day map[string]any
			if json.Unmarshal(raw, &day) == nil {
				if r, ok := day["remaining_protein_g"].(float64); ok {
					answer = fmt.Sprintf("По текущему дневнику осталось примерно %.0f г белка.", r)
				}
			}
		}
	}
	return CoachResponse{Message: answer, Model: p.Model(), Provider: p.Name()}, nil
}
func (p *LocalProvider) WeeklyReport(_ context.Context, s WeeklyStats) (WeeklyReport, error) {
	wins := []string{}
	if s.CompletedWorkouts > 0 {
		wins = append(wins, fmt.Sprintf("Выполнено тренировок: %d", s.CompletedWorkouts))
	}
	if s.NewPRs > 0 {
		wins = append(wins, fmt.Sprintf("Новых рекордов: %d", s.NewPRs))
	}
	focus := []string{"Продолжай вести тренировки и питание — больше данных даст точнее адаптацию."}
	if s.RecoveryCheckInDays > 0 {
		wins = append(wins, fmt.Sprintf("Recovery check-in: %d дн., средний readiness %.0f/100", s.RecoveryCheckInDays, s.AvgReadiness))
	} else {
		focus = append(focus, "Заполняй короткий recovery check-in, чтобы программа могла учитывать восстановление.")
	}
	return WeeklyReport{Stats: s, Summary: fmt.Sprintf("За неделю выполнено %d тренировок, общий объём %.0f кг.", s.CompletedWorkouts, s.TrainingVolume), Wins: wins, Focus: focus, NextActions: []string{"Выполни следующую тренировку по плану", "Записывай питание и recovery check-in ежедневно"}, Model: p.Model(), Provider: p.Name()}, nil
}
