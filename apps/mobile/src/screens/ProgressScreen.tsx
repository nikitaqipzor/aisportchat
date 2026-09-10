import React, {useCallback, useEffect, useMemo, useState} from 'react';
import {ActivityIndicator, Pressable, RefreshControl, ScrollView, StyleSheet, Text, TextInput, View} from 'react-native';
import {api, BodyMeasurement, HealthInsights, NutritionCorrelation, ProgressSummary, WorkoutProgressSummary} from '../api/client';
import {AppButton} from '../components/AppButton';
import {currentLocalDate} from '../domain/date';
import {colors, control, radius, spacing} from '../theme/tokens';

type AnalyticsTab = 'overview' | 'body' | 'nutrition';
const parseOptional = (value: string) => value.trim() ? Number(value.replace(',', '.')) : undefined;

function Delta({value, unit}: {value?: number; unit: string}) {
  if (value === undefined) return <Text style={styles.small}>Недостаточно данных</Text>;
  return <Text accessibilityLabel={`Изменение ${value > 0 ? 'плюс ' : ''}${value.toFixed(1)} ${unit}`} style={styles.delta}>{value > 0 ? '+' : ''}{value.toFixed(1)} {unit}</Text>;
}

function MetricCard({label, value, hint}: {label: string; value: string; hint?: string}) {
  return <View accessible accessibilityLabel={`${label}: ${value}${hint ? `. ${hint}` : ''}`} style={styles.metricCard}><Text accessible={false} style={styles.label}>{label}</Text><Text accessible={false} style={styles.metricValue}>{value}</Text>{hint ? <Text accessible={false} style={styles.small}>{hint}</Text> : null}</View>;
}

function AnalyticsTabs({value, onChange}: {value: AnalyticsTab; onChange: (tab: AnalyticsTab) => void}) {
  const items: Array<[AnalyticsTab, string]> = [['overview', 'Обзор'], ['body', 'Тело'], ['nutrition', 'Питание']];
  return <View accessibilityRole="tablist" style={styles.tabs}>{items.map(([id, title]) => <Pressable key={id} accessibilityRole="tab" accessibilityState={{selected: value === id}} testID={`analytics-tab-${id}`} onPress={() => onChange(id)} style={[styles.tab, value === id && styles.tabActive]}><Text style={[styles.tabText, value === id && styles.tabTextActive]}>{title}</Text></Pressable>)}</View>;
}

