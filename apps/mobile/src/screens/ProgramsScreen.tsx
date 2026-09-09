import React, {useCallback, useEffect, useState} from 'react';
import {ActivityIndicator, Pressable, RefreshControl, ScrollView, StyleSheet, Text, View} from 'react-native';
import {api, ProgramAnalytics, TrainingProgram} from '../api/client';

export function ProgramsScreen({accessToken,onBack,onCreate,onOpen}: {
  accessToken:string; onBack:()=>void; onCreate:()=>void; onOpen:(programId:string)=>void;
}) {
  const [active,setActive]=useState<TrainingProgram|undefined>();
  const [analytics,setAnalytics]=useState<ProgramAnalytics|undefined>();
  const [history,setHistory]=useState<TrainingProgram[]>([]);
  const [loading,setLoading]=useState(true);
  const load=useCallback(async()=>{
    setLoading(true);
    try{
      const current=await api.activeProgram(accessToken); setActive(current);
      if(current) setAnalytics(await api.programAnalytics(accessToken,current.program.id)); else setAnalytics(undefined);
      const old=await api.programHistory(accessToken,10); setHistory(old.items.filter(x=>x.program.status!=='active'));
    }finally{setLoading(false)}
  },[accessToken]);
  useEffect(()=>{void load()},[load]);
  return <ScrollView refreshControl={<RefreshControl refreshing={loading} onRefresh={load}/>} contentContainerStyle={styles.container}>
    <View style={styles.top}><Pressable onPress={onBack}><Text style={styles.back}>← Главная</Text></Pressable><Pressable onPress={onCreate}><Text style={styles.create}>+ Новая</Text></Pressable></View>
    <Text style={styles.eyebrow}>FITNESS 2.0</Text><Text style={styles.title}>Программа</Text>
    {loading && !active ? <ActivityIndicator/> : null}
    {!active ? <View style={styles.empty}><Text style={styles.emptyTitle}>Активной программы пока нет</Text><Text style={styles.muted}>Создай план на 4, 8 или 12 недель. Он будет отслеживать пропуски, разгрузки и выполнение.</Text><Pressable onPress={onCreate} style={styles.primary}><Text style={styles.primaryText}>Создать программу</Text></Pressable></View> : <Pressable onPress={()=>onOpen(active.program.id)} style={styles.hero}>
      <View style={styles.heroTop}><View><Text style={styles.heroKicker}>АКТИВНАЯ</Text><Text style={styles.heroTitle}>{active.program.title}</Text></View><Text style={styles.arrow}>→</Text></View>
      <View style={styles.stats}><Stat label="Неделя" value={`${analytics?.current_week ?? 1}/${active.program.weeks}`}/><Stat label="Adherence" value={`${analytics?.adherence_percent ?? 0}%`}/><Stat label="Готово" value={`${analytics?.completed_sessions ?? 0}`}/></View>
      {analytics?.next_session ? <Text style={styles.next}>Следующая: {new Date(analytics.next_session.planned_date).toLocaleDateString()} · {analytics.next_session.muscle}{analytics.next_session.is_deload?' · deload':''}</Text>:null}
    </Pressable>}
    {history.length>0?<><Text style={styles.section}>Прошлые программы</Text>{history.map(item=><Pressable key={item.program.id} onPress={()=>onOpen(item.program.id)} style={styles.card}><View><Text style={styles.cardTitle}>{item.program.title}</Text><Text style={styles.muted}>{item.program.status} · {item.program.environment}</Text></View><Text>→</Text></Pressable>)}</>:null}
  </ScrollView>
}
function Stat({label,value}:{label:string;value:string}){return <View style={styles.stat}><Text style={styles.statValue}>{value}</Text><Text style={styles.statLabel}>{label}</Text></View>}
const styles=StyleSheet.create({container:{padding:20,gap:16},top:{flexDirection:'row',justifyContent:'space-between'},back:{fontWeight:'800'},create:{fontWeight:'900'},eyebrow:{fontSize:11,fontWeight:'900',letterSpacing:1.4,opacity:.45},title:{fontSize:32,fontWeight:'900'},hero:{backgroundColor:'#111',borderRadius:22,padding:18,gap:16},heroTop:{flexDirection:'row',justifyContent:'space-between',alignItems:'center'},heroKicker:{color:'#fff',fontSize:10,fontWeight:'900',letterSpacing:1.4,opacity:.55},heroTitle:{color:'#fff',fontSize:22,fontWeight:'900',marginTop:4},arrow:{color:'#fff',fontSize:28},stats:{flexDirection:'row',gap:8},stat:{flex:1,backgroundColor:'#222',borderRadius:14,padding:12},statValue:{color:'#fff',fontSize:18,fontWeight:'900'},statLabel:{color:'#fff',fontSize:10,opacity:.55,marginTop:2},next:{color:'#fff',fontSize:13,opacity:.7},empty:{borderWidth:1,borderRadius:20,padding:18,gap:10},emptyTitle:{fontSize:20,fontWeight:'900'},muted:{fontSize:13,lineHeight:19,opacity:.58},primary:{backgroundColor:'#111',padding:14,borderRadius:14,alignItems:'center'},primaryText:{color:'#fff',fontWeight:'900'},section:{fontSize:18,fontWeight:'900',marginTop:4},card:{borderWidth:1,borderRadius:16,padding:15,flexDirection:'row',justifyContent:'space-between',alignItems:'center'},cardTitle:{fontWeight:'900',fontSize:16}})
