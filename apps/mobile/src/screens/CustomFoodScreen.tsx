import React, {useState} from 'react';
import {Pressable, ScrollView, StyleSheet, Text, TextInput, View} from 'react-native';
import {api, FoodItem} from '../api/client';

const fields = [
  ['kcal','Ккал / 100 г'], ['protein','Белки / 100 г'], ['fat','Жиры / 100 г'], ['carbs','Углеводы / 100 г'], ['fiber','Клетчатка / 100 г'], ['serving','Порция, г'],
] as const;

export function CustomFoodScreen({accessToken,onBack,onSaved}:{accessToken:string;onBack:()=>void;onSaved:(food:FoodItem)=>void}) {
  const [name,setName]=useState(''); const [brand,setBrand]=useState(''); const [barcode,setBarcode]=useState('');
  const [values,setValues]=useState<Record<string,string>>({kcal:'',protein:'',fat:'',carbs:'',fiber:'0',serving:'100'}); const [error,setError]=useState(''); const [saving,setSaving]=useState(false);
  const num=(key:string)=>Number((values[key]||'0').replace(',','.'));
  async function save(){try{setSaving(true);setError('');const food=await api.createCustomFood(accessToken,{name,brand,barcode,kcal_per_100g:num('kcal'),protein_per_100g:num('protein'),fat_per_100g:num('fat'),carbs_per_100g:num('carbs'),fiber_per_100g:num('fiber'),serving_g:num('serving')});onSaved(food);}catch(e){setError(e instanceof Error?e.message:'Не удалось сохранить продукт');}finally{setSaving(false)}}
  return <ScrollView contentContainerStyle={styles.container} keyboardShouldPersistTaps="handled">
    <Pressable onPress={onBack}><Text style={styles.back}>← Назад</Text></Pressable><Text style={styles.kicker}>СВОЙ ПРОДУКТ</Text><Text style={styles.title}>Добавить продукт</Text>
    <Text style={styles.label}>Название *</Text><TextInput value={name} onChangeText={setName} placeholder="Например, домашние сырники" style={styles.input}/>
    <Text style={styles.label}>Бренд</Text><TextInput value={brand} onChangeText={setBrand} placeholder="Необязательно" style={styles.input}/>
    <Text style={styles.label}>Штрихкод</Text><TextInput value={barcode} onChangeText={setBarcode} keyboardType="number-pad" placeholder="Необязательно" style={styles.input}/>
    <View style={styles.grid}>{fields.map(([key,title])=><View key={key} style={styles.field}><Text style={styles.label}>{title}</Text><TextInput value={values[key]} onChangeText={v=>setValues(current=>({...current,[key]:v}))} keyboardType="decimal-pad" style={styles.input}/></View>)}</View>
    <Text style={styles.hint}>Значения сохраняются как факты продукта. AI позже сможет предложить распознавание, но не будет незаметно менять введённые тобой БЖУ.</Text>
    <Pressable disabled={saving} onPress={()=>void save()} style={styles.primary}><Text style={styles.primaryText}>{saving?'Сохраняю…':'Сохранить продукт'}</Text></Pressable>{error?<Text style={styles.error}>{error}</Text>:null}
  </ScrollView>;
}
const styles=StyleSheet.create({container:{padding:20,gap:11},back:{fontWeight:'800'},kicker:{fontSize:11,fontWeight:'900',letterSpacing:1.3,opacity:.5},title:{fontSize:30,fontWeight:'900'},label:{fontSize:12,fontWeight:'800',opacity:.6},input:{borderWidth:1,borderRadius:13,paddingHorizontal:12,paddingVertical:11,fontWeight:'750'},grid:{flexDirection:'row',flexWrap:'wrap',gap:10},field:{width:'48%',gap:5},hint:{fontSize:12,lineHeight:18,opacity:.55},primary:{backgroundColor:'#111',borderRadius:16,paddingVertical:15,alignItems:'center',marginTop:5},primaryText:{color:'#fff',fontWeight:'900'},error:{fontWeight:'700',fontSize:12}});
