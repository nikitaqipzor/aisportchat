import React, {useEffect, useMemo, useState} from 'react';
import {ActivityIndicator, ScrollView, StyleSheet, Text, View} from 'react-native';
import {api, TechniqueExercise, TechniqueResult} from '../api/client';
import {AppButton} from '../components/AppButton';
import {PoseFrame, techniqueVideo} from '../native/techniqueVideo';
import {techniqueLive} from '../native/techniqueLive';
import {PoseSkeletonPreview} from '../components/PoseSkeletonPreview';
import {TechniqueWorkoutContext, techniqueKeyForExercise} from '../domain/technique';

type Props = {
  accessToken: string;
  onBack: () => void;
  workoutContext?: TechniqueWorkoutContext | null;
  onUseLinkedResult?: (result: TechniqueResult) => void;
};

export function TechniqueScreen({accessToken, onBack, workoutContext, onUseLinkedResult}: Props) {
  const [exercises, setExercises] = useState<TechniqueExercise[]>([]);
  const [selected, setSelected] = useState<TechniqueExercise | null>(null);
  const [history, setHistory] = useState<TechniqueResult[]>([]);
  const [result, setResult] = useState<TechniqueResult | null>(null);
  const [busy, setBusy] = useState(false);
  const [stage, setStage] = useState('');
  const [error, setError] = useState('');
  const [posePreview, setPosePreview] = useState<PoseFrame | null>(null);
  const [nativeLiveCount, setNativeLiveCount] = useState<number | null>(null);

  const linkedKey = useMemo(() => workoutContext ? techniqueKeyForExercise(workoutContext.exerciseId) : null, [workoutContext]);

  useEffect(() => { void load(); }, [workoutContext?.workoutExerciseId]);
  async function load() {
    try {
      const [catalog, past] = await Promise.all([api.techniqueExercises(accessToken), api.techniqueHistory(accessToken, 8)]);
      setExercises(catalog.items);
      setHistory(past.items);
      if (linkedKey) setSelected(catalog.items.find(item => item.key === linkedKey) ?? null);
      else if (catalog.items.length) setSelected(current => current ?? catalog.items[0]);
    } catch (e) { setError(e instanceof Error ? e.message : 'Не удалось загрузить анализ техники'); }
  }

  function payloadContext() {
    if (!workoutContext) return {};
    return {
      workout_id: workoutContext.workoutId,
      workout_exercise_id: workoutContext.workoutExerciseId,
      set_number: workoutContext.setNumber,
    };
  }

  async function submitPose(captureMode: 'recorded'|'live', pose: {duration_ms: number; frames: PoseFrame[]; valid_frames: number}, liveCount?: number) {
    if (!selected) return;
    if (pose.frames.length) setPosePreview(pose.frames[Math.floor(pose.frames.length / 2)]);
    if (typeof liveCount === 'number') setNativeLiveCount(liveCount);
    setStage(`Распознано ${pose.valid_frames} кадров. Проверяю повторения, амплитуду и стабильность…`);
    const analyzed = await api.analyzeTechnique(accessToken, {
      exercise_key: selected.key,
      capture_mode: captureMode,
      ...payloadContext(),
      duration_ms: pose.duration_ms,
      frames: pose.frames,
    });
    setResult(analyzed);
    setHistory(current => [analyzed, ...current.filter(x => x.id !== analyzed.id)].slice(0, 8));
    setStage('');
  }

  async function recordAndAnalyze() {
    if (!selected) return;
    setBusy(true); setError(''); setResult(null); setPosePreview(null); setNativeLiveCount(null); let path = '';
    try {
      setStage('Запишите 3–8 спокойных повторений. Всё тело должно быть видно в кадре.');
      path = await techniqueVideo.record(45);
      setStage('Ищу точки тела на устройстве…');
      const pose = await techniqueVideo.analyze(path, 120);
      await submitPose('recorded', pose);
    } catch (e) { setError(e instanceof Error ? e.message : 'Анализ не удался'); setStage(''); }
    finally { if (path) await techniqueVideo.cleanup(path).catch(() => undefined); setBusy(false); }
  }

  async function liveAndAnalyze() {
    if (!selected) return;
    setBusy(true); setError(''); setResult(null); setPosePreview(null); setNativeLiveCount(null);
    try {
      setStage('Открываю live-камеру. Скелет и счётчик повторений появятся прямо поверх изображения.');
      const pose = await techniqueLive.start(selected.key, 45);
      setStage('Сверяю live-счётчик с серверным Technique Engine…');
      await submitPose('live', pose, pose.live_rep_count);
    } catch (e) { setError(e instanceof Error ? e.message : 'Live-анализ не удался'); setStage(''); }
    finally { setBusy(false); }
  }

  const linkedUnsupported = Boolean(workoutContext && !linkedKey);

  return <ScrollView contentContainerStyle={styles.container}>
    <AppButton title="← Назад" variant="secondary" onPress={onBack} testID="technique-back" />
    <Text style={styles.title}>Анализ техники</Text>
    <Text style={styles.lead}>Pose Landmarks извлекаются на телефоне. На сервер отправляются координаты суставов и время кадров — исходное live-видео не загружается.</Text>

    {workoutContext ? <View style={styles.linkedCard}>
      <Text style={styles.linkedEyebrow}>ТЕКУЩИЙ ПОДХОД</Text>
      <Text style={styles.linkedTitle}>{workoutContext.exerciseName}</Text>
      <Text style={styles.muted}>Подход {workoutContext.setNumber} · результат можно вернуть в тренировку как число повторений.</Text>
    </View> : null}

    <Text style={styles.section}>Упражнение</Text>
    {linkedUnsupported ? <Text style={styles.warning}>Для этого упражнения Technique CV пока не откалиброван. Сейчас поддерживаются первые 5 паттернов.</Text> :
      <View style={styles.grid}>{exercises.map(item => <AppButton key={item.key} title={item.name} variant={selected?.key === item.key ? 'primary' : 'secondary'} onPress={() => !workoutContext && setSelected(item)} disabled={Boolean(workoutContext && item.key !== linkedKey)} />)}</View>}

    <View style={styles.tip}><Text style={styles.tipText}>Live-режим подсказывает, если тело выходит из кадра, рисует скелет и считает повторения прямо во время подхода. Сервер после завершения повторно пересчитывает метрики — его результат считается итоговым.</Text></View>

    <View style={styles.actions}>
      <AppButton title={busy ? 'Анализирую…' : '● Live-анализ'} onPress={() => void liveAndAnalyze()} disabled={busy || !selected || !techniqueLive.available()} testID="technique-live" />
      <AppButton title="Записать видео" variant="secondary" onPress={() => void recordAndAnalyze()} disabled={busy || !selected || !techniqueVideo.available()} testID="technique-record" />
    </View>
    {!techniqueLive.available() && <Text style={styles.warning}>Live-режим требует Android native build с CameraX.</Text>}
    {!!stage && <View style={styles.status}><ActivityIndicator/><Text style={styles.statusText}>{stage}</Text></View>}
    {!!error && <Text style={styles.error}>{error}</Text>}
    {posePreview && <PoseSkeletonPreview frame={posePreview} />}
    {result && <ResultCard result={result} nativeLiveCount={nativeLiveCount} />}
    {result && workoutContext && onUseLinkedResult ? <AppButton title={`Использовать ${result.rep_count} повторений в подходе`} onPress={() => onUseLinkedResult(result)} testID="technique-use-result" /> : null}

    <Text style={styles.section}>Последние анализы</Text>
    {history.length === 0 ? <Text style={styles.muted}>Пока нет записей.</Text> : history.map(item => <View key={item.id} style={styles.history}><View style={{flex: 1}}><Text style={styles.historyTitle}>{item.exercise_name}</Text><Text style={styles.muted}>{new Date(item.created_at).toLocaleString()} · {item.rep_count} повт. · {item.capture_mode === 'live' ? 'LIVE' : 'VIDEO'}</Text></View><Text style={styles.scoreSmall}>{item.technique_score}</Text></View>)}
  </ScrollView>;
}

