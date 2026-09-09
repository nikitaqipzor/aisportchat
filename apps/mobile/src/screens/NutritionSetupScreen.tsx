import React, {useState} from 'react';
import {Pressable, ScrollView, StyleSheet, Text, TextInput, View} from 'react-native';
import {api, NutritionProfile} from '../api/client';

type Goal = NutritionProfile['goal'];
type Activity = NutritionProfile['activity_level'];

const goals: Array<{id: Goal; title: string; subtitle: string}> = [
  {id: 'lose', title: 'Снижение веса', subtitle: 'Умеренный дефицит'},
  {id: 'recomp', title: 'Рекомпозиция', subtitle: 'Сохранение мышц + постепенное снижение жира'},
  {id: 'maintain', title: 'Поддержание', subtitle: 'Стабильный вес'},
  {id: 'gain', title: 'Набор массы', subtitle: 'Умеренный профицит'},
  {id: 'strength', title: 'Сила', subtitle: 'Небольшой профицит для силовой работы'},
  {id: 'endurance', title: 'Выносливость', subtitle: 'Дополнительная энергия под объём'},
];

const activities: Array<{id: Activity; title: string}> = [
  {id: 'low', title: 'Низкая'},
  {id: 'light', title: 'Лёгкая'},
  {id: 'moderate', title: 'Средняя'},
  {id: 'high', title: 'Высокая'},
  {id: 'athlete', title: 'Очень высокая'},
];

