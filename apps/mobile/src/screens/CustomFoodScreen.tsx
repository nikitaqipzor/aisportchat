import React, {useMemo, useState} from 'react';
import {Pressable, ScrollView, StyleSheet, Text, TextInput, View} from 'react-native';
import {api, FoodItem} from '../api/client';
import {AppButton} from '../components/AppButton';
import {colors, control, radius, spacing} from '../theme/tokens';

const fields = [
  ['kcal','Ккал / 100 г'], ['protein','Белки / 100 г'], ['fat','Жиры / 100 г'],
  ['carbs','Углеводы / 100 г'], ['fiber','Клетчатка / 100 г'], ['serving','Порция, г'],
] as const;
const parseNumber=(value:string)=>Number(value.replace(',','.'));

export function CustomFoodScreen({accessToken,onBack,onSaved}:{accessToken:string;onBack:()=>void;onSaved:(food:FoodItem)=>void}) {
  const [name,setName]=useState(''); const [brand,setBrand]=useState(''); const [barcode,setBarcode]=useState('');
  const [values,setValues]=useState<Record<string,string>>({kcal:'',protein:'',fat:'',carbs:'',fiber:'0',serving:'100'});
  const [error,setError]=useState(''); const [saving,setSaving]=useState(false); const [submitted,setSubmitted]=useState(false);
  const validation=useMemo(()=>{
    const errors:Record<string,string>={};
    if(name.trim().length<2)errors.name='Введите название не короче 2 символов.';
    if(barcode.trim()&&!/^\d{8,14}$/.test(barcode.trim()))errors.barcode='Штрихкод должен содержать 8–14 цифр.';
    fields.forEach(([key])=>{const n=parseNumber(values[key]??'');const max=key==='kcal'?2000:5000;if(!Number.isFinite(n)||n<0||n>max)errors[key]=`Введите число от 0 до ${max}.`;});
    if(parseNumber(values.serving)<=0)errors.serving='Порция должна быть больше 0 г.';
    return errors;
  },[barcode,name,values]);
  const canSave=Object.keys(validation).length===0&&!saving;
  async function save(){
    setSubmitted(true); if(!canSave)return;
    try{setSaving(true);setError('');const food=await api.createCustomFood(accessToken,{name:name.trim(),brand:brand.trim(),barcode:barcode.trim(),kcal_per_100g:parseNumber(values.kcal),protein_per_100g:parseNumber(values.protein),fat_per_100g:parseNumber(values.fat),carbs_per_100g:parseNumber(values.carbs),fiber_per_100g:parseNumber(values.fiber),serving_g:parseNumber(values.serving)});onSaved(food);}
    catch(e){setError(e instanceof Error?e.message:'Не удалось сохранить продукт');}
    finally{setSaving(false);}
  }
  const fieldError=(key:string)=>submitted?validation[key]:undefined;
  return <ScrollView contentContainerStyle={styles.container} keyboardShouldPersistTaps="handled">
    <Pressable accessibilityRole="button" accessibilityLabel="Вернуться к поиску" hitSlop={8} onPress={onBack} style={styles.navButton}><Text style={styles.back}>← Поиск</Text></Pressable>
    <Text style={styles.kicker}>СВОЙ ПРОДУКТ</Text><Text accessibilityRole="header" style={styles.title}>Добавить продукт</Text>
    <Text style={styles.subtitle}>Перенесите пищевую ценность с упаковки. Мы не будем заменять введённые значения оценкой AI.</Text>
    <Text style={styles.label}>Название *</Text><TextInput accessibilityLabel="Название продукта" value={name} onChangeText={setName} maxLength={100} placeholder="Например, домашние сырники" placeholderTextColor={colors.textMuted} style={[styles.input,fieldError('name')&&styles.invalid]}/>{fieldError('name')?<Text style={styles.fieldError}>{fieldError('name')}</Text>:null}
    <Text style={styles.label}>Бренд</Text><TextInput accessibilityLabel="Бренд продукта" value={brand} onChangeText={setBrand} maxLength={100} placeholder="Необязательно" placeholderTextColor={colors.textMuted} style={styles.input}/>
    <Text style={styles.label}>Штрихкод</Text><TextInput accessibilityLabel="Штрихкод продукта" value={barcode} onChangeText={setBarcode} keyboardType="number-pad" maxLength={14} placeholder="8–14 цифр, необязательно" placeholderTextColor={colors.textMuted} style={[styles.input,fieldError('barcode')&&styles.invalid]}/>{fieldError('barcode')?<Text style={styles.fieldError}>{fieldError('barcode')}</Text>:null}
    <View style={styles.grid}>{fields.map(([key,title])=><View key={key} style={styles.field}><Text style={styles.label}>{title}</Text><TextInput accessibilityLabel={title} value={values[key]} onChangeText={v=>setValues(current=>({...current,[key]:v}))} keyboardType="decimal-pad" maxLength={8} style={[styles.input,fieldError(key)&&styles.invalid]}/>{fieldError(key)?<Text style={styles.fieldError}>{fieldError(key)}</Text>:null}</View>)}</View>
    <View style={styles.info}><Text style={styles.infoTitle}>Проверьте этикетку</Text><Text style={styles.hint}>Калории и БЖУ обычно указываются на 100 г, а не на одну порцию. Эти данные будут использоваться в дневнике как введённые вами факты.</Text></View>
    <AppButton label="Сохранить продукт" loading={saving} disabled={submitted&&!canSave} onPress={()=>void save()}/>
    {error?<Text accessibilityRole="alert" style={styles.error}>{error}</Text>:null}
  </ScrollView>;
}
const styles=StyleSheet.create({container:{padding:spacing.lg,gap:spacing.sm,backgroundColor:colors.background},navButton:{minHeight:control.minTouch,justifyContent:'center',alignSelf:'flex-start'},back:{fontWeight:'800',color:colors.text},kicker:{fontSize:11,fontWeight:'900',letterSpacing:1.3,color:colors.textMuted},title:{fontSize:30,fontWeight:'900',color:colors.text},subtitle:{fontSize:14,lineHeight:20,color:colors.textMuted,marginBottom:spacing.xs},label:{fontSize:12,fontWeight:'800',color:colors.textMuted},input:{minHeight:control.minTouch,borderWidth:1,borderColor:colors.border,borderRadius:radius.md,paddingHorizontal:12,fontWeight:'700',color:colors.text,backgroundColor:colors.surface},invalid:{borderColor:colors.danger},fieldError:{fontSize:11,lineHeight:16,color:colors.danger},grid:{flexDirection:'row',flexWrap:'wrap',gap:spacing.sm},field:{width:'48%',flexGrow:1,gap:spacing.xs},info:{backgroundColor:colors.surfaceMuted,borderRadius:radius.md,padding:spacing.md,gap:spacing.xs,marginVertical:spacing.xs},infoTitle:{fontWeight:'900',color:colors.text},hint:{fontSize:12,lineHeight:18,color:colors.textMuted},error:{fontWeight:'700',fontSize:12,color:colors.danger}});
