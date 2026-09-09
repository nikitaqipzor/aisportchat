import React, {useEffect, useState} from 'react';
import {ActivityIndicator, Pressable, ScrollView, StyleSheet, Text, View} from 'react-native';
import {api, TrainingProgram} from '../api/client';
import {AppButton} from '../components/AppButton';
import {colors, control, radius, spacing} from '../theme/tokens';

type Env = 'home' | 'gym' | 'band';
const weekOptions: Array<4 | 8 | 12> = [4, 8, 12];
const envLabels: Record<Env, string> = {home: 'Дома', gym: 'Зал', band: 'Резина'};
const workoutOptions = [1, 2, 3, 4, 5, 6, 7];

export function ProgramSetupScreen({accessToken, onBack, onCreated}: {
  accessToken: string;
  onBack: () => void;
  onCreated: (program: TrainingProgram) => void;
}) {
  const [weeks, setWeeks] = useState<4 | 8 | 12>(8);
  const [workouts, setWorkouts] = useState(4);
  const [environment, setEnvironment] = useState<Env>('gym');
  const [allowed, setAllowed] = useState<Env[]>(['gym']);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [loadingProfile, setLoadingProfile] = useState(true);
  const [profileError, setProfileError] = useState('');
  const [loadAttempt, setLoadAttempt] = useState(0);

  useEffect(() => {
    setLoadingProfile(true); setProfileError('');
    api.getProfile(accessToken).then(profile => {
      const prefs = profile.training_preferences;
      if (!prefs) return;
      const envs = prefs.environments as Env[];
      if (envs.length) {
        setAllowed(envs);
        setEnvironment(envs[0]);
      }
      setWorkouts(Math.min(7, Math.max(1, prefs.workouts_per_week || 4)));
    }).catch(e => setProfileError(e instanceof Error ? e.message : 'Не удалось загрузить настройки тренировок.'))
      .finally(() => setLoadingProfile(false));
  }, [accessToken, loadAttempt]);

  async function create() {
    if (!allowed.includes(environment) || workouts < 1 || workouts > 7) {
      setError('Проверь настройки программы.'); return;
    }
    try {
      setSaving(true); setError('');
      const program = await api.generateProgram(accessToken, {weeks, workouts_per_week: workouts, environment});
      onCreated(program);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось создать программу.');
    } finally { setSaving(false); }
  }

  return <ScrollView testID="program-setup-screen" contentContainerStyle={styles.container} contentInsetAdjustmentBehavior="automatic">
    <Pressable accessibilityRole="button" accessibilityLabel="Вернуться к программам" hitSlop={8} onPress={onBack} style={styles.backAction}><Text style={styles.back}>← Программы</Text></Pressable>
    <Text style={styles.eyebrow}>PROGRAM ENGINE</Text>
    <Text accessibilityRole="header" style={styles.title}>Новая программа</Text>
    <Text style={styles.subtitle}>Движок использует твою цель и настройки профиля. Каждая 4-я неделя автоматически становится разгрузочной.</Text>

    {loadingProfile ? <View accessibilityRole="progressbar" accessibilityLabel="Загрузка настроек профиля" style={styles.stateCard}><ActivityIndicator color={colors.primary}/><Text style={styles.muted}>Загружаем настройки тренировок…</Text></View> : profileError ? <View accessibilityRole="alert" style={styles.errorCard}><Text style={styles.infoTitle}>Настройки недоступны</Text><Text style={styles.muted}>{profileError}</Text><AppButton label="Повторить" variant="secondary" testID="program-setup-retry" onPress={() => setLoadAttempt(value => value + 1)}/></View> : <>
    <View accessibilityRole="radiogroup" accessibilityLabel="Продолжительность программы" style={styles.field}><Text style={styles.label}>Продолжительность</Text>
    <View style={styles.row}>{weekOptions.map(value => <Pressable accessibilityRole="radio" accessibilityLabel={`${value} недель`} accessibilityState={{selected: weeks === value}} testID={`program-weeks-${value}`} key={value} onPress={() => {setWeeks(value); setError('');}} style={[styles.choice, weeks === value && styles.active]}><Text style={[styles.choiceText, weeks === value && styles.activeText]}>{value} недель</Text></Pressable>)}</View></View>

    <View accessibilityRole="radiogroup" accessibilityLabel="Количество тренировок в неделю" style={styles.field}><Text style={styles.label}>Тренировок в неделю</Text>
    <View style={styles.wrap}>{workoutOptions.map(value => <Pressable accessibilityRole="radio" accessibilityLabel={`${value} тренировок в неделю`} accessibilityState={{selected: workouts === value}} testID={`program-workouts-${value}`} key={value} onPress={() => {setWorkouts(value); setError('');}} style={[styles.smallChoice, workouts === value && styles.active]}><Text style={[styles.choiceText, workouts === value && styles.activeText]}>{value}</Text></Pressable>)}</View></View>

    <View accessibilityRole="radiogroup" accessibilityLabel="Место тренировки" style={styles.field}><Text style={styles.label}>Среда</Text>
    {allowed.length ? <View style={styles.row}>{allowed.map(value => <Pressable accessibilityRole="radio" accessibilityLabel={envLabels[value]} accessibilityState={{selected: environment === value}} testID={`program-environment-${value}`} key={value} onPress={() => {setEnvironment(value); setError('');}} style={[styles.choice, environment === value && styles.active]}><Text style={[styles.choiceText, environment === value && styles.activeText]}>{envLabels[value]}</Text></Pressable>)}</View> : <View style={styles.emptyCard}><Text style={styles.infoTitle}>Среда не выбрана</Text><Text style={styles.muted}>Сначала добавь место тренировки в настройках профиля.</Text></View>}</View>

    <View style={styles.info}><Text style={styles.infoTitle}>Что будет создано</Text><Text style={styles.infoText}>• календарь на {weeks} недель{`\n`}• {workouts} тренировок в неделю{`\n`}• плановый объём по мышцам{`\n`}• deload на неделях 4{weeks >= 8 ? ', 8' : ''}{weeks >= 12 ? ', 12' : ''}{`\n`}• adherence и автоперенос пропусков</Text></View>
    {error ? <Text accessibilityRole="alert" style={styles.error}>{error}</Text> : null}
    <AppButton label="Создать программу" accessibilityLabel={`Создать программу на ${weeks} недель, ${workouts} тренировок в неделю`} testID="program-setup-create" loading={saving} disabled={!allowed.length} onPress={() => void create()}/>
    </>}
  </ScrollView>;
}

