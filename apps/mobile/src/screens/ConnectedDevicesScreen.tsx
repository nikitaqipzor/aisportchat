import React, {useCallback, useEffect, useState} from 'react';
import {ActivityIndicator, Pressable, ScrollView, StyleSheet, Text, View} from 'react-native';
import {api, HealthDailySnapshot, HealthInsights} from '../api/client';
import {AppButton} from '../components/AppButton';
import {currentLocalDate} from '../domain/date';
import {healthConnect, HealthConnectStatus} from '../native/healthConnect';
import {flushPassiveHealthSnapshots, syncHealthDays} from '../health/sync';
import {colors, radius, spacing} from '../theme/tokens';
import {sessionStorage} from '../storage/session';

function metric(value: number | undefined, suffix: string) { return value === undefined ? '—' : `${Math.round(value)}${suffix}`; }
function sleepLabel(minutes: number | undefined) { if (!minutes) return '—'; const h=Math.floor(minutes/60); const m=Math.round(minutes%60); return `${h} ч ${m} мин`; }
function freshnessLabel(value: HealthInsights['freshness']['status']) { return value==='fresh'?'Свежие данные':value==='aging'?'Нужно обновить скоро':value==='stale'?'Данные устарели':'Нет данных'; }
function confidenceLabel(value: HealthInsights['confidence']) { return value==='high'?'Высокая':value==='medium'?'Средняя':'Низкая'; }
function signedPercent(value?: number) { if (value===undefined) return '—'; const rounded=Math.round(value); return `${rounded>0?'+':''}${rounded}%`; }

