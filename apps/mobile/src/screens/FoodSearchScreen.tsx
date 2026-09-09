import React, {useEffect, useMemo, useState} from 'react';
import {ActivityIndicator, Pressable, ScrollView, StyleSheet, Text, TextInput, View} from 'react-native';
import {api, FoodEntry, FoodItem, NutritionDay} from '../api/client';
import {AppButton} from '../components/AppButton';
import {colors, control, radius, spacing} from '../theme/tokens';

const meals: Array<{id: FoodEntry['meal_type']; title: string}> = [
  {id: 'breakfast', title: 'Завтрак'}, {id: 'lunch', title: 'Обед'},
  {id: 'dinner', title: 'Ужин'}, {id: 'snack', title: 'Перекус'},
];
const parseAmount = (value: string) => Number(value.replace(',', '.'));

export function FoodSearchScreen({accessToken,onBack,onLogged,onCustom,onRecipes}:{accessToken:string;onBack:()=>void;onLogged:(day:NutritionDay)=>void;onCustom:()=>void;onRecipes:()=>void}) {
  const [query,setQuery]=useState('');
  const [barcode,setBarcode]=useState('');
  const [items,setItems]=useState<FoodItem[]>([]);
  const [selected,setSelected]=useState<FoodItem|null>(null);
  const [meal,setMeal]=useState<FoodEntry['meal_type']>('lunch');
  const [grams,setGrams]=useState('200');
  const [error,setError]=useState('');
  const [loading,setLoading]=useState(false);
  const [adding,setAdding]=useState(false);
  const [hasSearched,setHasSearched]=useState(false);

  useEffect(()=>{void search('')},[accessToken]);

  async function search(value=query){
    try { setLoading(true); setError(''); setHasSearched(true); setItems((await api.searchFoods(accessToken,value.trim())).items); }
    catch(e){setError(e instanceof Error?e.message:'Не удалось выполнить поиск');}
    finally{setLoading(false);}
  }
  async function lookupBarcode(){
    const value=barcode.replace(/\s/g,'');
    if(!/^\d{8,14}$/.test(value)){setError('Введите от 8 до 14 цифр штрихкода.');return;}
    try { setLoading(true); setError(''); const food=await api.foodByBarcode(accessToken,value); setSelected(food); setItems([food]); setGrams(String(Math.round(food.serving_g||100))); }
    catch(e){setError(e instanceof Error?e.message:'Продукт с таким штрихкодом не найден');}
    finally{setLoading(false);}
  }
  const amount=parseAmount(grams);
  const validAmount=Number.isFinite(amount)&&amount>0&&amount<=5000;
  const preview=useMemo(()=>{if(!selected||!validAmount)return null;const factor=amount/100;return{kcal:Math.round(selected.kcal_per_100g*factor),p:Math.round(selected.protein_per_100g*factor),f:Math.round(selected.fat_per_100g*factor),c:Math.round(selected.carbs_per_100g*factor)}},[selected,amount,validAmount]);
  async function add(){
    if(!selected||!validAmount||adding)return;
    try { setAdding(true); setError(''); onLogged(await api.logFood(accessToken,{food_id:selected.id,meal_type:meal,quantity_g:amount})); }
    catch(e){setError(e instanceof Error?e.message:'Не удалось добавить продукт');}
    finally{setAdding(false);}
  }

  return <ScrollView contentContainerStyle={styles.container} keyboardShouldPersistTaps="handled">
    <Pressable accessibilityRole="button" accessibilityLabel="Вернуться к питанию" hitSlop={8} onPress={onBack} style={styles.navButton}><Text style={styles.back}>← Питание</Text></Pressable>
    <Text style={styles.kicker}>ДОБАВИТЬ ЕДУ</Text><Text accessibilityRole="header" style={styles.title}>Что добавить?</Text>
    <Text style={styles.subtitle}>Найдите продукт в каталоге или создайте свой. Значения указаны на 100 г.</Text>
    <View style={styles.actions}><AppButton label="Свой продукт" variant="secondary" onPress={onCustom} style={styles.action}/><AppButton label="Рецепты" variant="secondary" onPress={onRecipes} style={styles.action}/></View>
    <View style={styles.searchRow}><TextInput accessibilityLabel="Поиск продукта" value={query} onChangeText={setQuery} onSubmitEditing={()=>void search()} returnKeyType="search" placeholder="Творог, рис, курица…" placeholderTextColor={colors.textMuted} style={styles.input}/><AppButton label="Найти" onPress={()=>void search()} disabled={loading}/></View>
    <View style={styles.searchRow}><TextInput accessibilityLabel="Штрихкод продукта" value={barcode} onChangeText={setBarcode} onSubmitEditing={()=>void lookupBarcode()} returnKeyType="search" keyboardType="number-pad" maxLength={14} placeholder="Штрихкод" placeholderTextColor={colors.textMuted} style={styles.input}/><AppButton label="Проверить" variant="secondary" onPress={()=>void lookupBarcode()} disabled={loading||!barcode.trim()}/></View>
    <Text style={styles.label}>Приём пищи</Text><View style={styles.meals} accessibilityRole="radiogroup">{meals.map(x=><Pressable key={x.id} accessibilityRole="radio" accessibilityState={{selected:meal===x.id}} onPress={()=>setMeal(x.id)} style={[styles.meal,meal===x.id&&styles.mealActive]}><Text style={[styles.mealText,meal===x.id&&styles.white]}>{x.title}</Text></Pressable>)}</View>
    {loading?<View style={styles.stateRow} accessibilityLiveRegion="polite"><ActivityIndicator color={colors.text}/><Text style={styles.muted}>Ищем продукты…</Text></View>:null}
    {!loading&&hasSearched&&items.length===0&&!error?<View style={styles.emptyCard}><Text style={styles.emptyTitle}>Ничего не найдено</Text><Text style={styles.muted}>Проверьте запрос или добавьте продукт вручную.</Text><AppButton label="Создать продукт" variant="secondary" onPress={onCustom}/></View>:null}
    <View style={styles.list} accessibilityRole="radiogroup">{items.map(item=><Pressable key={item.id} accessibilityRole="radio" accessibilityState={{selected:selected?.id===item.id}} accessibilityLabel={`${item.name}, ${Math.round(item.kcal_per_100g)} килокалорий на 100 граммов`} onPress={()=>{setSelected(item);setGrams(String(Math.round(item.serving_g||100)));setError('')}} style={[styles.food,selected?.id===item.id&&styles.foodSelected]}><View style={styles.flex}><Text style={[styles.foodTitle,selected?.id===item.id&&styles.white]}>{item.name}</Text><Text style={[styles.foodMeta,selected?.id===item.id&&styles.whiteSoft]}>{item.source==='custom'?'Мой продукт · ':''}{Math.round(item.kcal_per_100g)} kcal · Б {item.protein_per_100g} · Ж {item.fat_per_100g} · У {item.carbs_per_100g}</Text></View></Pressable>)}</View>
    {selected?<View style={styles.editor}><Text style={styles.editorTitle}>{selected.name}</Text><Text style={styles.label}>Количество, г</Text><TextInput accessibilityLabel="Количество продукта в граммах" value={grams} onChangeText={setGrams} keyboardType="decimal-pad" maxLength={7} style={[styles.grams,!validAmount&&styles.inputInvalid]}/>{!validAmount?<Text style={styles.fieldError}>Введите значение от 1 до 5000 г.</Text>:null}<Text accessibilityLiveRegion="polite" style={styles.preview}>{preview?`${preview.kcal} kcal · Б ${preview.p} · Ж ${preview.f} · У ${preview.c}`:'—'}</Text><AppButton label="Добавить в дневник" loading={adding} disabled={!validAmount} onPress={()=>void add()}/></View>:null}
    {error?<View style={styles.errorCard}><Text accessibilityRole="alert" style={styles.error}>{error}</Text><AppButton label="Повторить поиск" variant="text" onPress={()=>void search()}/></View>:null}
  </ScrollView>;
}

