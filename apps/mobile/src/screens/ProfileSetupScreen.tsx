import React, {useState} from 'react';
import {Pressable, StyleSheet, Text, TextInput, View} from 'react-native';
import {AppButton} from '../components/AppButton';
import {colors,control,radius,spacing} from '../theme/tokens';

const levels = [['beginner', 'Новичок'],['intermediate', 'Средний'],['advanced', 'Продвинутый']] as const;

export function ProfileSetupScreen({onContinue}: {onContinue: (data: {height: number; weight: number; level: string}) => Promise<void>}) {
  const [height, setHeight] = useState('');
  const [weight, setWeight] = useState('');
  const [level, setLevel] = useState('beginner');
  const [loading,setLoading]=useState(false);const[error,setError]=useState('');
  const valid = Number(height) >= 100 && Number(height)<=250 && Number(weight) >= 30 && Number(weight)<=350;
  async function submit(){try{setLoading(true);setError('');await onContinue({height:Number(height),weight:Number(weight),level})}catch(e){setError(e instanceof Error?e.message:'Не удалось сохранить профиль')}finally{setLoading(false)}}

  return <View style={styles.container} testID="profile-setup-screen">
    <Text style={styles.step}>ШАГ 2 ИЗ 3</Text><Text accessibilityRole="header" style={styles.title}>Расскажи о себе</Text>
    <View style={styles.row}>
      <View style={styles.field}><Text style={styles.label}>Рост, см</Text><TextInput accessibilityLabel="Рост в сантиметрах" testID="profile-height" keyboardType="numeric" value={height} onChangeText={setHeight} style={styles.input} placeholder="193" placeholderTextColor={colors.textMuted}/></View>
      <View style={styles.field}><Text style={styles.label}>Вес, кг</Text><TextInput accessibilityLabel="Вес в килограммах" testID="profile-weight" keyboardType="decimal-pad" value={weight} onChangeText={setWeight} style={styles.input} placeholder="97" placeholderTextColor={colors.textMuted}/></View>
    </View>
    <Text style={styles.label}>Опыт тренировок</Text>
    <View style={styles.list} accessibilityRole="radiogroup">{levels.map(([id,label])=><Pressable key={id} testID={`profile-level-${id}`} accessibilityRole="radio" accessibilityLabel={label} accessibilityState={{selected:level===id}} onPress={()=>setLevel(id)} style={({pressed})=>[styles.card,level===id&&styles.selected,pressed&&styles.pressed]}><Text style={[styles.cardText,level===id&&styles.selectedText]}>{label}</Text></Pressable>)}</View>
    <View style={{flex:1}}/>{!!error&&<Text accessibilityRole="alert" style={styles.error}>{error}</Text>}<AppButton label="Продолжить" testID="profile-continue" loading={loading} disabled={!valid} onPress={()=>void submit()}/>
  </View>;
}
const styles=StyleSheet.create({container:{flex:1,padding:spacing.xl,paddingTop:54,backgroundColor:colors.background},step:{fontSize:12,fontWeight:'800',color:colors.textMuted,letterSpacing:1.4},title:{fontSize:32,fontWeight:'900',marginTop:12,marginBottom:26,color:colors.text},row:{flexDirection:'row',gap:12},field:{flex:1},label:{fontSize:14,fontWeight:'800',marginBottom:8,color:colors.text},input:{borderWidth:1,borderColor:colors.border,borderRadius:radius.md,padding:15,fontSize:18,marginBottom:22,color:colors.text,backgroundColor:colors.surface,minHeight:control.minTouch},list:{gap:10},card:{borderWidth:1,borderColor:colors.border,borderRadius:radius.md,padding:16,minHeight:control.minTouch,justifyContent:'center',backgroundColor:colors.surface},selected:{backgroundColor:colors.primary,borderColor:colors.primary},cardText:{fontSize:16,fontWeight:'700',color:colors.text},selectedText:{color:colors.inverse},pressed:{opacity:.72},error:{color:colors.danger,fontWeight:'700',marginBottom:spacing.sm}});
