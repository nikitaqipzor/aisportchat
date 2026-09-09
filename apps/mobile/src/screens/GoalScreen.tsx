import React, {useState} from 'react';
import {Pressable, StyleSheet, Text, View} from 'react-native';
import {AppButton} from '../components/AppButton';
import {colors, control, radius, spacing} from '../theme/tokens';

const goals = [
  ['muscle_gain', 'Набрать мышцы'],
  ['fat_loss', 'Снизить вес'],
  ['recomposition', 'Рекомпозиция'],
  ['strength', 'Стать сильнее'],
  ['maintenance', 'Поддерживать форму'],
  ['endurance', 'Развить выносливость'],
] as const;

export function GoalScreen({onContinue}: {onContinue: (goal: string) => Promise<void>}) {
  const [selected, setSelected] = useState('muscle_gain');
  const [loading, setLoading] = useState(false);
  const [error,setError]=useState('');

  async function submit(){
    try{setLoading(true);setError('');await onContinue(selected)}
    catch(e){setError(e instanceof Error?e.message:'Не удалось сохранить цель')}
    finally{setLoading(false)}
  }

  return (
    <View style={styles.container} testID="goal-screen">
      <Text style={styles.step}>ШАГ 1 ИЗ 3</Text>
      <Text accessibilityRole="header" style={styles.title}>Какая у тебя цель?</Text>
      <Text style={styles.subtitle}>Это определит объём, интенсивность и прогрессию тренировок.</Text>
      <View style={styles.list} accessibilityRole="radiogroup">
        {goals.map(([id, label]) => (
          <Pressable key={id} testID={`goal-${id}`} accessibilityRole="radio" accessibilityLabel={label} accessibilityState={{selected:selected===id}} hitSlop={3} onPress={() => setSelected(id)} style={({pressed})=>[styles.card, selected === id && styles.selected,pressed&&styles.pressed]}>
            <Text style={[styles.cardText, selected === id && styles.selectedText]}>{label}</Text>
          </Pressable>
        ))}
      </View>
      {!!error&&<Text accessibilityRole="alert" style={styles.error}>{error}</Text>}
      <AppButton label="Продолжить" testID="goal-continue" loading={loading} onPress={()=>void submit()}/>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {flex: 1, padding: spacing.xl, paddingTop: 54, backgroundColor:colors.background},
  step: {fontSize: 12, fontWeight: '800', letterSpacing: 1.4, color:colors.textMuted},
  title: {fontSize: 32, fontWeight: '900', marginTop: 12,color:colors.text},
  subtitle: {fontSize: 16, lineHeight: 22, color:colors.textMuted, marginTop: 8},
  list: {gap: 10, marginTop: 26, flex: 1},
  card: {borderWidth: 1, borderColor: colors.border, borderRadius: radius.lg, padding: 18,minHeight:control.minTouch,justifyContent:'center',backgroundColor:colors.surface},
  selected: {backgroundColor: colors.primary, borderColor: colors.primary},pressed:{opacity:.72},
  cardText: {fontSize: 17, fontWeight: '700',color:colors.text},selectedText: {color: colors.inverse},error:{color:colors.danger,fontWeight:'700',marginBottom:spacing.sm},
});