function ResultCard({result, nativeLiveCount}: {result: TechniqueResult; nativeLiveCount: number | null}) {
  const countDiff = nativeLiveCount !== null && nativeLiveCount !== result.rep_count;
  return <View style={styles.result}>
    <View style={styles.resultHeader}><Text style={styles.resultTitle}>{result.exercise_name}</Text><Text style={styles.mode}>{result.capture_mode === 'live' ? 'LIVE' : 'VIDEO'}</Text></View>
    <Text style={styles.score}>{result.technique_score}<Text style={styles.scoreOf}> / 100</Text></Text>
    <Text style={styles.reps}>{result.rep_count} повторений · уверенность {Math.round(result.confidence * 100)}%</Text>
    {result.rep_count > 0 ? <Text style={styles.phaseSummary}>Средний темп: эксцентрическая {(result.average_eccentric_ms / 1000).toFixed(1)}с · концентрическая {(result.average_concentric_ms / 1000).toFixed(1)}с · {result.average_rep_rpm.toFixed(1)} повт/мин</Text> : null}
    {countDiff ? <Text style={styles.liveNote}>Live-счётчик показал {nativeLiveCount}; итоговый движок после полного ряда landmarks определил {result.rep_count}.</Text> : null}
    <View style={styles.metrics}><Metric name="Амплитуда" value={result.rom_score}/><Metric name="Темп" value={result.tempo_score}/><Metric name="Симметрия" value={result.symmetry_score}/><Metric name="Стабильность" value={result.stability_score}/></View>
    {result.reps.slice(0, 8).map(rep => <Text key={rep.number} style={styles.repLine}>#{rep.number} · {(rep.duration_ms / 1000).toFixed(1)}с · эксц {(rep.eccentric_ms / 1000).toFixed(1)}с · конц {(rep.concentric_ms / 1000).toFixed(1)}с · ROM {Math.round(rep.rom_degrees)}°</Text>)}
    {result.feedback.map((line, i) => <Text key={i} style={styles.feedback}>• {line}</Text>)}
    <Text style={styles.disclaimer}>Оценка предназначена для тренировочной обратной связи и не является медицинской или клинической оценкой движения.</Text>
  </View>;
}
function Metric({name,value}:{name:string;value:number}) {return <View style={styles.metric}><Text style={styles.metricValue}>{value}</Text><Text style={styles.metricName}>{name}</Text></View>}
const styles=StyleSheet.create({container:{padding:20,paddingBottom:42,gap:14,backgroundColor:'#fff'},title:{fontSize:30,fontWeight:'900',color:'#111'},lead:{fontSize:15,lineHeight:22,color:'#444'},section:{fontSize:19,fontWeight:'800',color:'#111',marginTop:8},grid:{gap:8},actions:{gap:8},tip:{backgroundColor:'#f4f4f5',borderRadius:16,padding:14},tipText:{fontSize:14,lineHeight:20,color:'#333'},linkedCard:{borderWidth:1,borderColor:'#d8d8d8',borderRadius:18,padding:14,gap:4},linkedEyebrow:{fontSize:11,fontWeight:'900',letterSpacing:1,color:'#777'},linkedTitle:{fontSize:20,fontWeight:'900',color:'#111'},warning:{color:'#8a5a00',lineHeight:20},status:{flexDirection:'row',alignItems:'center',gap:10,padding:14},statusText:{flex:1,color:'#333'},error:{color:'#b42318',fontWeight:'600'},result:{backgroundColor:'#111',borderRadius:24,padding:20,gap:10},resultHeader:{flexDirection:'row',alignItems:'center',justifyContent:'space-between'},resultTitle:{color:'#fff',fontSize:20,fontWeight:'800'},mode:{color:'#111',backgroundColor:'#fff',fontSize:11,fontWeight:'900',paddingHorizontal:9,paddingVertical:4,borderRadius:10},score:{color:'#fff',fontSize:48,fontWeight:'900'},scoreOf:{fontSize:18,color:'#aaa'},reps:{color:'#ddd'},phaseSummary:{color:'#bbb',fontSize:12,lineHeight:18},liveNote:{color:'#d7f7e2',fontSize:12,lineHeight:18},metrics:{flexDirection:'row',flexWrap:'wrap',gap:8},metric:{backgroundColor:'#222',borderRadius:14,padding:12,minWidth:'46%'},metricValue:{color:'#fff',fontWeight:'900',fontSize:22},metricName:{color:'#aaa',fontSize:12},repLine:{color:'#bbb',fontSize:12},feedback:{color:'#eee',lineHeight:20},disclaimer:{color:'#999',fontSize:11,lineHeight:16,marginTop:6},history:{borderWidth:1,borderColor:'#e5e5e5',borderRadius:16,padding:14,flexDirection:'row',gap:10,justifyContent:'space-between',alignItems:'center'},historyTitle:{fontWeight:'800',color:'#111'},muted:{color:'#777'},scoreSmall:{fontSize:24,fontWeight:'900',color:'#111'}});
