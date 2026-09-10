import React, {useEffect, useState} from 'react';
import {ActivityIndicator, Alert, AppState, BackHandler, StatusBar, StyleSheet, Text, View} from 'react-native';
import {SafeAreaProvider, SafeAreaView} from 'react-native-safe-area-context';
import {api, AuthTokens, FinishResult, OnboardingStatus, TechniqueResult, WorkoutView} from './src/api/client';
import {ActiveWorkoutScreen} from './src/screens/ActiveWorkoutScreen';
import {AuthScreen} from './src/screens/AuthScreen';
import {GoalScreen} from './src/screens/GoalScreen';
import {HistoryScreen} from './src/screens/HistoryScreen';
import {MuscleDetailScreen} from './src/screens/MuscleDetailScreen';
import {HomeScreen} from './src/screens/HomeScreen';
import {ProfileSetupScreen} from './src/screens/ProfileSetupScreen';
import {TrainingSetupScreen} from './src/screens/TrainingSetupScreen';
import {WorkoutDetailScreen} from './src/screens/WorkoutDetailScreen';
import {WorkoutPreviewScreen} from './src/screens/WorkoutPreviewScreen';
import {WorkoutSummaryScreen} from './src/screens/WorkoutSummaryScreen';
import {NutritionScreen} from './src/screens/NutritionScreen';
import {NutritionSetupScreen} from './src/screens/NutritionSetupScreen';
import {FoodSearchScreen} from './src/screens/FoodSearchScreen';
import {CustomFoodScreen} from './src/screens/CustomFoodScreen';
import {RecipesScreen} from './src/screens/RecipesScreen';
import {ProgressScreen} from './src/screens/ProgressScreen';
import {AIFoodInputScreen} from './src/screens/AIFoodInputScreen';
import {FoodPhotoScreen} from './src/screens/FoodPhotoScreen';
import {AICoachScreen} from './src/screens/AICoachScreen';
import {WeeklyAIReportScreen} from './src/screens/WeeklyAIReportScreen';
import {ProgramsScreen} from './src/screens/ProgramsScreen';
import {ProgramSetupScreen} from './src/screens/ProgramSetupScreen';
import {ProgramDetailScreen} from './src/screens/ProgramDetailScreen';
import {BodyScanScreen} from './src/screens/BodyScanScreen';
import {TechniqueScreen} from './src/screens/TechniqueScreen';
import {RecoveryScreen} from './src/screens/RecoveryScreen';
import {ConnectedDevicesScreen} from './src/screens/ConnectedDevicesScreen';
import {ManualWorkoutScreen} from './src/screens/ManualWorkoutScreen';
import {AthleteProfileScreen} from './src/screens/AthleteProfileScreen';
import {sessionStorage} from './src/storage/session';
import {flushOfflineQueue} from './src/storage/sync';
import {MuscleId} from './src/domain/muscles';
import {TechniqueWorkoutContext} from './src/domain/technique';
import {BottomNavigation, MainTab} from './src/components/BottomNavigation';
import {refreshTodayIfPossible} from './src/health/sync';
import {healthConnect} from './src/native/healthConnect';

type Step = 'auth' | 'goal' | 'profile' | 'training' | 'profileSettings' | 'home' | 'muscle' | 'manualWorkout' | 'preview' | 'active' | 'summary' | 'history' | 'workoutDetail' | 'nutrition' | 'nutritionSetup' | 'foodSearch' | 'customFood' | 'recipes' | 'progress' | 'aiFood' | 'foodPhoto' | 'aiCoach' | 'weeklyAI' | 'programs' | 'programSetup' | 'programDetail' | 'bodyScan' | 'technique' | 'recovery' | 'devices';
type Environment = 'home' | 'gym' | 'band';

function nextOnboardingStep(status: OnboardingStatus): Step {
  if (status.completed) return 'home';
  if (!status.goal_completed) return 'goal';
  if (!status.profile_completed) return 'profile';
  return 'training';
}

