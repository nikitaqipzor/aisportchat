import React, {useEffect, useRef, useState} from 'react';
import {Pressable, ScrollView, StyleSheet, Text, TextInput, View} from 'react-native';
import {api, AIChatMessage, AIStatus} from '../api/client';
import {AppButton} from '../components/AppButton';
import {sessionStorage} from '../storage/session';
import {colors, control, radius, spacing} from '../theme/tokens';

type FailedRequest = {text: string; history: AIChatMessage[]};

export function AICoachScreen({accessToken,onBack,onWeekly}:{accessToken:string;onBack:()=>void;onWeekly:()=>void}) {
  const [messages,setMessages]=useState<AIChatMessage[]>([]);
  const [input,setInput]=useState('');
  const [status,setStatus]=useState<AIStatus|null>(null);
  const [busy,setBusy]=useState(false);
  const [error,setError]=useState('');
  const [failed,setFailed]=useState<FailedRequest|null>(null);
  const controllerRef=useRef<AbortController|null>(null);
  const scrollRef=useRef<ScrollView|null>(null);

  useEffect(()=>{
    api.aiStatus(accessToken).then(setStatus).catch(()=>undefined);
    sessionStorage.loadAIChat().then(setMessages).catch(()=>undefined);
    return ()=>controllerRef.current?.abort();
  },[accessToken]);

  useEffect(()=>{
    const timer=setTimeout(()=>scrollRef.current?.scrollToEnd({animated:true}),30);
    return ()=>clearTimeout(timer);
  },[messages,busy,error]);

  async function requestAssistant(text:string,history:AIChatMessage[],visibleMessages:AIChatMessage[]) {
    const controller=new AbortController();
    controllerRef.current=controller;
    try {
      setBusy(true);setError('');setFailed(null);
      const res=await api.aiChat(accessToken,text,history,controller.signal);
      const complete=[...visibleMessages,{role:'assistant' as const,content:res.message}];
      setMessages(complete);
      await sessionStorage.saveAIChat(complete);
      setStatus({provider:res.provider,model:res.model});
    } catch(e) {
      if (controller.signal.aborted) {
        setError('Ответ остановлен. Можно повторить запрос.');
      } else {
        setError(e instanceof Error?e.message:'AI Coach недоступен');
      }
      setFailed({text,history});
    } finally {
      if (controllerRef.current===controller) controllerRef.current=null;
      setBusy(false);
    }
  }

  async function send() {
    const text=input.trim();if(!text||busy)return;
    const history=messages.slice(-10);
    const withUser=[...messages,{role:'user' as const,content:text}];
    setMessages(withUser);await sessionStorage.saveAIChat(withUser);setInput('');
    await requestAssistant(text,history,withUser);
  }

  async function retry() {
    if (!failed||busy) return;
    await requestAssistant(failed.text,failed.history,messages);
  }

  function stop() { controllerRef.current?.abort(); }

  return <View style={styles.screen} testID="ai-coach-screen">
    <View style={styles.header}>
      <Pressable testID="ai-coach-back" accessibilityRole="button" accessibilityLabel="Вернуться на главную" hitSlop={10} onPress={onBack}><Text style={styles.back}>← Главная</Text></Pressable>
      <Pressable testID="ai-coach-weekly-report" accessibilityRole="button" accessibilityLabel="Открыть недельный отчёт" hitSlop={10} onPress={onWeekly}><Text style={styles.report}>Отчёт недели</Text></Pressable>
    </View>
    <View style={styles.hero}><Text style={styles.kicker}>AI COACH</Text><Text accessibilityRole="header" style={styles.title}>Персональный тренер</Text><Text style={styles.status}>{status?`${status.provider} · ${status.model}`:'подключаю…'}</Text></View>
    <ScrollView ref={scrollRef} style={styles.chat} contentContainerStyle={styles.chatInner} keyboardShouldPersistTaps="handled">
      {messages.length===0?<View style={styles.hint}><Text style={styles.hintTitle}>Можно спросить</Text><Text style={styles.hintText}>«Сколько белка осталось?» · «Что я тренировал недавно?» · «Почему readiness сегодня ниже?»</Text></View>:messages.map((m,i)=><View key={`${m.role}-${i}`} accessibilityLabel={`${m.role==='user'?'Вы':'AI Coach'}: ${m.content}`} style={[styles.bubble,m.role==='user'?styles.user:styles.ai]}><Text style={[styles.bubbleText,m.role==='user'&&styles.userText]}>{m.content}</Text></View>)}
      {busy?<View style={styles.busyRow}><Text style={styles.typing}>AI анализирует твои данные…</Text><AppButton label="Остановить" variant="text" testID="ai-coach-stop" onPress={stop}/></View>:null}
      {error?<View style={styles.errorBox}><Text accessibilityRole="alert" style={styles.error}>{error}</Text>{failed&&!busy?<AppButton label="Повторить" variant="secondary" testID="ai-coach-retry" onPress={()=>void retry()}/>:null}</View>:null}
    </ScrollView>
    <View style={styles.composer}>
      <TextInput accessibilityLabel="Сообщение AI Coach" testID="ai-coach-input" value={input} onChangeText={setInput} placeholder="Спроси AI Coach…" placeholderTextColor={colors.textMuted} style={styles.input} multiline editable={!busy}/>
      <Pressable testID="ai-coach-send" accessibilityRole="button" accessibilityLabel="Отправить сообщение" accessibilityState={{disabled:busy||!input.trim()}} disabled={busy||!input.trim()} hitSlop={4} onPress={()=>void send()} style={({pressed})=>[styles.send,(busy||!input.trim())&&styles.sendDisabled,pressed&&styles.sendPressed]}><Text style={styles.sendText}>↑</Text></Pressable>
    </View>
  </View>;
}

