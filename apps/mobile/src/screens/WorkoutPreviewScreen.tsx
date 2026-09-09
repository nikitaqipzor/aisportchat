import React, {useState} from 'react';
import {Pressable, ScrollView, StyleSheet, Text, View} from 'react-native';
import {api, Exercise, WorkoutView} from '../api/client';
import {AppButton} from '../components/AppButton';
import {ExerciseGuide} from '../components/ExerciseGuide';
import {colors,radius,spacing} from '../theme/tokens';

const muscleNames: Record<string, string> = {
  chest: 'Грудь', back: 'Спина', shoulders: 'Плечи', biceps: 'Бицепс', triceps: 'Трицепс',
  forearms: 'Предплечья', core: 'Кор', quads: 'Квадрицепс', hamstrings: 'Бицепс бедра', glutes: 'Ягодицы', calves: 'Икры',
};

export function WorkoutPreviewScreen({accessToken,workout,onBack,onStart}:{accessToken:string;workout:WorkoutView;onBack:()=>void;onStart:(workout:WorkoutView)=>void}) {
  const [current,setCurrent]=useState(workout);const[busyId,setBusyId]=useState<string|null>(null);const[starting,setStarting]=useState(false);const[error,setError]=useState('');const[guide,setGuide]=useState<Exercise|null>(null);
  async function replace(workoutExerciseId:string){try{setBusyId(workoutExerciseId);setError('');setCurrent(await api.replaceExercise(accessToken,current.workout.id,workoutExerciseId,'preview_replace'))}catch(e){setError(e instanceof Error?e.message:'Не удалось заменить упражнение')}finally{setBusyId(null)}}
  async function start(){try{setStarting(true);setError('');onStart(await api.startWorkout(accessToken,current.workout.id))}catch(e){setError(e instanceof Error?e.message:'Не удалось начать тренировку')}finally{setStarting(false)}}
  return <ScrollView testID="workout-preview-screen" contentContainerStyle={styles.container}>
    <Pressable testID="workout-preview-back" accessibilityRole="button" accessibilityLabel="Вернуться назад" hitSlop={10} onPress={onBack}><Text style={styles.back}>← Назад</Text></Pressable>
    <Text style={styles.eyebrow}>ТРЕНИРОВКА ГОТОВА</Text><Text accessibilityRole="header" style={styles.title}>{muscleNames[current.workout.muscle]??current.workout.muscle}</Text><Text style={styles.subtitle}>{current.workout.duration_minutes} минут · {current.exercises.length} упражнений</Text>
    {error?<Text accessibilityRole="alert" style={styles.error}>{error}</Text>:null}
    <View style={styles.list}>{current.exercises.map((item,index)=><View key={item.workout_exercise.id} style={styles.card} testID={`workout-preview-exercise-${index+1}`}><Text style={styles.number}>{index+1}</Text><View style={styles.content}><Text style={styles.name}>{item.exercise.name}</Text><Text style={styles.meta}>{item.workout_exercise.target_sets} × {item.workout_exercise.target_reps_min}–{item.workout_exercise.target_reps_max}{item.workout_exercise.target_weight?` · ${item.workout_exercise.target_weight} кг`:''}</Text>{item.workout_exercise.progression_note?<Text style={styles.note}>{item.workout_exercise.progression_note}</Text>:null}<View style={styles.actions}><Pressable testID={`workout-preview-technique-${index+1}`} accessibilityRole="button" accessibilityLabel={`Открыть технику: ${item.exercise.name}`} hitSlop={6} onPress={()=>setGuide(item.exercise)}><Text style={styles.guide}>▶ Техника</Text></Pressable><Pressable testID={`workout-preview-replace-${index+1}`} accessibilityRole="button" accessibilityLabel={`Заменить упражнение ${item.exercise.name}`} accessibilityState={{disabled:busyId!==null,busy:busyId===item.workout_exercise.id}} disabled={busyId!==null} hitSlop={6} onPress={()=>void replace(item.workout_exercise.id)}><Text style={styles.replace}>{busyId===item.workout_exercise.id?'Подбираю…':'Заменить упражнение'}</Text></Pressable></View></View></View>)}</View>
    <ExerciseGuide exercise={guide} visible={guide!==null} onClose={()=>setGuide(null)}/>
    <AppButton label="НАЧАТЬ ТРЕНИРОВКУ" testID="workout-preview-start" loading={starting} onPress={()=>void start()}/>
  </ScrollView>;
}
const styles=StyleSheet.create({container:{padding:spacing.lg,gap:spacing.md,backgroundColor:colors.background},back:{fontWeight:'700',color:colors.text},eyebrow:{fontSize:12,letterSpacing:1.4,fontWeight:'800',marginTop:8,color:colors.textMuted},title:{fontSize:32,fontWeight:'800',color:colors.text},subtitle:{fontSize:16,color:colors.textMuted},list:{gap:10},card:{padding:16,borderWidth:1,borderColor:colors.border,borderRadius:radius.lg,flexDirection:'row',gap:14,backgroundColor:colors.surface},number:{fontSize:18,fontWeight:'800',color:colors.textMuted},content:{flex:1,gap:5},name:{fontSize:18,fontWeight:'700',color:colors.text},meta:{fontSize:15,fontWeight:'600',color:colors.text},note:{fontSize:13,lineHeight:18,color:colors.textMuted},actions:{gap:8,marginTop:5},guide:{fontSize:13,fontWeight:'800',color:colors.text},replace:{fontSize:13,fontWeight:'700',color:colors.textMuted},error:{padding:12,borderWidth:1,borderColor:colors.danger,borderRadius:12,color:colors.danger,fontWeight:'700'}});
