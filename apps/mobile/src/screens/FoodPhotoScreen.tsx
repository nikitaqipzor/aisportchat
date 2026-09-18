import React, {useState} from 'react';
import {Image, Pressable, ScrollView, StyleSheet, Text, View} from 'react-native';
import {api, AIFoodDraft, FoodEntry} from '../api/client';
import {AIFoodDraftView} from '../components/AIFoodDraftView';
import {foodPhotoErrorMessage, foodPhotoPicker, isFoodPhotoPickCancelled} from '../native/foodPhoto';
import {AppButton} from '../components/AppButton';
import {colors, control, radius, spacing} from '../theme/tokens';

export function FoodPhotoScreen({accessToken,onBack,onDone}:{accessToken:string;onBack:()=>void;onDone:()=>void}) {
  const [draft,setDraft]=useState<AIFoodDraft|null>(null);
  const [selectedImage,setSelectedImage]=useState('');
  const [busy,setBusy]=useState(false);
  const [error,setError]=useState('');

  async function analyze(dataUrl:string) {
    if (busy) return;
    try {
      setBusy(true);
      setError('');
      setDraft(await api.aiParseFoodPhoto(accessToken,dataUrl));
    } catch (cause) {
      setError(foodPhotoErrorMessage(cause));
    } finally {
      setBusy(false);
    }
  }

  async function pick() {
    if (!foodPhotoPicker.available||busy) return;
    try {
      setError('');
      const dataUrl=await foodPhotoPicker.pickImage();
      setSelectedImage(dataUrl);
      setDraft(null);
      await analyze(dataUrl);
    } catch (cause) {
      if (!isFoodPhotoPickCancelled(cause)) setError(foodPhotoErrorMessage(cause));
    }
  }

  async function confirm(meal:FoodEntry['meal_type'],items:Array<{food_id:string;quantity_g:number}>) {
    try {
      setBusy(true);
      setError('');
      await api.aiConfirmFood(accessToken,meal,items);
      setSelectedImage('');
      onDone();
    } catch (cause) {
      setError(foodPhotoErrorMessage(cause));
    } finally {
      setBusy(false);
    }
  }

  return <ScrollView contentContainerStyle={styles.container} keyboardShouldPersistTaps="handled">
    <Pressable accessibilityRole="button" accessibilityLabel="Вернуться к AI-вводу" hitSlop={8} onPress={onBack} style={styles.navButton}><Text style={styles.back}>← AI-ввод</Text></Pressable>
    <Text style={styles.kicker}>AI VISION · FOOD</Text>
    <Text accessibilityRole="header" style={styles.title}>Фото еды</Text>
    <Text style={styles.subtitle}>Выберите чёткую фотографию блюда сверху. AI предложит состав и примерные порции, но перед записью их нужно проверить.</Text>
    <View style={styles.info}><Text style={styles.infoTitle}>Приватность и точность</Text><Text style={styles.note}>Фото отправляется защищённому AI-провайдеру только для распознавания и не сохраняется в дневнике. По снимку нельзя точно определить вес, калории или аллергены.</Text></View>
    {selectedImage?<View style={styles.previewCard}><Image accessibilityLabel="Выбранная фотография еды" source={{uri:selectedImage}} resizeMode="cover" style={styles.preview}/><Text style={styles.previewNote}>{busy?'Анализируем снимок…':'Проверьте, что блюдо хорошо видно.'}</Text></View>:null}
    <AppButton label={selectedImage?'Выбрать другое фото':'Выбрать фото'} loading={busy&&!draft} disabled={!foodPhotoPicker.available||busy} onPress={()=>void pick()}/>
    {!foodPhotoPicker.available?<View style={styles.unavailable}><Text style={styles.unavailableTitle}>Фото недоступно в этой сборке</Text><Text style={styles.note}>Вернитесь к текстовому вводу или обновите приложение. Ваши данные не потеряны.</Text></View>:null}
    {draft?<><View style={styles.draftHeader}><Text style={styles.draftTitle}>Проверьте распознавание</Text></View><AIFoodDraftView draft={draft} onConfirm={(meal,items)=>void confirm(meal,items)} onManualEntry={onBack} busy={busy}/></>:null}
    {error?<View style={styles.errorBox}><Text accessibilityRole="alert" style={styles.error}>{error}</Text>{selectedImage?<AppButton label="Повторить анализ" variant="secondary" disabled={busy} onPress={()=>void analyze(selectedImage)}/>:null}<AppButton label="Выбрать другое фото" variant="text" disabled={!foodPhotoPicker.available||busy} onPress={()=>void pick()}/><AppButton label="Описать еду текстом" variant="text" disabled={busy} onPress={onBack}/></View>:null}
  </ScrollView>;
}

const styles=StyleSheet.create({container:{padding:spacing.lg,gap:spacing.md,backgroundColor:colors.background},navButton:{minHeight:control.minTouch,justifyContent:'center',alignSelf:'flex-start'},back:{fontWeight:'800',color:colors.text},kicker:{fontSize:11,fontWeight:'900',letterSpacing:1.4,color:colors.textMuted},title:{fontSize:30,fontWeight:'900',color:colors.text},subtitle:{fontSize:14,lineHeight:20,color:colors.textMuted},info:{backgroundColor:colors.surfaceMuted,borderRadius:radius.md,padding:spacing.md,gap:spacing.xs},infoTitle:{fontWeight:'900',color:colors.text},note:{fontSize:12,lineHeight:18,color:colors.textMuted},previewCard:{borderWidth:1,borderColor:colors.border,borderRadius:radius.md,overflow:'hidden'},preview:{width:'100%',height:220,backgroundColor:colors.surfaceMuted},previewNote:{fontSize:12,lineHeight:18,color:colors.textMuted,padding:spacing.sm},unavailable:{borderWidth:1,borderColor:colors.border,borderRadius:radius.md,padding:spacing.md,gap:spacing.xs},unavailableTitle:{fontWeight:'900',color:colors.text},draftHeader:{flexDirection:'row',alignItems:'center',justifyContent:'space-between',gap:spacing.sm},draftTitle:{fontSize:18,fontWeight:'900',color:colors.text,flex:1},errorBox:{borderWidth:1,borderColor:colors.danger,borderRadius:radius.md,padding:spacing.md,gap:spacing.sm},error:{fontSize:12,lineHeight:18,fontWeight:'700',color:colors.danger}});