const styles=StyleSheet.create({
  container:{padding:spacing.lg,paddingBottom:spacing.xl,gap:14,backgroundColor:colors.background,flexGrow:1}, backAction:{minHeight:control.minTouch,justifyContent:'center',alignSelf:'flex-start'}, back:{fontSize:14,fontWeight:'800',color:colors.text}, eyebrow:{fontSize:11,fontWeight:'900',letterSpacing:1.4,color:colors.textMuted,marginTop:8}, title:{fontSize:30,lineHeight:36,fontWeight:'900',color:colors.text}, subtitle:{fontSize:15,lineHeight:22,color:colors.textMuted}, field:{gap:8}, label:{fontSize:14,fontWeight:'900',marginTop:8,color:colors.text}, row:{flexDirection:'row',flexWrap:'wrap',gap:8}, wrap:{flexDirection:'row',flexWrap:'wrap',gap:8}, choice:{flexGrow:1,flexBasis:90,minHeight:control.minTouch,borderWidth:1,borderColor:colors.border,borderRadius:14,paddingHorizontal:8,alignItems:'center',justifyContent:'center'}, smallChoice:{width:control.minTouch,minHeight:control.minTouch,borderWidth:1,borderColor:colors.border,borderRadius:14,alignItems:'center',justifyContent:'center'}, active:{backgroundColor:colors.primary,borderColor:colors.primary}, choiceText:{fontWeight:'800',color:colors.text,textAlign:'center'}, activeText:{color:colors.inverse}, info:{borderWidth:1,borderColor:colors.border,borderRadius:18,padding:16,gap:6}, infoTitle:{fontSize:16,fontWeight:'900',color:colors.text}, infoText:{fontSize:14,lineHeight:22,color:colors.textMuted}, stateCard:{minHeight:160,borderWidth:1,borderColor:colors.border,borderRadius:radius.lg,alignItems:'center',justifyContent:'center',gap:spacing.sm}, errorCard:{borderWidth:1,borderColor:colors.danger,borderRadius:radius.lg,padding:spacing.md,gap:spacing.sm}, emptyCard:{borderWidth:1,borderColor:colors.border,borderRadius:radius.md,padding:spacing.md,gap:spacing.sm}, muted:{fontSize:13,lineHeight:19,color:colors.textMuted}, error:{fontSize:13,lineHeight:19,fontWeight:'700',color:colors.danger}
});