const styles=StyleSheet.create({
  container:{padding:spacing.lg,gap:spacing.md,backgroundColor:colors.background},navButton:{minHeight:control.minTouch,justifyContent:'center',alignSelf:'flex-start'},back:{fontWeight:'800',color:colors.text},kicker:{fontSize:11,fontWeight:'900',letterSpacing:1.3,color:colors.textMuted},title:{fontSize:30,fontWeight:'900',color:colors.text},subtitle:{fontSize:14,lineHeight:20,color:colors.textMuted},actions:{flexDirection:'row',gap:spacing.sm},action:{flex:1},searchRow:{flexDirection:'row',gap:spacing.sm,alignItems:'stretch'},input:{flex:1,minHeight:control.minTouch,borderWidth:1,borderColor:colors.border,borderRadius:radius.md,paddingHorizontal:13,fontSize:15,color:colors.text,backgroundColor:colors.surface},label:{fontSize:12,fontWeight:'800',color:colors.textMuted},meals:{flexDirection:'row',flexWrap:'wrap',gap:spacing.sm},meal:{minHeight:control.minTouch,borderWidth:1,borderColor:colors.border,borderRadius:radius.pill,paddingHorizontal:12,justifyContent:'center'},mealActive:{backgroundColor:colors.primary,borderColor:colors.primary},mealText:{fontSize:12,fontWeight:'800',color:colors.text},white:{color:colors.inverse},whiteSoft:{color:colors.inverse,opacity:.78},stateRow:{minHeight:control.minTouch,flexDirection:'row',alignItems:'center',gap:spacing.sm},muted:{color:colors.textMuted,lineHeight:19},emptyCard:{borderWidth:1,borderColor:colors.border,borderRadius:radius.lg,padding:spacing.md,gap:spacing.sm},emptyTitle:{fontSize:16,fontWeight:'900',color:colors.text},list:{gap:spacing.sm},food:{minHeight:control.minTouch,borderWidth:1,borderColor:colors.border,borderRadius:radius.md,padding:13,flexDirection:'row'},foodSelected:{backgroundColor:colors.primary,borderColor:colors.primary},flex:{flex:1},foodTitle:{fontSize:15,fontWeight:'800',color:colors.text},foodMeta:{fontSize:11,lineHeight:16,color:colors.textMuted,marginTop:4},editor:{borderWidth:1,borderColor:colors.border,borderRadius:radius.lg,padding:spacing.md,gap:spacing.sm},editorTitle:{fontSize:18,fontWeight:'900',color:colors.text},grams:{minHeight:control.minTouch,borderWidth:1,borderColor:colors.border,borderRadius:radius.md,paddingHorizontal:12,fontSize:18,fontWeight:'800',color:colors.text},inputInvalid:{borderColor:colors.danger},fieldError:{fontSize:12,color:colors.danger},preview:{fontWeight:'800',color:colors.text},errorCard:{borderWidth:1,borderColor:colors.danger,borderRadius:radius.md,padding:spacing.sm,gap:spacing.xs},error:{fontSize:13,fontWeight:'700',color:colors.danger},
});
