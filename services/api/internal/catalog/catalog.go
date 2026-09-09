package catalog

import "github.com/example/ai-fitness-os/services/api/internal/model"

var Muscles = []model.Muscle{
	{ID: "chest", Name: "Грудь"},
	{ID: "back", Name: "Спина"},
	{ID: "shoulders", Name: "Плечи"},
	{ID: "biceps", Name: "Бицепс"},
	{ID: "triceps", Name: "Трицепс"},
	{ID: "forearms", Name: "Предплечья"},
	{ID: "core", Name: "Кор"},
	{ID: "quads", Name: "Квадрицепс"},
	{ID: "hamstrings", Name: "Задняя поверхность бедра"},
	{ID: "glutes", Name: "Ягодицы"},
	{ID: "calves", Name: "Икры"},
}

var Equipment = []model.Equipment{
	{ID: "bodyweight", Name: "Собственный вес", Category: "home"},
	{ID: "dumbbells", Name: "Гантели", Category: "free_weight"},
	{ID: "barbell", Name: "Штанга", Category: "free_weight"},
	{ID: "bench", Name: "Скамья", Category: "gym"},
	{ID: "rack", Name: "Стойка", Category: "gym"},
	{ID: "pullup_bar", Name: "Турник", Category: "home"},
	{ID: "cable_machine", Name: "Блочный тренажёр", Category: "machine"},
	{ID: "leg_machine", Name: "Тренажёр для ног", Category: "machine"},
	{ID: "resistance_band", Name: "Резинка", Category: "band"},
	{ID: "kettlebell", Name: "Гиря", Category: "free_weight"},
}

