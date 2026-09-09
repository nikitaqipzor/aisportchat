INSERT INTO muscles (id, name, body_region) VALUES
('chest','Грудь','upper'),('back','Спина','upper'),('shoulders','Плечи','upper'),
('biceps','Бицепс','upper'),('triceps','Трицепс','upper'),('forearms','Предплечья','upper'),
('core','Кор','core'),('quads','Квадрицепс','lower'),('hamstrings','Задняя поверхность бедра','lower'),
('glutes','Ягодицы','lower'),('calves','Икры','lower')
ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, body_region=EXCLUDED.body_region;

INSERT INTO equipment (id, name, category) VALUES
('bodyweight','Собственный вес','home'),('dumbbells','Гантели','free_weight'),('barbell','Штанга','free_weight'),
('bench','Скамья','gym'),('rack','Стойка','gym'),('pullup_bar','Турник','home'),
('cable_machine','Блочный тренажёр','machine'),('resistance_band','Резинка','band')
ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, category=EXCLUDED.category;

INSERT INTO exercises (id,name,difficulty,movement_pattern,exercise_type,compound,default_sets,rep_min,rep_max,rest_seconds) VALUES
('pullup','Подтягивания','intermediate','vertical_pull','strength',true,4,6,10,120),
('lat_pulldown','Тяга верхнего блока','beginner','vertical_pull','strength',true,3,8,12,90),
('barbell_row','Тяга штанги в наклоне','intermediate','horizontal_pull','strength',true,4,6,10,120),
('seated_cable_row','Горизонтальная тяга блока','beginner','horizontal_pull','strength',true,3,10,12,90),
('band_row','Тяга резины к поясу','beginner','horizontal_pull','strength',true,4,10,15,75),
('band_lat_pulldown','Вертикальная тяга резины','beginner','vertical_pull','strength',true,4,10,15,75),
('band_face_pull','Face Pull с резиной','beginner','horizontal_pull','strength',false,3,12,20,60),
('pushup','Отжимания','beginner','horizontal_push','strength',true,4,8,15,90),
('bench_press','Жим штанги лёжа','intermediate','horizontal_push','strength',true,4,6,10,120),
('incline_dumbbell_press','Жим гантелей под углом','intermediate','horizontal_push','strength',true,3,8,12,90),
('band_chest_press','Жим резины от груди','beginner','horizontal_push','strength',true,4,10,15,75),
('band_fly','Сведение рук с резиной','beginner','horizontal_push','strength',false,3,12,18,60),
('bodyweight_squat','Приседания с собственным весом','beginner','squat','strength',true,4,12,20,90),
('barbell_squat','Приседания со штангой','intermediate','squat','strength',true,4,5,8,150),
('band_squat','Приседания с резиной','beginner','squat','strength',true,4,10,15,90)
ON CONFLICT (id) DO UPDATE SET
name=EXCLUDED.name,difficulty=EXCLUDED.difficulty,movement_pattern=EXCLUDED.movement_pattern,
exercise_type=EXCLUDED.exercise_type,compound=EXCLUDED.compound,default_sets=EXCLUDED.default_sets,
rep_min=EXCLUDED.rep_min,rep_max=EXCLUDED.rep_max,rest_seconds=EXCLUDED.rest_seconds;

INSERT INTO exercise_muscles (exercise_id,muscle_id,role) VALUES
('pullup','back','primary'),('pullup','biceps','secondary'),
('lat_pulldown','back','primary'),('lat_pulldown','biceps','secondary'),
('barbell_row','back','primary'),('barbell_row','biceps','secondary'),('barbell_row','core','secondary'),
('seated_cable_row','back','primary'),('seated_cable_row','biceps','secondary'),
('band_row','back','primary'),('band_row','biceps','secondary'),
('band_lat_pulldown','back','primary'),('band_lat_pulldown','biceps','secondary'),
('band_face_pull','back','primary'),('band_face_pull','shoulders','secondary'),
('pushup','chest','primary'),('pushup','triceps','secondary'),('pushup','shoulders','secondary'),
('bench_press','chest','primary'),('bench_press','triceps','secondary'),('bench_press','shoulders','secondary'),
('incline_dumbbell_press','chest','primary'),('incline_dumbbell_press','triceps','secondary'),('incline_dumbbell_press','shoulders','secondary'),
('band_chest_press','chest','primary'),('band_chest_press','triceps','secondary'),('band_chest_press','shoulders','secondary'),
('band_fly','chest','primary'),('band_fly','shoulders','secondary'),
('bodyweight_squat','quads','primary'),('bodyweight_squat','glutes','secondary'),('bodyweight_squat','core','secondary'),
('barbell_squat','quads','primary'),('barbell_squat','glutes','secondary'),('barbell_squat','hamstrings','secondary'),('barbell_squat','core','secondary'),
('band_squat','quads','primary'),('band_squat','glutes','secondary')
ON CONFLICT DO NOTHING;

INSERT INTO exercise_environments (exercise_id,environment) VALUES
('pullup','gym'),('pullup','home'),('lat_pulldown','gym'),('barbell_row','gym'),('seated_cable_row','gym'),
('band_row','band'),('band_row','home'),('band_lat_pulldown','band'),('band_lat_pulldown','home'),
('band_face_pull','band'),('band_face_pull','home'),('pushup','home'),('pushup','gym'),('bench_press','gym'),
('incline_dumbbell_press','gym'),('incline_dumbbell_press','home'),('band_chest_press','band'),('band_chest_press','home'),
('band_fly','band'),('band_fly','home'),('bodyweight_squat','home'),('bodyweight_squat','gym'),('barbell_squat','gym'),
('band_squat','band'),('band_squat','home')
ON CONFLICT DO NOTHING;

INSERT INTO exercise_equipment (exercise_id,equipment_id) VALUES
('pullup','pullup_bar'),('lat_pulldown','cable_machine'),('barbell_row','barbell'),('seated_cable_row','cable_machine'),
('band_row','resistance_band'),('band_lat_pulldown','resistance_band'),('band_face_pull','resistance_band'),
('pushup','bodyweight'),('bench_press','barbell'),('bench_press','bench'),('incline_dumbbell_press','dumbbells'),
('incline_dumbbell_press','bench'),('band_chest_press','resistance_band'),('band_fly','resistance_band'),
('bodyweight_squat','bodyweight'),('barbell_squat','barbell'),('barbell_squat','rack'),('band_squat','resistance_band')
ON CONFLICT DO NOTHING;