function localFinished(workout: WorkoutView): FinishResult {
  let volume = 0;
  let targetSets = 0;
  let completedSets = 0;
  for (const exercise of workout.exercises) {
    targetSets += exercise.workout_exercise.target_sets;
    completedSets += Math.min(exercise.sets.length, exercise.workout_exercise.target_sets);
    for (const set of exercise.sets) {
      if (set.weight !== undefined) volume += set.weight * set.repetitions;
    }
  }
  const completionPercent = targetSets > 0 ? Math.round((completedSets / targetSets) * 10000) / 100 : 0;
  return {
    workout: {
      ...workout,
      workout: {
        ...workout.workout,
        status: 'completed',
        completed_at: new Date().toISOString(),
        total_volume: volume,
        completion_percent: completionPercent,
        ended_early: completedSets < targetSets,
      },
    },
    personal_records: [],
    next_recommendations: [],
  };
}

function savedSessionMayRunOffline(tokens: AuthTokens | null) {
  if (!tokens) return false;
  const expires = Date.parse(tokens.access_expires_at);
  return Number.isFinite(expires) && Date.now() - expires < 7 * 24 * 60 * 60 * 1000;
}

export default function App() {
  const [step, setStep] = useState<Step>('auth');
  const [tokens, setTokens] = useState<AuthTokens | null>(null);
  const [workout, setWorkout] = useState<WorkoutView | null>(null);
  const [summary, setSummary] = useState<FinishResult | null>(null);
  const [booting, setBooting] = useState(true);
  const [systemMessage, setSystemMessage] = useState('');
  const [selectedMuscle, setSelectedMuscle] = useState<MuscleId | null>(null);
  const [selectedEnvironment, setSelectedEnvironment] = useState<Environment>('gym');
  const [selectedWorkoutId, setSelectedWorkoutId] = useState<string | null>(null);
  const [selectedProgramId, setSelectedProgramId] = useState<string | null>(null);
  const [techniqueContext, setTechniqueContext] = useState<TechniqueWorkoutContext | null>(null);
  const [techniquePrefill, setTechniquePrefill] = useState<{workoutExerciseId: string; setNumber: number; repCount: number; analysisId: string} | null>(null);
  const [devicesOrigin, setDevicesOrigin] = useState<'home' | 'progress'>('home');

  const access = tokens?.access_token ?? '';

  async function revokeLocalSession(clearEnvironmentCache = false) {
    const ownerUserId = await sessionStorage.currentUserId();
    if (ownerUserId) {
      try { await healthConnect.clearPassiveSync(ownerUserId); } catch { /* credential revocation must still complete */ }
      if (clearEnvironmentCache) await sessionStorage.clearTrainingEnvironments(ownerUserId);
    }
    if (clearEnvironmentCache) await sessionStorage.clearTokens();
    else await sessionStorage.revokeCurrentSession();
  }

  useEffect(() => {
    api.configureAuthSession({
      loadTokens: () => sessionStorage.loadTokens(),
      saveTokens: tokens => sessionStorage.saveTokens(tokens),
      clearTokens: () => revokeLocalSession(false),
      onTokensChanged: next => {
        setTokens(next);
        if (!next) {
          setWorkout(null);
          setSummary(null);
          setStep('auth');
        }
      },
    });
    void bootstrap();
    return () => api.configureAuthSession(null);
  }, []);

  useEffect(() => {
    const subscription = AppState.addEventListener('change', state => {
      if (state === 'active' && tokens) void (async()=>{const latest=(await syncPending(tokens))??tokens;await refreshTodayIfPossible(latest.access_token)})();
    });
    return () => subscription.remove();
  }, [tokens]);

  useEffect(() => {
    const subscription = BackHandler.addEventListener('hardwareBackPress', () => {
      if (step === 'preview') {
        setStep(selectedMuscle ? 'muscle' : 'home');
        return true;
      }
      if (step === 'manualWorkout') { setStep('home'); return true; }
      if (step === 'muscle' || step === 'history' || step === 'nutrition' || step === 'progress' || step === 'aiCoach' || step === 'programs') {
        setStep('home');
        return true;
      }
      if (step === 'nutritionSetup' || step === 'foodSearch' || step === 'customFood' || step === 'recipes' || step === 'aiFood') {
        setStep('nutrition');
        return true;
      }
      if (step === 'foodPhoto') { setStep('aiFood'); return true; }
      if (step === 'weeklyAI') { setStep('aiCoach'); return true; }
      if (step === 'bodyScan') { setStep('progress'); return true; }
      if (step === 'devices') { setStep(devicesOrigin); return true; }
      if (step === 'recovery' || step === 'profileSettings') { setStep('home'); return true; }
      if (step === 'technique') { setStep(techniqueContext ? 'active' : 'progress'); setTechniqueContext(null); return true; }
      if (step === 'programSetup' || step === 'programDetail') { setStep('programs'); return true; }
      if (step === 'workoutDetail') {
        setStep('history');
        return true;
      }
      if (step === 'summary') {
        setSummary(null);
        setWorkout(null);
        setStep('home');
        return true;
      }
      if (step === 'active') return true;
      return false;
    });
    return () => subscription.remove();
  }, [step, selectedMuscle, techniqueContext, devicesOrigin]);

  async function getOnboarding(current: AuthTokens) {
    const status = await api.onboardingStatus(current.access_token);
    const latest = (await sessionStorage.loadTokens()) ?? current;
    return {status, tokens: latest};
  }

  async function bootstrap() {
    let saved: AuthTokens | null = null;
    try {
      setBooting(true);
      saved = await sessionStorage.loadTokens();
      if (!saved) {
        setStep('auth');
        return;
      }
      setTokens(saved);
      api.setCurrentTokens(saved);
      const bootstrapOwner = await sessionStorage.currentUserId();
      if (bootstrapOwner) { try { await healthConnect.bindSessionOwner(bootstrapOwner); } catch { /* health integration must not block auth */ } }
      const synced = await syncPending(saved);
      const current = synced ?? saved;
      await refreshTodayIfPossible(current.access_token);
      const onboarding = await getOnboarding(current);
      if (!onboarding.status.completed) {
        setStep(nextOnboardingStep(onboarding.status));
        return;
      }
      try {
        const active = await api.activeWorkout(onboarding.tokens.access_token);
        if (active) {
          setWorkout(active);
          await sessionStorage.saveActiveWorkout(active);
          setStep('active');
          return;
        }
        await sessionStorage.saveActiveWorkout(null);
        setStep('home');
      } catch (error) {
        if (api.isNetworkError(error)) {
          const cached = await sessionStorage.loadActiveWorkout();
          if (cached?.workout.status === 'active') {
            setWorkout(cached);
            setSystemMessage('Работаем офлайн. Активная тренировка восстановлена с телефона.');
            setStep('active');
            return;
          }
        }
        throw error;
      }
    } catch (error) {
      if (api.isUnauthorized(error)) {
        await revokeLocalSession(false);
        api.setCurrentTokens(null);
        setTokens(null);
        setStep('auth');
      } else {
        const cached = await sessionStorage.loadActiveWorkout();
        if (cached?.workout.status === 'active') {
          setWorkout(cached);
          setSystemMessage('Сервер недоступен. Тренировка открыта в офлайн-режиме.');
          setStep('active');
        } else if (api.isNetworkError(error) && savedSessionMayRunOffline(saved)) {
          setSystemMessage('Нет сети. Вы остались в профиле; данные синхронизируются позже.');
          setStep('home');
        } else {
          setSystemMessage('Не удалось связаться с сервером. Повторим синхронизацию при возвращении приложения.');
          setStep('auth');
        }
      }
    } finally {
      setBooting(false);
    }
  }

  async function syncPending(current: AuthTokens): Promise<AuthTokens | null> {
    try {
      const result = await flushOfflineQueue(current);
      const latest = (await sessionStorage.loadTokens()) ?? result.tokens;
      if (latest.access_token !== current.access_token) {
        setTokens(latest);
        api.setCurrentTokens(latest);
      }
      if (result.workout?.workout.status === 'active') {
        setWorkout(result.workout);
      }
      if (result.remaining === 0) setSystemMessage('');
      else setSystemMessage(`${result.remaining} действий ожидают синхронизации.`);
      return latest;
    } catch (error) {
      if (api.isNetworkError(error)) {
        const queue = await sessionStorage.loadQueue();
        if (queue.length > 0) setSystemMessage(`${queue.length} действий сохранены офлайн.`);
        return current;
      }
      return current;
    }
  }

  async function handleAuth(mode: 'register' | 'login', email: string, password: string) {
    const result = mode === 'register' ? await api.register(email, password) : await api.login(email, password);
    setTokens(result.tokens);
    api.setCurrentTokens(result.tokens);
    await sessionStorage.saveTokens(result.tokens, result.user.id);
    try { await healthConnect.bindSessionOwner(result.user.id); } catch { /* health integration must not block login */ }
    if (mode === 'register') {
      setStep('goal');
      return;
    }
    const onboarding = await getOnboarding(result.tokens);
    if (onboarding.status.completed) {
      const active = await api.activeWorkout(onboarding.tokens.access_token);
      if (active) {
        setWorkout(active);
        await sessionStorage.saveActiveWorkout(active);
        setStep('active');
      } else {
        setStep('home');
      }
      return;
    }
    setStep(nextOnboardingStep(onboarding.status));
  }

  async function updateWorkout(next: WorkoutView) {
    setWorkout(next);
    await sessionStorage.saveActiveWorkout(next.workout.status === 'active' ? next : null);
  }

  async function finishWorkout() {
    if (!workout || !tokens) return;
    const currentTokens = (await syncPending(tokens)) ?? tokens;
    try {
      const result = await api.finishWorkout(currentTokens.access_token, workout.workout.id);
      setSummary(result);
      setWorkout(result.workout);
      await sessionStorage.saveActiveWorkout(null);
      setTechniquePrefill(null);
      setStep('summary');
    } catch (error) {
      if (!api.isNetworkError(error)) {
        setSystemMessage(error instanceof Error ? error.message : 'Не удалось завершить тренировку.');
        return;
      }
      await sessionStorage.enqueueFinish(workout.workout.id);
      const result = localFinished(workout);
      setSummary(result);
      setWorkout(result.workout);
      await sessionStorage.saveActiveWorkout(null);
      setTechniquePrefill(null);
      setSystemMessage('Тренировка завершена офлайн. Она появится в истории после синхронизации.');
      setStep('summary');
    }
  }

  async function cancelWorkout() {
    if (!workout || !tokens) return;
    const workoutID = workout.workout.id;
    const currentTokens = (await syncPending(tokens)) ?? tokens;
    try {
      await api.cancelWorkout(currentTokens.access_token, workoutID);
      await sessionStorage.removeWorkoutOperations(workoutID);
      await sessionStorage.saveActiveWorkout(null);
      setWorkout(null);
      setSelectedMuscle(null);
      setTechniquePrefill(null);
      setSystemMessage('Тренировка отменена.');
      setStep('home');
    } catch (error) {
      if (!api.isNetworkError(error)) {
        setSystemMessage(error instanceof Error ? error.message : 'Не удалось отменить тренировку.');
        return;
      }
      await sessionStorage.enqueueCancel(workoutID);
      await sessionStorage.saveActiveWorkout(null);
      setWorkout(null);
      setSelectedMuscle(null);
      setTechniquePrefill(null);
      setSystemMessage('Отмена сохранена офлайн и будет синхронизирована при появлении сети.');
      setStep('home');
    }
  }

  async function performLogout() {
    if (tokens) void api.logout(tokens.refresh_token).catch(() => undefined);
    await revokeLocalSession(true);
    api.setCurrentTokens(null);
    setTokens(null);
    setWorkout(null);
    setSummary(null);
    setSystemMessage('');
    setTechniqueContext(null);
    setTechniquePrefill(null);
    setStep('auth');
  }

  async function logout() {
    const pending = await sessionStorage.loadQueue();
    const active = await sessionStorage.loadActiveWorkout();
    if (pending.length || active) {
      Alert.alert('Выйти из аккаунта?', `${pending.length ? `Не синхронизировано действий: ${pending.length}. ` : ''}${active ? 'Есть активная тренировка. ' : ''}Данные останутся на телефоне.`, [{text:'Отмена',style:'cancel'},{text:'Выйти',style:'destructive',onPress:()=>void performLogout()}]);
      return;
    }
    await performLogout();
  }

  const mainTab: MainTab | null =
    step === 'home' ? 'home' :
    step === 'programs' ? 'training' :
    step === 'nutrition' ? 'nutrition' :
    step === 'progress' ? 'progress' :
    step === 'aiCoach' ? 'ai' : null;

  function navigateMain(tab: MainTab) {
    if (tab === 'home') setStep('home');
    else if (tab === 'training') setStep('programs');
    else if (tab === 'nutrition') setStep('nutrition');
    else if (tab === 'progress') setStep('progress');
    else setStep('aiCoach');
  }

  if (booting) {
    return <SafeAreaProvider><SafeAreaView style={styles.loading} edges={['top', 'bottom']}><ActivityIndicator /><Text style={styles.loadingText}>Восстанавливаю профиль и тренировку…</Text></SafeAreaView></SafeAreaProvider>;
  }

  return (
    <SafeAreaProvider>
      <SafeAreaView style={styles.shell} edges={['top', 'left', 'right', 'bottom']}>
      <StatusBar barStyle="dark-content" />
      {systemMessage ? <View style={styles.banner}><Text style={styles.bannerText}>{systemMessage}</Text></View> : null}
      <View style={styles.content}>
      {step === 'auth' && <AuthScreen onSubmit={handleAuth} />}
      {step === 'goal' && <GoalScreen onContinue={async goal => {await api.setGoal(access, goal); setStep('profile');}} />}
      {step === 'profile' && <ProfileSetupScreen onContinue={async data => {
        await api.updateProfile(access, {height_cm: data.height, weight_kg: data.weight, age_years: data.age, experience_level: data.level, injuries: data.injuries, limitations: data.limitations, unit_system: 'metric'});
        setStep('training');
      }} />}
      {step === 'training' && <TrainingSetupScreen onFinish={async data => {
        await api.setTrainingPreferences(access, {environments: data.environments, equipment_ids: data.equipment, workouts_per_week: data.workouts, session_minutes: data.minutes});
        await api.completeOnboarding(access);
        setStep('home');
      }} />}
      {step === 'home' && <HomeScreen accessToken={access} onMuscle={(muscle, environment) => {setSelectedMuscle(muscle); setSelectedEnvironment(environment); setStep('muscle');}} onManualWorkout={() => {setSelectedMuscle(null); setStep('manualWorkout');}} onHistory={() => setStep('history')} onPrograms={() => setStep('programs')} onAI={() => setStep('aiCoach')} onRecovery={() => setStep('recovery')} onDevices={() => {setDevicesOrigin('home');setStep('devices')}} onProfile={() => setStep('profileSettings')} />}
      {step === 'manualWorkout' && <ManualWorkoutScreen accessToken={access} onBack={() => setStep('home')} onCreated={next => {setWorkout(next);setSelectedMuscle(null);setSelectedEnvironment(next.workout.environment);setStep('preview')}} />}
      {step === 'profileSettings' && <AthleteProfileScreen accessToken={access} onBack={() => setStep('home')} onLogout={logout} />}
      {step === 'muscle' && selectedMuscle && <MuscleDetailScreen accessToken={access} muscle={selectedMuscle} environment={selectedEnvironment} onBack={() => setStep('home')} onWorkout={next => {setWorkout(next); setStep('preview');}} />}
      {step === 'preview' && workout && <WorkoutPreviewScreen accessToken={access} workout={workout} onBack={() => setStep(selectedMuscle ? 'muscle' : 'home')} onStart={next => {void updateWorkout(next); setStep('active');}} />}
      {step === 'active' && workout && <ActiveWorkoutScreen accessToken={access} workout={workout} onWorkoutChange={next => {void updateWorkout(next);}} onFinish={finishWorkout} onCancel={cancelWorkout} techniquePrefill={techniquePrefill} onTechnique={context => {setTechniqueContext(context); setStep('technique');}} />}
      {step === 'summary' && summary && <WorkoutSummaryScreen result={summary} onDone={() => {setWorkout(null); setSummary(null); setSelectedMuscle(null); setStep('home');}} />}
      {step === 'history' && <HistoryScreen accessToken={access} onBack={() => setStep('home')} onSelect={workoutId => {setSelectedWorkoutId(workoutId); setStep('workoutDetail');}} />}
      {step === 'workoutDetail' && selectedWorkoutId && <WorkoutDetailScreen accessToken={access} workoutId={selectedWorkoutId} onBack={() => setStep('history')} onRepeat={next => {setWorkout(next); setSelectedMuscle(next.workout.muscle as MuscleId); setSelectedEnvironment(next.workout.environment); setStep('preview');}} />}
      {step === 'nutrition' && <NutritionScreen accessToken={access} onBack={() => setStep('home')} onSetup={() => setStep('nutritionSetup')} onAddFood={() => setStep('foodSearch')} onRecipes={() => setStep('recipes')} onAI={() => setStep('aiFood')} />}
      {step === 'nutritionSetup' && <NutritionSetupScreen accessToken={access} onBack={() => setStep('nutrition')} onSaved={() => setStep('nutrition')} />}
      {step === 'foodSearch' && <FoodSearchScreen accessToken={access} onBack={() => setStep('nutrition')} onLogged={() => setStep('nutrition')} onCustom={() => setStep('customFood')} onRecipes={() => setStep('recipes')} />}
      {step === 'customFood' && <CustomFoodScreen accessToken={access} onBack={() => setStep('foodSearch')} onSaved={() => setStep('foodSearch')} />}
      {step === 'recipes' && <RecipesScreen accessToken={access} onBack={() => setStep('nutrition')} onLogged={() => setStep('nutrition')} />}
      {/* Keep the analytics tree mounted behind its tools. Legacy navigation contract: onDevices={() => setStep('devices')} */}
      {(step === 'progress' || step === 'devices' || step === 'bodyScan' || (step === 'technique' && !techniqueContext)) && <View style={step === 'progress' ? styles.content : styles.hidden}><ProgressScreen accessToken={access} onBack={() => setStep('home')} onBodyScan={() => setStep('bodyScan')} onTechnique={() => {setTechniqueContext(null);setStep('technique')}} onDevices={() => {setDevicesOrigin('progress');setStep('devices')}} /></View>}
      {step === 'recovery' && <RecoveryScreen accessToken={access} onBack={() => setStep('home')} />}
      {step === 'devices' && <ConnectedDevicesScreen accessToken={access} onBack={() => setStep(devicesOrigin)} />}
      {step === 'bodyScan' && <BodyScanScreen accessToken={access} onBack={() => setStep('progress')} />}
      {step === 'technique' && <TechniqueScreen accessToken={access} workoutContext={techniqueContext} onBack={() => {setStep(techniqueContext ? 'active' : 'progress'); setTechniqueContext(null);}} onUseLinkedResult={(result: TechniqueResult) => {if (!techniqueContext) return; setTechniquePrefill({workoutExerciseId: techniqueContext.workoutExerciseId, setNumber: techniqueContext.setNumber, repCount: result.rep_count, analysisId: result.id}); setTechniqueContext(null); setStep('active');}} />}
      {step === 'aiFood' && <AIFoodInputScreen accessToken={access} onBack={() => setStep('nutrition')} onPhoto={() => setStep('foodPhoto')} onDone={() => setStep('nutrition')} />}
      {step === 'foodPhoto' && <FoodPhotoScreen accessToken={access} onBack={() => setStep('aiFood')} onDone={() => setStep('nutrition')} />}
      {step === 'aiCoach' && <AICoachScreen accessToken={access} onBack={() => setStep('home')} onWeekly={() => setStep('weeklyAI')} />}
      {step === 'weeklyAI' && <WeeklyAIReportScreen accessToken={access} onBack={() => setStep('aiCoach')} />}
      {step === 'programs' && <ProgramsScreen accessToken={access} onBack={() => setStep('home')} onCreate={() => setStep('programSetup')} onOpen={programId => {setSelectedProgramId(programId); setStep('programDetail');}} />}
      {step === 'programSetup' && <ProgramSetupScreen accessToken={access} onBack={() => setStep('programs')} onCreated={program => {setSelectedProgramId(program.program.id); setStep('programDetail');}} />}
      {step === 'programDetail' && selectedProgramId && <ProgramDetailScreen accessToken={access} programId={selectedProgramId} onBack={() => setStep('programs')} onArchived={() => {setSelectedProgramId(null); setStep('programs');}} onWorkout={(next, muscle) => {setWorkout(next); setSelectedMuscle(muscle); setSelectedEnvironment(next.workout.environment); setStep('preview');}} />}
      </View>
      {mainTab ? <BottomNavigation current={mainTab} onSelect={navigateMain} /> : null}
      </SafeAreaView>
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  shell: {flex: 1},
  content: {flex: 1},
  hidden: {display: 'none'},
  loading: {flex: 1, alignItems: 'center', justifyContent: 'center', gap: 12},
  loadingText: {fontWeight: '700', opacity: 0.6},
  banner: {paddingHorizontal: 16, paddingVertical: 9, borderBottomWidth: 1},
  bannerText: {fontSize: 12, lineHeight: 17, fontWeight: '600'},
});
