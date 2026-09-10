import React, {useCallback, useEffect, useState} from 'react';
import {ActivityIndicator, Pressable, ScrollView, StyleSheet, Text, View} from 'react-native';
import {api, HealthDailySnapshot, HealthInsights} from '../api/client';
import {AppButton} from '../components/AppButton';
import {currentLocalDate} from '../domain/date';
import {healthConnect, HealthConnectStatus} from '../native/healthConnect';
import {flushPassiveHealthSnapshots, syncHealthDays} from '../health/sync';
import {colors, control, radius, spacing} from '../theme/tokens';
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
  const [initialLoading,setInitialLoading]=useState(true);
  const [loadError,setLoadError]=useState('');
  const [busy,setBusy]=useState(false);
  const [busyLabel,setBusyLabel]=useState('');
  const [message,setMessage]=useState('');
  const [messageTone,setMessageTone]=useState<'success'|'error'|'info'>('info');
  const [syncProgress,setSyncProgress]=useState('');

  const load=useCallback(async()=>{
    setInitialLoading(true);
    setLoadError('');
    try {
      const nativeStatus=await healthConnect.status();
      setStatus(nativeStatus);
      const [server,analysis]=await Promise.allSettled([
        api.healthToday(accessToken,currentLocalDate()),
        api.healthInsights(accessToken,currentLocalDate()),
      ]);
      setSnapshot(server.status==='fulfilled'?(server.value??null):null);
      setInsights(analysis.status==='fulfilled'?analysis.value:null);
      if(server.status==='rejected'||analysis.status==='rejected'){
        setLoadError('Health Connect доступен, но данные сервера загрузились не полностью. Можно повторить проверку.');
      }
    } catch(e) {
      setLoadError(e instanceof Error?e.message:'Не удалось проверить Health Connect.');
    } finally {
      setInitialLoading(false);
    }
  },[accessToken]);
  useEffect(()=>{void load()},[load]);

  async function permissions(){
    try { setBusy(true); setBusyLabel('Запрашиваю разрешения…'); setMessage(''); const next=await healthConnect.requestPermissions(); setStatus(next); setMessageTone(next.permissions_granted?'success':'info'); setMessage(next.permissions_granted?'Доступ Health Connect разрешён. Теперь можно синхронизировать данные Mi Fitness.':'Разрешения не выданы. Без них синхронизация недоступна.'); }
    catch(e){setMessageTone('error');setMessage(e instanceof Error?e.message:'Не удалось запросить разрешения.')} finally{setBusy(false);setBusyLabel('')}
  }
  async function backgroundPermission(){
    try{setBusy(true);setBusyLabel('Запрашиваю фоновый доступ…');setMessage('');const next=await healthConnect.requestBackgroundPermission();setStatus(next);setMessageTone(next.background_read_granted?'success':'info');setMessage(next.background_read_granted?'Фоновое чтение Health Connect разрешено.':'Фоновое разрешение не выдано.');}
    catch(e){setMessageTone('error');setMessage(e instanceof Error?e.message:'Не удалось запросить фоновое разрешение.')}finally{setBusy(false);setBusyLabel('')}
  }
  async function togglePassive(enabled:boolean){
    try{setBusy(true);setBusyLabel(enabled?'Включаю пассивную синхронизацию…':'Выключаю пассивную синхронизацию…');setMessage('');const ownerUserId=await sessionStorage.currentUserId();if(!ownerUserId)throw new Error('Сессия пользователя не найдена');const next=await healthConnect.setPassiveSyncEnabled(ownerUserId,enabled);setStatus(next);setMessageTone('success');setMessage(enabled?'Пассивная синхронизация включена. Android будет периодически обновлять защищённый локальный кэш.':'Пассивная синхронизация выключена.');}
    catch(e){setMessageTone('error');setMessage(e instanceof Error?e.message:'Не удалось изменить пассивную синхронизацию.')}finally{setBusy(false);setBusyLabel('')}
  }
  async function sync(days:number){
    try {
      setBusy(true); setBusyLabel('Читаю Health Connect…'); setMessage(''); setSyncProgress('');
      const drained=await flushPassiveHealthSnapshots(accessToken);
      const result=await syncHealthDays(accessToken,days,(processed,total)=>setSyncProgress(`${processed}/${total} дней`));
      const [analysis,today,nativeStatus]=await Promise.all([api.healthInsights(accessToken,currentLocalDate()),api.healthToday(accessToken,currentLocalDate()).catch(()=>undefined),healthConnect.status()]);
      setStatus(nativeStatus); setSnapshot(today??result.latest??drained.latest??null); setInsights(analysis);
      setMessageTone('success');setMessage(`Импортировано дней с данными: ${result.imported}${result.empty?` · пустых пропущено: ${result.empty}`:''}${drained.imported?` · из фонового кэша: ${drained.imported}`:''}. Recovery пересчитан.`);
    } catch(e){setMessageTone('error');setMessage(e instanceof Error?e.message:'Синхронизация не удалась.')} finally{setBusy(false);setBusyLabel('');setSyncProgress('')}
  }

  async function openSettings(){
    try{setBusy(true);setBusyLabel('Открываю настройки…');setMessage('');await healthConnect.openSettings();}
    catch(e){setMessageTone('error');setMessage(e instanceof Error?e.message:'Не удалось открыть настройки Health Connect.')}finally{setBusy(false);setBusyLabel('')}
  }

  const available=status?.sdk_status==='available'; const granted=status?.permissions_granted===true;
  const b28=insights?.baseline_28d;
  const miFitnessState=!status?'Статус Mi Fitness пока неизвестен':status.mi_fitness_installed?'Mi Fitness найден на телефоне':'Mi Fitness не найден';
  const connectionStep=!status?.mi_fitness_installed?1:!available?2:!granted?3:!snapshot?4:5;
  return <ScrollView testID="connected-devices-screen" contentContainerStyle={styles.container} showsVerticalScrollIndicator={false}>
    <Pressable
      onPress={onBack}
      accessibilityRole="button"
      accessibilityLabel="Вернуться на главную"
      accessibilityState={{disabled:busy}}
      disabled={busy}
      style={({pressed})=>[styles.backButton,pressed&&styles.pressed]}>
      <Text style={styles.back}>← Главная</Text>
    </Pressable>
    <Text style={styles.kicker}>УСТРОЙСТВА</Text>
    <Text accessibilityRole="header" style={styles.title}>Подключённые устройства</Text>
    <Text style={styles.intro}>Импортируй активность и сон из Health Connect. Конкретная модель часов определяется приложением-поставщиком и может быть неизвестна.</Text>

    <View style={styles.connectionCard} testID="health-connection-progress">
      <View style={styles.rowBetween}><View style={styles.flex}><Text style={styles.connectionEyebrow}>ПОДКЛЮЧЕНИЕ</Text><Text style={styles.connectionTitle}>{connectionStep===5?'Часы подключены':`Шаг ${connectionStep} из 4`}</Text></View><Text style={[styles.connectionBadge,connectionStep===5&&styles.connectionBadgeReady]}>{connectionStep===5?'ГОТОВО':'НАСТРОЙКА'}</Text></View>
      <ConnectionStep number={1} title="Приложение часов" detail={status?.mi_fitness_installed?'Mi Fitness найден. Модель часов не определяется':'Подключите часы к совместимому приложению, например Mi Fitness'} done={connectionStep>1}/>
      <ConnectionStep number={2} title="Health Connect" detail={available?'Доступен':'Установите или обновите Health Connect'} done={connectionStep>2}/>
      <ConnectionStep number={3} title="Разрешения" detail={granted?'Доступ выдан':'Разрешите чтение выбранных показателей'} done={connectionStep>3}/>
      <ConnectionStep number={4} title="Первая синхронизация" detail={snapshot?'Данные получены':'Синхронизируйте данные за сегодня'} done={connectionStep>4}/>
    </View>

    {initialLoading?<View accessibilityLiveRegion="polite" style={styles.loadingCard}><ActivityIndicator color={colors.text}/><Text style={styles.meta}>Проверяем устройство и данные…</Text></View>:null}
    {loadError?<View style={styles.errorCard}><Text accessibilityRole="alert" style={styles.errorText}>{loadError}</Text><AppButton label="Повторить проверку" variant="secondary" disabled={initialLoading||busy} loading={initialLoading} onPress={()=>void load()} testID="health-retry-load"/></View>:null}

    <View style={styles.deviceCard} accessible accessibilityLabel={`Источник данных Health Connect. Модель устройства неизвестна. ${miFitnessState}`}>
      <View style={styles.icon} accessible={false}><Text style={styles.iconText}>⌚</Text></View>
      <View style={styles.flex} accessible={false}>
        <Text style={styles.deviceName}>Устройство через Health Connect</Text>
        <Text style={styles.meta}>{status?.mi_fitness_installed?'Mi Fitness найден · модель часов неизвестна':'Поставщик и модель пока неизвестны'}</Text>
        <Text style={styles.state}>{miFitnessState}</Text>
      </View>
    </View>

    <View style={styles.card}>
      <Text accessibilityRole="header" style={styles.section}>Health Connect</Text>
      <Text style={styles.statusText}>{!status?'Статус пока неизвестен':available?'Доступен на устройстве':status.sdk_status==='update_required'?'Нужно установить или обновить Health Connect':'Health Connect недоступен на этом устройстве'}</Text>
      <Text style={styles.meta}>{granted?'Разрешения выданы':'Нужны разрешения на шаги, дистанцию, активные калории, сон, тренировки и пульс.'}</Text>
      {!granted&&available?<AppButton label="Разрешить доступ" accessibilityLabel="Разрешить Health Connect доступ к данным" onPress={()=>void permissions()} disabled={busy} loading={busy&&busyLabel==='Запрашиваю разрешения…'} testID="health-connect-permissions"/>:null}
      <AppButton label="Открыть настройки Health Connect" variant="secondary" onPress={()=>void openSettings()} disabled={!available||busy} testID="health-connect-settings" />
    </View>

    {granted&&status?.background_read_available?<View style={styles.card}>
      <Text accessibilityRole="header" style={styles.section}>Пассивная синхронизация</Text>
      <Text style={styles.meta}>Android может читать Health Connect примерно раз в час в фоне. Снимки шифруются Android Keystore и отправляются в API только после авторизованного запуска приложения.</Text>
      <Text style={styles.statusText}>{status.background_read_granted?'Фоновое чтение разрешено':'Нужно разрешить фоновое чтение'} · {status.passive_sync.enabled?'синхронизация включена':'синхронизация выключена'}{status.passive_sync.last_success_at?` · последнее обновление ${new Date(status.passive_sync.last_success_at).toLocaleString()}`:''}</Text>
      {status.passive_sync.last_error?<Text accessibilityRole="alert" style={styles.errorText}>{status.passive_sync.last_error}</Text>:null}
      <View style={styles.actions}>
        {!status.background_read_granted?<AppButton label="Разрешить фоновое чтение" variant="secondary" onPress={()=>void backgroundPermission()} disabled={busy}/>:null}
        <AppButton label={status.passive_sync.enabled?'Выключить пассивную синхронизацию':'Включить пассивную синхронизацию'} variant={status.passive_sync.enabled?'secondary':'primary'} onPress={()=>void togglePassive(!status.passive_sync.enabled)} disabled={busy||!status.background_read_granted}/>
      </View>
    </View>:null}

    {granted?<View style={styles.actions} accessibilityLabel="Действия синхронизации">
      <AppButton label="Синхронизировать сегодня" onPress={()=>void sync(1)} disabled={busy} testID="health-sync-today"/>
      <AppButton label="Синхронизировать 7 дней" variant="secondary" onPress={()=>void sync(7)} disabled={busy} testID="health-sync-week"/>
      <AppButton label="Рассчитать норму за 28 дней" variant="secondary" onPress={()=>void sync(28)} disabled={busy} testID="health-sync-baseline"/>
    </View>:null}
    {busy?<View accessibilityLiveRegion="polite" style={styles.loading}><ActivityIndicator color={colors.text}/><Text style={styles.meta}>{busyLabel} {syncProgress}</Text></View>:null}
    {message?<Text accessibilityRole={messageTone==='error'?'alert':'text'} accessibilityLiveRegion={messageTone==='error'?'assertive':'polite'} style={[styles.message,messageTone==='error'&&styles.messageError,messageTone==='success'&&styles.messageSuccess]}>{message}</Text>:null}

    {granted&&!initialLoading&&!snapshot&&!insights?<View style={styles.emptyCard}><Text style={styles.emptyTitle}>Данных пока нет</Text><Text style={styles.meta}>Запусти синхронизацию за сегодня или 7 дней. Если показатели не появятся, проверь разрешения Health Connect и синхронизацию Mi Fitness.</Text></View>:null}

    {insights?<View style={styles.card}><View style={styles.rowBetween}><View style={styles.flex}><Text accessibilityRole="header" style={styles.section}>Качество данных</Text><Text style={styles.meta}>{freshnessLabel(insights.freshness.status)}</Text></View><View accessible accessibilityLabel={`Надёжность данных ${insights.confidence_percent} процентов, ${confidenceLabel(insights.confidence)}`} style={styles.confidencePill}><Text accessible={false} style={styles.confidenceValue}>{insights.confidence_percent}%</Text><Text accessible={false} style={styles.confidenceLabel}>{confidenceLabel(insights.confidence)}</Text></View></View><View style={styles.grid}><Metric label="Дней с данными" value={`${b28?.available_days??0}/28`}/><Metric label="Покрытие" value={`${b28?.coverage_percent??0}%`}/><Metric label="Синхронизация" value={insights.freshness.status==='missing'?'—':`${Math.max(0,Math.round(insights.freshness.sync_age_minutes/60))} ч назад`}/><Metric label="Источник" value={snapshot?.source_label??'—'}/></View>{insights.reasons.map((reason,i)=><Text key={i} style={styles.reason}>• {reason}</Text>)}</View>:null}

    {b28?<View style={styles.card}><Text accessibilityRole="header" style={styles.section}>Твоя норма · 28 дней</Text><Text style={styles.meta}>Это персональная база активности по предыдущим дням. Сегодняшний день в среднее не входит.</Text><View style={styles.grid}><Metric label="Сон" value={sleepLabel(b28.sleep_minutes?.average)}/><Metric label="Шаги" value={metric(b28.steps?.average,'')}/><Metric label="Активные ккал" value={metric(b28.active_calories_kcal?.average,'')}/><Metric label="Тренировки" value={metric(b28.exercise_minutes?.average,' мин/д')}/></View></View>:null}

    {snapshot?<View style={styles.card}><Text accessibilityRole="header" style={styles.section}>Сегодня</Text><Text style={styles.source}>{snapshot.source_label}</Text><View style={styles.grid}><Metric label="Шаги" value={metric(snapshot.steps,'')}/><Metric label="Относительно нормы" value={signedPercent(insights?.deviations.steps?.delta_percent)}/><Metric label="Дистанция" value={metric(snapshot.distance_m/1000,' км')}/><Metric label="Активные ккал" value={metric(snapshot.active_calories_kcal,'')}/><Metric label="Сон" value={sleepLabel(snapshot.sleep_minutes)}/><Metric label="Сон относительно нормы" value={signedPercent(insights?.deviations.sleep_minutes?.delta_percent)}/><Metric label="Глубокий сон" value={sleepLabel(snapshot.deep_sleep_minutes)}/><Metric label="REM-сон" value={sleepLabel(snapshot.rem_sleep_minutes)}/><Metric label="Тренировки" value={`${snapshot.exercise_sessions} · ${snapshot.exercise_minutes} мин`}/><Metric label="Пульс тренировки" value={snapshot.exercise_heart_rate_avg?`${Math.round(snapshot.exercise_heart_rate_avg)} ср. / ${Math.round(snapshot.exercise_heart_rate_max??0)} макс.`:'—'}/></View></View>:null}

    {insights?<View style={styles.card}><Text accessibilityRole="header" style={styles.section}>Происхождение данных</Text>{insights.conflict_resolved?<View style={styles.conflictBox}><Text style={styles.conflictTitle}>Найдено несколько источников</Text><Text style={styles.meta}>Сервер сохранил источники раздельно и выбрал приоритетный, чтобы данные не перезаписывали друг друга.</Text></View>:null}{insights.sources.map(source=><View key={source.source_package} accessible accessibilityLabel={`${source.source_label||source.source_package}. ${source.selected?'Выбранный источник':'Резервный источник'}. ${source.data_types.join(', ')||'Нет доступных метрик'}`} style={styles.sourceRow}><View accessible={false} style={styles.flex}><Text style={styles.provenanceMetric}>{source.source_label||source.source_package}</Text><Text style={styles.meta}>{source.data_types.join(' · ')||'нет доступных метрик'}</Text></View><Text accessible={false} style={[styles.sourceBadge,source.selected&&styles.sourceBadgeSelected]}>{source.selected?'ВЫБРАН':'РЕЗЕРВ'}</Text></View>)}{insights.provenance.map(item=><View key={item.metric} accessible accessibilityLabel={`${provenanceName(item.metric)}: ${item.available?item.source_label??item.source_package:'Нет записи'}`} style={styles.provenanceRow}><Text accessible={false} style={styles.provenanceMetric}>{provenanceName(item.metric)}</Text><Text accessible={false} style={[styles.provenanceState,!item.available&&styles.muted]}>{item.available?item.source_label??item.source_package:'Нет записи'}</Text></View>)}</View>:null}
    <View style={styles.info}><Text style={styles.infoText}>Мы читаем только разрешённые данные Health Connect. Персональная норма — fitness-тренд, а не медицинская диагностика. Если синхронизация устарела, Recovery Engine не должен слепо доверять старому сну.</Text></View>
  </ScrollView>;
}
function Metric({label,value}:{label:string;value:string}){return <View accessible accessibilityLabel={`${label}: ${value}`} style={styles.metric}><Text accessible={false} style={styles.metricLabel}>{label}</Text><Text accessible={false} style={styles.metricValue}>{value}</Text></View>}
function ConnectionStep({number,title,detail,done}:{number:number;title:string;detail:string;done:boolean}){return <View accessible accessibilityLabel={`${title}. ${detail}`} style={styles.connectionStep}><View accessible={false} style={[styles.stepDot,done&&styles.stepDotDone]}><Text style={[styles.stepNumber,done&&styles.stepNumberDone]}>{done?'✓':number}</Text></View><View accessible={false} style={styles.flex}><Text style={styles.stepTitle}>{title}</Text><Text style={styles.stepDetail}>{detail}</Text></View></View>}
function provenanceName(value:string){return value==='steps'?'Шаги':value==='distance'?'Дистанция':value==='active_calories'?'Активные калории':value==='sleep'?'Сон':value==='exercise'?'Тренировки':value==='heart_rate'?'Пульс тренировки':'Пульс покоя'}
const styles=StyleSheet.create({
  container:{flexGrow:1,padding:spacing.lg,paddingBottom:spacing.xl,gap:spacing.md,backgroundColor:colors.background},
  backButton:{minHeight:control.minTouch,minWidth:control.minTouch,alignSelf:'flex-start',justifyContent:'center',paddingRight:spacing.sm},
  back:{fontSize:15,fontWeight:'800',color:colors.text},
  kicker:{fontSize:11,fontWeight:'900',letterSpacing:1.4,color:colors.textMuted},
  title:{fontSize:30,lineHeight:37,fontWeight:'900',color:colors.text},
  intro:{fontSize:14,lineHeight:21,color:colors.textMuted},
  connectionCard:{borderRadius:radius.lg,padding:spacing.md,gap:11,backgroundColor:colors.primary},
  connectionEyebrow:{fontSize:10,fontWeight:'900',letterSpacing:1.2,color:'#AFAFAF'},
  connectionTitle:{fontSize:20,lineHeight:26,fontWeight:'900',color:colors.inverse,marginTop:2},
  connectionBadge:{fontSize:9,fontWeight:'900',color:colors.text,backgroundColor:'#E5E5E5',borderRadius:radius.pill,paddingHorizontal:9,paddingVertical:6},
  connectionBadgeReady:{color:colors.success,backgroundColor:'#E8F5ED'},
  connectionStep:{flexDirection:'row',alignItems:'center',gap:11,minHeight:44},
  stepDot:{width:30,height:30,borderRadius:15,alignItems:'center',justifyContent:'center',borderWidth:1,borderColor:'#666'},
  stepDotDone:{backgroundColor:colors.inverse,borderColor:colors.inverse},
  stepNumber:{fontSize:12,fontWeight:'900',color:'#AAA'},
  stepNumberDone:{color:colors.success},
  stepTitle:{fontSize:13,lineHeight:18,fontWeight:'900',color:colors.inverse},
  stepDetail:{fontSize:11,lineHeight:16,color:'#BDBDBD'},
  loadingCard:{minHeight:72,borderRadius:radius.md,padding:spacing.md,backgroundColor:colors.surfaceMuted,flexDirection:'row',alignItems:'center',gap:spacing.sm},
  errorCard:{borderWidth:1,borderColor:colors.danger,borderRadius:radius.md,padding:spacing.md,gap:spacing.sm,backgroundColor:colors.surface},
  deviceCard:{borderWidth:1,borderColor:colors.border,borderRadius:radius.lg,padding:spacing.md,flexDirection:'row',gap:13,alignItems:'center',backgroundColor:colors.surface},
  icon:{width:52,height:52,borderRadius:18,backgroundColor:colors.primary,alignItems:'center',justifyContent:'center'},
  iconText:{fontSize:25},
  flex:{flex:1},
  deviceName:{fontSize:19,lineHeight:24,fontWeight:'900',color:colors.text},
  meta:{fontSize:12,lineHeight:18,color:colors.textMuted},
  state:{fontSize:12,lineHeight:17,fontWeight:'800',marginTop:5,color:colors.text},
  statusText:{fontSize:13,lineHeight:19,fontWeight:'800',color:colors.text},
  card:{borderWidth:1,borderColor:colors.border,borderRadius:radius.lg,padding:spacing.md,gap:10,backgroundColor:colors.surface},
  section:{fontSize:18,lineHeight:24,fontWeight:'900',color:colors.text},
  actions:{gap:9},
  loading:{minHeight:control.minTouch,flexDirection:'row',gap:spacing.sm,alignItems:'center'},
  message:{fontSize:13,lineHeight:19,fontWeight:'700',color:colors.text,padding:12,borderRadius:radius.sm,backgroundColor:colors.surfaceMuted},
  messageError:{color:colors.danger,borderWidth:1,borderColor:colors.danger,backgroundColor:colors.surface},
  messageSuccess:{color:colors.success},
  emptyCard:{borderWidth:1,borderColor:colors.border,borderRadius:radius.lg,padding:spacing.lg,gap:spacing.sm,alignItems:'flex-start',backgroundColor:colors.surfaceMuted},
  emptyTitle:{fontSize:17,lineHeight:23,fontWeight:'900',color:colors.text},
  source:{fontSize:12,fontWeight:'800',color:colors.textMuted},
  grid:{flexDirection:'row',flexWrap:'wrap',gap:9},
  metric:{width:'48%',flexGrow:1,minWidth:130,backgroundColor:colors.surfaceMuted,borderRadius:radius.md,padding:12,minHeight:74},
  metricLabel:{fontSize:11,lineHeight:16,fontWeight:'800',color:colors.textMuted},
  metricValue:{fontSize:16,lineHeight:21,fontWeight:'900',color:colors.text,marginTop:6},
  info:{borderRadius:radius.md,backgroundColor:colors.surfaceMuted,padding:14},
  infoText:{fontSize:12,lineHeight:18,color:colors.textMuted},
  rowBetween:{flexDirection:'row',justifyContent:'space-between',alignItems:'center',gap:10},
  confidencePill:{minWidth:78,borderRadius:radius.md,paddingVertical:8,paddingHorizontal:10,backgroundColor:colors.primary,alignItems:'center'},
  confidenceValue:{fontSize:20,lineHeight:25,fontWeight:'900',color:colors.inverse},
  confidenceLabel:{fontSize:10,lineHeight:14,fontWeight:'800',color:colors.inverse,opacity:.75},
  reason:{fontSize:12,lineHeight:17,color:colors.textMuted},
  provenanceRow:{minHeight:control.minTouch,flexDirection:'row',justifyContent:'space-between',alignItems:'center',gap:12,paddingVertical:8,borderBottomWidth:1,borderColor:colors.border},
  provenanceMetric:{flexShrink:1,fontWeight:'800',color:colors.text},
  provenanceState:{flex:1,textAlign:'right',fontSize:12,lineHeight:17,fontWeight:'700',color:colors.text},
  muted:{color:colors.textMuted},
  conflictBox:{borderRadius:radius.sm,padding:11,backgroundColor:colors.surfaceMuted,gap:3},
  conflictTitle:{fontWeight:'900',color:colors.text},
  sourceRow:{minHeight:control.minTouch,flexDirection:'row',alignItems:'center',gap:10,paddingVertical:9,borderBottomWidth:1,borderColor:colors.border},
  sourceBadge:{fontSize:9,fontWeight:'900',color:colors.textMuted,borderWidth:1,borderColor:colors.border,borderRadius:radius.pill,paddingHorizontal:8,paddingVertical:5},
  sourceBadgeSelected:{backgroundColor:colors.primary,color:colors.inverse,borderColor:colors.primary},
  errorText:{fontSize:12,lineHeight:18,fontWeight:'700',color:colors.danger},
  pressed:{opacity:.72},
});
