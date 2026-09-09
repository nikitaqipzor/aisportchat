import React, {useState} from 'react';
import {Pressable, ScrollView, StyleSheet, Text, View} from 'react-native';
import {AppButton} from '../components/AppButton';
import {colors, control, radius, spacing} from '../theme/tokens';

const goals = [
  ['muscle_gain', 'Набрать мышцы', 'Увеличить мышечную массу'],
  ['fat_loss', 'Снизить вес', 'Постепенно уменьшать жировую массу'],
  ['recomposition', 'Рекомпозиция', 'Снижать жир и сохранять мышцы'],
  ['strength', 'Стать сильнее', 'Развивать силовые показатели'],
  ['maintenance', 'Поддерживать форму', 'Сохранить текущий уровень'],
  ['endurance', 'Развить выносливость', 'Лучше переносить длительную нагрузку'],
] as const;

export function GoalScreen({onContinue}: {onContinue: (goal: string) => Promise<void>}) {
  const [selected, setSelected] = useState<(typeof goals)[number][0]>('muscle_gain');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  async function submit() {
    if (loading) {
      return;
    }
    try {
      setLoading(true);
      setError('');
      await onContinue(selected);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось сохранить цель. Попробуйте ещё раз.');
    } finally {
      setLoading(false);
    }
  }

  return (
    <ScrollView
      testID="goal-screen"
      contentContainerStyle={styles.container}
      showsVerticalScrollIndicator={false}>
      <View style={styles.content}>
        <Text style={styles.step}>ШАГ 1 ИЗ 3</Text>
        <Text accessibilityRole="header" style={styles.title}>Какая у тебя цель?</Text>
        <Text style={styles.subtitle}>
          Мы используем её для объёма, интенсивности и прогрессии тренировок. Цель можно изменить позже.
        </Text>
        <View style={styles.list} accessibilityRole="radiogroup">
          {goals.map(([id, label, description]) => {
            const isSelected = selected === id;
            return (
              <Pressable
                key={id}
                testID={`goal-${id}`}
                accessibilityRole="radio"
                accessibilityLabel={label}
                accessibilityHint={description}
                accessibilityState={{selected: isSelected, disabled: loading}}
                disabled={loading}
                onPress={() => {
                  setSelected(id);
                  setError('');
                }}
                style={({pressed}) => [
                  styles.card,
                  isSelected && styles.selected,
                  pressed && styles.pressed,
                ]}>
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
          accessibilityLabel={`Продолжить с целью «${goals.find(([id]) => id === selected)?.[1]}»`}
          testID="goal-continue"
          loading={loading}
          onPress={() => void submit()}
        />
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flexGrow: 1,
    padding: spacing.xl,
    paddingTop: 54,
    backgroundColor: colors.background,
  },
  content: {width: '100%', maxWidth: 600, alignSelf: 'center'},
  step: {fontSize: 12, fontWeight: '800', letterSpacing: 1.4, color: colors.textMuted},
  title: {fontSize: 32, lineHeight: 39, fontWeight: '900', marginTop: 12, color: colors.text},
  subtitle: {fontSize: 16, lineHeight: 23, color: colors.textMuted, marginTop: spacing.sm},
  list: {gap: 10, marginTop: spacing.lg},
  card: {
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: radius.lg,
    paddingHorizontal: 18,
    paddingVertical: 14,
    minHeight: control.minTouch,
    justifyContent: 'center',
    backgroundColor: colors.surface,
  },
  selected: {backgroundColor: colors.primary, borderColor: colors.primary},
  pressed: {opacity: 0.72},
  cardText: {fontSize: 17, lineHeight: 22, fontWeight: '800', color: colors.text},
  selectedText: {color: colors.inverse},
  cardDescription: {fontSize: 13, lineHeight: 18, color: colors.textMuted, marginTop: 3},
  selectedDescription: {color: colors.inverse, opacity: 0.76},
  footer: {width: '100%', maxWidth: 600, alignSelf: 'center', marginTop: 'auto', paddingTop: spacing.lg},
  error: {color: colors.danger, fontWeight: '700', lineHeight: 20, marginBottom: spacing.sm},
});
