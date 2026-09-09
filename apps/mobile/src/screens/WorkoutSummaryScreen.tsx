import React from 'react';
import {ScrollView, StyleSheet, Text, View} from 'react-native';
import {AppButton} from '../components/AppButton';
import {FinishResult} from '../api/client';

export function WorkoutSummaryScreen({result, onDone}: {result: FinishResult; onDone: () => void}) {
  const completedSets = result.workout.exercises.reduce((sum, item) => sum + item.sets.length, 0);
  return (
    <ScrollView contentContainerStyle={styles.container}>
      <Text style={styles.eyebrow}>ГОТОВО</Text>
      <Text style={styles.title}>Тренировка завершена</Text>
      <View style={styles.stats}>
        <View style={styles.stat}><Text style={styles.value}>{completedSets}</Text><Text style={styles.label}>подходов</Text></View>
        <View style={styles.stat}><Text style={styles.value}>{Math.round(result.workout.workout.total_volume)}</Text><Text style={styles.label}>кг объёма</Text></View>
        <View style={styles.stat}><Text style={styles.value}>{Math.round(result.workout.workout.completion_percent)}%</Text><Text style={styles.label}>плана</Text></View>
      </View>
      {result.workout.workout.ended_early ? <Text style={styles.early}>Тренировка завершена досрочно — это отмечено в истории и аналитике.</Text> : null}

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
      {result.next_recommendations.map(item => (
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
  container: {padding: 20, gap: 16},
  eyebrow: {fontSize: 12, fontWeight: '800', letterSpacing: 1.4},
  title: {fontSize: 32, lineHeight: 38, fontWeight: '800'},
  stats: {flexDirection: 'row', gap: 8},
  stat: {flex: 1, borderWidth: 1, borderRadius: 18, padding: 14},
  value: {fontSize: 22, fontWeight: '800'},
  label: {fontSize: 12, opacity: 0.55, marginTop: 3},
  section: {fontSize: 20, fontWeight: '800', marginTop: 6},
  card: {borderWidth: 1, borderRadius: 16, padding: 15, gap: 5},
  name: {fontSize: 16, fontWeight: '800'},
  message: {fontSize: 14, lineHeight: 20, opacity: 0.7},
  recordCard: {borderWidth: 1, borderRadius: 16, padding: 15, gap: 4},
  recordValue: {fontSize: 16, fontWeight: '900'},
  previous: {fontSize: 12, opacity: 0.55},
  early: {fontSize: 13, lineHeight: 19, fontWeight: '700', opacity: 0.65},
});
