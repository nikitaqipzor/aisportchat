import React, {useCallback, useEffect, useRef, useState} from 'react';
import {ActivityIndicator, Pressable, RefreshControl, ScrollView, StyleSheet, Text, View} from 'react-native';
import {api, ProgramAnalytics, TrainingProgram} from '../api/client';
import {AppButton} from '../components/AppButton';
import {colors, control, radius, spacing} from '../theme/tokens';
import {formatCalendarDate} from '../domain/date';

const statusLabels: Record<string, string> = {archived: 'В архиве', completed: 'Завершена', draft: 'Черновик'};
const environmentLabels: Record<string, string> = {gym: 'Зал', home: 'Дома', band: 'Резинки'};

export function ProgramsScreen({accessToken, onBack, onCreate, onOpen}: {
  accessToken: string; onBack: () => void; onCreate: () => void; onOpen: (programId: string) => void;
}) {
  const [active, setActive] = useState<TrainingProgram>();
  const [analytics, setAnalytics] = useState<ProgramAnalytics>();
  const [history, setHistory] = useState<TrainingProgram[]>([]);
  const [nextCursor, setNextCursor] = useState('');
  const [hasMore, setHasMore] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [moreError, setMoreError] = useState('');
  const loadGeneration = useRef(0);
  const loadingMoreRef = useRef(false);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState('');
  const [analyticsError, setAnalyticsError] = useState(false);

  const load = useCallback(async (refresh = false) => {
    const generation = ++loadGeneration.current;
    refresh ? setRefreshing(true) : setLoading(true);
    setMoreError('');
    setError('');
    try {
      const current = await api.activeProgram(accessToken);
      setActive(current);
      setAnalytics(undefined); setAnalyticsError(false);
      const [stats, old] = await Promise.all([
        current ? api.programAnalytics(accessToken, current.program.id).catch(() => {if (generation === loadGeneration.current) setAnalyticsError(true); return undefined;}) : Promise.resolve(undefined),
        api.programHistory(accessToken, 50),
      ]);
      if (generation !== loadGeneration.current) return;
      setAnalytics(stats);
      setHistory(old.items.filter(item => item.program.status !== 'active'));
      setNextCursor(old.next_cursor);
      setHasMore(old.has_more);
    } catch (e) {
      if (generation !== loadGeneration.current) return;
      setError(e instanceof Error ? e.message : 'Не удалось загрузить программы.');
    } finally {
      if (generation === loadGeneration.current) { setLoading(false); setRefreshing(false); }
    }
  }, [accessToken]);

  const loadMore = useCallback(async () => {
    if (!hasMore || loadingMoreRef.current || loading || refreshing) return;
    loadingMoreRef.current = true;
    const generation = loadGeneration.current;
    setLoadingMore(true); setMoreError('');
    try {
      const page = await api.programHistory(accessToken, 50, nextCursor);
      if (generation !== loadGeneration.current) return;
      setHistory(previous => {
        const ids = new Set(previous.map(item => item.program.id));
        return [...previous, ...page.items.filter(item => item.program.status !== 'active' && !ids.has(item.program.id))];
      });
      setNextCursor(page.next_cursor);
      setHasMore(page.has_more);
    } catch (e) {
      if (generation === loadGeneration.current) setMoreError(e instanceof Error ? e.message : 'Не удалось загрузить прошлые программы.');
    } finally {
      loadingMoreRef.current = false;
      if (generation === loadGeneration.current) setLoadingMore(false);
    }
  }, [accessToken, hasMore, nextCursor, loading, refreshing]);

  useEffect(() => { void load(); return () => { loadGeneration.current++; }; }, [load]);

  return (
    <ScrollView testID="programs-screen" contentContainerStyle={styles.container}
      refreshControl={<RefreshControl refreshing={refreshing} onRefresh={() => void load(true)} />}>
      <View style={styles.top}>
        <Pressable accessibilityRole="button" accessibilityLabel="Вернуться на главную" hitSlop={8} onPress={onBack} style={({pressed}) => [styles.topAction, pressed && styles.pressed]}>
          <Text style={styles.back}>← Главная</Text>
        </Pressable>
        <Pressable accessibilityRole="button" accessibilityLabel="Создать новую программу" hitSlop={8} onPress={onCreate} style={({pressed}) => [styles.topAction, pressed && styles.pressed]}>
          <Text style={styles.create}>+ Новая</Text>
        </Pressable>
      </View>
      <View>
        <Text style={styles.eyebrow}>ПРОГРАММЫ</Text>
        <Text accessibilityRole="header" style={styles.title}>Твой план тренировок</Text>
        <Text style={styles.subtitle}>Система следит за выполнением, разгрузочными неделями и переносами.</Text>
      </View>
      {loading ? (
        <View accessibilityRole="progressbar" accessibilityLabel="Загрузка программ" style={styles.stateCard}>
          <ActivityIndicator color={colors.primary} /><Text style={styles.muted}>Загружаем программу…</Text>
        </View>
      ) : error ? (
        <View accessibilityRole="alert" style={styles.errorCard}>
          <Text style={styles.errorTitle}>Не удалось загрузить программы</Text><Text style={styles.muted}>{error}</Text>
          <AppButton label="Повторить" variant="secondary" testID="programs-retry" onPress={() => void load()} />
        </View>
      ) : !active ? (
        <View style={styles.empty}>
          <Text style={styles.emptyTitle}>Активной программы пока нет</Text>
          <Text style={styles.muted}>Создай план на 4, 8 или 12 недель — с понятным расписанием и автоматической адаптацией нагрузки.</Text>
          <AppButton label="Создать программу" testID="programs-create-empty" onPress={onCreate} />
        </View>
      ) : (
        <Pressable accessibilityRole="button" accessibilityLabel={`Открыть активную программу ${active.program.title}`} testID="programs-active" onPress={() => onOpen(active.program.id)} style={({pressed}) => [styles.hero, pressed && styles.heroPressed]}>
          <View style={styles.heroTop}><View style={styles.flex}><Text style={styles.heroKicker}>АКТИВНАЯ ПРОГРАММА</Text><Text style={styles.heroTitle}>{active.program.title}</Text></View><Text accessibilityElementsHidden style={styles.arrow}>→</Text></View>
          <View style={styles.stats}><Stat label="Неделя" value={analytics?`${analytics.current_week}/${active.program.weeks}`:'—'} /><Stat label="Выполнение" value={analytics?`${analytics.adherence_percent}%`:'—'} /><Stat label="Готово" value={analytics?`${analytics.completed_sessions}`:'—'} /></View>
          {analyticsError ? <Text accessibilityRole="alert" style={styles.next}>Аналитика недоступна. Открой программу для повторной загрузки.</Text> : analytics?.next_session ? <Text style={styles.next}>Следующая: {formatCalendarDate(analytics.next_session.planned_date)} · {analytics.next_session.muscle}{analytics.next_session.is_deload ? ' · разгрузка' : ''}</Text> : <Text style={styles.next}>Все ближайшие тренировки выполнены.</Text>}
        </Pressable>
      )}
      {!loading && !error && (history.length > 0 || hasMore) ? <View style={styles.history}>
        <Text style={styles.section}>Прошлые программы</Text>
        {history.map(item => <Pressable key={item.program.id} accessibilityRole="button" accessibilityLabel={`Открыть программу ${item.program.title}`} onPress={() => onOpen(item.program.id)} style={({pressed}) => [styles.card, pressed && styles.pressed]}>
          <View style={styles.flex}><Text style={styles.cardTitle}>{item.program.title}</Text><Text style={styles.muted}>{statusLabels[item.program.status] ?? item.program.status} · {environmentLabels[item.program.environment] ?? item.program.environment}</Text></View><Text accessibilityElementsHidden style={styles.cardArrow}>→</Text>
        </Pressable>)}
        {moreError ? <Text accessibilityRole="alert" style={styles.muted}>{moreError}</Text> : null}
        {hasMore ? <AppButton label={moreError ? 'Повторить загрузку' : 'Показать ещё'} variant="secondary" loading={loadingMore} disabled={loadingMore} testID="programs-load-more" onPress={() => void loadMore()} /> : null}
      </View> : null}
    </ScrollView>
  );
}