const styles=StyleSheet.create({
  screen:{flex:1,padding:spacing.lg,gap:spacing.sm,backgroundColor:colors.background},
  header:{flexDirection:'row',justifyContent:'space-between',minHeight:control.minTouch,alignItems:'center'},back:{fontWeight:'850',color:colors.text},report:{fontWeight:'850',color:colors.text},
  hero:{gap:3},kicker:{fontSize:11,fontWeight:'900',letterSpacing:1.4,color:colors.textMuted},title:{fontSize:27,fontWeight:'950',color:colors.text},status:{fontSize:11,color:colors.textMuted},
  chat:{flex:1},chatInner:{gap:9,paddingVertical:8},hint:{borderWidth:1,borderColor:colors.border,borderRadius:radius.lg,padding:15},hintTitle:{fontWeight:'900',color:colors.text},hintText:{fontSize:13,lineHeight:19,color:colors.textMuted,marginTop:5},
  bubble:{maxWidth:'88%',borderRadius:radius.lg,padding:12},user:{backgroundColor:colors.primary,alignSelf:'flex-end'},ai:{borderWidth:1,borderColor:colors.border,alignSelf:'flex-start',backgroundColor:colors.surface},bubbleText:{fontSize:14,lineHeight:20,color:colors.text},userText:{color:colors.inverse},
  busyRow:{gap:6},typing:{fontSize:12,color:colors.textMuted},errorBox:{gap:8},error:{fontSize:12,fontWeight:'750',color:colors.danger},
  composer:{flexDirection:'row',alignItems:'flex-end',gap:8},input:{flex:1,borderWidth:1,borderColor:colors.border,borderRadius:radius.lg,paddingHorizontal:13,paddingVertical:10,maxHeight:100,minHeight:control.minTouch,color:colors.text,backgroundColor:colors.surface},
  send:{width:control.minTouch,height:control.minTouch,borderRadius:control.minTouch/2,backgroundColor:colors.primary,alignItems:'center',justifyContent:'center'},sendDisabled:{opacity:.4},sendPressed:{opacity:.72},sendText:{color:colors.inverse,fontSize:22,fontWeight:'900'},
});
