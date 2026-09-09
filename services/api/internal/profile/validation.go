package profile

import "errors"

var allowedGoals = map[string]bool{"muscle_gain": true, "fat_loss": true, "recomposition": true, "strength": true, "maintenance": true, "endurance": true}
var allowedLevels = map[string]bool{"beginner": true, "intermediate": true, "advanced": true}
var allowedEnvironments = map[string]bool{"home": true, "gym": true, "band": true}

func ValidateGoal(goal string) error {
	if !allowedGoals[goal] {
		return errors.New("unsupported goal_type")
	}
	return nil
}
func ValidateLevel(level string) error {
	if !allowedLevels[level] {
		return errors.New("unsupported experience_level")
	}
	return nil
}
func ValidateEnvironments(items []string) error {
	if len(items) == 0 {
		return errors.New("at least one environment is required")
	}
	for _, v := range items {
		if !allowedEnvironments[v] {
			return errors.New("unsupported environment: " + v)
		}
	}
	return nil
}
