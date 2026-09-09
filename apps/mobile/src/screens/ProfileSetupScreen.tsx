import React, {useRef, useState} from 'react';
import {Pressable, ScrollView, StyleSheet, Text, TextInput, type TextInputInstance, useWindowDimensions, View} from 'react-native';
import {AppButton} from '../components/AppButton';
import {colors, control, radius, spacing} from '../theme/tokens';

const levels = [
  ['beginner', 'Новичок', 'Тренируюсь меньше 6 месяцев'],
  ['intermediate', 'Средний', 'Регулярно тренируюсь 6–24 месяца'],
  ['advanced', 'Продвинутый', 'Системно тренируюсь более 2 лет'],
] as const;

function parseNumber(value: string) {
  return Number(value.replace(',', '.'));
}

export function ProfileSetupScreen({onContinue}: {onContinue: (data: {height: number; weight: number; level: string}) => Promise<void>}) {
  const {width: viewportWidth} = useWindowDimensions();
  const weightRef = useRef<TextInputInstance>(null);
  const [height, setHeight] = useState('');
  const [weight, setWeight] = useState('');
  const [level, setLevel] = useState<(typeof levels)[number][0]>('beginner');
  const [heightTouched, setHeightTouched] = useState(false);
  const [weightTouched, setWeightTouched] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const heightNumber = parseNumber(height);
  const weightNumber = parseNumber(weight);
  const heightError = !height.trim()
    ? 'Укажите рост'
    : !Number.isInteger(heightNumber) || heightNumber < 100 || heightNumber > 250
      ? 'Допустимый рост: от 100 до 250 см'
      : '';
  const weightError = !weight.trim()
    ? 'Укажите вес'
    : !Number.isFinite(weightNumber) || weightNumber < 30 || weightNumber > 350
      ? 'Допустимый вес: от 30 до 350 кг'
      : '';
  const valid = !heightError && !weightError;

  async function submit() {
    setHeightTouched(true);
    setWeightTouched(true);
    if (!valid || loading) {
      return;
    }
    try {
      setLoading(true);
      setError('');
      await onContinue({height: heightNumber, weight: weightNumber, level});
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось сохранить профиль. Попробуйте ещё раз.');
    } finally {
      setLoading(false);
    }
  }

  return (
    <ScrollView
      testID="profile-setup-screen"
      contentContainerStyle={styles.container}
      keyboardShouldPersistTaps="handled"
      showsVerticalScrollIndicator={false}>
      <View style={styles.content}>
        <Text style={styles.step}>ШАГ 2 ИЗ 3</Text>
        <Text accessibilityRole="header" style={styles.title}>Расскажи о себе</Text>
        <Text style={styles.subtitle}>Параметры нужны для безопасной стартовой нагрузки. Их можно обновить в профиле.</Text>

        <View style={[styles.row, viewportWidth < 360 && styles.rowNarrow]}>
          <View style={styles.field}>
            <Text nativeID="profile-height-label" style={styles.label}>Рост, см</Text>
            <TextInput
              accessibilityLabelledBy="profile-height-label"
              accessibilityHint="От 100 до 250 сантиметров"
              testID="profile-height"
              keyboardType="number-pad"
              returnKeyType="next"
              value={height}
              editable={!loading}
              maxLength={3}
              onBlur={() => setHeightTouched(true)}
              onChangeText={value => {
                setHeight(value);
                setError('');
              }}
              onSubmitEditing={() => weightRef.current?.focus()}
              style={[styles.input, heightTouched && heightError ? styles.inputError : null]}
              placeholder="Например, 180"
              placeholderTextColor={colors.textMuted}
            />
            {heightTouched && heightError ? <Text accessibilityRole="alert" style={styles.fieldError}>{heightError}</Text> : null}
          </View>
          <View style={styles.field}>
            <Text nativeID="profile-weight-label" style={styles.label}>Вес, кг</Text>
            <TextInput
              ref={weightRef}
              accessibilityLabelledBy="profile-weight-label"
              accessibilityHint="От 30 до 350 килограммов"
              testID="profile-weight"
              keyboardType="decimal-pad"
              returnKeyType="done"
              value={weight}
              editable={!loading}
              maxLength={6}
              onBlur={() => setWeightTouched(true)}
              onChangeText={value => {
                setWeight(value);
                setError('');
              }}
              onSubmitEditing={() => void submit()}
              style={[styles.input, weightTouched && weightError ? styles.inputError : null]}
              placeholder="Например, 75,5"
              placeholderTextColor={colors.textMuted}
            />
            {weightTouched && weightError ? <Text accessibilityRole="alert" style={styles.fieldError}>{weightError}</Text> : null}
          </View>
        </View>

        <Text style={styles.section}>Опыт тренировок</Text>
        <View style={styles.list} accessibilityRole="radiogroup">
          {levels.map(([id, label, description]) => {
            const isSelected = level === id;
            return (
              <Pressable
                key={id}
                testID={`profile-level-${id}`}
                accessibilityRole="radio"
                accessibilityLabel={label}
                accessibilityHint={description}
                accessibilityState={{selected: isSelected, disabled: loading}}
                disabled={loading}
                onPress={() => {
                  setLevel(id);
                  setError('');
                }}
                style={({pressed}) => [styles.card, isSelected && styles.selected, pressed && styles.pressed]}>
                <Text style={[styles.cardText, isSelected && styles.selectedText]}>{label}</Text>
                <Text style={[styles.cardDescription, isSelected && styles.selectedDescription]}>{description}</Text>
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
          label="Продолжить"
          accessibilityLabel="Сохранить параметры профиля и продолжить"
          testID="profile-continue"
          loading={loading}
          disabled={!valid}
          onPress={() => void submit()}
        />
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {flexGrow: 1, padding: spacing.xl, paddingTop: 54, backgroundColor: colors.background},
  content: {width: '100%', maxWidth: 600, alignSelf: 'center'},
  step: {fontSize: 12, fontWeight: '800', color: colors.textMuted, letterSpacing: 1.4},
  title: {fontSize: 32, lineHeight: 39, fontWeight: '900', marginTop: 12, color: colors.text},
  subtitle: {fontSize: 15, lineHeight: 22, color: colors.textMuted, marginTop: spacing.sm, marginBottom: spacing.lg},
  row: {flexDirection: 'row', gap: 12},
  rowNarrow: {flexDirection: 'column'},
  field: {flex: 1},
  label: {fontSize: 14, lineHeight: 20, fontWeight: '800', marginBottom: spacing.sm, color: colors.text},
  input: {
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: radius.md,
    paddingHorizontal: 14,
    paddingVertical: 13,
    fontSize: 16,
    color: colors.text,
    backgroundColor: colors.surface,
    minHeight: control.minTouch,
  },
  inputError: {borderColor: colors.danger},
  fieldError: {fontSize: 12, lineHeight: 17, color: colors.danger, marginTop: spacing.xs},
  section: {fontSize: 17, lineHeight: 23, fontWeight: '800', marginTop: spacing.lg, marginBottom: spacing.sm, color: colors.text},
  list: {gap: 10},
  card: {
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: radius.md,
    paddingHorizontal: spacing.md,
    paddingVertical: 13,
    minHeight: control.minTouch,
    justifyContent: 'center',
    backgroundColor: colors.surface,
  },
  selected: {backgroundColor: colors.primary, borderColor: colors.primary},
  cardText: {fontSize: 16, lineHeight: 21, fontWeight: '800', color: colors.text},
  selectedText: {color: colors.inverse},
  cardDescription: {fontSize: 12, lineHeight: 17, color: colors.textMuted, marginTop: 3},
  selectedDescription: {color: colors.inverse, opacity: 0.76},
  pressed: {opacity: 0.72},
  footer: {width: '100%', maxWidth: 600, alignSelf: 'center', marginTop: 'auto', paddingTop: spacing.lg},
  error: {color: colors.danger, fontWeight: '700', lineHeight: 20, marginBottom: spacing.sm},
});
