import React, {useEffect, useMemo, useState} from 'react';
import {Pressable, StyleSheet, Text, TextInput, View} from 'react-native';
import {AIFoodDraft, FoodEntry} from '../api/client';
import {catalogMacrosForQuantity, parseFoodQuantity} from '../domain/nutrition';

const meals: Array<{id: FoodEntry['meal_type']; title: string}> = [
  {id:'breakfast',title:'Завтрак'},{id:'lunch',title:'Обед'},{id:'dinner',title:'Ужин'},{id:'snack',title:'Перекус'},
];

export function AIFoodDraftView({draft,onConfirm,onManualEntry,busy=false}:{draft:AIFoodDraft;onConfirm:(meal:FoodEntry['meal_type'],items:Array<{food_id:string;quantity_g:number}>)=>void;onManualEntry?:()=>void;busy?:boolean}) {
  const [meal,setMeal]=useState<FoodEntry['meal_type']>('snack');
  const [quantities,setQuantities]=useState<Record<number,string>>(()=>Object.fromEntries(draft.items.map((x,i)=>[i,String(Math.round(x.quantity_g))])));
  const [included,setIncluded]=useState<Record<number,boolean>>(()=>Object.fromEntries(draft.items.map((item,i)=>[i,Boolean(item.matched_food)])));
  useEffect(() => {
    setQuantities(Object.fromEntries(draft.items.map((item, index) => [index, String(Math.round(item.quantity_g))])));
    setIncluded(Object.fromEntries(draft.items.map((item,index)=>[index,Boolean(item.matched_food)])));
  }, [draft]);
  const confirmable=useMemo(()=>draft.items.map((item,i)=>({item,i})).filter(x=>Boolean(x.item.matched_food)&&included[x.i]),[draft,included]);
  const parsedQuantities=confirmable.map(({i})=>parseFoodQuantity(quantities[i]??''));
  const hasInvalidQuantity=parsedQuantities.some(result=>result.value === null);
  return <View style={styles.wrap}>
    {draft.warning?<Text style={styles.warning}>{draft.warning}</Text>:null}
    {draft.items.map((item,i)=><View key={`${item.name}-${i}`} style={styles.card}>
      <View style={styles.row}><Text style={styles.name}>{item.name}</Text><Text style={styles.conf}>{Math.round(item.confidence*100)}%</Text></View>
      {item.matched_food?(()=>{const parsed=parseFoodQuantity(quantities[i]??'');const macros=parsed.value===null?null:catalogMacrosForQuantity(item,parsed.value);return <><Text style={styles.match}>↳ {item.matched_food.name}</Text><TextInput accessibilityLabel={`Вес ${item.matched_food.name} в граммах`} keyboardType="decimal-pad" value={quantities[i]??''} onChangeText={v=>setQuantities(q=>({...q,[i]:v}))} style={[styles.input,parsed.error&&styles.inputError]} placeholder="граммы"/>{parsed.error?<Text accessibilityRole="alert" style={styles.error}>{parsed.error}</Text>:null}{macros?<Text style={styles.macros}>≈ {Math.round(macros.calories)} kcal · Б {Math.round(macros.proteinG)} · Ж {Math.round(macros.fatG)} · У {Math.round(macros.carbsG)}</Text>:null}</>})():<Text style={styles.unmatched}>Нет точного продукта в каталоге — добавь его вручную.</Text>}
      {item.matched_food?<Pressable accessibilityRole="checkbox" accessibilityState={{checked:Boolean(included[i])}} onPress={()=>setIncluded(current=>({...current,[i]:!current[i]}))} style={styles.include}><Text style={styles.includeText}>{included[i]?'✓ Добавить в дневник':'Не добавлять'}</Text></Pressable>:null}
      {item.notes?<Text style={styles.notes}>{item.notes}</Text>:null}
    </View>)}
    {onManualEntry?<Pressable accessibilityRole="button" onPress={onManualEntry} style={styles.manual}><Text style={styles.manualText}>Исправить состав текстом</Text></Pressable>:null}
    <Text style={styles.label}>Приём пищи</Text><View style={styles.meals}>{meals.map(x=><Pressable accessibilityRole="radio" accessibilityState={{selected:meal===x.id}} key={x.id} onPress={()=>setMeal(x.id)} style={[styles.meal,meal===x.id&&styles.mealActive]}><Text style={[styles.mealText,meal===x.id&&styles.mealTextActive]}>{x.title}</Text></Pressable>)}</View>
    <Pressable accessibilityRole="button" accessibilityState={{disabled:busy||confirmable.length===0||hasInvalidQuantity,busy}} disabled={busy||confirmable.length===0||hasInvalidQuantity} onPress={()=>onConfirm(meal,confirmable.map(({item},index)=>({food_id:item.matched_food!.id,quantity_g:parsedQuantities[index].value!})))} style={[styles.primary,(busy||confirmable.length===0||hasInvalidQuantity)&&styles.disabled]}><Text style={styles.primaryText}>{busy?'Сохраняю…':`Добавить ${confirmable.length} поз.`}</Text></Pressable>
  </View>;
}
const styles=StyleSheet.create({wrap:{gap:10},warning:{fontSize:12,lineHeight:17,opacity:.65},card:{borderWidth:1,borderRadius:17,padding:13,gap:6},row:{flexDirection:'row',justifyContent:'space-between',gap:10},name:{fontSize:16,fontWeight:'900',flex:1},conf:{fontWeight:'800',opacity:.45},match:{fontWeight:'700',opacity:.65},input:{borderWidth:1,borderRadius:12,paddingHorizontal:12,paddingVertical:9,fontWeight:'800'},inputError:{borderColor:'#a12626'},error:{fontSize:11,fontWeight:'700',color:'#a12626'},macros:{fontSize:12,fontWeight:'700',opacity:.6},unmatched:{fontSize:12,fontWeight:'700',opacity:.55},notes:{fontSize:11,opacity:.5},include:{minHeight:44,justifyContent:'center',alignSelf:'flex-start'},includeText:{fontSize:12,fontWeight:'800'},manual:{minHeight:44,justifyContent:'center',alignItems:'center',borderWidth:1,borderRadius:14},manualText:{fontWeight:'800'},label:{fontSize:13,fontWeight:'900'},meals:{flexDirection:'row',flexWrap:'wrap',gap:6},meal:{minHeight:44,justifyContent:'center',borderWidth:1,borderRadius:99,paddingHorizontal:12,paddingVertical:7},mealActive:{backgroundColor:'#111'},mealText:{fontSize:11,fontWeight:'800'},mealTextActive:{color:'#fff'},primary:{minHeight:48,justifyContent:'center',backgroundColor:'#111',borderRadius:16,paddingVertical:14,alignItems:'center'},disabled:{opacity:.35},primaryText:{color:'#fff',fontWeight:'900'}});
