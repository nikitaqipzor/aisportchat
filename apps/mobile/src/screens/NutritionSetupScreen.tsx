import React, {useState} from 'react';
import {Pressable, ScrollView, StyleSheet, Text, TextInput, View} from 'react-native';
import {api, NutritionProfile} from '../api/client';
import {AppButton} from '../components/AppButton';
import {colors, control, radius, spacing} from '../theme/tokens';

type Goal = NutritionProfile['goal'];
type Activity = NutritionProfile['activity_level'];

const goals: Array<{id: Goal; title: string; subtitle: string}> = [
  {id: 'lose', title: 'Снижение веса', subtitle: 'Умеренный дефицит калорий'},
  {id: 'recomp', title: 'Рекомпозиция', subtitle: 'Сохранять мышцы и постепенно снижать жир'},
  {id: 'maintain', title: 'Поддержание', subtitle: 'Сохранять текущий вес'},
  {id: 'gain', title: 'Набор массы', subtitle: 'Умеренный профицит калорий'},
  {id: 'strength', title: 'Сила', subtitle: 'Поддерживать энергию для силовой работы'},
  {id: 'endurance', title: 'Выносливость', subtitle: 'Больше энергии для объёмной работы'},
];

const activities: Array<{id: Activity; title: string; hint: string}> = [
  {id: 'low', title: 'Низкая', hint: 'В основном сидячий образ жизни'},
  {id: 'light', title: 'Лёгкая', hint: 'Небольшая активность 1–2 раза в неделю'},
  {id: 'moderate', title: 'Средняя', hint: 'Активность или тренировки 3–4 раза в неделю'},
  {id: 'high', title: 'Высокая', hint: 'Интенсивная активность 5–6 раз в неделю'},
  {id: 'athlete', title: 'Очень высокая', hint: 'Ежедневные тяжёлые тренировки или физическая работа'},
];

function numberFromInput(value: string) {
  return Number(value.replace(',', '.'));
}

function macroError(value: string) {
  const number = numberFromInput(value);
  if (!value.trim()) {
    return 'Заполните поле';
  }
  if (!Number.isFinite(number) || number < 0 || number > 1000) {
    return 'Введите значение от 0 до 1000 г';
  }
  return '';
}

