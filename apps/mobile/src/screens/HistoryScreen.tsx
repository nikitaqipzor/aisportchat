import React, {useEffect, useState} from 'react';
import {ActivityIndicator, Pressable, ScrollView, StyleSheet, Text, View} from 'react-native';
import {api, HistoryFilters, WorkoutProgressSummary, WorkoutView} from '../api/client';
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
  const [summary, setSummary] = useState<WorkoutProgressSummary | null>(null);
  const [environment, setEnvironment] = useState<EnvironmentFilter>('all');
  const [status, setStatus] = useState<StatusFilter>('completed');
  const [favoritesOnly, setFavoritesOnly] = useState(false);
  const [loading, setLoading] = useState(true);
  const [summaryLoading, setSummaryLoading] = useState(true);
  const [error, setError] = useState('');
  const [summaryError, setSummaryError] = useState(false);

  useEffect(() => {
    void loadHistory();
  }, [accessToken, environment, status, favoritesOnly]);

  useEffect(() => {
    void loadSummary();
  }, [accessToken]);

  async function loadHistory() {
    try {
      setLoading(true);
      setError('');
      const filters: HistoryFilters = {limit: 50};
      if (environment !== 'all') filters.environment = environment;
      if (status !== 'all') filters.status = status;
      if (favoritesOnly) filters.favorite = true;
      const history = await api.workoutHistory(accessToken, filters);
      setItems(history.items);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось загрузить историю');
    } finally {
      setLoading(false);
    }
  }

  async function loadSummary() {
    try {
      setSummaryLoading(true);
      setSummaryError(false);
      setSummary(await api.workoutProgressSummary(accessToken, 28));
    } catch {
      setSummary(null);
      setSummaryError(true);
    } finally {
      setSummaryLoading(false);
    }
  }

  return (
    <ScrollView contentContainerStyle={styles.container}>
      <Pressable accessibilityRole="button" accessibilityLabel="Вернуться на главную" hitSlop={8} onPress={onBack} style={styles.navButton}><Text style={styles.back}>← Главная</Text></Pressable>
      <Text accessibilityRole="header" style={styles.title}>История</Text>
      <Text style={styles.subtitle}>Фильтруй тренировки и открывай подробности каждого занятия.</Text>

      {summary && summary.completed_workouts > 0 ? <View style={styles.summaryCard} testID="history-progress-summary">
        <View style={styles.summaryHeader}><View><Text style={styles.summaryEyebrow}>ПОСЛЕДНИЕ 28 ДНЕЙ</Text><Text style={styles.summaryTitle}>Твой тренировочный ритм</Text></View><Text style={styles.streak}>{summary.weekly_streak} нед. подряд</Text></View>
        <View style={styles.summaryGrid}><SummaryMetric value={String(summary.completed_workouts)} label="тренировок"/><SummaryMetric value={summary.workouts_per_week.toFixed(1)} label="в неделю"/><SummaryMetric value={String(summary.total_sets)} label="подходов"/><SummaryMetric value={`${Math.round(summary.total_volume)} кг`} label="объёма"/></View>
        <VolumeComparison recent={summary.recent_volume_7d} previous={summary.previous_volume_7d}/>
        {summary.personal_record_count > 0 ? <Text style={styles.recordLine}>🏆 Новых рекордов за период: {summary.personal_record_count}</Text> : <Text style={styles.summaryHint}>Новый PR появится после улучшения веса, повторений или расчётного 1ПМ.</Text>}
        {summary.exercise_bests.length > 0 ? <View style={styles.bests}><Text style={styles.bestTitle}>Лучшие рабочие результаты</Text>{summary.exercise_bests.map(item=><View key={item.exercise_id} style={styles.bestRow}><Text numberOfLines={1} style={styles.bestName}>{item.exercise_name}</Text><Text style={styles.bestValue}>{item.max_weight !== undefined?`${item.max_weight} кг · `:''}{item.max_reps} повт.</Text></View>)}</View>:null}
      </View> : !summaryLoading && !summaryError ? <View style={styles.progressEmpty}><Text style={styles.emptyTitle}>Прогресс начнётся с первой тренировки</Text><Text style={styles.muted}>Заверши занятие и запиши подходы — приложение посчитает метрики по фактическим данным.</Text></View> : null}
      {summaryError ? <View style={styles.progressEmpty}><Text style={styles.emptyTitle}>Аналитика временно недоступна</Text><Text style={styles.muted}>История тренировок продолжает работать. Метрики обновятся при следующем открытии экрана.</Text><AppButton label="Обновить аналитику" variant="secondary" onPress={() => void loadSummary()}/></View> : null}

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
      {error ? <View style={styles.errorBox}><Text accessibilityRole="alert" style={styles.error}>{error}</Text><AppButton label="Повторить" variant="secondary" onPress={()=>void loadHistory()}/></View> : null}
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

function SummaryMetric({value,label}:{value:string;label:string}) { return <View style={styles.summaryMetric}><Text style={styles.summaryValue}>{value}</Text><Text style={styles.summaryLabel}>{label}</Text></View>; }
function VolumeComparison({recent,previous}:{recent:number;previous:number}) { const delta=previous>0?Math.round(((recent-previous)/previous)*100):undefined; return <View style={styles.volumeCompare}><Text style={styles.bestTitle}>Объём за 7 дней</Text><Text style={styles.volumeValue}>{Math.round(recent)} кг</Text><Text style={styles.summaryHint}>{delta===undefined?'Сравнение появится после предыдущей недели':`${delta>0?'+':''}${delta}% к предыдущим 7 дням`}</Text></View>; }

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
  summaryCard:{backgroundColor:colors.primary,borderRadius:radius.lg,padding:spacing.md,gap:spacing.md},summaryHeader:{flexDirection:'row',justifyContent:'space-between',alignItems:'flex-start',gap:spacing.sm},summaryEyebrow:{fontSize:10,fontWeight:'900',letterSpacing:1.2,color:'#aaa'},summaryTitle:{fontSize:19,fontWeight:'900',color:colors.inverse,marginTop:3},streak:{fontSize:11,fontWeight:'900',color:colors.text,backgroundColor:colors.inverse,paddingHorizontal:9,paddingVertical:6,borderRadius:12},summaryGrid:{flexDirection:'row',flexWrap:'wrap',gap:8},summaryMetric:{width:'48%',backgroundColor:'#222',borderRadius:radius.md,padding:11},summaryValue:{fontSize:21,fontWeight:'900',color:colors.inverse},summaryLabel:{fontSize:11,color:'#aaa',marginTop:2},volumeCompare:{borderTopWidth:StyleSheet.hairlineWidth,borderTopColor:'#555',paddingTop:spacing.sm},volumeValue:{fontSize:23,fontWeight:'900',color:colors.inverse,marginTop:4},summaryHint:{fontSize:12,lineHeight:17,color:'#aaa'},recordLine:{color:colors.inverse,fontWeight:'800'},bests:{gap:7},bestTitle:{color:'#bbb',fontSize:12,fontWeight:'800'},bestRow:{flexDirection:'row',justifyContent:'space-between',gap:10},bestName:{flex:1,color:colors.inverse,fontSize:13},bestValue:{color:colors.inverse,fontSize:13,fontWeight:'800'},progressEmpty:{backgroundColor:colors.surfaceMuted,borderRadius:radius.md,padding:spacing.md,gap:spacing.xs},
});
