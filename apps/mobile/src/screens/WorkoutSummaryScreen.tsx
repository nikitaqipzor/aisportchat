import React from 'react';
import {ScrollView, StyleSheet, Text, View} from 'react-native';
import {AppButton} from '../components/AppButton';
import {FinishResult} from '../api/client';
import {colors, radius, spacing} from '../theme/tokens';

export function WorkoutSummaryScreen({result, onDone}: {result: FinishResult; onDone: () => void}) {
  const completedSets = result.workout.exercises.reduce((sum, item) => sum + item.sets.length, 0);
  return (
    <ScrollView testID="workout-summary-screen" contentContainerStyle={styles.container}>
      <Text style={styles.eyebrow}>ГОТОВО</Text>
      <Text accessibilityRole="header" style={styles.title}>Тренировка завершена</Text>
      <Text style={styles.lead}>Результат сохранён в истории. Отличная работа — восстановись перед следующей нагрузкой.</Text>
      <View style={styles.stats}>
        <View style={styles.stat}><Text style={styles.value}>{completedSets}</Text><Text style={styles.label}>подходов</Text></View>
        <View style={styles.stat}><Text style={styles.value}>{Math.round(result.workout.workout.total_volume)}</Text><Text style={styles.label}>кг объёма</Text></View>
        <View style={styles.stat}><Text style={styles.value}>{Math.round(result.workout.workout.completion_percent)}%</Text><Text style={styles.label}>плана</Text></View>
      </View>
      {result.workout.workout.ended_early ? <Text accessibilityRole="alert" style={styles.early}>Тренировка завершена досрочно — это отмечено в истории и аналитике.</Text> : null}

      {result.personal_records.length > 0 ? (
        <>
          <Text style={styles.section}>🏆 Новые рекорды</Text>
          {result.personal_records.map(record => {
            const exercise = result.workout.exercises.find(item => item.exercise.id === record.exercise_id)?.exercise;
            const label = record.record_type === 'max_weight' ? 'макс. вес' : record.record_type === 'max_reps' ? 'макс. повторы' : 'расчётный 1ПМ';
            const suffix = record.record_type === 'max_reps' ? ' повт.' : ' кг';
            return (
              <View key={record.id} style={styles.recordCard}>
                <Text style={styles.name}>{exercise?.name ?? record.exercise_id}</Text>
                <Text style={styles.recordValue}>{label}: {record.value}{suffix}</Text>
                {record.previous_value !== undefined ? <Text style={styles.previous}>Было: {record.previous_value}{suffix}</Text> : <Text style={styles.previous}>Первый зафиксированный результат</Text>}
              </View>
            );
          })}
        </>
      ) : null}

      <Text style={styles.section}>Следующая прогрессия</Text>
      {result.next_recommendations.length===0?<View style={styles.empty}><Text style={styles.name}>Рекомендации появятся позже</Text><Text style={styles.message}>Системе нужно больше выполненных подходов, чтобы предложить следующую прогрессию.</Text></View>:result.next_recommendations.map(item => (
        <View key={item.exercise_id} style={styles.card}>
          <Text style={styles.name}>{item.exercise_name}</Text>
          <Text style={styles.message}>{item.message}</Text>
        </View>
      ))}

      <AppButton label="На главную" testID="summary-home" onPress={onDone} />
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {padding: spacing.lg, paddingBottom: spacing.xl, gap: spacing.md, backgroundColor:colors.background,flexGrow:1},
  eyebrow: {fontSize: 12, fontWeight: '800', letterSpacing: 1.4,color:colors.textMuted},
  title: {fontSize: 32, lineHeight: 38, fontWeight: '800',color:colors.text},
  lead:{fontSize:14,lineHeight:20,color:colors.textMuted},
  stats: {flexDirection: 'row', gap: 8},
  stat: {flex: 1, borderWidth: 1,borderColor:colors.border,borderRadius: radius.lg, padding: 14},
  value: {fontSize: 22, fontWeight: '800',color:colors.text},
  label: {fontSize: 12,color:colors.textMuted,marginTop: 3},
  section: {fontSize: 20, fontWeight: '800', marginTop: 6,color:colors.text},
  card: {borderWidth: 1,borderColor:colors.border,borderRadius: radius.md, padding: 15, gap: 5},
  name: {fontSize: 16, fontWeight: '800',color:colors.text},
  message: {fontSize: 14, lineHeight: 20,color:colors.textMuted},
  recordCard: {borderWidth: 1,borderColor:colors.border,borderRadius: radius.md, padding: 15, gap: 4},
  recordValue: {fontSize: 16, fontWeight: '900',color:colors.text},
  previous: {fontSize: 12,color:colors.textMuted},
  early: {fontSize: 13, lineHeight: 19, fontWeight: '700',color:colors.danger,borderWidth:1,borderColor:colors.danger,borderRadius:radius.md,padding:12},
  empty:{borderWidth:1,borderColor:colors.border,borderRadius:radius.md,padding:spacing.md,gap:spacing.xs},
});