export function NutritionSetupScreen({accessToken, onBack, onSaved}: {accessToken: string; onBack: () => void; onSaved: (profile: NutritionProfile) => void}) {
  const [goal, setGoal] = useState<Goal>('recomp');
  const [activity, setActivity] = useState<Activity>('moderate');
  const [mode, setMode] = useState<'auto' | 'manual'>('auto');
  const [calories, setCalories] = useState('2500');
  const [protein, setProtein] = useState('180');
  const [fat, setFat] = useState('75');
  const [carbs, setCarbs] = useState('280');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  const calorieNumber = numberFromInput(calories);
  const calorieError = !calories.trim()
    ? 'Заполните поле'
    : !Number.isInteger(calorieNumber) || calorieNumber < 800 || calorieNumber > 8000
      ? 'Введите целое число от 800 до 8000 ккал'
      : '';
  const proteinError = macroError(protein);
  const fatError = macroError(fat);
  const carbsError = macroError(carbs);
  const manualValid = !calorieError && !proteinError && !fatError && !carbsError;

  async function save() {
    if (saving || (mode === 'manual' && !manualValid)) {
      return;
    }
    try {
      setSaving(true);
      setError('');
      const manual = mode === 'manual';
      const profile = await api.setNutritionProfile(accessToken, {
        goal,
        activity_level: activity,
        calculation_mode: mode,
        ...(manual ? {
          calorie_target: calorieNumber,
          protein_target_g: numberFromInput(protein),
          fat_target_g: numberFromInput(fat),
          carb_target_g: numberFromInput(carbs),
        } : {}),
      });
      onSaved(profile);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось сохранить настройки питания. Попробуйте ещё раз.');
    } finally {
      setSaving(false);
    }
  }

  function clearServerError() {
    if (error) {
      setError('');
    }
  }

  return (
    <ScrollView
      testID="nutrition-setup-screen"
      contentContainerStyle={styles.container}
      keyboardShouldPersistTaps="handled"
      showsVerticalScrollIndicator={false}>
      <View style={styles.content}>
        <View style={styles.header}>
          <Pressable
            accessibilityRole="button"
            accessibilityLabel="Вернуться к питанию"
            accessibilityState={{disabled: saving}}
            disabled={saving}
            onPress={onBack}
            style={({pressed}) => [styles.backButton, pressed && styles.pressed]}>
            <Text style={styles.back}>← Назад</Text>
          </Pressable>
          <Text style={styles.kicker}>ПИТАНИЕ</Text>
        </View>

        <Text accessibilityRole="header" style={styles.title}>Настроим питание</Text>
        <Text style={styles.subtitle}>
          Первый расчёт — ориентир, а не медицинское назначение. Позже его можно изменить по реальной динамике веса.
        </Text>

        <Text style={styles.section}>Цель</Text>
        <View style={styles.stack} accessibilityRole="radiogroup">
          {goals.map(item => {
            const selected = goal === item.id;
            return (
              <Pressable
                key={item.id}
                testID={`nutrition-goal-${item.id}`}
                accessibilityRole="radio"
                accessibilityLabel={item.title}
                accessibilityHint={item.subtitle}
                accessibilityState={{selected, disabled: saving}}
                disabled={saving}
                onPress={() => {
                  setGoal(item.id);
                  clearServerError();
                }}
                style={({pressed}) => [styles.card, selected && styles.cardActive, pressed && styles.pressed]}>
                <Text style={[styles.cardTitle, selected && styles.activeText]}>{item.title}</Text>
                <Text style={[styles.cardSubtitle, selected && styles.activeSub]}>{item.subtitle}</Text>
              </Pressable>
            );
          })}
        </View>

        <Text style={styles.section}>Общая активность</Text>
        <View style={styles.wrap} accessibilityRole="radiogroup">
          {activities.map(item => {
            const selected = activity === item.id;
            return (
              <Pressable
                key={item.id}
                testID={`nutrition-activity-${item.id}`}
                accessibilityRole="radio"
                accessibilityLabel={item.title}
                accessibilityHint={item.hint}
                accessibilityState={{selected, disabled: saving}}
                disabled={saving}
                onPress={() => {
                  setActivity(item.id);
                  clearServerError();
                }}
                style={({pressed}) => [styles.pill, selected && styles.pillActive, pressed && styles.pressed]}>
                <Text style={[styles.pillText, selected && styles.activeText]}>{item.title}</Text>
              </Pressable>
            );
          })}
        </View>
        <Text style={styles.selectionHint}>{activities.find(item => item.id === activity)?.hint}</Text>

        <Text style={styles.section}>Режим расчёта</Text>
        <View style={styles.modeRow} accessibilityRole="radiogroup">
          {(['auto', 'manual'] as const).map(item => {
            const selected = mode === item;
            const label = item === 'auto' ? 'Автоматически' : 'Вручную';
            return (
              <Pressable
                key={item}
                testID={`nutrition-mode-${item}`}
                accessibilityRole="radio"
                accessibilityLabel={label}
                accessibilityState={{selected, disabled: saving}}
                disabled={saving}
                onPress={() => {
                  setMode(item);
                  clearServerError();
                }}
                style={({pressed}) => [styles.modeButton, selected && styles.modeButtonActive, pressed && styles.pressed]}>
                <Text style={[styles.modeText, selected && styles.activeText]}>{label}</Text>
              </Pressable>
            );
          })}
        </View>

        {mode === 'manual' ? (
          <View style={styles.manualGrid}>
            <ManualField
              id="calories"
              label="Калории, ккал"
              value={calories}
              error={calorieError}
              helper="800–8000 ккал"
              integer
              disabled={saving}
              onChange={value => { setCalories(value); clearServerError(); }}
            />
            <ManualField id="protein" label="Белок, г" value={protein} error={proteinError} helper="0–1000 г" disabled={saving} onChange={value => { setProtein(value); clearServerError(); }} />
            <ManualField id="fat" label="Жиры, г" value={fat} error={fatError} helper="0–1000 г" disabled={saving} onChange={value => { setFat(value); clearServerError(); }} />
            <ManualField id="carbs" label="Углеводы, г" value={carbs} error={carbsError} helper="0–1000 г" disabled={saving} onChange={value => { setCarbs(value); clearServerError(); }} />
          </View>
        ) : null}

        <View style={styles.note}>
          <Text style={styles.noteTitle}>{mode === 'auto' ? 'Как работает расчёт' : 'Проверьте значения'}</Text>
          <Text style={styles.noteText}>
            {mode === 'auto'
              ? 'Мы используем вес из профиля, цель и уровень активности. Результат можно скорректировать вручную.'
              : 'Сохраним введённые цели без автоматического пересчёта. При резких изменениях рациона проконсультируйтесь со специалистом.'}
          </Text>
        </View>

        {error ? (
          <Text accessibilityRole="alert" accessibilityLiveRegion="assertive" style={styles.error}>{error}</Text>
        ) : null}
        <AppButton
          label={mode === 'auto' ? 'Рассчитать питание' : 'Сохранить цели'}
          accessibilityLabel={mode === 'auto' ? 'Рассчитать и сохранить питание' : 'Сохранить ручные цели питания'}
          testID="nutrition-save"
          loading={saving}
          disabled={mode === 'manual' && !manualValid}
          onPress={() => void save()}
        />
      </View>
    </ScrollView>
  );
}

