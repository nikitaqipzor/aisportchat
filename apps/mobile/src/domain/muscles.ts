export type MuscleId =
  | 'chest'
  | 'back'
  | 'shoulders'
  | 'biceps'
  | 'triceps'
  | 'forearms'
  | 'core'
  | 'quads'
  | 'hamstrings'
  | 'glutes'
  | 'calves';

export const muscleMeta: Record<MuscleId, {title: string; short: string; side: 'front' | 'back' | 'both'}> = {
  chest: {title: 'Грудь', short: 'Грудь', side: 'front'},
  back: {title: 'Спина', short: 'Спина', side: 'back'},
  shoulders: {title: 'Плечи', short: 'Плечи', side: 'both'},
  biceps: {title: 'Бицепс', short: 'Бицепс', side: 'front'},
  triceps: {title: 'Трицепс', short: 'Трицепс', side: 'back'},
  forearms: {title: 'Предплечья', short: 'Предпл.', side: 'both'},
  core: {title: 'Кор и пресс', short: 'Кор', side: 'front'},
  quads: {title: 'Квадрицепс', short: 'Квадрицепс', side: 'front'},
  hamstrings: {title: 'Задняя поверхность бедра', short: 'Бицепс бедра', side: 'back'},
  glutes: {title: 'Ягодицы', short: 'Ягодицы', side: 'back'},
  calves: {title: 'Икры', short: 'Икры', side: 'both'},
};

export const muscleIds = Object.keys(muscleMeta) as MuscleId[];