var Exercises = []model.Exercise{
	// Back
	ex("pullup", "Подтягивания", "back", []string{"biceps"}, []string{"gym", "home"}, []string{"pullup_bar"}, "intermediate", "vertical_pull", true, 4, 6, 10, 120),
	ex("lat_pulldown", "Тяга верхнего блока", "back", []string{"biceps"}, []string{"gym"}, []string{"cable_machine"}, "beginner", "vertical_pull", true, 3, 8, 12, 90),
	ex("barbell_row", "Тяга штанги в наклоне", "back", []string{"biceps", "core"}, []string{"gym"}, []string{"barbell"}, "intermediate", "horizontal_pull", true, 4, 6, 10, 120),
	ex("dumbbell_row", "Тяга гантели одной рукой", "back", []string{"biceps"}, []string{"gym", "home"}, []string{"dumbbells"}, "beginner", "horizontal_pull", true, 3, 8, 12, 90),
	ex("seated_cable_row", "Горизонтальная тяга блока", "back", []string{"biceps"}, []string{"gym"}, []string{"cable_machine"}, "beginner", "horizontal_pull", true, 3, 10, 12, 90),
	ex("chest_supported_row", "Тяга гантелей с упором грудью", "back", []string{"biceps"}, []string{"gym"}, []string{"dumbbells", "bench"}, "beginner", "horizontal_pull", true, 3, 8, 12, 90),
	ex("straight_arm_pulldown", "Тяга прямыми руками на блоке", "back", nil, []string{"gym"}, []string{"cable_machine"}, "beginner", "vertical_pull", false, 3, 10, 15, 75),
	ex("cable_face_pull", "Face Pull на блоке", "back", []string{"shoulders"}, []string{"gym"}, []string{"cable_machine"}, "beginner", "horizontal_pull", false, 3, 12, 20, 60),
	ex("band_row", "Тяга резины к поясу", "back", []string{"biceps"}, []string{"band", "home"}, []string{"resistance_band"}, "beginner", "horizontal_pull", true, 4, 10, 15, 75),
	ex("band_lat_pulldown", "Вертикальная тяга резины", "back", []string{"biceps"}, []string{"band", "home"}, []string{"resistance_band"}, "beginner", "vertical_pull", true, 4, 10, 15, 75),
	ex("band_face_pull", "Face Pull с резиной", "back", []string{"shoulders"}, []string{"band", "home"}, []string{"resistance_band"}, "beginner", "horizontal_pull", false, 3, 12, 20, 60),

	// Chest
	ex("pushup", "Отжимания", "chest", []string{"triceps", "shoulders"}, []string{"home", "gym"}, []string{"bodyweight"}, "beginner", "horizontal_push", true, 4, 8, 15, 90),
	ex("bench_press", "Жим штанги лёжа", "chest", []string{"triceps", "shoulders"}, []string{"gym"}, []string{"barbell", "bench"}, "intermediate", "horizontal_push", true, 4, 6, 10, 120),
	ex("incline_dumbbell_press", "Жим гантелей под углом", "chest", []string{"triceps", "shoulders"}, []string{"gym", "home"}, []string{"dumbbells", "bench"}, "intermediate", "horizontal_push", true, 3, 8, 12, 90),
	ex("dumbbell_fly", "Разведение гантелей лёжа", "chest", []string{"shoulders"}, []string{"gym", "home"}, []string{"dumbbells", "bench"}, "beginner", "horizontal_push", false, 3, 10, 15, 75),
	ex("cable_fly", "Сведение рук в кроссовере", "chest", []string{"shoulders"}, []string{"gym"}, []string{"cable_machine"}, "beginner", "horizontal_push", false, 3, 10, 15, 75),
	ex("band_chest_press", "Жим резины от груди", "chest", []string{"triceps", "shoulders"}, []string{"band", "home"}, []string{"resistance_band"}, "beginner", "horizontal_push", true, 4, 10, 15, 75),
	ex("band_fly", "Сведение рук с резиной", "chest", []string{"shoulders"}, []string{"band", "home"}, []string{"resistance_band"}, "beginner", "horizontal_push", false, 3, 12, 18, 60),

	// Shoulders
	ex("overhead_press", "Жим штанги над головой", "shoulders", []string{"triceps"}, []string{"gym"}, []string{"barbell"}, "intermediate", "vertical_push", true, 4, 6, 10, 120),
	ex("dumbbell_shoulder_press", "Жим гантелей сидя", "shoulders", []string{"triceps"}, []string{"gym", "home"}, []string{"dumbbells"}, "beginner", "vertical_push", true, 3, 8, 12, 90),
	ex("lateral_raise", "Подъём гантелей в стороны", "shoulders", nil, []string{"gym", "home"}, []string{"dumbbells"}, "beginner", "abduction", false, 3, 12, 18, 60),
	ex("cable_lateral_raise", "Подъём руки в сторону на блоке", "shoulders", nil, []string{"gym"}, []string{"cable_machine"}, "beginner", "abduction", false, 3, 12, 18, 60),
	ex("band_shoulder_press", "Жим резины над головой", "shoulders", []string{"triceps"}, []string{"band", "home"}, []string{"resistance_band"}, "beginner", "vertical_push", true, 4, 10, 15, 75),
	ex("band_lateral_raise", "Подъём рук в стороны с резиной", "shoulders", nil, []string{"band", "home"}, []string{"resistance_band"}, "beginner", "abduction", false, 3, 12, 20, 60),

	// Biceps
	ex("barbell_curl", "Сгибание рук со штангой", "biceps", []string{"forearms"}, []string{"gym"}, []string{"barbell"}, "beginner", "elbow_flexion", false, 3, 8, 12, 75),
	ex("dumbbell_curl", "Сгибание рук с гантелями", "biceps", []string{"forearms"}, []string{"gym", "home"}, []string{"dumbbells"}, "beginner", "elbow_flexion", false, 3, 8, 12, 75),
	ex("hammer_curl", "Молотковые сгибания", "biceps", []string{"forearms"}, []string{"gym", "home"}, []string{"dumbbells"}, "beginner", "elbow_flexion", false, 3, 10, 14, 75),
	ex("cable_curl", "Сгибание рук на нижнем блоке", "biceps", []string{"forearms"}, []string{"gym"}, []string{"cable_machine"}, "beginner", "elbow_flexion", false, 3, 10, 15, 60),
	ex("band_curl", "Сгибание рук с резиной", "biceps", []string{"forearms"}, []string{"band", "home"}, []string{"resistance_band"}, "beginner", "elbow_flexion", false, 4, 10, 18, 60),

	// Triceps
	ex("close_grip_pushup", "Узкие отжимания", "triceps", []string{"chest"}, []string{"home", "gym"}, []string{"bodyweight"}, "beginner", "horizontal_push", true, 3, 8, 15, 75),
	ex("close_grip_bench", "Жим узким хватом", "triceps", []string{"chest"}, []string{"gym"}, []string{"barbell", "bench"}, "intermediate", "horizontal_push", true, 3, 6, 10, 105),
	ex("triceps_pushdown", "Разгибание рук на блоке", "triceps", nil, []string{"gym"}, []string{"cable_machine"}, "beginner", "elbow_extension", false, 3, 10, 15, 60),
	ex("overhead_dumbbell_extension", "Разгибание гантели из-за головы", "triceps", nil, []string{"gym", "home"}, []string{"dumbbells"}, "beginner", "elbow_extension", false, 3, 10, 15, 75),
	ex("band_triceps_pushdown", "Разгибание рук с резиной", "triceps", nil, []string{"band", "home"}, []string{"resistance_band"}, "beginner", "elbow_extension", false, 4, 12, 20, 60),

	// Quads
	ex("bodyweight_squat", "Приседания с собственным весом", "quads", []string{"glutes", "core"}, []string{"home", "gym"}, []string{"bodyweight"}, "beginner", "squat", true, 4, 12, 20, 90),
	ex("barbell_squat", "Приседания со штангой", "quads", []string{"glutes", "hamstrings", "core"}, []string{"gym"}, []string{"barbell", "rack"}, "intermediate", "squat", true, 4, 5, 8, 150),
	ex("goblet_squat", "Гоблет-присед", "quads", []string{"glutes", "core"}, []string{"gym", "home"}, []string{"dumbbells"}, "beginner", "squat", true, 4, 8, 12, 90),
	ex("split_squat", "Болгарские выпады", "quads", []string{"glutes"}, []string{"gym", "home"}, []string{"dumbbells"}, "intermediate", "lunge", true, 3, 8, 12, 90),
	ex("leg_extension", "Разгибание ног в тренажёре", "quads", nil, []string{"gym"}, []string{"leg_machine"}, "beginner", "knee_extension", false, 3, 10, 15, 75),
	ex("band_squat", "Приседания с резиной", "quads", []string{"glutes"}, []string{"band", "home"}, []string{"resistance_band"}, "beginner", "squat", true, 4, 10, 15, 90),

	// Hamstrings
	ex("romanian_deadlift", "Румынская тяга", "hamstrings", []string{"glutes", "back"}, []string{"gym"}, []string{"barbell"}, "intermediate", "hinge", true, 4, 6, 10, 120),
	ex("dumbbell_rdl", "Румынская тяга с гантелями", "hamstrings", []string{"glutes"}, []string{"gym", "home"}, []string{"dumbbells"}, "beginner", "hinge", true, 3, 8, 12, 90),
	ex("leg_curl", "Сгибание ног в тренажёре", "hamstrings", nil, []string{"gym"}, []string{"leg_machine"}, "beginner", "knee_flexion", false, 3, 10, 15, 75),
	ex("band_leg_curl", "Сгибание ног с резиной", "hamstrings", nil, []string{"band", "home"}, []string{"resistance_band"}, "beginner", "knee_flexion", false, 4, 12, 20, 60),
	ex("single_leg_rdl", "Румынская тяга на одной ноге", "hamstrings", []string{"glutes", "core"}, []string{"home", "gym"}, []string{"dumbbells"}, "intermediate", "hinge", true, 3, 8, 12, 75),

	// Glutes
	ex("hip_thrust", "Ягодичный мост со штангой", "glutes", []string{"hamstrings"}, []string{"gym"}, []string{"barbell", "bench"}, "intermediate", "hip_extension", true, 4, 8, 12, 105),
	ex("glute_bridge", "Ягодичный мост", "glutes", []string{"hamstrings"}, []string{"home", "gym"}, []string{"bodyweight"}, "beginner", "hip_extension", true, 4, 12, 20, 75),
	ex("dumbbell_hip_thrust", "Ягодичный мост с гантелью", "glutes", []string{"hamstrings"}, []string{"home", "gym"}, []string{"dumbbells", "bench"}, "beginner", "hip_extension", true, 4, 10, 15, 90),
	ex("band_glute_bridge", "Ягодичный мост с резиной", "glutes", []string{"hamstrings"}, []string{"band", "home"}, []string{"resistance_band"}, "beginner", "hip_extension", true, 4, 12, 20, 75),
	ex("band_abduction", "Отведение ноги с резиной", "glutes", nil, []string{"band", "home"}, []string{"resistance_band"}, "beginner", "abduction", false, 3, 15, 25, 45),

	// Calves
	ex("standing_calf_raise", "Подъём на носки стоя", "calves", nil, []string{"home", "gym"}, []string{"bodyweight"}, "beginner", "plantar_flexion", false, 4, 12, 20, 60),
	ex("dumbbell_calf_raise", "Подъём на носки с гантелями", "calves", nil, []string{"home", "gym"}, []string{"dumbbells"}, "beginner", "plantar_flexion", false, 4, 10, 18, 60),
	ex("band_calf_raise", "Подъём на носки с резиной", "calves", nil, []string{"band", "home"}, []string{"resistance_band"}, "beginner", "plantar_flexion", false, 4, 15, 25, 60),

	// Core
	ex("plank", "Планка", "core", nil, []string{"home", "gym"}, []string{"bodyweight"}, "beginner", "anti_extension", false, 3, 30, 60, 45),
	ex("dead_bug", "Dead Bug", "core", nil, []string{"home", "gym"}, []string{"bodyweight"}, "beginner", "anti_extension", false, 3, 8, 14, 45),
	ex("hanging_knee_raise", "Подъём коленей в висе", "core", []string{"forearms"}, []string{"gym", "home"}, []string{"pullup_bar"}, "intermediate", "flexion", false, 3, 8, 15, 60),
	ex("cable_crunch", "Скручивания на верхнем блоке", "core", nil, []string{"gym"}, []string{"cable_machine"}, "beginner", "flexion", false, 3, 10, 15, 60),
	ex("band_pallof_press", "Pallof Press с резиной", "core", nil, []string{"band", "home"}, []string{"resistance_band"}, "beginner", "anti_rotation", false, 3, 10, 15, 45),

	// Forearms
	ex("wrist_curl", "Сгибание кистей с гантелями", "forearms", nil, []string{"home", "gym"}, []string{"dumbbells"}, "beginner", "wrist_flexion", false, 3, 12, 20, 45),
	ex("reverse_wrist_curl", "Разгибание кистей с гантелями", "forearms", nil, []string{"home", "gym"}, []string{"dumbbells"}, "beginner", "wrist_extension", false, 3, 12, 20, 45),
	ex("farmer_carry", "Фермерская прогулка", "forearms", []string{"core", "shoulders"}, []string{"gym", "home"}, []string{"dumbbells"}, "beginner", "carry", true, 4, 30, 60, 60),
	ex("band_wrist_curl", "Сгибание кистей с резиной", "forearms", nil, []string{"band", "home"}, []string{"resistance_band"}, "beginner", "wrist_flexion", false, 3, 15, 25, 45),
}

