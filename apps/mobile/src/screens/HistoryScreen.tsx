import React, {useEffect, useState} from 'react';
import {ActivityIndicator, Pressable, ScrollView, StyleSheet, Text, View} from 'react-native';
import {api, HistoryFilters, WorkoutView} from '../api/client';
import {MuscleId, muscleMeta} from '../domain/muscles';
import {AppButton} from '../components/AppButton';
import {colors, control, radius, spacing} from '../theme/tokens';

type EnvironmentFilter = 'all' | 'gym' | 'home' | 'band';

type StatusFilter = 'all' | 'completed' | 'planned' | 'active';
const statusTitles:Record<string,string>={completed:'Завершена',planned:'Запланирована',active:'Активна'};
const environmentTitles:Record<string,string>={gym:'Зал',home:'Дом',band:'Резина'};

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
      <Pressable accessibilityRole="button" accessibilityLabel="Вернуться на главную" hitSlop={8} onPress={onBack} style={styles.navButton}><Text style={styles.back}>← Главная</Text></Pressable>
      <Text accessibilityRole="header" style={styles.title}>История</Text>
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
        <Chip active={favoritesOnly} label="★ Избранные" role="checkbox" onPress={() => setFavoritesOnly(value => !value)} />
      </View>

      {loading ? <View style={styles.loading} accessibilityLiveRegion="polite"><ActivityIndicator color={colors.text}/><Text style={styles.muted}>Загружаем тренировки…</Text></View> : null}
      {error ? <View style={styles.errorBox}><Text accessibilityRole="alert" style={styles.error}>{error}</Text><AppButton label="Повторить" variant="secondary" onPress={()=>void load()}/></View> : null}
      {!loading && !error && items.length === 0 ? <View style={styles.emptyCard}><Text style={styles.emptyTitle}>Нет тренировок</Text><Text style={styles.muted}>По выбранным фильтрам ничего не найдено. Попробуйте сбросить фильтры.</Text><AppButton label="Сбросить фильтры" variant="secondary" onPress={()=>{setEnvironment('all');setStatus('completed');setFavoritesOnly(false)}}/></View> : null}

      {items.map(item => {
        const muscle = item.workout.muscle as MuscleId;
        return (
          <Pressable key={item.workout.id} accessibilityRole="button" accessibilityLabel={`Открыть тренировку: ${muscleMeta[muscle]?.title ?? item.workout.muscle}`} style={({pressed})=>[styles.card,pressed&&styles.pressed]} onPress={() => onSelect(item.workout.id)}>
            <View style={styles.row}>
              <Text style={styles.name}>{muscleMeta[muscle]?.title ?? item.workout.muscle}</Text>
              <Text style={styles.status}>{item.workout.favorite ? '★ ' : ''}{statusTitles[item.workout.status]??item.workout.status}</Text>
            </View>
            <Text style={styles.meta}>{new Date(item.workout.completed_at ?? item.workout.created_at).toLocaleString('ru-RU')} · {environmentTitles[item.workout.environment]??item.workout.environment}</Text>
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

function Chip({active, label, onPress,role='radio'}: {active: boolean; label: string; onPress: () => void;role?:'radio'|'checkbox'}) {
  return <Pressable accessibilityRole={role} accessibilityState={role==='checkbox'?{checked:active}:{selected:active}} onPress={onPress} style={({pressed})=>[styles.chip, active && styles.chipActive,pressed&&styles.pressed]}><Text style={[styles.chipText, active && styles.chipTextActive]}>{label}</Text></Pressable>;
}

const styles = StyleSheet.create({
  container: {padding: spacing.lg, gap: 12, backgroundColor: colors.background},
  navButton:{minHeight:control.minTouch,justifyContent:'center',alignSelf:'flex-start'},
  back: {fontWeight: '800',color:colors.text},
  title: {fontSize: 32, fontWeight: '900', marginTop: 6,color:colors.text},
  subtitle: {fontSize: 14, lineHeight: 20,color:colors.textMuted},
  filterTitle: {fontSize: 12, fontWeight: '800',color:colors.textMuted, marginTop: 4},
  chips: {flexDirection: 'row', flexWrap: 'wrap', gap: 7},
  chip: {minHeight:control.minTouch,borderWidth: 1,borderColor:colors.border, borderRadius: radius.md, paddingHorizontal: 11,justifyContent:'center'},
  chipActive: {backgroundColor: colors.primary,borderColor:colors.primary},
  chipText: {fontSize: 12, fontWeight: '800',color:colors.text},
  chipTextActive: {color: colors.inverse},
  card: {minHeight:control.minTouch,borderWidth: 1,borderColor:colors.border, borderRadius: radius.lg, padding: 15, gap: 6},
  row: {flexDirection: 'row', justifyContent: 'space-between', gap: 10},
  name: {fontSize: 18, fontWeight: '900', flex: 1},
  status: {fontSize: 12, fontWeight: '800',color:colors.textMuted},
  meta: {fontSize: 13,color:colors.textMuted},
  metrics: {flexDirection: 'row', justifyContent: 'space-between', marginTop: 4},
  volume: {fontSize: 13, fontWeight: '800'},
  pr: {fontSize: 13, fontWeight: '800'},
  loading:{minHeight:control.minTouch,flexDirection:'row',alignItems:'center',gap:spacing.sm},
  muted:{fontSize:13,lineHeight:19,color:colors.textMuted},emptyCard:{backgroundColor:colors.surfaceMuted,borderRadius:radius.md,padding:spacing.md,gap:spacing.sm},emptyTitle:{fontSize:16,fontWeight:'900',color:colors.text},
  errorBox:{borderWidth:1,borderColor:colors.danger,borderRadius:radius.md,padding:spacing.md,gap:spacing.sm},error: {color: colors.danger,fontSize:12,fontWeight:'700'},pressed:{opacity:.68},
});