function ManualField({id, label, value, error, helper, integer = false, disabled, onChange}: {
  id: string;
  label: string;
  value: string;
  error: string;
  helper: string;
  integer?: boolean;
  disabled: boolean;
  onChange: (value: string) => void;
}) {
  const showError = Boolean(value.trim()) && Boolean(error);
  return (
    <View style={styles.manualField}>
      <Text nativeID={`nutrition-${id}-label`} style={styles.label}>{label}</Text>
      <TextInput
        accessibilityLabelledBy={`nutrition-${id}-label`}
        accessibilityHint={helper}
        testID={`nutrition-${id}`}
        value={value}
        editable={!disabled}
        maxLength={7}
        onChangeText={onChange}
        keyboardType={integer ? 'number-pad' : 'decimal-pad'}
        selectTextOnFocus
        style={[styles.manualInput, showError && styles.inputError]}
      />
      <Text accessibilityRole={showError ? 'alert' : 'text'} style={[styles.helper, showError && styles.fieldError]}>
        {showError ? error : helper}
      </Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {flexGrow: 1, padding: spacing.lg, backgroundColor: colors.background},
  content: {width: '100%', maxWidth: 680, alignSelf: 'center', gap: spacing.md},
  header: {minHeight: control.minTouch, flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center'},
  backButton: {minHeight: control.minTouch, minWidth: control.minTouch, justifyContent: 'center', paddingRight: spacing.sm},
  back: {fontSize: 15, fontWeight: '800', color: colors.text},
  kicker: {fontSize: 11, fontWeight: '900', letterSpacing: 1.4, color: colors.textMuted},
  title: {fontSize: 30, lineHeight: 37, fontWeight: '900', color: colors.text},
  subtitle: {fontSize: 15, lineHeight: 22, color: colors.textMuted},
  section: {fontSize: 18, lineHeight: 24, fontWeight: '900', marginTop: spacing.sm, color: colors.text},
  stack: {gap: spacing.sm},
  card: {borderWidth: 1, borderColor: colors.border, borderRadius: radius.lg, paddingHorizontal: 14, paddingVertical: 12, minHeight: 64, justifyContent: 'center', backgroundColor: colors.surface},
  cardActive: {backgroundColor: colors.primary, borderColor: colors.primary},
  cardTitle: {fontSize: 16, lineHeight: 21, fontWeight: '800', color: colors.text},
  cardSubtitle: {fontSize: 12, lineHeight: 17, marginTop: 3, color: colors.textMuted},
  activeText: {color: colors.inverse},
  activeSub: {color: colors.inverse, opacity: 0.76},
  wrap: {flexDirection: 'row', flexWrap: 'wrap', gap: spacing.sm},
  pill: {borderWidth: 1, borderColor: colors.border, borderRadius: radius.pill, paddingHorizontal: 14, minHeight: control.minTouch, justifyContent: 'center', backgroundColor: colors.surface},
  pillActive: {backgroundColor: colors.primary, borderColor: colors.primary},
  pillText: {fontWeight: '800', color: colors.text},
  selectionHint: {fontSize: 12, lineHeight: 18, color: colors.textMuted, marginTop: -spacing.sm},
  modeRow: {flexDirection: 'row', gap: spacing.sm},
  modeButton: {flex: 1, minHeight: control.minTouch, borderWidth: 1, borderColor: colors.border, borderRadius: radius.md, alignItems: 'center', justifyContent: 'center', paddingHorizontal: spacing.sm, backgroundColor: colors.surface},
  modeButtonActive: {backgroundColor: colors.primary, borderColor: colors.primary},
  modeText: {fontWeight: '800', color: colors.text, textAlign: 'center'},
  manualGrid: {flexDirection: 'row', flexWrap: 'wrap', gap: spacing.sm},
  manualField: {width: '48%', flexGrow: 1, minWidth: 132},
  label: {fontSize: 12, lineHeight: 17, fontWeight: '800', color: colors.text, marginBottom: spacing.xs},
  manualInput: {minHeight: control.minTouch, borderWidth: 1, borderColor: colors.border, borderRadius: radius.md, paddingHorizontal: 12, paddingVertical: 10, color: colors.text, backgroundColor: colors.surface, fontWeight: '800', fontSize: 16},
  inputError: {borderColor: colors.danger},
  helper: {fontSize: 11, lineHeight: 16, color: colors.textMuted, marginTop: spacing.xs},
  fieldError: {color: colors.danger},
  note: {borderWidth: 1, borderColor: colors.border, borderRadius: radius.md, padding: 14, marginTop: spacing.xs, backgroundColor: colors.surfaceMuted},
  noteTitle: {fontSize: 13, lineHeight: 18, fontWeight: '900', color: colors.text, marginBottom: spacing.xs},
  noteText: {fontSize: 13, lineHeight: 19, color: colors.textMuted},
  error: {fontSize: 13, lineHeight: 19, fontWeight: '700', color: colors.danger},
  pressed: {opacity: 0.72},
});