export function ProgressScreen({accessToken, onBack, onBodyScan, onTechnique, onDevices}: {accessToken: string; onBack: () => void; onBodyScan: () => void; onTechnique: () => void; onDevices: () => void}) {
  const [tab, setTab] = useState<AnalyticsTab>('overview');
  const [summary, setSummary] = useState<ProgressSummary | null>(null);
  const [history, setHistory] = useState<BodyMeasurement[]>([]);
  const [correlation, setCorrelation] = useState<NutritionCorrelation | null>(null);
  const [workouts, setWorkouts] = useState<WorkoutProgressSummary | null>(null);
  const [health, setHealth] = useState<HealthInsights | null>(null);
  const [weight, setWeight] = useState('');
  const [waist, setWaist] = useState('');
  const [arm, setArm] = useState('');
  const [refreshing, setRefreshing] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [formError, setFormError] = useState('');

  const load = useCallback(async () => {
    setError('');
    const results = await Promise.allSettled([
      api.progressSummary(accessToken, 30),
      api.measurementHistory(accessToken, 30),
      api.nutritionCorrelation(accessToken, 30),
      api.workoutProgressSummary(accessToken, 28),
      api.healthInsights(accessToken, currentLocalDate()),
    ]);
    if (results[0].status === 'fulfilled') setSummary(results[0].value);
    if (results[1].status === 'fulfilled') setHistory(results[1].value.items);
    if (results[2].status === 'fulfilled') setCorrelation(results[2].value);
    if (results[3].status === 'fulfilled') setWorkouts(results[3].value);
    if (results[4].status === 'fulfilled') setHealth(results[4].value);
    const failed = results.filter(result => result.status === 'rejected').length;
    if (failed === results.length) setError('Не удалось загрузить аналитику. Проверьте соединение и повторите попытку.');
    else if (failed > 0) setError(`Часть данных временно недоступна (${failed} из ${results.length}). Доступные показатели показаны ниже.`);
    setLoading(false);
  }, [accessToken]);

  useEffect(() => { void load(); }, [load]);
  async function refresh() { setRefreshing(true); await load(); setRefreshing(false); }

  const values = {weight: parseOptional(weight), waist: parseOptional(waist), arm: parseOptional(arm)};
  const hasValue = Object.values(values).some(value => value !== undefined);
  function validate() {
    if (!hasValue) return 'Введите хотя бы один замер.';
    if (Object.values(values).some(value => value !== undefined && !Number.isFinite(value))) return 'Используйте только числа.';
    if (values.weight !== undefined && (values.weight < 20 || values.weight > 400)) return 'Вес должен быть от 20 до 400 кг.';
    if (values.waist !== undefined && (values.waist < 30 || values.waist > 300)) return 'Талия должна быть от 30 до 300 см.';
    if (values.arm !== undefined && (values.arm < 10 || values.arm > 100)) return 'Обхват руки должен быть от 10 до 100 см.';
    return '';
  }

  async function save() {
    const invalid = validate();
    setFormError(invalid);
    if (invalid) return;
    try {
      setSaving(true);
      setError('');
      await api.logMeasurement(accessToken, {weight_kg: values.weight, waist_cm: values.waist, arm_cm: values.arm});
      setWeight(''); setWaist(''); setArm('');
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось сохранить замер');
    } finally { setSaving(false); }
  }

  const weights = useMemo(() => history.filter(item => item.weight_kg !== undefined), [history]);
  const min = weights.length ? Math.min(...weights.map(item => item.weight_kg ?? 0)) : 0;
  const max = weights.length ? Math.max(...weights.map(item => item.weight_kg ?? 0)) : 1;
  const range = Math.max(1, max - min);
  const visibleWeights = weights.slice(-14);
  const snapshot = health?.snapshot;

  return <ScrollView testID="analytics-screen" contentContainerStyle={styles.container} refreshControl={<RefreshControl refreshing={refreshing} onRefresh={refresh}/>} keyboardShouldPersistTaps="handled">
    <Pressable accessibilityRole="button" accessibilityLabel="Вернуться на главную" hitSlop={8} onPress={onBack} style={styles.navButton}><Text style={styles.back}>← Главная</Text></Pressable>
    <Text style={styles.kicker}>ATHLETICA ANALYTICS</Text>
    <Text accessibilityRole="header" style={styles.title}>Аналитика</Text>
    <Text style={styles.subtitle}>Тренировки, тело, питание и данные часов — только по фактическим записям.</Text>
    <AnalyticsTabs value={tab} onChange={setTab}/>

    {loading ? <View style={styles.loading} accessibilityLiveRegion="polite"><ActivityIndicator color={colors.text}/><Text style={styles.muted}>Собираем показатели…</Text></View> : null}
    {error ? <View style={styles.errorBox}><Text accessibilityRole="alert" style={styles.error}>{error}</Text><AppButton label="Обновить данные" variant="secondary" onPress={() => void load()}/></View> : null}

    {!loading && tab === 'overview' ? <>
      <View style={styles.heroDark}><Text style={styles.heroEyebrow}>ПОСЛЕДНИЕ 28 ДНЕЙ</Text><Text style={styles.heroDarkValue}>{workouts?.completed_workouts ?? 0}</Text><Text style={styles.heroDarkLabel}>завершённых тренировок</Text><View style={styles.heroRow}><Text style={styles.heroPill}>{workouts?.workouts_per_week.toFixed(1) ?? '0.0'} в неделю</Text><Text style={styles.heroPill}>{workouts?.weekly_streak ?? 0} нед. подряд</Text></View></View>
      <View style={styles.metricGrid}><MetricCard label="Объём" value={`${Math.round(workouts?.total_volume ?? 0)} кг`} hint="за 28 дней"/><MetricCard label="Подходы" value={String(workouts?.total_sets ?? 0)} hint="завершённые"/><MetricCard label="Вес" value={summary?.weight_current_kg !== undefined ? `${summary.weight_current_kg.toFixed(1)} кг` : '—'} hint="последний замер"/><MetricCard label="Рекорды" value={String(workouts?.personal_record_count ?? 0)} hint="за период"/></View>
      <Text style={styles.section}>Часы и восстановление</Text>
      {snapshot ? <View style={styles.card}><View style={styles.metricGrid}><MetricCard label="Шаги" value={String(Math.round(snapshot.steps))}/><MetricCard label="Сон" value={`${Math.floor(snapshot.sleep_minutes / 60)} ч ${Math.round(snapshot.sleep_minutes % 60)} мин`}/><MetricCard label="Активность" value={`${Math.round(snapshot.active_calories_kcal)} kcal`}/><MetricCard label="Надёжность" value={`${health?.confidence_percent ?? 0}%`}/></View><Text style={styles.disclaimer}>Источник: {snapshot.source_label}. Данные используются как fitness-тренд, а не медицинская диагностика.</Text></View> : <View style={styles.empty}><Text style={styles.emptyTitle}>Нет данных часов</Text><Text style={styles.muted}>Подключите Mi Fitness через Health Connect, чтобы видеть сон, шаги и активность рядом с тренировками.</Text><AppButton label="Подключить часы" testID="analytics-connect-device" onPress={onDevices}/></View>}
      {workouts?.exercise_bests.length ? <><Text style={styles.section}>Лучшие результаты</Text><View style={styles.card}>{workouts.exercise_bests.slice(0, 5).map(item => <View key={item.exercise_id} style={styles.bestRow}><Text numberOfLines={1} style={styles.bestName}>{item.exercise_name}</Text><Text style={styles.bestValue}>{item.max_weight !== undefined ? `${item.max_weight} кг · ` : ''}{item.max_reps} повт.</Text></View>)}</View></> : null}
    </> : null}

    {!loading && tab === 'body' ? <>
      <View style={styles.hero}><View><Text style={styles.label}>Текущий вес</Text><Text style={styles.heroValue}>{summary?.weight_current_kg?.toFixed(1) ?? '—'} <Text style={styles.unit}>кг</Text></Text></View><Delta value={summary?.weight_delta_kg} unit="кг за 30 дней"/></View>
      <View style={styles.metricGrid}><MetricCard label="Талия" value={summary?.waist_current_cm !== undefined ? `${summary.waist_current_cm.toFixed(1)} см` : '—'} hint={summary?.waist_delta_cm !== undefined ? `${summary.waist_delta_cm > 0 ? '+' : ''}${summary.waist_delta_cm.toFixed(1)} см` : undefined}/><MetricCard label="Замеры" value={String(summary?.points ?? 0)} hint="за 30 дней"/></View>
      <Text style={styles.section}>Вес · 30 дней</Text>
      {visibleWeights.length > 1 ? <View style={styles.chart} accessibilityLabel={`График веса: от ${visibleWeights[0].weight_kg?.toFixed(1)} до ${visibleWeights[visibleWeights.length - 1].weight_kg?.toFixed(1)} килограмма`}>{visibleWeights.map((item, index) => { const height = 18 + (((item.weight_kg ?? min) - min) / range) * 70; return <View key={item.id} style={styles.barWrap}><View style={[styles.bar, {height}]}/><Text style={styles.barLabel}>{index === 0 || index === visibleWeights.length - 1 ? (item.weight_kg ?? 0).toFixed(1) : ''}</Text></View>; })}</View> : <View style={styles.empty}><Text style={styles.emptyTitle}>Нужны два замера</Text><Text style={styles.muted}>Добавьте ещё один замер веса, чтобы увидеть тренд.</Text></View>}
      <Text style={styles.section}>Новый замер</Text>
      <View style={styles.form}><TextInput accessibilityLabel="Вес в килограммах" testID="analytics-weight" value={weight} onChangeText={value => { setWeight(value); setFormError(''); }} keyboardType="decimal-pad" maxLength={7} placeholder="Вес, кг" placeholderTextColor={colors.textMuted} style={styles.input}/><TextInput accessibilityLabel="Обхват талии в сантиметрах" testID="analytics-waist" value={waist} onChangeText={value => { setWaist(value); setFormError(''); }} keyboardType="decimal-pad" maxLength={7} placeholder="Талия, см" placeholderTextColor={colors.textMuted} style={styles.input}/><TextInput accessibilityLabel="Обхват руки в сантиметрах" testID="analytics-arm" value={arm} onChangeText={value => { setArm(value); setFormError(''); }} keyboardType="decimal-pad" maxLength={7} placeholder="Бицепс, см" placeholderTextColor={colors.textMuted} style={styles.input}/>{formError ? <Text accessibilityRole="alert" style={styles.error}>{formError}</Text> : null}<AppButton label="Сохранить замер" loading={saving} disabled={!hasValue} onPress={() => void save()}/></View>
      <Text style={styles.section}>Инструменты</Text><View style={styles.card}><Text style={styles.cardTitle}>Фото-прогресс формы</Text><Text style={styles.muted}>Приватные ракурсы в одинаковых условиях помогают сравнивать контрольные точки.</Text><AppButton label="Открыть Body Scan" testID="progress-body-scan" onPress={onBodyScan}/></View><View style={styles.card}><Text style={styles.cardTitle}>Pose-анализ подхода</Text><Text style={styles.muted}>Ориентировочная оценка повторений, амплитуды, темпа и симметрии.</Text><AppButton label="Анализировать технику" testID="progress-technique" onPress={onTechnique}/></View>
    </> : null}

    {!loading && tab === 'nutrition' ? <>
      <Text style={styles.section}>Питание × тренировки</Text>
      {correlation && (correlation.logged_training_days + correlation.logged_rest_days > 0) ? <View style={styles.card}><Text style={styles.disclaimer}>Наблюдаемая связь за 30 дней — не доказательство причины.</Text><Text style={styles.cardTitle}>Тренировочные дни</Text><Text style={styles.corrValue}>{Math.round(correlation.avg_training_calories)} kcal · Б {Math.round(correlation.avg_training_protein_g)} г</Text><Text style={styles.small}>{correlation.logged_training_days} дней с записями · попадание в цель {Math.round(correlation.training_calorie_adherence_percent)}%</Text><View style={styles.divider}/><Text style={styles.cardTitle}>Дни отдыха</Text><Text style={styles.corrValue}>{Math.round(correlation.avg_rest_calories)} kcal · Б {Math.round(correlation.avg_rest_protein_g)} г</Text><Text style={styles.small}>{correlation.logged_rest_days} дней с записями · попадание в цель {Math.round(correlation.rest_calorie_adherence_percent)}%</Text></View> : <View style={styles.empty}><Text style={styles.emptyTitle}>Недостаточно записей</Text><Text style={styles.muted}>Ведите дневник питания и завершайте тренировки — сравнение появится автоматически.</Text></View>}
    </> : null}
  </ScrollView>;
}

