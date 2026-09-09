import React, {useCallback, useEffect, useMemo, useState} from 'react';
import {Alert, Image, Pressable, ScrollView, StyleSheet, Text, View} from 'react-native';
import {api, BodyScanComparison, BodyScanDetails} from '../api/client';
import {AppButton} from '../components/AppButton';
import {bodyPhotoCapture} from '../native/bodyPhoto';
import {colors, spacing} from '../theme/tokens';

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
  const [busy, setBusy] = useState<ViewID | 'start' | 'complete' | null>(null);
  const [error, setError] = useState('');

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
    }
  }, [accessToken]);

  useEffect(() => { void load(); }, [load]);

  async function ensureScan() {
    if (current) return current;
    setBusy('start');
    try {
      const next = await api.createBodyScan(accessToken);
      setCurrent(next);
      return next;
    } finally { setBusy(null); }
  }

  async function capture(view: ViewID) {
    try {
      setError('');
      setBusy(view);
      const scan = current ?? await ensureScan();
      const image = await bodyPhotoCapture.captureImage();
      setPreviews(prev => ({...prev, [view]: image}));
      const next = await api.putBodyScanPhoto(accessToken, scan.scan.id, view, image);
      setCurrent(next);
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Не удалось сохранить фотографию';
      setError(message);
      Alert.alert('Body Scan', message);
    } finally { setBusy(null); }
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

  const photoByView = useMemo(() => Object.fromEntries((current?.photos ?? []).map(p => [p.view, p])), [current]);
  const ready = views.every(v => Boolean(photoByView[v.id]));
  const completed = items.filter(x => x.scan.status === 'completed');

  return (
    <ScrollView contentContainerStyle={styles.container}>
      <Pressable accessibilityRole="button" onPress={onBack} hitSlop={10}><Text style={styles.back}>← Прогресс</Text></Pressable>
      <Text style={styles.kicker}>AI FITNESS · PHASE 3</Text>
      <Text style={styles.title}>Body Scan</Text>
      <Text style={styles.subtitle}>Три одинаково снятых ракурса создают контрольную точку формы. Фото приватные и доступны только вашему аккаунту.</Text>

      {!current ? (
        <View style={styles.startCard}>
          <Text style={styles.cardTitle}>Новая контрольная точка</Text>
          <Text style={styles.muted}>Лучше снимать в одинаковом месте, освещении и одежде. Камера должна стоять примерно на уровне талии.</Text>
          <AppButton label="Начать Body Scan" testID="body-scan-start" loading={busy === 'start'} onPress={() => void ensureScan()} />
        </View>
      ) : (
        <>
          <View style={styles.progressRow}>
            <Text style={styles.cardTitle}>Текущий скан</Text>
            <Text style={styles.progress}>{current.photos.length}/3</Text>
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
                    {saved.quality_issues?.map(issue => <Text key={issue} style={styles.warning}>• {issue}</Text>)}
                  </View>
                ) : null}
                <AppButton label={saved ? 'Переснять' : `Снять ${item.title.toLowerCase()}`} testID={`body-scan-capture-${item.id}`} loading={busy === item.id} onPress={() => void capture(item.id)} variant={saved ? 'secondary' : 'primary'} />
              </View>
            );
          })}
          <AppButton label="Удалить черновик" testID="body-scan-delete-draft" variant="danger" disabled={busy !== null} onPress={() => void discard()} />
          <AppButton label="Завершить Body Scan" testID="body-scan-complete" disabled={!ready} loading={busy === 'complete'} onPress={() => void complete()} />
        </>
      )}

      {comparison ? (
        <View style={styles.compareCard}>
          <Text style={styles.section}>Сравнимость последних сканов</Text>
          <Text style={styles.score}>{comparison.capture_consistency_score.toFixed(0)}<Text style={styles.scoreUnit}> / 100</Text></Text>
          <Text style={styles.muted}>{comparison.days_between} дней между контрольными точками · разница освещения {comparison.lighting_delta.toFixed(0)}</Text>
          {comparison.weight_delta_kg !== undefined ? <Text style={styles.metric}>Вес: {comparison.weight_delta_kg > 0 ? '+' : ''}{comparison.weight_delta_kg.toFixed(1)} кг</Text> : null}
          {comparison.waist_delta_cm !== undefined ? <Text style={styles.metric}>Талия: {comparison.waist_delta_cm > 0 ? '+' : ''}{comparison.waist_delta_cm.toFixed(1)} см</Text> : null}
          {comparison.warnings?.map(w => <Text key={w} style={styles.warning}>• {w}</Text>)}
          <Text style={styles.cvPending}>Визуальный анализ пропорций будет подключён в следующем CV-слое; текущий score оценивает качество и сопоставимость съёмки.</Text>
        </View>
      ) : null}

      <Text style={styles.section}>История</Text>
      {completed.length === 0 ? <Text style={styles.muted}>Пока нет завершённых сканов.</Text> : completed.map(item => (
        <View key={item.scan.id} style={styles.historyCard}>
          <View><Text style={styles.cardTitle}>{new Date(item.scan.completed_at ?? item.scan.created_at).toLocaleDateString('ru-RU')}</Text><Text style={styles.muted}>{item.photos.length} ракурса · приватное хранение</Text></View>
          <Text style={styles.done}>Готов</Text>
        </View>
      ))}
      {error ? <Text style={styles.error}>{error}</Text> : null}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {padding: spacing.lg, gap: spacing.md, backgroundColor: colors.surface},
  back: {fontWeight: '850', color: colors.text},
  kicker: {fontSize: 11, fontWeight: '900', letterSpacing: 1.4, color: colors.textMuted},
  title: {fontSize: 34, fontWeight: '950', color: colors.text},
  subtitle: {fontSize: 14, lineHeight: 20, color: colors.textMuted},
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
  score: {fontSize: 38, fontWeight: '950', color: colors.text},
  scoreUnit: {fontSize: 15, fontWeight: '800', color: colors.textMuted},
  metric: {fontSize: 15, fontWeight: '850', color: colors.text},
  cvPending: {fontSize: 11, lineHeight: 17, color: colors.textMuted, marginTop: spacing.xs},
  historyCard: {borderWidth: 1, borderColor: colors.border, borderRadius: 16, padding: spacing.md, flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center'},
  done: {fontSize: 11, fontWeight: '900', color: colors.text},
  error: {fontSize: 12, lineHeight: 18, fontWeight: '700', color: colors.text},
});
