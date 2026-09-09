import React, {useEffect, useMemo, useState} from 'react';
import {ActivityIndicator, Pressable, ScrollView, StyleSheet, Text, View} from 'react-native';
import {api, MuscleStats, Readiness, WorkoutView} from '../api/client';
import {MuscleId, muscleMeta} from '../domain/muscles';
import {currentLocalDate} from '../domain/date';

type Environment = 'home' | 'gym' | 'band';

function readiness(stats: MuscleStats | null) {
  if (!stats || stats.days_since_last_workout === undefined) return {label: 'Нет истории', detail: 'Можно начать с базовой тренировки.'};
  if (stats.days_since_last_workout === 0) return {label: 'Недавно тренировалась', detail: 'Эта метка основана только на истории тренировок.'};
  if (stats.days_since_last_workout === 1) return {label: 'Средняя пауза', detail: 'Оцени самочувствие перед высокой нагрузкой.'};
  return {label: 'Давно не тренировалась', detail: 'По истории нагрузки пауза составляет 2+ дня.'};
}

export function MuscleDetailScreen({
  accessToken,
  muscle,
  environment,
  onBack,
  onWorkout,
}: {
  accessToken: string;
  muscle: MuscleId;
  environment: Environment;
  onBack: () => void;
  onWorkout: (workout: WorkoutView) => void;
}) {
  const [stats, setStats] = useState<MuscleStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const [error, setError] = useState('');
  const [dailyReadiness, setDailyReadiness] = useState<Readiness | null>(null);
  const status = useMemo(() => readiness(stats), [stats]);

  useEffect(() => {
    setLoading(true);
    Promise.all([api.muscleStats(accessToken, muscle), api.recoveryToday(accessToken, currentLocalDate()).catch(() => null)])
      .then(([nextStats, recovery]) => { setStats(nextStats); setDailyReadiness(recovery); })
      .catch(e => setError(e instanceof Error ? e.message : 'Не удалось загрузить статистику мышцы.'))
      .finally(() => setLoading(false));
  }, [accessToken, muscle]);

  async function generate() {
    try {
      setError('');
      setGenerating(true);
      onWorkout(await api.generateWorkout(accessToken, muscle, environment, currentLocalDate()));
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось создать тренировку.');
    } finally {
      setGenerating(false);
    }
  }

  return (
    <ScrollView contentContainerStyle={styles.container}>
      <Pressable onPress={onBack}><Text style={styles.back}>← Карта тела</Text></Pressable>
      <Text style={styles.eyebrow}>ГРУППА МЫШЦ</Text>
      <Text style={styles.title}>{muscleMeta[muscle].title}</Text>
      <Text style={styles.subtitle}>Режим: {environment === 'gym' ? 'зал' : environment === 'home' ? 'дома' : 'резинки'}</Text>

      {loading ? <ActivityIndicator /> : (
        <>
          <View style={styles.heroCard}>
            <Text style={styles.heroLabel}>По истории тренировок</Text>
            <Text style={styles.heroValue}>{status.label}</Text>
            <Text style={styles.heroNote}>{status.detail}</Text>
          </View>


          {dailyReadiness ? (() => { const r = dailyReadiness.muscles.find(item => item.muscle === muscle); return r ? <View style={styles.recoveryCard}><Text style={styles.heroLabel}>Recovery сегодня</Text><Text style={styles.recoveryValue}>{r.score}/100 · {r.status === 'ready' ? 'готова' : r.status === 'moderate' ? 'частично восстановлена' : 'утомлена'}</Text><Text style={styles.heroNote}>Soreness {r.soreness}/5 · {r.recent_sets_48h} экв. подходов за 48 часов. Генератор автоматически учтёт этот показатель.</Text></View> : null; })() : null}

          <View style={styles.statsGrid}>
            <Stat value={stats?.completed_workouts ?? 0} label="тренировок" />
            <Stat value={stats?.total_sets ?? 0} label="подходов" />
            <Stat value={Math.round(stats?.total_volume ?? 0)} label="кг объёма" />
            <Stat value={stats?.personal_records_count ?? 0} label="PR-событий" />
          </View>

          <View style={styles.card}>
            <Text style={styles.cardTitle}>Последняя нагрузка</Text>
            <Text style={styles.row}>Последняя тренировка: {stats?.days_since_last_workout === undefined ? 'ещё не было' : stats.days_since_last_workout === 0 ? 'сегодня' : `${stats.days_since_last_workout} дн. назад`}</Text>
            <Text style={styles.row}>Объём последней: {Math.round(stats?.last_workout_volume ?? 0)} кг</Text>
            <Text style={styles.row}>Объём последних 4: {Math.round(stats?.recent_volume ?? 0)} кг</Text>
          </View>
        </>
      )}

      {error ? <Text style={styles.error}>{error}</Text> : null}
      <Pressable style={styles.primary} onPress={generate} disabled={generating}>
        {generating ? <ActivityIndicator color="#fff" /> : <Text style={styles.primaryText}>СОЗДАТЬ ТРЕНИРОВКУ</Text>}
      </Pressable>
    </ScrollView>
  );
}

function Stat({value, label}: {value: number; label: string}) {
  return <View style={styles.stat}><Text style={styles.statValue}>{value}</Text><Text style={styles.statLabel}>{label}</Text></View>;
}

const styles = StyleSheet.create({
  container: {padding: 20, gap: 14},
  back: {fontWeight: '800'},
  eyebrow: {fontSize: 11, fontWeight: '800', letterSpacing: 1.5, opacity: 0.5, marginTop: 8},
  title: {fontSize: 36, lineHeight: 41, fontWeight: '900'},
  subtitle: {fontSize: 15, opacity: 0.55},
  heroCard: {borderWidth: 1, borderRadius: 22, padding: 18, gap: 6},
  heroLabel: {fontSize: 12, fontWeight: '800', opacity: 0.5},
  heroValue: {fontSize: 23, fontWeight: '900'},
  heroNote: {fontSize: 13, lineHeight: 18, opacity: 0.58},
  recoveryCard: {borderWidth: 1, borderRadius: 18, padding: 15, gap: 5},
  recoveryValue: {fontSize: 18, fontWeight: '900'},
  statsGrid: {flexDirection: 'row', flexWrap: 'wrap', gap: 10},
  stat: {width: '48%', borderWidth: 1, borderRadius: 18, padding: 14},
  statValue: {fontSize: 24, fontWeight: '900'},
  statLabel: {fontSize: 12, opacity: 0.55, marginTop: 3},
  card: {borderWidth: 1, borderRadius: 18, padding: 16, gap: 7},
  cardTitle: {fontSize: 16, fontWeight: '900', marginBottom: 2},
  row: {fontSize: 14, opacity: 0.68},
  primary: {backgroundColor: '#111', minHeight: 54, borderRadius: 16, alignItems: 'center', justifyContent: 'center', marginTop: 6},
  primaryText: {color: '#fff', fontWeight: '900'},
  error: {color: '#8b1e1e'},
});