export function NutritionSetupScreen({accessToken, onBack, onSaved}: {accessToken: string; onBack: () => void; onSaved: (profile: NutritionProfile) => void}) {
  const [goal, setGoal] = useState<Goal>('recomp');
  const [activity, setActivity] = useState<Activity>('moderate');
  const [mode, setMode] = useState<'auto' | 'manual'>('auto');
  const [calories, setCalories] = useState('2500');
  const [protein, setProtein] = useState('180');
  const [fat, setFat] = useState('75');
  const [carbs, setCarbs] = useState('280');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  async function save() {
    try {
      setSaving(true);
      setError('');
      const manual = mode === 'manual';
      const profile = await api.setNutritionProfile(accessToken, {
        goal,
        activity_level: activity,
        calculation_mode: mode,
        ...(manual ? {
          calorie_target: Number(calories),
          protein_target_g: Number(protein),
          fat_target_g: Number(fat),
          carb_target_g: Number(carbs),
        } : {}),
      });
      onSaved(profile);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось сохранить настройки питания');
    } finally {
      setSaving(false);
    }
  }

  return (
    <ScrollView contentContainerStyle={styles.container}>
      <View style={styles.header}><Pressable onPress={onBack}><Text style={styles.back}>← Назад</Text></Pressable><Text style={styles.kicker}>ПИТАНИЕ</Text></View>
      <Text style={styles.title}>Настроим цель</Text>
      <Text style={styles.subtitle}>Первый расчёт — ориентир. Позже его можно вручную изменить или адаптировать по реальной динамике веса.</Text>

      <Text style={styles.section}>Цель</Text>
      <View style={styles.stack}>
        {goals.map(item => <Pressable key={item.id} onPress={() => setGoal(item.id)} style={[styles.card, goal === item.id && styles.cardActive]}>
          <Text style={[styles.cardTitle, goal === item.id && styles.activeText]}>{item.title}</Text>
          <Text style={[styles.cardSubtitle, goal === item.id && styles.activeSub]}>{item.subtitle}</Text>
        </Pressable>)}
      </View>

      <Text style={styles.section}>Общая активность</Text>
      <View style={styles.wrap}>
        {activities.map(item => <Pressable key={item.id} onPress={() => setActivity(item.id)} style={[styles.pill, activity === item.id && styles.pillActive]}>
          <Text style={[styles.pillText, activity === item.id && styles.activeText]}>{item.title}</Text>
        </Pressable>)}
      </View>

      <Text style={styles.section}>Режим расчёта</Text>
      <View style={styles.modeRow}>
        <Pressable onPress={() => setMode('auto')} style={[styles.modeButton, mode === 'auto' && styles.modeButtonActive]}><Text style={[styles.modeText, mode === 'auto' && styles.activeText]}>Авто</Text></Pressable>
        <Pressable onPress={() => setMode('manual')} style={[styles.modeButton, mode === 'manual' && styles.modeButtonActive]}><Text style={[styles.modeText, mode === 'manual' && styles.activeText]}>Вручную</Text></Pressable>
      </View>
      {mode === 'manual' ? <View style={styles.manualGrid}>
        <View style={styles.manualField}><Text style={styles.label}>Ккал</Text><TextInput value={calories} onChangeText={setCalories} keyboardType="number-pad" style={styles.manualInput} /></View>
        <View style={styles.manualField}><Text style={styles.label}>Белок, г</Text><TextInput value={protein} onChangeText={setProtein} keyboardType="decimal-pad" style={styles.manualInput} /></View>
        <View style={styles.manualField}><Text style={styles.label}>Жиры, г</Text><TextInput value={fat} onChangeText={setFat} keyboardType="decimal-pad" style={styles.manualInput} /></View>
        <View style={styles.manualField}><Text style={styles.label}>Углеводы, г</Text><TextInput value={carbs} onChangeText={setCarbs} keyboardType="decimal-pad" style={styles.manualInput} /></View>
      </View> : null}

      <View style={styles.note}><Text style={styles.noteText}>Авто-режим использует текущий вес из профиля как стартовый ориентир. Все цели можно переключить на ручной режим.</Text></View>
      {error ? <Text style={styles.error}>{error}</Text> : null}
      <Pressable disabled={saving} onPress={save} style={styles.primary}><Text style={styles.primaryText}>{saving ? 'Сохраняю…' : 'Рассчитать питание'}</Text></Pressable>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {padding: 20, gap: 14},
  header: {flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center'},
  back: {fontWeight: '800'}, kicker: {fontSize: 11, fontWeight: '900', letterSpacing: 1.4, opacity: 0.55},
  title: {fontSize: 30, fontWeight: '900', marginTop: 8}, subtitle: {fontSize: 15, lineHeight: 21, opacity: 0.65},
  section: {fontSize: 18, fontWeight: '900', marginTop: 8}, stack: {gap: 8},
  card: {borderWidth: 1, borderRadius: 18, padding: 14}, cardActive: {backgroundColor: '#111', borderColor: '#111'},
  cardTitle: {fontSize: 16, fontWeight: '850'}, cardSubtitle: {fontSize: 12, marginTop: 4, opacity: 0.55}, activeText: {color: '#fff'}, activeSub: {color: '#fff', opacity: 0.72},
  wrap: {flexDirection: 'row', flexWrap: 'wrap', gap: 8}, pill: {borderWidth: 1, borderRadius: 999, paddingHorizontal: 14, paddingVertical: 10}, pillActive: {backgroundColor: '#111'}, pillText: {fontWeight: '750'},
  modeRow: {flexDirection: 'row', gap: 8}, modeButton: {flex: 1, borderWidth: 1, borderRadius: 14, paddingVertical: 11, alignItems: 'center'}, modeButtonActive: {backgroundColor: '#111'}, modeText: {fontWeight: '850'},
  manualGrid: {flexDirection: 'row', flexWrap: 'wrap', gap: 8}, manualField: {width: '48%', gap: 5}, label: {fontSize: 11, fontWeight: '750', opacity: 0.55}, manualInput: {borderWidth: 1, borderRadius: 12, paddingHorizontal: 12, paddingVertical: 10, fontWeight: '850', fontSize: 16},
  note: {borderWidth: 1, borderRadius: 16, padding: 14, marginTop: 4}, noteText: {fontSize: 13, lineHeight: 18, opacity: 0.68},
  error: {fontSize: 13, fontWeight: '700'}, primary: {backgroundColor: '#111', borderRadius: 18, paddingVertical: 15, alignItems: 'center', marginTop: 4}, primaryText: {color: '#fff', fontWeight: '900', fontSize: 16},
});
