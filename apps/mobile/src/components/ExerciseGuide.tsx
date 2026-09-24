import React from 'react';
import {Modal, Pressable, ScrollView, StyleSheet, Text, View} from 'react-native';
import {Exercise} from '../api/client';
import {colors, control, radius, spacing} from '../theme/tokens';

export function ExerciseGuide({exercise, visible, onClose}: {exercise: Exercise | null; visible: boolean; onClose: () => void}) {
  if (!exercise) return null;
  return (
    <Modal visible={visible} animationType="slide" presentationStyle="pageSheet" onRequestClose={onClose} accessibilityViewIsModal>
      <ScrollView contentContainerStyle={styles.container}>
        <View style={styles.header}>
          <View style={{flex: 1}}>
            <Text style={styles.eyebrow}>ТЕХНИКА</Text>
            <Text style={styles.title}>{exercise.name}</Text>
          </View>
          <Pressable accessibilityRole="button" accessibilityLabel="Закрыть инструкцию" testID="exercise-guide-close" onPress={onClose} style={styles.close}><Text style={styles.closeText}>✕</Text></Pressable>
        </View>
        <Text style={styles.description}>{exercise.description}</Text>
        {exercise.technique_tips.length > 0 ? (
          <View style={styles.keyTip}>
            <Text style={styles.keyTipLabel}>ГЛАВНЫЙ ОРИЕНТИР</Text>
            <Text style={styles.keyTipText}>{exercise.technique_tips[0]}</Text>
          </View>
        ) : null}
        <Section title="Как выполнять" items={exercise.instructions} numbered />
        <Section title="Частые ошибки" items={exercise.common_mistakes} />
        <Section title="Другие подсказки" items={exercise.technique_tips.slice(1)} />
        <View style={styles.note}><Text style={styles.noteText}>Если движение вызывает резкую или необычную боль — прекрати подход и не пытайся «продавить» амплитуду.</Text></View>
      </ScrollView>
    </Modal>
  );
}

function Section({title, items, numbered = false}: {title: string; items: string[]; numbered?: boolean}) {
  if (items.length === 0 && !numbered) return null;
  return <View style={styles.section}>
    <Text accessibilityRole="header" style={styles.sectionTitle}>{title}</Text>
    {items.length === 0 ? <Text style={styles.item}>Для этого упражнения шаги пока не добавлены.</Text> : null}
    {items.map((item, index) => numbered ? (
      <View key={`${title}-${index}`} style={styles.step}>
        <View style={styles.stepNumber} accessibilityElementsHidden importantForAccessibility="no"><Text style={styles.stepNumberText}>{index + 1}</Text></View>
        <Text style={styles.stepText} accessibilityLabel={`Шаг ${index + 1}. ${item}`}>{item}</Text>
      </View>
    ) : <Text key={`${title}-${index}`} style={styles.item}>• {item}</Text>)}
  </View>;
}

const styles = StyleSheet.create({
  container: {padding: spacing.lg, gap: spacing.md, backgroundColor: colors.background},
  header: {flexDirection: 'row', alignItems: 'flex-start', gap: 12},
  eyebrow: {fontSize: 12, fontWeight: '800', letterSpacing: 1.4, color: colors.textMuted},
  title: {fontSize: 26, lineHeight: 32, fontWeight: '900', marginTop: 4, color: colors.text},
  close: {width: control.minTouch, height: control.minTouch, borderWidth: 1, borderColor: colors.border, borderRadius: radius.pill, alignItems: 'center', justifyContent: 'center'},
  closeText: {fontSize: 18, fontWeight: '800'},
  description: {fontSize: 16, lineHeight: 24, color: colors.textMuted},
  keyTip: {borderWidth: 1, borderColor: colors.border, borderRadius: radius.md, backgroundColor: colors.surfaceMuted, padding: spacing.md, gap: spacing.xs},
  keyTipLabel: {fontSize: 11, fontWeight: '900', letterSpacing: 1, color: colors.textMuted},
  keyTipText: {fontSize: 16, lineHeight: 23, fontWeight: '700', color: colors.text},
  section: {gap: spacing.sm, borderTopWidth: 1, borderTopColor: colors.border, paddingTop: spacing.md},
  sectionTitle: {fontSize: 19, fontWeight: '800', color: colors.text},
  step: {borderWidth: 1, borderColor: colors.border, borderRadius: radius.md, padding: spacing.sm, flexDirection: 'row', alignItems: 'center', gap: spacing.sm, minHeight: control.minTouch},
  stepNumber: {width: 32, height: 32, borderRadius: radius.pill, backgroundColor: colors.primary, alignItems: 'center', justifyContent: 'center'},
  stepNumberText: {fontSize: 15, fontWeight: '800', color: colors.inverse},
  stepText: {flex: 1, fontSize: 15, lineHeight: 22, color: colors.text},
  item: {fontSize: 15, lineHeight: 22, color: colors.text},
  note: {borderWidth: 1, borderColor: colors.border, borderRadius: radius.md, padding: spacing.md},
  noteText: {fontSize: 14, lineHeight: 20, fontWeight: '600', color: colors.text},
});
