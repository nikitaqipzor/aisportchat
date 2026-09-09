import React, {useEffect, useState} from 'react';
import {ActivityIndicator, Pressable, ScrollView, StyleSheet, Text, View} from 'react-native';
import {api, HistoryFilters, WorkoutView} from '../api/client';
import {MuscleId, muscleMeta} from '../domain/muscles';

type EnvironmentFilter = 'all' | 'gym' | 'home' | 'band';

type StatusFilter = 'all' | 'completed' | 'planned' | 'active';

export function HistoryScreen({
  accessToken,
  onBack,
  onSelect,
}: {
  accessToken: string;
  onBack: () => void;
  onSelect: (workoutId: string) => void;
}) {
  const [items, setItems] = useState<WorkoutView[]>([]);
  const [environment, setEnvironment] = useState<EnvironmentFilter>('all');
  const [status, setStatus] = useState<StatusFilter>('completed');
  const [favoritesOnly, setFavoritesOnly] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    void load();
  }, [accessToken, environment, status, favoritesOnly]);

  async function load() {
    try {
      setLoading(true);
      setError('');
      const filters: HistoryFilters = {limit: 50};
      if (environment !== 'all') filters.environment = environment;
      if (status !== 'all') filters.status = status;
      if (favoritesOnly) filters.favorite = true;
      setItems((await api.workoutHistory(accessToken, filters)).items);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось загрузить историю');
    } finally {
      setLoading(false);
    }
  }

  return (
    <ScrollView contentContainerStyle={styles.container}>
      <Pressable onPress={onBack}><Text style={styles.back}>← Главная</Text></Pressable>
      <Text style={styles.title}>История</Text>
      <Text style={styles.subtitle}>Фильтруй тренировки и открывай подробности каждого занятия.</Text>

      <Text style={styles.filterTitle}>Место</Text>
      <View style={styles.chips}>
        {([['all', 'Все'], ['gym', 'Зал'], ['home', 'Дом'], ['band', 'Резина']] as const).map(([id, label]) => (
          <Chip key={id} active={environment === id} label={label} onPress={() => setEnvironment(id)} />
        ))}
      </View>
      <Text style={styles.filterTitle}>Статус</Text>
      <View style={styles.chips}>
        {([['completed', 'Готовые'], ['all', 'Все'], ['planned', 'Запланированные'], ['active', 'Активные']] as const).map(([id, label]) => (
          <Chip key={id} active={status === id} label={label} onPress={() => setStatus(id)} />
        ))}
        <Chip active={favoritesOnly} label="★ Избранные" onPress={() => setFavoritesOnly(value => !value)} />
      </View>

      {loading ? <ActivityIndicator /> : null}
      {error ? <Text style={styles.error}>{error}</Text> : null}
      {!loading && items.length === 0 ? <Text style={styles.empty}>По выбранным фильтрам тренировок пока нет.</Text> : null}

      {items.map(item => {
        const muscle = item.workout.muscle as MuscleId;
        return (
          <Pressable key={item.workout.id} style={styles.card} onPress={() => onSelect(item.workout.id)}>
            <View style={styles.row}>
              <Text style={styles.name}>{muscleMeta[muscle]?.title ?? item.workout.muscle}</Text>
              <Text style={styles.status}>{item.workout.favorite ? '★ ' : ''}{item.workout.status}</Text>
            </View>
            <Text style={styles.meta}>{new Date(item.workout.completed_at ?? item.workout.created_at).toLocaleString('ru-RU')} · {item.workout.environment}</Text>
            <View style={styles.metrics}>
              <Text style={styles.volume}>{Math.round(item.workout.total_volume)} кг объёма</Text>
              <Text style={styles.pr}>{item.personal_records.length > 0 ? `🏆 ${item.personal_records.length} PR` : `${item.exercises.length} упр.`}</Text>
            </View>
          </Pressable>
        );
      })}
    </ScrollView>
  );
}

function Chip({active, label, onPress}: {active: boolean; label: string; onPress: () => void}) {
  return <Pressable onPress={onPress} style={[styles.chip, active && styles.chipActive]}><Text style={[styles.chipText, active && styles.chipTextActive]}>{label}</Text></Pressable>;
}

const styles = StyleSheet.create({
  container: {padding: 20, gap: 12},
  back: {fontWeight: '800'},
  title: {fontSize: 32, fontWeight: '900', marginTop: 6},
  subtitle: {fontSize: 14, lineHeight: 20, opacity: 0.58},
  filterTitle: {fontSize: 12, fontWeight: '800', opacity: 0.5, marginTop: 4},
  chips: {flexDirection: 'row', flexWrap: 'wrap', gap: 7},
  chip: {borderWidth: 1, borderRadius: 14, paddingHorizontal: 11, paddingVertical: 8},
  chipActive: {backgroundColor: '#111'},
  chipText: {fontSize: 12, fontWeight: '800'},
  chipTextActive: {color: '#fff'},
  card: {borderWidth: 1, borderRadius: 18, padding: 15, gap: 6},
  row: {flexDirection: 'row', justifyContent: 'space-between', gap: 10},
  name: {fontSize: 18, fontWeight: '900', flex: 1},
  status: {fontSize: 12, fontWeight: '800', opacity: 0.55},
  meta: {fontSize: 13, opacity: 0.55},
  metrics: {flexDirection: 'row', justifyContent: 'space-between', marginTop: 4},
  volume: {fontSize: 13, fontWeight: '800'},
  pr: {fontSize: 13, fontWeight: '800'},
  empty: {fontSize: 15, opacity: 0.6, paddingVertical: 12},
  error: {color: '#8b1e1e'},
});