export function ConnectedDevicesScreen({accessToken,onBack}:{accessToken:string;onBack:()=>void}) {
  const [status,setStatus]=useState<HealthConnectStatus|null>(null);
  const [snapshot,setSnapshot]=useState<HealthDailySnapshot|null>(null);
  const [insights,setInsights]=useState<HealthInsights|null>(null);
  const [busy,setBusy]=useState(false);
  const [message,setMessage]=useState('');
  const [syncProgress,setSyncProgress]=useState('');

  const load=useCallback(async()=>{
    const [nativeStatus,server,analysis]=await Promise.all([
      healthConnect.status(),
      api.healthToday(accessToken,currentLocalDate()).catch(()=>undefined),
      api.healthInsights(accessToken,currentLocalDate()).catch(()=>undefined),
    ]);
    setStatus(nativeStatus); setSnapshot(server??null); setInsights(analysis??null);
  },[accessToken]);
  useEffect(()=>{void load().catch(e=>setMessage(e instanceof Error?e.message:'Не удалось проверить Health Connect'))},[load]);

  async function permissions(){
    try { setBusy(true); setMessage(''); const next=await healthConnect.requestPermissions(); setStatus(next); if(next.permissions_granted)setMessage('Доступ Health Connect разрешён. Теперь можно синхронизировать данные Mi Fitness.'); }
    catch(e){setMessage(e instanceof Error?e.message:'Не удалось запросить разрешения')} finally{setBusy(false)}
  }
  async function backgroundPermission(){
    try{setBusy(true);setMessage('');const next=await healthConnect.requestBackgroundPermission();setStatus(next);setMessage(next.background_read_granted?'Фоновое чтение Health Connect разрешено.':'Фоновое разрешение не выдано.');}
    catch(e){setMessage(e instanceof Error?e.message:'Не удалось запросить фоновое разрешение')}finally{setBusy(false)}
  }
  async function togglePassive(enabled:boolean){
    try{setBusy(true);setMessage('');const ownerUserId=await sessionStorage.currentUserId();if(!ownerUserId)throw new Error('Сессия пользователя не найдена');const next=await healthConnect.setPassiveSyncEnabled(ownerUserId,enabled);setStatus(next);setMessage(enabled?'Пассивная синхронизация включена. Android будет периодически обновлять защищённый локальный кэш.':'Пассивная синхронизация выключена.');}
    catch(e){setMessage(e instanceof Error?e.message:'Не удалось изменить пассивную синхронизацию')}finally{setBusy(false)}
  }
  async function sync(days:number){
    try {
      setBusy(true); setMessage(''); setSyncProgress('');
      const drained=await flushPassiveHealthSnapshots(accessToken);
      const result=await syncHealthDays(accessToken,days,(processed,total)=>setSyncProgress(`${processed}/${total} дней`));
      const [analysis,today,nativeStatus]=await Promise.all([api.healthInsights(accessToken,currentLocalDate()),api.healthToday(accessToken,currentLocalDate()).catch(()=>undefined),healthConnect.status()]);
      setStatus(nativeStatus); setSnapshot(today??result.latest??drained.latest??null); setInsights(analysis);
      setMessage(`Импортировано дней с данными: ${result.imported}${result.empty?` · пустых пропущено: ${result.empty}`:''}${drained.imported?` · из фонового кэша: ${drained.imported}`:''}. Recovery пересчитан.`);
    } catch(e){setMessage(e instanceof Error?e.message:'Синхронизация не удалась')} finally{setBusy(false);setSyncProgress('')}
  }

  const available=status?.sdk_status==='available'; const granted=status?.permissions_granted===true;
  const b28=insights?.baseline_28d;
  return <ScrollView contentContainerStyle={styles.container}>
    <Pressable onPress={onBack} accessibilityRole="button" hitSlop={10}><Text style={styles.back}>← Главная</Text></Pressable>
    <Text style={styles.kicker}>DEVICES</Text><Text style={styles.title}>Подключённые устройства</Text>
    <View style={styles.deviceCard}><View style={styles.icon}><Text style={styles.iconText}>⌚</Text></View><View style={styles.flex}><Text style={styles.deviceName}>Xiaomi Watch S3</Text><Text style={styles.meta}>Mi Fitness → Health Connect → AI Fitness OS</Text><Text style={styles.state}>{status?.mi_fitness_installed?'Mi Fitness найден на телефоне':'Mi Fitness не найден'}</Text></View></View>

    <View style={styles.card}><Text style={styles.section}>Health Connect</Text><Text style={styles.meta}>{!status?'Проверяю…':available?'Доступен на устройстве':status.sdk_status==='update_required'?'Нужно установить/обновить Health Connect':'Health Connect недоступен'}</Text><Text style={styles.meta}>{granted?'Разрешения выданы':'Нужны разрешения на шаги, дистанцию, активные калории, сон, тренировки и пульс'}</Text>
      {!granted&&available?<AppButton label="Разрешить доступ" onPress={()=>void permissions()} disabled={busy} testID="health-connect-permissions"/>:null}
      <AppButton label="Открыть настройки Health Connect" variant="secondary" onPress={()=>void healthConnect.openSettings()} disabled={!available} />
    </View>

    {granted&&status?.background_read_available?<View style={styles.card}><Text style={styles.section}>Пассивная синхронизация</Text><Text style={styles.meta}>Android может читать Health Connect примерно раз в час в фоне. Снимки шифруются Android Keystore и отправляются в наш API только после авторизованного запуска приложения.</Text><Text style={styles.meta}>{status.background_read_granted?'Фоновое чтение разрешено':'Нужно отдельное разрешение фонового чтения'} · {status.passive_sync.enabled?'включено':'выключено'}{status.passive_sync.last_success_at?` · последний read ${new Date(status.passive_sync.last_success_at).toLocaleString()}`:''}</Text>{status.passive_sync.last_error?<Text style={styles.errorText}>{status.passive_sync.last_error}</Text>:null}<View style={styles.actions}>{!status.background_read_granted?<AppButton label="Разрешить фоновое чтение" variant="secondary" onPress={()=>void backgroundPermission()} disabled={busy}/>:null}<AppButton label={status.passive_sync.enabled?'Выключить пассивную синхронизацию':'Включить пассивную синхронизацию'} variant={status.passive_sync.enabled?'secondary':'primary'} onPress={()=>void togglePassive(!status.passive_sync.enabled)} disabled={busy||!status.background_read_granted}/></View></View>:null}

    {granted?<View style={styles.actions}><AppButton label="Синхронизировать сегодня" onPress={()=>void sync(1)} disabled={busy} testID="health-sync-today"/><AppButton label="Синхронизировать 7 дней" variant="secondary" onPress={()=>void sync(7)} disabled={busy} testID="health-sync-week"/><AppButton label="Обучить норму · 28 дней" variant="secondary" onPress={()=>void sync(28)} disabled={busy} testID="health-sync-baseline"/></View>:null}
    {busy?<View style={styles.loading}><ActivityIndicator/><Text style={styles.meta}>Читаю Health Connect… {syncProgress}</Text></View>:null}
    {message?<Text style={styles.message}>{message}</Text>:null}

    {insights?<View style={styles.card}><View style={styles.rowBetween}><View><Text style={styles.section}>Качество данных</Text><Text style={styles.meta}>{freshnessLabel(insights.freshness.status)}</Text></View><View style={styles.confidencePill}><Text style={styles.confidenceValue}>{insights.confidence_percent}%</Text><Text style={styles.confidenceLabel}>{confidenceLabel(insights.confidence)}</Text></View></View><View style={styles.grid}><Metric label="Покрытие 28 дней" value={`${b28?.available_days??0}/28`}/><Metric label="Покрытие" value={`${b28?.coverage_percent??0}%`}/><Metric label="Синхронизация" value={insights.freshness.status==='missing'?'—':`${Math.max(0,Math.round(insights.freshness.sync_age_minutes/60))} ч назад`}/><Metric label="Источник" value={snapshot?.source_label??'—'}/></View>{insights.reasons.map((reason,i)=><Text key={i} style={styles.reason}>• {reason}</Text>)}</View>:null}

    {b28?<View style={styles.card}><Text style={styles.section}>Твоя норма · 28 дней</Text><Text style={styles.meta}>Это персональная fitness-база по предыдущим дням. Сегодняшний день в среднее не входит.</Text><View style={styles.grid}><Metric label="Сон" value={sleepLabel(b28.sleep_minutes?.average)}/><Metric label="Шаги" value={metric(b28.steps?.average,'')}/><Metric label="Активные ккал" value={metric(b28.active_calories_kcal?.average,'')}/><Metric label="Тренировки" value={metric(b28.exercise_minutes?.average,' мин/д')}/></View></View>:null}

    {snapshot?<View style={styles.card}><Text style={styles.section}>Сегодня</Text><Text style={styles.source}>{snapshot.source_label}</Text><View style={styles.grid}><Metric label="Шаги" value={metric(snapshot.steps,'')}/><Metric label="vs твоя норма" value={signedPercent(insights?.deviations.steps?.delta_percent)}/><Metric label="Дистанция" value={metric(snapshot.distance_m/1000,' км')}/><Metric label="Активные ккал" value={metric(snapshot.active_calories_kcal,'')}/><Metric label="Сон" value={sleepLabel(snapshot.sleep_minutes)}/><Metric label="Сон vs норма" value={signedPercent(insights?.deviations.sleep_minutes?.delta_percent)}/><Metric label="Глубокий" value={sleepLabel(snapshot.deep_sleep_minutes)}/><Metric label="REM" value={sleepLabel(snapshot.rem_sleep_minutes)}/><Metric label="Тренировки" value={`${snapshot.exercise_sessions} · ${snapshot.exercise_minutes} мин`}/><Metric label="Пульс тренировки" value={snapshot.exercise_heart_rate_avg?`${Math.round(snapshot.exercise_heart_rate_avg)} ср. / ${Math.round(snapshot.exercise_heart_rate_max??0)} max`:'—'}/></View></View>:null}

    {insights?<View style={styles.card}><Text style={styles.section}>Происхождение данных</Text>{insights.conflict_resolved?<View style={styles.conflictBox}><Text style={styles.conflictTitle}>Найдено несколько источников</Text><Text style={styles.meta}>Сервер сохранил их раздельно и выбрал приоритетный источник вместо last-write-wins.</Text></View>:null}{insights.sources.map(source=><View key={source.source_package} style={styles.sourceRow}><View style={styles.flex}><Text style={styles.provenanceMetric}>{source.source_label||source.source_package}</Text><Text style={styles.meta}>{source.data_types.join(' · ')||'нет доступных метрик'}</Text></View><Text style={[styles.sourceBadge,source.selected&&styles.sourceBadgeSelected]}>{source.selected?'ВЫБРАН':'резерв'}</Text></View>)}{insights.provenance.map(item=><View key={item.metric} style={styles.provenanceRow}><Text style={styles.provenanceMetric}>{provenanceName(item.metric)}</Text><Text style={[styles.provenanceState,!item.available&&styles.muted]}>{item.available?item.source_label??item.source_package:'Нет записи'}</Text></View>)}</View>:null}
    <View style={styles.info}><Text style={styles.infoText}>Мы читаем только разрешённые данные Health Connect. Персональная норма — fitness-тренд, а не медицинская диагностика. Если синхронизация устарела, Recovery Engine не должен слепо доверять старому сну.</Text></View>
  </ScrollView>;
}
function Metric({label,value}:{label:string;value:string}){return <View style={styles.metric}><Text style={styles.metricLabel}>{label}</Text><Text style={styles.metricValue}>{value}</Text></View>}
function provenanceName(value:string){return value==='steps'?'Шаги':value==='distance'?'Дистанция':value==='active_calories'?'Активные калории':value==='sleep'?'Сон':value==='exercise'?'Тренировки':value==='heart_rate'?'Пульс тренировки':'Пульс покоя'}
const styles=StyleSheet.create({container:{padding:spacing.lg,gap:spacing.md,backgroundColor:colors.background},back:{fontWeight:'800',color:colors.text},kicker:{fontSize:11,fontWeight:'900',letterSpacing:1.4,color:colors.textMuted},title:{fontSize:30,fontWeight:'900',color:colors.text},deviceCard:{borderWidth:1,borderColor:colors.border,borderRadius:radius.lg,padding:16,flexDirection:'row',gap:13,alignItems:'center'},icon:{width:52,height:52,borderRadius:18,backgroundColor:colors.primary,alignItems:'center',justifyContent:'center'},iconText:{fontSize:25},flex:{flex:1},deviceName:{fontSize:19,fontWeight:'900',color:colors.text},meta:{fontSize:12,lineHeight:17,color:colors.textMuted},state:{fontSize:12,fontWeight:'800',marginTop:5,color:colors.text},card:{borderWidth:1,borderColor:colors.border,borderRadius:radius.lg,padding:16,gap:10},section:{fontSize:18,fontWeight:'900',color:colors.text},actions:{gap:9},loading:{flexDirection:'row',gap:9,alignItems:'center'},message:{fontSize:13,lineHeight:19,fontWeight:'700',color:colors.text},source:{fontSize:12,fontWeight:'800',color:colors.textMuted},grid:{flexDirection:'row',flexWrap:'wrap',gap:9},metric:{width:'48%',backgroundColor:'#f4f4f4',borderRadius:14,padding:12,minHeight:74},metricLabel:{fontSize:11,fontWeight:'800',color:colors.textMuted},metricValue:{fontSize:16,fontWeight:'900',color:colors.text,marginTop:6},info:{borderRadius:radius.md,backgroundColor:'#f3f3f3',padding:14},infoText:{fontSize:12,lineHeight:18,color:colors.textMuted},rowBetween:{flexDirection:'row',justifyContent:'space-between',alignItems:'center',gap:10},confidencePill:{minWidth:78,borderRadius:16,paddingVertical:8,paddingHorizontal:10,backgroundColor:colors.primary,alignItems:'center'},confidenceValue:{fontSize:20,fontWeight:'950',color:colors.inverse},confidenceLabel:{fontSize:10,fontWeight:'800',color:colors.inverse,opacity:.75},reason:{fontSize:12,lineHeight:17,color:colors.textMuted},provenanceRow:{flexDirection:'row',justifyContent:'space-between',gap:12,paddingVertical:8,borderBottomWidth:1,borderColor:colors.border},provenanceMetric:{fontWeight:'800',color:colors.text},provenanceState:{flex:1,textAlign:'right',fontSize:12,fontWeight:'700',color:colors.text},muted:{color:colors.textMuted},conflictBox:{borderRadius:12,padding:11,backgroundColor:'#f4f4f4',gap:3},conflictTitle:{fontWeight:'900',color:colors.text},sourceRow:{flexDirection:'row',alignItems:'center',gap:10,paddingVertical:9,borderBottomWidth:1,borderColor:colors.border},sourceBadge:{fontSize:9,fontWeight:'900',color:colors.textMuted,borderWidth:1,borderColor:colors.border,borderRadius:999,paddingHorizontal:8,paddingVertical:5},sourceBadgeSelected:{backgroundColor:colors.primary,color:colors.inverse,borderColor:colors.primary},errorText:{fontSize:12,fontWeight:'700',color:'#8b1e1e'}});