function Stat({label, value}: {label: string; value: string}) { return <View style={styles.stat}><Text style={styles.statValue}>{value}</Text><Text style={styles.statLabel}>{label}</Text></View>; }

const styles = StyleSheet.create({
  container: {padding: spacing.lg, paddingBottom: spacing.xl, gap: spacing.md, backgroundColor: colors.background, flexGrow: 1},
  top: {flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center'}, topAction: {minHeight: control.minTouch, justifyContent: 'center', paddingHorizontal: spacing.xs},
  back: {fontWeight: '800', color: colors.text}, create: {fontWeight: '900', color: colors.text}, eyebrow: {fontSize: 11, fontWeight: '900', letterSpacing: 1.4, color: colors.textMuted},
  title: {fontSize: 30, lineHeight: 36, fontWeight: '900', color: colors.text, marginTop: spacing.xs}, subtitle: {fontSize: 14, lineHeight: 20, color: colors.textMuted, marginTop: spacing.sm},
  stateCard: {minHeight: 150, borderWidth: 1, borderColor: colors.border, borderRadius: radius.lg, alignItems: 'center', justifyContent: 'center', gap: spacing.sm},
  errorCard: {borderWidth: 1, borderColor: colors.danger, borderRadius: radius.lg, padding: spacing.md, gap: spacing.sm}, errorTitle: {fontSize: 18, fontWeight: '900', color: colors.danger},
  empty: {borderWidth: 1, borderColor: colors.border, borderRadius: radius.lg, padding: spacing.md, gap: spacing.md}, emptyTitle: {fontSize: 20, fontWeight: '900', color: colors.text}, muted: {fontSize: 13, lineHeight: 19, color: colors.textMuted},
  hero: {backgroundColor: colors.primary, borderRadius: radius.lg, padding: 18, gap: spacing.md}, heroPressed: {opacity: 0.86, transform: [{scale: 0.995}]}, heroTop: {flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center'}, flex: {flex: 1},
  heroKicker: {color: colors.inverse, fontSize: 10, fontWeight: '900', letterSpacing: 1.4, opacity: 0.65}, heroTitle: {color: colors.inverse, fontSize: 22, lineHeight: 28, fontWeight: '900', marginTop: spacing.xs}, arrow: {color: colors.inverse, fontSize: 28, marginLeft: spacing.sm},
  stats: {flexDirection: 'row', gap: spacing.sm}, stat: {flex: 1, backgroundColor: '#242424', borderRadius: radius.md, padding: 12, minHeight: 70}, statValue: {color: colors.inverse, fontSize: 18, fontWeight: '900'}, statLabel: {color: colors.inverse, fontSize: 10, opacity: 0.65, marginTop: spacing.xs}, next: {color: colors.inverse, fontSize: 13, lineHeight: 18, opacity: 0.75},
  history: {gap: spacing.sm}, section: {fontSize: 19, fontWeight: '900', color: colors.text, marginTop: spacing.xs}, card: {minHeight: 72, borderWidth: 1, borderColor: colors.border, borderRadius: radius.md, padding: spacing.md, flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center'}, cardTitle: {fontWeight: '900', fontSize: 16, color: colors.text, marginBottom: spacing.xs}, cardArrow: {fontSize: 22, color: colors.text, marginLeft: spacing.sm}, pressed: {opacity: 0.65},
});
