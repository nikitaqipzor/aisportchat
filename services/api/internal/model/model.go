package model

type Muscle struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Equipment struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

type Exercise struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	PrimaryMuscle    string   `json:"primary_muscle"`
	SecondaryMuscles []string `json:"secondary_muscles,omitempty"`
	Environment      []string `json:"environment"`
	Equipment        []string `json:"equipment"`
	Difficulty       string   `json:"difficulty"`
	MovementPattern  string   `json:"movement_pattern"`
	Compound         bool     `json:"compound"`
	DefaultSets      int      `json:"default_sets"`
	RepMin           int      `json:"rep_min"`
	RepMax           int      `json:"rep_max"`
	RestSeconds      int      `json:"rest_seconds"`
	Description      string   `json:"description"`
	Instructions     []string `json:"instructions"`
	CommonMistakes   []string `json:"common_mistakes"`
	TechniqueTips    []string `json:"technique_tips"`
}

type GeneratedExercise struct {
	Exercise   Exercise `json:"exercise"`
	Sets       int      `json:"sets"`
	RepMin     int      `json:"rep_min"`
	RepMax     int      `json:"rep_max"`
	RestSecond int      `json:"rest_seconds"`
}

type Workout struct {
	Muscle          string              `json:"muscle"`
	Environment     string              `json:"environment"`
	DurationMinutes int                 `json:"duration_minutes"`
	Exercises       []GeneratedExercise `json:"exercises"`
}