func ex(id, name, primary string, secondary, environment, equipment []string, difficulty, pattern string, compound bool, sets, repMin, repMax, rest int) model.Exercise {
	description, instructions, mistakes, tips := guidance(name, pattern)
	return model.Exercise{
		ID: id, Name: name, PrimaryMuscle: primary, SecondaryMuscles: secondary,
		Environment: environment, Equipment: equipment, Difficulty: difficulty,
		MovementPattern: pattern, Compound: compound, DefaultSets: sets,
		RepMin: repMin, RepMax: repMax, RestSeconds: rest, Description: description,
		Instructions: instructions, CommonMistakes: mistakes, TechniqueTips: tips,
	}
}

func guidance(name, pattern string) (string, []string, []string, []string) {
	description := name + " — силовое упражнение. Работай в контролируемой амплитуде и прекращай подход при боли или потере техники."
	baseTips := []string{"Двигайся плавно без рывков.", "Сохраняй контролируемое дыхание и устойчивое положение корпуса."}
	switch pattern {
	case "horizontal_pull":
		return description, []string{"Зафиксируй корпус и начни движение лопатками.", "Тяни рукоять или вес к корпусу, ведя локти назад.", "В конечной точке сведи лопатки без подъёма плеч.", "Контролируемо верни вес в исходное положение."}, []string{"Раскачивание корпусом.", "Подъём плеч к ушам.", "Слишком короткая амплитуда."}, baseTips
	case "vertical_pull":
		return description, []string{"Стабилизируй корпус и опусти плечи.", "Тяни локти вниз по направлению к бокам.", "Не запрокидывай корпус назад ради повторения.", "Плавно вернись до контролируемого растяжения."}, []string{"Тяга только руками.", "Сильное отклонение корпуса.", "Рывок из нижней точки."}, baseTips
	case "horizontal_push":
		return description, []string{"Зафиксируй лопатки и устойчиво поставь опору.", "Опускай сопротивление под контролем.", "Выжимай по устойчивой траектории без потери положения плеч.", "Не блокируй суставы резким ударным движением."}, []string{"Локти чрезмерно разведены.", "Отрыв устойчивой опоры.", "Отскок или провал в нижней точке."}, baseTips
	case "vertical_push":
		return description, []string{"Напряги корпус и удерживай нейтральное положение позвоночника.", "Начни жим из устойчивой позиции.", "Проведи сопротивление вверх без чрезмерного прогиба поясницы.", "Вернись вниз под контролем."}, []string{"Сильный прогиб в пояснице.", "Рывок ногами без цели.", "Потеря контроля плеч."}, baseTips
	case "squat":
		return description, []string{"Поставь стопы устойчиво и напряги корпус.", "Начни движение одновременно в тазобедренных и коленных суставах.", "Колени направляй по линии стоп.", "Опустись до комфортной глубины и встань, сохраняя баланс."}, []string{"Колени заметно заваливаются внутрь.", "Пятки теряют опору.", "Слишком быстрое падение вниз."}, []string{"Дави всей стопой в пол.", "Выбирай глубину, в которой сохраняется контроль."}
	case "lunge":
		return description, []string{"Сделай устойчивую разножку.", "Опускай таз вертикально и контролируй колено передней ноги.", "Сохраняй корпус собранным.", "Вернись вверх через опорную стопу."}, []string{"Слишком узкая постановка ног.", "Потеря баланса.", "Колено заваливается внутрь."}, baseTips
	case "hinge":
		return description, []string{"Слегка согни колени и зафиксируй корпус.", "Отводи таз назад, сохраняя нейтральную спину.", "Опускай сопротивление до контролируемого растяжения задней поверхности бедра.", "Разогни тазобедренные суставы и вернись вверх."}, []string{"Округление спины.", "Превращение движения в присед.", "Вес уходит далеко от тела."}, []string{"Думай о движении таза назад, а не о наклоне вниз.", "Сохраняй сопротивление близко к ногам."}
	case "elbow_flexion":
		return description, []string{"Зафиксируй плечи рядом с корпусом.", "Согни локти без раскачивания туловища.", "Сделай короткую контролируемую паузу сверху.", "Полностью контролируй опускание."}, []string{"Раскачивание корпусом.", "Локти сильно уходят вперёд.", "Бросание веса вниз."}, baseTips
	case "elbow_extension":
		return description, []string{"Зафиксируй плечо и положение корпуса.", "Разогни локоть в полной комфортной амплитуде.", "Не помогай движению корпусом.", "Плавно вернись в исходное положение."}, []string{"Локти расходятся без необходимости.", "Раскачивание корпуса.", "Слишком большой вес снижает амплитуду."}, baseTips
	case "hip_extension":
		return description, []string{"Установи устойчивую опору и напряги корпус.", "Подними таз за счёт разгибания тазобедренных суставов.", "В верхней точке сократи ягодицы без переразгибания поясницы.", "Плавно опусти таз."}, []string{"Переразгибание поясницы.", "Толчок инерцией.", "Неустойчивая постановка стоп."}, baseTips
	case "abduction":
		return description, []string{"Зафиксируй таз и корпус.", "Отводи конечность в сторону без раскачивания.", "Сделай короткую паузу в конечной точке.", "Вернись под контролем."}, []string{"Раскачивание корпусом.", "Слишком большая амплитуда ценой положения таза.", "Рывок."}, baseTips
	case "knee_extension", "knee_flexion":
		return description, []string{"Настрой положение сустава относительно оси тренажёра или резины.", "Выполни движение в комфортной амплитуде.", "Коротко зафиксируй конечную позицию.", "Медленно верни сопротивление."}, []string{"Рывок в начале повторения.", "Слишком быстрый негатив.", "Неподходящая настройка тренажёра."}, baseTips
	case "plantar_flexion":
		return description, []string{"Установи устойчивую опору стопы.", "Поднимись на носки максимально высоко без потери баланса.", "Сделай короткую паузу сверху.", "Опустись под контролем до комфортного растяжения."}, []string{"Пружинящие повторения.", "Слишком короткая амплитуда.", "Завал стоп внутрь или наружу."}, baseTips
	case "anti_extension", "anti_rotation":
		return description, []string{"Прими стабильное исходное положение.", "Напряги мышцы корпуса до начала движения.", "Не позволяй пояснице или тазу менять положение.", "Дыши спокойно, сохраняя напряжение."}, []string{"Провал поясницы.", "Задержка дыхания на весь подход.", "Потеря положения таза."}, baseTips
	case "flexion":
		return description, []string{"Стабилизируй исходное положение.", "Выполни сгибание контролируемо, без рывка.", "Сохрани напряжение в целевых мышцах.", "Плавно вернись назад."}, []string{"Рывок за счёт инерции.", "Избыточное движение шеей.", "Слишком высокая скорость."}, baseTips
	case "wrist_flexion", "wrist_extension":
		return description, []string{"Зафиксируй предплечья.", "Двигай только кистями в комфортной амплитуде.", "Сделай паузу в конечной позиции.", "Медленно верни сопротивление."}, []string{"Движение всем предплечьем.", "Рывки кистью.", "Слишком тяжёлый вес."}, baseTips
	case "carry":
		return description, []string{"Возьми вес и выстрой высокий устойчивый корпус.", "Опусти плечи и удерживай лопатки стабильно.", "Иди короткими контролируемыми шагами.", "Сохраняй одинаковую высоту таза и плеч."}, []string{"Сутулость.", "Раскачивание из стороны в сторону.", "Слишком длинный шаг при потере стабильности."}, baseTips
	default:
		return description, []string{"Прими устойчивое исходное положение.", "Выполняй рабочую фазу в контролируемой амплитуде.", "Сохраняй положение корпуса.", "Плавно вернись в исходную позицию."}, []string{"Рывки.", "Потеря контроля корпуса.", "Слишком большая нагрузка для выбранной техники."}, baseTips
	}
}

func ExerciseByID(id string) (model.Exercise, bool) {
	for _, exercise := range Exercises {
		if exercise.ID == id {
			return exercise, true
		}
	}
	return model.Exercise{}, false
}