const styles = StyleSheet.create({
  container:{padding:spacing.lg,paddingBottom:spacing.xl,gap:spacing.md,backgroundColor:colors.background},navButton:{minHeight:control.minTouch,justifyContent:'center',alignSelf:'flex-start'},back:{fontWeight:'800',color:colors.text},kicker:{fontSize:11,fontWeight:'900',letterSpacing:1.3,color:colors.textMuted},title:{fontSize:32,fontWeight:'900',color:colors.text},subtitle:{fontSize:13,lineHeight:19,color:colors.textMuted},tabs:{flexDirection:'row',borderRadius:radius.md,padding:4,backgroundColor:colors.surfaceMuted},tab:{flex:1,minHeight:44,alignItems:'center',justifyContent:'center',borderRadius:radius.sm},tabActive:{backgroundColor:colors.primary},tabText:{fontSize:12,fontWeight:'800',color:colors.textMuted},tabTextActive:{color:colors.inverse},loading:{minHeight:120,alignItems:'center',justifyContent:'center',gap:spacing.sm},muted:{fontSize:12,lineHeight:18,color:colors.textMuted},errorBox:{borderWidth:1,borderColor:colors.danger,borderRadius:radius.md,padding:spacing.md,gap:spacing.sm},error:{fontSize:12,lineHeight:18,fontWeight:'700',color:colors.danger},heroDark:{borderRadius:radius.lg,padding:spacing.lg,backgroundColor:colors.primary},heroEyebrow:{fontSize:10,fontWeight:'900',letterSpacing:1.2,color:'#AAA'},heroDarkValue:{fontSize:52,lineHeight:58,fontWeight:'900',color:colors.inverse,marginTop:8},heroDarkLabel:{fontSize:14,fontWeight:'700',color:'#C8C8C8'},heroRow:{flexDirection:'row',flexWrap:'wrap',gap:8,marginTop:spacing.md},heroPill:{fontSize:11,fontWeight:'900',color:colors.text,backgroundColor:colors.inverse,borderRadius:radius.pill,paddingHorizontal:10,paddingVertical:7},metricGrid:{flexDirection:'row',flexWrap:'wrap',gap:spacing.sm},metricCard:{width:'48%',flexGrow:1,minWidth:130,borderWidth:1,borderColor:colors.border,borderRadius:radius.md,padding:spacing.md,gap:4,backgroundColor:colors.surface},label:{fontSize:11,fontWeight:'800',color:colors.textMuted},metricValue:{fontSize:21,fontWeight:'900',color:colors.text},small:{fontSize:10,lineHeight:15,color:colors.textMuted},section:{fontSize:19,fontWeight:'900',color:colors.text,marginTop:3},card:{borderWidth:1,borderColor:colors.border,borderRadius:radius.lg,padding:spacing.md,gap:spacing.sm,backgroundColor:colors.surface},cardTitle:{fontSize:14,fontWeight:'900',color:colors.text},disclaimer:{fontSize:11,lineHeight:16,color:colors.textMuted},empty:{backgroundColor:colors.surfaceMuted,borderRadius:radius.md,padding:spacing.md,gap:spacing.sm},emptyTitle:{fontWeight:'900',color:colors.text},bestRow:{minHeight:40,flexDirection:'row',alignItems:'center',justifyContent:'space-between',gap:10,borderBottomWidth:StyleSheet.hairlineWidth,borderColor:colors.border},bestName:{flex:1,fontSize:13,color:colors.text},bestValue:{fontSize:13,fontWeight:'800',color:colors.text},hero:{borderWidth:1,borderColor:colors.border,borderRadius:radius.lg,padding:spacing.md,flexDirection:'row',justifyContent:'space-between',alignItems:'flex-end',gap:spacing.sm},heroValue:{fontSize:38,fontWeight:'900',marginTop:4,color:colors.text},unit:{fontSize:15,fontWeight:'800'},delta:{fontSize:12,fontWeight:'800',color:colors.text},chart:{height:120,borderWidth:1,borderColor:colors.border,borderRadius:radius.lg,padding:12,flexDirection:'row',alignItems:'flex-end',gap:5},barWrap:{flex:1,alignItems:'center',justifyContent:'flex-end'},bar:{width:'100%',minWidth:5,backgroundColor:colors.primary,borderRadius:5},barLabel:{fontSize:8,fontWeight:'700',height:12,marginTop:3,color:colors.text},form:{gap:spacing.sm},input:{minHeight:control.minTouch,borderWidth:1,borderColor:colors.border,borderRadius:radius.md,paddingHorizontal:12,fontWeight:'700',color:colors.text},corrValue:{fontSize:17,fontWeight:'900',color:colors.text},divider:{height:1,backgroundColor:colors.border,marginVertical:spacing.sm},
});
