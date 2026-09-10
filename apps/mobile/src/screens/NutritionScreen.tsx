import React, {useCallback, useEffect, useRef, useState} from 'react';
import {ActivityIndicator, Pressable, RefreshControl, ScrollView, StyleSheet, Text, View} from 'react-native';
import {api} from '../api/client';
import type {FoodEntry, NutritionDay, NutritionHistoryItem} from '../api/client';
import {AppButton} from '../components/AppButton';
import {currentLocalDate, shiftLocalDate} from '../domain/date';
import {colors, control, radius, spacing} from '../theme/tokens';

function Metric({title, value, target, unit}: {title: string; value: number; target: number; unit: string}) {
  const pct = target > 0 ? Math.min(100, value / target * 100) : 0;
  return (
    <View style={styles.metric}>
      <View style={styles.metricTop}>
        <Text style={styles.metricTitle}>{title}</Text>
        <Text style={styles.metricValue}>{Math.round(value)} / {Math.round(target)} {unit}</Text>
      </View>
      <View accessibilityRole="progressbar" accessibilityLabel={title} accessibilityValue={{min:0,max:Math.max(0,Math.round(target)),now:target>0?Math.min(Math.round(target),Math.round(value)):0,text:`${Math.round(value)} из ${Math.round(target)} ${unit}`}} style={styles.track}><View style={[styles.fill, {width: `${pct}%`}]} /></View>
    </View>
  );
}

const mealTitles: Record<string, string> = {breakfast: 'Завтрак', lunch: 'Обед', dinner: 'Ужин', snack: 'Перекус'};

