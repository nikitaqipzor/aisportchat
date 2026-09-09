import React, {useMemo, useState} from 'react';
import {Pressable, StyleSheet, Text, TextInput, View} from 'react-native';
import {AIFoodDraft, FoodEntry} from '../api/client';

const meals: Array<{id: FoodEntry['meal_type']; title: string}> = [
  {id:'breakfast',title:'Завтрак'},{id:'lunch',title:'Обед'},{id:'dinner',title:'Ужин'},{id:'snack',title:'Перекус'},
];

export function AIFoodDraftView({draft,onConfirm,busy=false}:{draft:AIFoodDraft;onConfirm:(meal:FoodEntry['meal_type'],items:Array<{food_id:string;quantity_g:number}>)=>void;busy?:boolean}) {
  const [meal,setMeal]=useState<FoodEntry['meal_type']>('snack');
  const [quantities,setQuantities]=useState<Record<number,string>>(()=>Object.fromEntries(draft.items.map((x,i)=>[i,String(Math.round(x.quantity_g))])));
  const confirmable=useMemo(()=>draft.items.map((item,i)=>({item,i})).filter(x=>Boolean(x.item.matched_food)),[draft]);
  return <View style={styles.wrap}>
    {draft.warning?<Text style={styles.warning}>{draft.warning}</Text>:null}
    {draft.items.map((item,i)=><View key={`${item.name}-${i}`} style={styles.card}>
      <View style={styles.row}><Text style={styles.name}>{item.name}</Text><Text style={styles.conf}>{Math.round(item.confidence*100)}%</Text></View>
      {item.matched_food?<><Text style={styles.match}>↳ {item.matched_food.name}</Text><TextInput keyboardType="decimal-pad" value={quantities[i]??''} onChangeText={v=>setQuantities(q=>({...q,[i]:v}))} style={styles.input} placeholder="граммы"/><Text style={styles.macros}>≈ {Math.round(item.estimated_calories??0)} kcal · Б {Math.round(item.estimated_protein_g??0)} · Ж {Math.round(item.estimated_fat_g??0)} · У {Math.round(item.estimated_carbs_g??0)}</Text></>:<Text style={styles.unmatched}>Нет точного продукта в каталоге — добавь его вручную.</Text>}
      {item.notes?<Text style={styles.notes}>{item.notes}</Text>:null}
    </View>)}
    <Text style={styles.label}>Приём пищи</Text><View style={styles.meals}>{meals.map(x=><Pressable key={x.id} onPress={()=>setMeal(x.id)} style={[styles.meal,meal===x.id&&styles.mealActive]}><Text style={[styles.mealText,meal===x.id&&styles.mealTextActive]}>{x.title}</Text></Pressable>)}</View>
    <Pressable disabled={busy||confirmable.length===0} onPress={()=>onConfirm(meal,confirmable.map(({item,i})=>({food_id:item.matched_food!.id,quantity_g:Number((quantities[i]??'').replace(',','.'))||item.quantity_g})))} style={[styles.primary,(busy||confirmable.length===0)&&styles.disabled]}><Text style={styles.primaryText}>{busy?'Сохраняю…':`Добавить ${confirmable.length} поз.`}</Text></Pressable>
  </View>;
}
const styles=StyleSheet.create({wrap:{gap:10},warning:{fontSize:12,lineHeight:17,opacity:.65},card:{borderWidth:1,borderRadius:17,padding:13,gap:6},row:{flexDirection:'row',justifyContent:'space-between',gap:10},name:{fontSize:16,fontWeight:'900',flex:1},conf:{fontWeight:'800',opacity:.45},match:{fontWeight:'750',opacity:.65},input:{borderWidth:1,borderRadius:12,paddingHorizontal:12,paddingVertical:9,fontWeight:'800'},macros:{fontSize:12,fontWeight:'750',opacity:.6},unmatched:{fontSize:12,fontWeight:'700',opacity:.55},notes:{fontSize:11,opacity:.5},label:{fontSize:13,fontWeight:'900'},meals:{flexDirection:'row',flexWrap:'wrap',gap:6},meal:{borderWidth:1,borderRadius:99,paddingHorizontal:10,paddingVertical:7},mealActive:{backgroundColor:'#111'},mealText:{fontSize:11,fontWeight:'800'},mealTextActive:{color:'#fff'},primary:{backgroundColor:'#111',borderRadius:16,paddingVertical:14,alignItems:'center'},disabled:{opacity:.35},primaryText:{color:'#fff',fontWeight:'900'}});
