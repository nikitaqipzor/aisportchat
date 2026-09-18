import React, {useCallback, useEffect, useMemo, useState} from 'react';
import {ActivityIndicator, Alert, Image, Pressable, ScrollView, StyleSheet, Text, View} from 'react-native';
import {api, BodyScanComparison, BodyScanDetails} from '../api/client';
import {AppButton} from '../components/AppButton';
import {bodyPhotoCapture} from '../native/bodyPhoto';
import {colors, control, radius, spacing} from '../theme/tokens';

type ViewID = 'front' | 'side' | 'back';
const views: Array<{id: ViewID; title: string; hint: string}> = [
  {id: 'front', title: 'Спереди', hint: 'Стойте прямо, руки слегка от корпуса, всё тело в кадре.'},
  {id: 'side', title: 'Сбоку', hint: 'Повернитесь на 90°, сохраняйте нейтральную стойку.'},
  {id: 'back', title: 'Сзади', hint: 'Спина к камере, одинаковое расстояние и положение ног.'},
];

export function BodyScanScreen({accessToken, onBack}: {accessToken: string; onBack: () => void}) {
  const [items, setItems] = useState<BodyScanDetails[]>([]);
  const [current, setCurrent] = useState<BodyScanDetails | null>(null);
  const [comparison, setComparison] = useState<BodyScanComparison | null>(null);
  const [previews, setPreviews] = useState<Partial<Record<ViewID, string>>>({});
  const [busy, setBusy] = useState<string | null>(null);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    try {
      setError('');
      const list = await api.bodyScans(accessToken, 20);
      setItems(list.items);
      const draft = list.items.find(item => item.scan.status === 'draft');
      setCurrent(draft ?? null);
      try { setComparison(await api.latestBodyScanComparison(accessToken)); } catch { setComparison(null); }
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось загрузить Body Scan');
    } finally { setLoading(false); }
  }, [accessToken]);

  useEffect(() => { void load(); }, [load]);

  async function ensureScan(manageBusy = true) {
    if (current) return current;
    if (manageBusy) setBusy('start');
    try {
      const next = await api.createBodyScan(accessToken);
      setCurrent(next);
      return next;
    } finally { if (manageBusy) setBusy(null); }
  }

  async function capture(view: ViewID) {
    try {
      setError(''); setNotice('');
      setBusy(view);
      const scan = current ?? await ensureScan(false);
      const image = await bodyPhotoCapture.captureImage();
      const next = await api.putBodyScanPhoto(accessToken, scan.scan.id, view, image);
      setPreviews(prev => ({...prev, [view]: image}));
      setCurrent(next);
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Не удалось сохранить фотографию';
      if (/BODY_PHOTO_CANCELLED|Съёмка отменена/i.test(message)) {
        setNotice('Съёмка отменена. Сохранённые фотографии не изменились.');
        return;
      }
      setError(message);
      Alert.alert('Body Scan', message);
    } finally { setBusy(null); }
  }

  async function startScan() {
    try { setError(''); await ensureScan(); }
    catch (e) { setError(e instanceof Error ? e.message : 'Не удалось начать Body Scan'); }
  }

  async function discard() {
    if (!current) return;
    Alert.alert('Удалить Body Scan?', 'Черновик и уже снятые фотографии будут удалены безвозвратно.', [
      {text: 'Отмена', style: 'cancel'},
      {text: 'Удалить', style: 'destructive', onPress: () => void (async () => {
        try { setBusy('complete'); await api.deleteBodyScan(accessToken, current.scan.id); setCurrent(null); setPreviews({}); await load(); }
        catch (e) { setError(e instanceof Error ? e.message : 'Не удалось удалить Body Scan'); }
        finally { setBusy(null); }
      })()},
    ]);
  }

  async function complete() {
    if (!current) return;
    try {
      setBusy('complete');
      await api.completeBodyScan(accessToken, current.scan.id);
      setCurrent(null);
      setPreviews({});
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось завершить Body Scan');
    } finally { setBusy(null); }
  }

  function removeCompleted(item: BodyScanDetails) {
    const date = new Date(item.scan.completed_at ?? item.scan.created_at).toLocaleDateString('ru-RU');
    Alert.alert('Удалить контрольную точку?', `Body Scan от ${date} и все три приватных фото будут удалены безвозвратно.`, [
      {text: 'Отмена', style: 'cancel'},
      {text: 'Удалить', style: 'destructive', onPress: () => void (async () => {
        try {
          setError(''); setBusy(`delete-${item.scan.id}`);
          await api.deleteBodyScan(accessToken, item.scan.id);
          await load();
        } catch (e) { setError(e instanceof Error ? e.message : 'Не удалось удалить Body Scan'); }
        finally { setBusy(null); }
      })()},
    ]);
  }

  const photoByView = useMemo(() => Object.fromEntries((current?.photos ?? []).map(p => [p.view, p])), [current]);
  const ready = views.every(v => Boolean(photoByView[v.id]));
  const completed = items.filter(x => x.scan.status === 'completed');
  const gradeLabel = comparison?.capture_grade === 'excellent' ? 'Отлично сопоставимы' : comparison?.capture_grade === 'good' ? 'Хорошо сопоставимы' : comparison?.capture_grade === 'retake_recommended' ? 'Лучше переснять' : 'Базовая оценка сопоставимости';
  const viewTitle = (view: ViewID) => views.find(item => item.id === view)?.title ?? view;
  const issueLabel = (issue: string) => ({
    'unusual framing; keep full body centered': 'Необычное соотношение кадра — разместите всё тело по центру',
    'low contrast': 'Низкий контраст — выберите более однородный фон',
  }[issue] ?? issue);

  return (
    <ScrollView contentContainerStyle={styles.container}>
      <Pressable accessibilityRole="button" accessibilityLabel="Вернуться к прогрессу" onPress={onBack} hitSlop={8} style={styles.navButton}><Text style={styles.back}>← Прогресс</Text></Pressable>
      <Text style={styles.kicker}>AI FITNESS · PHASE 3</Text>
      <Text accessibilityRole="header" style={styles.title}>Body Scan</Text>
      <Text style={styles.subtitle}>Три одинаково снятых ракурса создают контрольную точку формы. Фото приватные и доступны только вашему аккаунту.</Text>
      <View style={styles.privacyCard}><Text style={styles.privacyTitle}>Важно</Text><Text style={styles.muted}>Это визуальное сравнение, а не измерение состава тела или медицинская оценка. Снимайте только себя и не добавляйте в кадр личные документы.</Text></View>

      {loading ? <View style={styles.loading} accessibilityLiveRegion="polite"><ActivityIndicator color={colors.text}/><Text style={styles.muted}>Загружаем Body Scan…</Text></View> : null}
      {!loading && error && items.length === 0 ? <View style={styles.errorBox}><Text accessibilityRole="alert" style={styles.error}>{error}</Text><AppButton label="Повторить" variant="secondary" onPress={()=>{setLoading(true);void load()}}/></View> : null}

      {!loading && !(error && items.length === 0) && !current ? (
        <View style={styles.startCard}>
          <Text style={styles.cardTitle}>Новая контрольная точка</Text>
          <Text style={styles.muted}>Лучше снимать в одинаковом месте, освещении и одежде. Камера должна стоять примерно на уровне талии.</Text>
          <AppButton label="Начать Body Scan" testID="body-scan-start" loading={busy === 'start'} disabled={!bodyPhotoCapture.available} onPress={() => void startScan()} />
          {!bodyPhotoCapture.available ? <Text style={styles.warning}>Камера Body Scan недоступна в этой сборке. Обновите приложение или используйте поддерживаемое Android-устройство.</Text> : null}
        </View>
      ) : !loading && current ? (
        <>
          <View style={styles.progressRow}>
            <Text style={styles.cardTitle}>Текущий скан</Text>
            <Text style={styles.progress}>{current?.photos.length ?? 0}/3</Text>
          </View>
          {views.map(item => {
            const saved = photoByView[item.id];
            const preview = previews[item.id];
            return (
              <View key={item.id} style={styles.captureCard}>
                <View style={styles.captureHeader}>
                  <View style={{flex: 1}}><Text style={styles.captureTitle}>{item.title}</Text><Text style={styles.hint}>{item.hint}</Text></View>
                  <Text style={[styles.status, saved && styles.statusDone]}>{saved ? '✓' : '○'}</Text>
                </View>
                {preview ? <Image source={{uri: preview}} style={styles.preview} resizeMode="cover" /> : null}
                {saved ? (
                  <View style={styles.quality}>
                    <Text style={styles.qualityText}>{saved.width}×{saved.height} · свет {saved.brightness.toFixed(0)} · контраст {saved.contrast.toFixed(0)}</Text>
                    {saved.quality_issues?.map(issue => <Text key={issue} style={styles.warning}>• {issueLabel(issue)}</Text>)}
                  </View>
                ) : null}
                <AppButton label={saved ? 'Переснять' : `Снять ${item.title.toLowerCase()}`} testID={`body-scan-capture-${item.id}`} loading={busy === item.id} disabled={!bodyPhotoCapture.available || (busy !== null && busy !== item.id)} onPress={() => void capture(item.id)} variant={saved ? 'secondary' : 'primary'} />
              </View>
            );
          })}
          <AppButton label="Удалить черновик" testID="body-scan-delete-draft" variant="danger" disabled={busy !== null} onPress={() => void discard()} />
          <AppButton label="Завершить Body Scan" testID="body-scan-complete" disabled={!ready} loading={busy === 'complete'} onPress={() => void complete()} />
        </>
      ) : null}

      {comparison ? (
        <View style={styles.compareCard}>
          <Text style={styles.section}>Сравнимость последних сканов</Text>
          <Text style={styles.score}>{comparison.capture_consistency_score.toFixed(0)}<Text style={styles.scoreUnit}> / 100</Text></Text>
          <Text style={styles.grade}>{gradeLabel}</Text>
          <Text style={styles.muted}>{comparison.days_between} дней между контрольными точками</Text>
          <View style={styles.metricGrid}>
            <Text style={styles.captureMetric}>Свет Δ {comparison.lighting_delta.toFixed(0)}</Text>
            {comparison.contrast_delta !== undefined ? <Text style={styles.captureMetric}>Контраст Δ {comparison.contrast_delta.toFixed(0)}</Text> : null}
            {comparison.resolution_delta_percent !== undefined ? <Text style={styles.captureMetric}>Разрешение Δ {comparison.resolution_delta_percent.toFixed(0)}%</Text> : null}
          </View>
          {(comparison.view_metrics ?? []).map(metric => (
            <View key={metric.view} style={styles.viewMetricRow}>
              <Text style={styles.viewMetricTitle}>{viewTitle(metric.view)}</Text>
              <Text style={styles.viewMetricScore}>{metric.consistency_score.toFixed(0)}/100</Text>
            </View>
          ))}
          {comparison.weight_delta_kg !== undefined ? <Text style={styles.metric}>Вес: {comparison.weight_delta_kg > 0 ? '+' : ''}{comparison.weight_delta_kg.toFixed(1)} кг</Text> : null}
          {comparison.waist_delta_cm !== undefined ? <Text style={styles.metric}>Талия: {comparison.waist_delta_cm > 0 ? '+' : ''}{comparison.waist_delta_cm.toFixed(1)} см</Text> : null}
          {comparison.warnings?.map(w => <Text key={w} style={styles.warning}>• {w}</Text>)}
          <Text style={styles.cvPending}>Оценка отражает только сопоставимость условий съёмки. Она не анализирует фигуру, состав тела, здоровье или медицинские изменения.</Text>
        </View>
      ) : null}

      <Text style={styles.section}>История</Text>
      {completed.length === 0 ? <Text style={styles.muted}>Пока нет завершённых сканов.</Text> : completed.map(item => (
        <View key={item.scan.id} style={styles.historyCard}>
          <View style={styles.historyText}><Text style={styles.cardTitle}>{new Date(item.scan.completed_at ?? item.scan.created_at).toLocaleDateString('ru-RU')}</Text><Text style={styles.muted}>{item.photos.length} ракурса · приватное хранение</Text></View>
          <Pressable accessibilityRole="button" accessibilityLabel={`Удалить Body Scan от ${new Date(item.scan.completed_at ?? item.scan.created_at).toLocaleDateString('ru-RU')}`} disabled={busy !== null} onPress={() => removeCompleted(item)} hitSlop={8}><Text style={styles.delete}>{busy === `delete-${item.scan.id}` ? '…' : 'Удалить'}</Text></Pressable>
        </View>
      ))}
      {error && items.length > 0 ? <Text accessibilityRole="alert" style={styles.error}>{error}</Text> : null}
      {notice ? <Text accessibilityLiveRegion="polite" style={styles.muted}>{notice}</Text> : null}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {padding: spacing.lg, gap: spacing.md, backgroundColor: colors.surface},
  navButton:{minHeight:control.minTouch,justifyContent:'center',alignSelf:'flex-start'},back: {fontWeight: '800', color: colors.text},
  kicker: {fontSize: 11, fontWeight: '900', letterSpacing: 1.4, color: colors.textMuted},
  title: {fontSize: 34, fontWeight: '900', color: colors.text},
  subtitle: {fontSize: 14, lineHeight: 20, color: colors.textMuted},
  privacyCard:{backgroundColor:colors.surfaceMuted,borderRadius:radius.md,padding:spacing.md,gap:spacing.xs},privacyTitle:{fontWeight:'900',color:colors.text},loading:{minHeight:120,alignItems:'center',justifyContent:'center',gap:spacing.sm},errorBox:{borderWidth:1,borderColor:colors.danger,borderRadius:radius.md,padding:spacing.md,gap:spacing.sm},
  startCard: {borderWidth: 1, borderColor: colors.border, borderRadius: 22, padding: spacing.lg, gap: spacing.md},
  cardTitle: {fontSize: 17, fontWeight: '900', color: colors.text},
  muted: {fontSize: 12, lineHeight: 18, color: colors.textMuted},
  progressRow: {flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between'},
  progress: {fontSize: 14, fontWeight: '900', color: colors.textMuted},
  captureCard: {borderWidth: 1, borderColor: colors.border, borderRadius: 20, padding: spacing.md, gap: spacing.sm},
  captureHeader: {flexDirection: 'row', gap: spacing.sm, alignItems: 'flex-start'},
  captureTitle: {fontSize: 17, fontWeight: '900', color: colors.text},
  hint: {fontSize: 11, lineHeight: 16, color: colors.textMuted, marginTop: 3},
  status: {fontSize: 22, fontWeight: '900', color: colors.textMuted},
  statusDone: {color: colors.text},
  preview: {height: 260, width: '100%', borderRadius: 16, backgroundColor: colors.surfaceMuted},
  quality: {backgroundColor: colors.surfaceMuted, borderRadius: 12, padding: spacing.sm, gap: 3},
  qualityText: {fontSize: 11, fontWeight: '800', color: colors.text},
  warning: {fontSize: 11, lineHeight: 16, color: colors.textMuted},
  compareCard: {borderWidth: 1, borderColor: colors.border, borderRadius: 22, padding: spacing.lg, gap: spacing.sm},
  section: {fontSize: 19, fontWeight: '900', color: colors.text, marginTop: spacing.sm},
  score: {fontSize: 38, fontWeight: '900', color: colors.text},
  scoreUnit: {fontSize: 15, fontWeight: '800', color: colors.textMuted},
  grade: {fontSize: 15, fontWeight: '900', color: colors.text},
  metricGrid: {flexDirection: 'row', flexWrap: 'wrap', gap: spacing.xs},
  captureMetric: {fontSize: 11, fontWeight: '800', color: colors.text, backgroundColor: colors.surfaceMuted, borderRadius: 10, paddingHorizontal: spacing.sm, paddingVertical: spacing.xs},
  viewMetricRow: {flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', borderTopWidth: 1, borderTopColor: colors.border, paddingTop: spacing.sm},
  viewMetricTitle: {fontSize: 13, fontWeight: '800', color: colors.text},
  viewMetricScore: {fontSize: 13, fontWeight: '900', color: colors.text},
  metric: {fontSize: 15, fontWeight: '800', color: colors.text},
  cvPending: {fontSize: 11, lineHeight: 17, color: colors.textMuted, marginTop: spacing.xs},
  historyCard: {borderWidth: 1, borderColor: colors.border, borderRadius: 16, padding: spacing.md, flexDirection: 'row', gap: spacing.sm, justifyContent: 'space-between', alignItems: 'center'},
  historyText: {flex: 1},
  delete: {fontSize: 11, fontWeight: '900', color: colors.danger, paddingVertical: spacing.sm},
  error: {fontSize: 12, lineHeight: 18, fontWeight: '700', color: colors.danger},
});