export function NutritionScreen({
  accessToken,
  onBack,
  onSetup,
  onAddFood,
  onRecipes,
  onAI,
}: {
  accessToken: string;
  onBack: () => void;
  onSetup: () => void;
  onAddFood: () => void;
  onRecipes: () => void;
  onAI: () => void;
}) {
  const [day, setDay] = useState<NutritionDay | null>(null);
  const [selectedDate, setSelectedDate] = useState(currentLocalDate());
  const [history, setHistory] = useState<NutritionHistoryItem[]>([]);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState('');
  const [deletingId, setDeletingId] = useState('');
  const [repeatingId, setRepeatingId] = useState('');
  const [undoEntry, setUndoEntry] = useState<FoodEntry | null>(null);
  const undoTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const load = useCallback(async () => {
    try {
      setError('');
      const profile = await api.nutritionProfile(accessToken);
      if (!profile) {
        onSetup();
        return;
      }
      const [selectedDay, h] = await Promise.all([api.nutritionToday(accessToken, selectedDate), api.nutritionHistory(accessToken, 7)]);
      setDay(selectedDay);
      setHistory(h.items);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось загрузить питание');
    }
  }, [accessToken, onSetup, selectedDate]);

  useEffect(() => { void load(); }, [load]);
  useEffect(() => () => { if (undoTimer.current) clearTimeout(undoTimer.current); }, []);

  async function refresh() {
    setRefreshing(true);
    await load();
    setRefreshing(false);
  }

  function openDate(date: string) {
    if (date === selectedDate) return;
    setDay(null);
    setError('');
    setSelectedDate(date);
  }

  async function remove(entry: FoodEntry) {
    try {
      setDeletingId(entry.id);
      setError('');
      await api.deleteFoodEntry(accessToken, entry.id);
      setUndoEntry(entry);
      if (undoTimer.current) clearTimeout(undoTimer.current);
      undoTimer.current = setTimeout(() => setUndoEntry(null), 6000);
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось удалить запись');
    } finally {
      setDeletingId('');
    }
  }

  async function undoDelete() {
    if (!undoEntry) return;
    const entry = undoEntry;
    try {
      setUndoEntry(null);
      if (undoTimer.current) clearTimeout(undoTimer.current);
      await api.logFood(accessToken, {
        food_id: entry.food_id,
        meal_type: entry.meal_type,
        quantity_g: entry.quantity_g,
        logged_at: entry.logged_at,
      });
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось вернуть запись');
    }
  }

  async function repeat(entry: FoodEntry) {
    try {
      setRepeatingId(entry.id);
      setDay(await api.repeatFoodEntry(accessToken, entry.id, entry.meal_type));
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось повторить запись');
    } finally {
      setRepeatingId('');
    }
  }

  if (!day) {
    return (
      <ScrollView contentContainerStyle={styles.container}>
        <Pressable accessibilityRole="button" accessibilityLabel="Назад" testID="nutrition-back-loading" onPress={onBack} style={styles.backButton}><Text style={styles.back}>← Назад</Text></Pressable>
        <Text accessibilityRole="header" style={styles.title}>Питание</Text>
        {error ? <View style={styles.errorCard}><Text accessibilityRole="alert" style={styles.error}>{error}</Text><AppButton label="Повторить" variant="secondary" onPress={()=>void load()}/></View> : <View style={styles.loading} accessibilityLiveRegion="polite"><ActivityIndicator color={colors.text}/><Text style={styles.muted}>Загружаем дневник…</Text></View>}
      </ScrollView>
    );
  }

  const grouped = day.entries.reduce<Record<string, typeof day.entries>>((acc, item) => {
    (acc[item.meal_type] ??= []).push(item);
    return acc;
  }, {});
  const isToday = selectedDate === currentLocalDate();
  const dateTitle = isToday ? 'Сегодня' : new Date(`${selectedDate}T12:00:00`).toLocaleDateString('ru-RU', {day: 'numeric', month: 'long'});

  return (
    <ScrollView contentContainerStyle={styles.container} refreshControl={<RefreshControl refreshing={refreshing} onRefresh={refresh} />}>
      <View style={styles.header}>
        <Pressable accessibilityRole="button" accessibilityLabel="Назад на главную" testID="nutrition-back" hitSlop={4} onPress={onBack} style={styles.backButton}><Text style={styles.back}>← Главная</Text></Pressable>
        <Pressable accessibilityRole="button" accessibilityLabel="Настроить цели питания" testID="nutrition-settings" hitSlop={4} onPress={onSetup} style={styles.backButton}><Text style={styles.edit}>Настроить</Text></Pressable>
      </View>

      <Text style={styles.kicker}>FITNESS 2.0 · ПИТАНИЕ</Text>
      <View style={styles.dateNavigation}>
        <Pressable accessibilityRole="button" accessibilityLabel="Предыдущий день" testID="nutrition-previous-day" onPress={() => openDate(shiftLocalDate(selectedDate, -1))} style={styles.dateButton}><Text style={styles.dateArrow}>‹</Text></Pressable>
        <View style={styles.dateCenter}><Text accessibilityRole="header" style={styles.title}>{dateTitle}</Text><Text style={styles.dateCaption}>{selectedDate}</Text></View>
        <Pressable accessibilityRole="button" accessibilityLabel="Следующий день" accessibilityState={{disabled: isToday}} testID="nutrition-next-day" disabled={isToday} onPress={() => openDate(shiftLocalDate(selectedDate, 1))} style={[styles.dateButton, isToday && styles.dateButtonDisabled]}><Text style={styles.dateArrow}>›</Text></Pressable>
      </View>
      <View style={styles.hero}>
        <Text style={styles.heroValue}>{Math.round(day.consumed_calories)}</Text>
        <Text style={styles.heroTarget}>/ {day.profile.calorie_target} kcal</Text>
        <Text style={styles.heroRemain}>{day.remaining_calories >= 0 ? `Осталось ${Math.round(day.remaining_calories)} kcal` : `Выше цели на ${Math.abs(Math.round(day.remaining_calories))} kcal`}</Text>
        {day.training_day ? <Text style={styles.training}>Тренировочный день · {day.completed_workouts} трен.</Text> : null}
      </View>

      <Metric title="Белок" value={day.consumed_protein_g} target={day.profile.protein_target_g} unit="г" />
      <Metric title="Жиры" value={day.consumed_fat_g} target={day.profile.fat_target_g} unit="г" />
      <Metric title="Углеводы" value={day.consumed_carbs_g} target={day.profile.carb_target_g} unit="г" />

      {isToday ? <>
        <AppButton label="✨ Добавить через AI" testID="nutrition-ai-add" onPress={onAI} />
        <View style={styles.quickSecondary}>
          <AppButton label="Поиск" variant="secondary" testID="nutrition-search" onPress={onAddFood} style={styles.quickHalf} />
          <AppButton label="Рецепты" variant="secondary" testID="nutrition-recipes" onPress={onRecipes} style={styles.quickHalf} />
        </View>
      </> : <View style={styles.pastDayNotice}><Text style={styles.pastDayText}>Просмотр прошедшего дня. Новые записи добавляются в сегодняшний дневник.</Text><AppButton label="Вернуться к сегодня" variant="secondary" testID="nutrition-return-today" onPress={() => openDate(currentLocalDate())}/></View>}

      <Text style={styles.section}>Дневник</Text>
      {(['breakfast', 'lunch', 'dinner', 'snack'] as const).map(meal => (
        <View key={meal} style={styles.mealCard}>
          <Text style={styles.mealTitle}>{mealTitles[meal]}</Text>
          {(grouped[meal] ?? []).length === 0 ? <Text style={styles.empty}>Пока ничего</Text> : (grouped[meal] ?? []).map(entry => (
            <View key={entry.id} style={styles.entry}>
              <View style={styles.entryBody}>
                <Text style={styles.entryTitle}>{entry.food_name} · {Math.round(entry.quantity_g)} г</Text>
                <Text style={styles.entryMeta}>{Math.round(entry.calories)} kcal · Б {Math.round(entry.protein_g)} · Ж {Math.round(entry.fat_g)} · У {Math.round(entry.carbs_g)}</Text>
              </View>
              <View style={styles.entryActions}>
                <Pressable
                  accessibilityRole="button"
                  accessibilityLabel={`Повторить ${entry.food_name}`}
                  accessibilityState={{disabled: repeatingId === entry.id}}
                  testID={`nutrition-repeat-${entry.id}`}
                  disabled={repeatingId === entry.id || deletingId === entry.id}
                  onPress={() => { void repeat(entry); }}
                  style={({pressed}) => [styles.iconButton, pressed && styles.pressed]}>
                  <Text style={styles.repeat}>{repeatingId === entry.id ? '…' : '＋'}</Text>
                </Pressable>
                <Pressable
                  accessibilityRole="button"
                  accessibilityLabel={`Удалить ${entry.food_name}`}
                  accessibilityState={{disabled: deletingId === entry.id}}
                  testID={`nutrition-delete-${entry.id}`}
                  disabled={deletingId === entry.id || repeatingId === entry.id}
                  onPress={() => { void remove(entry); }}
                  style={({pressed}) => [styles.iconButton, pressed && styles.pressed]}>
                  <Text style={styles.delete}>{deletingId === entry.id ? '…' : '×'}</Text>
                </Pressable>
              </View>
            </View>
          ))}
        </View>
      ))}

      {undoEntry ? (
        <View style={styles.undoBar} accessibilityLiveRegion="polite">
          <Text style={styles.undoText}>«{undoEntry.food_name}» удалено</Text>
          <Pressable accessibilityRole="button" accessibilityLabel="Отменить удаление" testID="nutrition-undo-delete" onPress={() => { void undoDelete(); }} style={styles.undoButton}>
            <Text style={styles.undoAction}>Вернуть</Text>
          </Pressable>
        </View>
      ) : null}

      <Text style={styles.section}>Последние 7 дней</Text>
      <View style={styles.history}>
        {history.length === 0 ? <Text style={styles.historyEmpty}>Пока недостаточно записей для истории.</Text> : null}
        {history.map(item => (
          <Pressable key={item.date} accessibilityRole="button" accessibilityLabel={`Открыть питание за ${item.date}`} accessibilityState={{selected: item.date === selectedDate}} onPress={() => openDate(item.date)} style={[styles.historyRow, item.date === selectedDate && styles.historyRowSelected]}>
            <View><Text style={styles.historyDate}>{item.date.slice(5)}</Text><Text style={styles.historyMeta}>{item.training_day ? '● тренировочный' : 'день отдыха'}</Text></View>
            <Text style={styles.historyCalories}>{Math.round(item.calories)} / {item.target_calories}</Text>
          </Pressable>
        ))}
      </View>
      {error ? <Text style={styles.error}>{error}</Text> : null}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {padding: spacing.lg, gap: 14},
  header: {flexDirection: 'row', justifyContent: 'space-between'},
  backButton: {minHeight: control.minTouch, justifyContent: 'center'},
  back: {fontWeight: '800'},
  edit: {fontWeight: '800'},
  kicker: {fontSize: 11, fontWeight: '900', letterSpacing: 1.4, color: colors.textMuted},
  title: {fontSize: 32, fontWeight: '900'},
  dateNavigation:{flexDirection:'row',alignItems:'center',justifyContent:'space-between',gap:spacing.sm},
  dateButton:{width:control.minTouch,height:control.minTouch,borderWidth:1,borderColor:colors.border,borderRadius:radius.md,alignItems:'center',justifyContent:'center'},
  dateButtonDisabled:{opacity:.3},
  dateArrow:{fontSize:30,lineHeight:34,fontWeight:'500',color:colors.text},
  dateCenter:{flex:1,alignItems:'center'},
  dateCaption:{fontSize:11,color:colors.textMuted,marginTop:2},
  muted: {color: colors.textMuted,lineHeight:19},
  loading:{minHeight:100,alignItems:'center',justifyContent:'center',gap:spacing.sm},
  hero: {borderWidth: 1, borderColor: colors.border, borderRadius: radius.lg, padding: 18},
  heroValue: {fontSize: 44, fontWeight: '900'},
  heroTarget: {fontSize: 18, fontWeight: '800', color: colors.textMuted, marginTop: -4},
  heroRemain: {fontSize: 13, fontWeight: '700', marginTop: 10},
  training: {fontSize: 12, fontWeight: '800', marginTop: 7, color: colors.textMuted},
  metric: {gap: 7},
  metricTop: {flexDirection: 'row', justifyContent: 'space-between'},
  metricTitle: {fontWeight: '800'},
  metricValue: {fontWeight: '700', color: colors.textMuted},
  track: {height: 9, borderRadius: 99, backgroundColor: '#e9e9e9', overflow: 'hidden'},
  fill: {height: '100%', backgroundColor: colors.primary, borderRadius: 99},
  quickSecondary: {flexDirection: 'row', gap: spacing.sm},
  quickHalf: {flex: 1},
  pastDayNotice:{borderRadius:radius.md,backgroundColor:colors.surfaceMuted,padding:spacing.md,gap:spacing.sm},
  pastDayText:{fontSize:12,lineHeight:18,color:colors.textMuted},
  section: {fontSize: 19, fontWeight: '900', marginTop: 4},
  mealCard: {borderWidth: 1, borderColor: colors.border, borderRadius: 18, padding: 14, gap: 9},
  mealTitle: {fontWeight: '900', fontSize: 16},
  empty: {fontSize: 12, color: colors.textMuted},
  entry: {flexDirection: 'row', alignItems: 'center', gap: 8, borderTopWidth: StyleSheet.hairlineWidth, paddingTop: 9},
  entryBody: {flex: 1},
  entryTitle: {fontWeight: '800'},
  entryMeta: {fontSize: 11, color: colors.textMuted, marginTop: 3},
  entryActions: {flexDirection: 'row', alignItems: 'center', gap: 4},
  iconButton: {minWidth: control.minTouch, minHeight: control.minTouch, alignItems: 'center', justifyContent: 'center', borderRadius: radius.md},
  repeat: {fontSize: 19, fontWeight: '700'},
  delete: {fontSize: 25, fontWeight: '300', color: colors.danger},
  undoBar: {backgroundColor: colors.primary, borderRadius: radius.md, paddingHorizontal: spacing.md, minHeight: control.minTouch, flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', gap: spacing.sm},
  undoText: {color: colors.inverse, fontWeight: '700', flex: 1},
  undoButton: {minHeight: control.minTouch, justifyContent: 'center', paddingHorizontal: spacing.sm},
  undoAction: {color: colors.inverse, fontWeight: '900'},
  history: {borderWidth: 1, borderColor: colors.border, borderRadius: 18, overflow: 'hidden'},
  historyEmpty:{padding:spacing.md,color:colors.textMuted,fontSize:12,lineHeight:18},
  historyRow: {flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', padding: 12, borderBottomWidth: StyleSheet.hairlineWidth},
  historyRowSelected:{backgroundColor:colors.surfaceMuted},
  historyDate: {fontWeight: '800'},
  historyMeta: {fontSize: 10, color: colors.textMuted, marginTop: 2},
  historyCalories: {fontWeight: '800'},
  error: {fontSize: 12, fontWeight: '700', color: colors.danger},
  errorCard:{borderWidth:1,borderColor:colors.danger,borderRadius:radius.md,padding:spacing.md,gap:spacing.sm},
  pressed: {opacity: 0.6},
});
