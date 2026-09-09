import React, {useState} from 'react';
import {Pressable, ScrollView, StyleSheet, Text} from 'react-native';
import {api, AIFoodDraft, FoodEntry} from '../api/client';
import {AIFoodDraftView} from '../components/AIFoodDraftView';
import {foodPhotoPicker} from '../native/foodPhoto';

export function FoodPhotoScreen({accessToken,onBack,onDone}:{accessToken:string;onBack:()=>void;onDone:()=>void}) {
  const[draft,setDraft]=useState<AIFoodDraft|null>(null);const[busy,setBusy]=useState(false);const[error,setError]=useState('');
  async function pick(){try{setBusy(true);setError('');const dataUrl=await foodPhotoPicker.pickImage();setDraft(await api.aiParseFoodPhoto(accessToken,dataUrl));}catch(e){setError(e instanceof Error?e.message:'Не удалось обработать фото');}finally{setBusy(false)}}
  async function confirm(meal:FoodEntry['meal_type'],items:Array<{food_id:string;quantity_g:number}>){try{setBusy(true);await api.aiConfirmFood(accessToken,meal,items);onDone();}catch(e){setError(e instanceof Error?e.message:'Не удалось сохранить');}finally{setBusy(false)}}
  return <ScrollView contentContainerStyle={styles.container}><Pressable onPress={onBack}><Text style={styles.back}>← AI-ввод</Text></Pressable><Text style={styles.kicker}>AI VISION · FOOD</Text><Text style={styles.title}>Фото еды</Text><Text style={styles.subtitle}>Выбери фотографию блюда. AI определит продукты и примерные порции; перед сохранением всё можно проверить и изменить.</Text><Pressable onPress={()=>void pick()} disabled={busy} style={styles.primary}><Text style={styles.primaryText}>{busy?'Анализирую…':'Выбрать фото'}</Text></Pressable>{!foodPhotoPicker.available?<Text style={styles.note}>Нативный модуль выбора фото недоступен в этой сборке. Проверь Android Codegen/сборку приложения.</Text>:null}{draft?<AIFoodDraftView draft={draft} onConfirm={(meal,items)=>void confirm(meal,items)} busy={busy}/>:null}{error?<Text style={styles.error}>{error}</Text>:null}</ScrollView>;
}
const styles=StyleSheet.create({container:{padding:20,gap:14},back:{fontWeight:'850'},kicker:{fontSize:11,fontWeight:'900',letterSpacing:1.4,opacity:.5},title:{fontSize:30,fontWeight:'950'},subtitle:{fontSize:14,lineHeight:20,opacity:.62},primary:{backgroundColor:'#111',borderRadius:16,padding:15,alignItems:'center'},primaryText:{color:'#fff',fontWeight:'900'},note:{fontSize:11,lineHeight:16,opacity:.5},error:{fontSize:12,fontWeight:'750'}});
