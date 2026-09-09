import React, {useState} from 'react';
import {Pressable, ScrollView, StyleSheet, Text, View} from 'react-native';
import {AppButton} from '../components/AppButton';
import {colors, control, radius, spacing} from '../theme/tokens';

const environments = [
  ['home', 'Дома'],
  ['gym', 'В зале'],
  ['band', 'С резинками'],
] as const;

const equipment = [
  ['bodyweight', 'Собственный вес'],
  ['dumbbells', 'Гантели'],
  ['barbell', 'Штанга'],
  ['bench', 'Скамья'],
  ['rack', 'Стойка'],
  ['pullup_bar', 'Турник'],
  ['cable_machine', 'Блочные тренажёры'],
  ['leg_machine', 'Тренажёры для ног'],
  ['kettlebell', 'Гиря'],
  ['resistance_band', 'Резинки'],
] as const;

export function TrainingSetupScreen({onFinish}: {onFinish: (data: {environments: string[]; equipment: string[]; workouts: number; minutes: number}) => Promise<void>}) {
  const [places, setPlaces] = useState<string[]>(['gym']);
  const [gear, setGear] = useState<string[]>(['bodyweight', 'dumbbells', 'barbell', 'bench', 'rack', 'cable_machine', 'leg_machine']);
  const [workouts, setWorkouts] = useState(4);
  const [minutes, setMinutes] = useState(60);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const valid = places.length > 0;

  function togglePlace(value: string) {
    setPlaces(current => current.includes(value) ? current.filter(item => item !== value) : [...current, value]);
    setError('');
  }

  function toggleEquipment(value: string) {
    setGear(current => current.includes(value) ? current.filter(item => item !== value) : [...current, value]);
    setError('');
  }

  async function submit() {
    if (!valid || loading) {
      return;
    }
    try {
      setLoading(true);
      setError('');
      await onFinish({environments: places, equipment: gear, workouts, minutes});
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось завершить настройку. Попробуйте ещё раз.');
    } finally {
      setLoading(false);
    }
  }

  return (
    <ScrollView
      testID="training-setup-screen"
      contentContainerStyle={styles.container}
      showsVerticalScrollIndicator={false}>
      <View style={styles.content}>
        <Text style={styles.step}>ШАГ 3 ИЗ 3</Text>
        <Text accessibilityRole="header" style={styles.title}>Настроим тренировки</Text>
        <Text style={styles.subtitle}>Выбери все доступные варианты — программа будет использовать только их.</Text>

        <Text style={styles.section}>Где и как ты тренируешься?</Text>
        <View style={styles.wrap} accessibilityRole="group" accessibilityLabel="Места тренировок">
          {environments.map(([id, label]) => (
            <ChoiceChip
              key={id}
              id={id}
              label={label}
              selected={places.includes(id)}
              disabled={loading}
              testPrefix="training-place"
              onPress={() => togglePlace(id)}
            />
          ))}
        </View>
        {!valid ? (
          <Text accessibilityRole="alert" accessibilityLiveRegion="polite" style={styles.validationError}>
            Выбери хотя бы один вариант тренировки.
          </Text>
        ) : null}

        <Text style={styles.section}>Доступное оборудование</Text>
        <Text style={styles.helper}>Можно выбрать несколько вариантов или оставить список пустым.</Text>
        <View style={styles.wrap} accessibilityRole="group" accessibilityLabel="Доступное оборудование">
          {equipment.map(([id, label]) => (
            <ChoiceChip
              key={id}
              id={id}
              label={label}
              selected={gear.includes(id)}
              disabled={loading}
              testPrefix="training-equipment"
              onPress={() => toggleEquipment(id)}
            />
          ))}
        </View>

        <Text style={styles.section}>Тренировок в неделю</Text>
        <View style={styles.wrap} accessibilityRole="radiogroup">
          {[2, 3, 4, 5, 6].map(value => {
            const selected = workouts === value;
            return (
              <Pressable
                key={value}
                testID={`training-frequency-${value}`}
                accessibilityRole="radio"
                accessibilityLabel={`${value} тренировок в неделю`}
                accessibilityState={{selected, disabled: loading}}
                disabled={loading}
                onPress={() => {
                  setWorkouts(value);
                  setError('');
                }}
                style={({pressed}) => [styles.number, selected && styles.active, pressed && styles.pressed]}>
                <Text style={[styles.chipText, selected && styles.activeText]}>{value}</Text>
              </Pressable>
            );
          })}
        </View>

        <Text style={styles.section}>Время на одну тренировку</Text>
        <View style={styles.wrap} accessibilityRole="radiogroup">
          {[30, 45, 60, 90].map(value => {
            const selected = minutes === value;
            return (
              <Pressable
                key={value}
                testID={`training-duration-${value}`}
                accessibilityRole="radio"
                accessibilityLabel={`${value} минут на тренировку`}
                accessibilityState={{selected, disabled: loading}}
                disabled={loading}
                onPress={() => {
                  setMinutes(value);
                  setError('');
                }}
                style={({pressed}) => [styles.chip, selected && styles.active, pressed && styles.pressed]}>
                <Text style={[styles.chipText, selected && styles.activeText]}>{value} мин</Text>
              </Pressable>
            );
          })}
        </View>
      </View>

      <View style={styles.footer}>
        {error ? (
          <Text accessibilityRole="alert" accessibilityLiveRegion="assertive" style={styles.error}>{error}</Text>
        ) : null}
        <AppButton
          label="Начать тренироваться"
          accessibilityLabel="Сохранить настройки и начать тренироваться"
          testID="training-finish"
          loading={loading}
          disabled={!valid}
          onPress={() => void submit()}
        />
      </View>
    </ScrollView>
  );
}

