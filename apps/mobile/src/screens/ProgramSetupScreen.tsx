import React, {useEffect, useState} from 'react';
import {ActivityIndicator, Pressable, ScrollView, StyleSheet, Text, View} from 'react-native';
import {api, TrainingProgram} from '../api/client';

type Env = 'home' | 'gym' | 'band';
const weekOptions: Array<4 | 8 | 12> = [4, 8, 12];
const envLabels: Record<Env, string> = {home: 'Дома', gym: 'Зал', band: 'Резина'};

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

  useEffect(() => {
    api.getProfile(accessToken).then(profile => {
      const prefs = profile.training_preferences;
      if (!prefs) return;
      const envs = prefs.environments as Env[];
      if (envs.length) {
        setAllowed(envs);
        setEnvironment(envs[0]);
      }
      setWorkouts(Math.min(7, Math.max(1, prefs.workouts_per_week || 4)));
    }).catch(() => undefined);
  }, [accessToken]);

  async function create() {
    try {
      setSaving(true); setError('');
      const program = await api.generateProgram(accessToken, {weeks, workouts_per_week: workouts, environment});
      onCreated(program);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось создать программу.');
    } finally { setSaving(false); }
  }

  return <ScrollView contentContainerStyle={styles.container}>
    <Pressable onPress={onBack}><Text style={styles.back}>← Программы</Text></Pressable>
    <Text style={styles.eyebrow}>PROGRAM ENGINE</Text>
    <Text style={styles.title}>Новая программа</Text>
    <Text style={styles.subtitle}>Движок использует твою цель и настройки профиля. Каждая 4-я неделя автоматически становится разгрузочной.</Text>

    <Text style={styles.label}>Продолжительность</Text>
    <View style={styles.row}>{weekOptions.map(value => <Pressable key={value} onPress={() => setWeeks(value)} style={[styles.choice, weeks === value && styles.active]}><Text style={[styles.choiceText, weeks === value && styles.activeText]}>{value} недель</Text></Pressable>)}</View>

    <Text style={styles.label}>Тренировок в неделю</Text>
    <View style={styles.wrap}>{[2,3,4,5,6].map(value => <Pressable key={value} onPress={() => setWorkouts(value)} style={[styles.smallChoice, workouts === value && styles.active]}><Text style={[styles.choiceText, workouts === value && styles.activeText]}>{value}</Text></Pressable>)}</View>

    <Text style={styles.label}>Среда</Text>
    <View style={styles.row}>{allowed.map(value => <Pressable key={value} onPress={() => setEnvironment(value)} style={[styles.choice, environment === value && styles.active]}><Text style={[styles.choiceText, environment === value && styles.activeText]}>{envLabels[value]}</Text></Pressable>)}</View>

    <View style={styles.info}><Text style={styles.infoTitle}>Что будет создано</Text><Text style={styles.infoText}>• календарь на {weeks} недель{`\n`}• {workouts} тренировок в неделю{`\n`}• плановый объём по мышцам{`\n`}• deload на неделях 4{weeks >= 8 ? ', 8' : ''}{weeks >= 12 ? ', 12' : ''}{`\n`}• adherence и автоперенос пропусков</Text></View>
    {error ? <Text style={styles.error}>{error}</Text> : null}
    <Pressable disabled={saving} onPress={create} style={styles.primary}>{saving ? <ActivityIndicator color="#fff"/> : <Text style={styles.primaryText}>Создать программу</Text>}</Pressable>
  </ScrollView>;
}

const styles=StyleSheet.create({
  container:{padding:20,gap:14}, back:{fontSize:14,fontWeight:'800'}, eyebrow:{fontSize:11,fontWeight:'900',letterSpacing:1.4,opacity:.45,marginTop:8}, title:{fontSize:30,fontWeight:'900'}, subtitle:{fontSize:15,lineHeight:22,opacity:.65}, label:{fontSize:14,fontWeight:'900',marginTop:8}, row:{flexDirection:'row',gap:8}, wrap:{flexDirection:'row',flexWrap:'wrap',gap:8}, choice:{flex:1,borderWidth:1,borderRadius:14,paddingVertical:12,alignItems:'center'}, smallChoice:{minWidth:52,borderWidth:1,borderRadius:14,paddingVertical:12,alignItems:'center'}, active:{backgroundColor:'#111'}, choiceText:{fontWeight:'800'}, activeText:{color:'#fff'}, info:{borderWidth:1,borderRadius:18,padding:16,gap:6}, infoTitle:{fontSize:16,fontWeight:'900'}, infoText:{fontSize:14,lineHeight:22,opacity:.68}, primary:{backgroundColor:'#111',borderRadius:16,padding:16,alignItems:'center',marginTop:8}, primaryText:{color:'#fff',fontWeight:'900'}, error:{fontSize:13,fontWeight:'700'}
});