function ChoiceChip({id, label, selected, disabled, testPrefix, onPress}: {
  id: string;
  label: string;
  selected: boolean;
  disabled: boolean;
  testPrefix: string;
  onPress: () => void;
}) {
  return (
    <Pressable
      testID={`${testPrefix}-${id}`}
      accessibilityRole="checkbox"
      accessibilityLabel={label}
      accessibilityState={{checked: selected, disabled}}
      disabled={disabled}
      onPress={onPress}
      style={({pressed}) => [styles.chip, selected && styles.active, pressed && styles.pressed]}>
      <Text style={[styles.chipText, selected && styles.activeText]}>{label}</Text>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  container: {flexGrow: 1, padding: spacing.xl, paddingTop: 54, backgroundColor: colors.background},
  content: {width: '100%', maxWidth: 680, alignSelf: 'center'},
  step: {fontSize: 12, fontWeight: '800', color: colors.textMuted, letterSpacing: 1.4},
  title: {fontSize: 32, lineHeight: 39, fontWeight: '900', marginTop: 12, color: colors.text},
  subtitle: {fontSize: 15, lineHeight: 22, color: colors.textMuted, marginTop: spacing.sm, marginBottom: spacing.md},
  section: {fontSize: 17, lineHeight: 23, fontWeight: '800', marginTop: spacing.lg, marginBottom: spacing.sm, color: colors.text},
  helper: {fontSize: 12, lineHeight: 18, color: colors.textMuted, marginTop: -spacing.xs, marginBottom: spacing.sm},
  wrap: {flexDirection: 'row', flexWrap: 'wrap', gap: 10},
  chip: {borderWidth: 1, borderColor: colors.border, borderRadius: radius.pill, paddingHorizontal: spacing.md, minHeight: control.minTouch, justifyContent: 'center', backgroundColor: colors.surface},
  number: {width: control.minTouch, height: control.minTouch, borderWidth: 1, borderColor: colors.border, borderRadius: radius.pill, alignItems: 'center', justifyContent: 'center', backgroundColor: colors.surface},
  active: {backgroundColor: colors.primary, borderColor: colors.primary},
  chipText: {fontSize: 14, lineHeight: 19, fontWeight: '800', color: colors.text},
  activeText: {color: colors.inverse},
  pressed: {opacity: 0.72},
  validationError: {fontSize: 12, lineHeight: 18, color: colors.danger, marginTop: spacing.sm},
  footer: {width: '100%', maxWidth: 680, alignSelf: 'center', marginTop: 'auto', paddingTop: spacing.xl},
  error: {color: colors.danger, fontWeight: '700', lineHeight: 20, marginBottom: spacing.sm},
});
