import React, {useEffect, useState} from 'react';
import {ActivityIndicator, Alert, Pressable, ScrollView, StyleSheet, Text, View} from 'react-native';
import {api, Readiness} from '../api/client';
import {currentLocalDate} from '../domain/date';
import {AppButton} from '../components/AppButton';
import {BodyMap} from '../components/BodyMap';
import {MuscleId, muscleIds, muscleMeta} from '../domain/muscles';
import {colors, control, radius, spacing} from '../theme/tokens';

type Environment = 'home' | 'gym' | 'band';

const environments: Array<{id: Environment; title: string}> = [
  {id: 'home', title: 'Дома'},
  {id: 'gym', title: 'Зал'},
  {id: 'band', title: 'Резина'},
];

export function HomeScreen({
  accessToken,
  onMuscle,
  onHistory,
  onPrograms,
  onAI,
  onRecovery,
  onDevices,
  onLogout,
}: {
  accessToken: string;
  onMuscle: (muscle: MuscleId, environment: Environment) => void;
  onHistory: () => void;
  onPrograms: () => void;
  onAI: () => void;
  onRecovery: () => void;
  onDevices: () => void;
  onLogout: () => void;
}) {
  const [environment, setEnvironment] = useState<Environment>('gym');
  const [allowedEnvironments, setAllowedEnvironments] = useState<Environment[]>(['gym']);
  const [readiness, setReadiness] = useState<Readiness | null>(null);
  const [profileLoading, setProfileLoading] = useState(true);
  const [readinessLoading, setReadinessLoading] = useState(true);
  const [readinessError, setReadinessError] = useState(false);

  useEffect(() => {
    setProfileLoading(true);
    api.getProfile(accessToken).then(profile => {
      const allowed = (profile.training_preferences?.environments ?? []) as Environment[];
      if (allowed.length > 0) {
        setAllowedEnvironments(allowed);
        if (!allowed.includes(environment)) setEnvironment(allowed[0]);
      }
    }).catch(() => undefined).finally(() => setProfileLoading(false));
    setReadinessLoading(true); setReadinessError(false);
    api.recoveryToday(accessToken, currentLocalDate()).then(setReadiness).catch(() => setReadinessError(true)).finally(() => setReadinessLoading(false));
  }, [accessToken]);

  function openProfileMenu() {
    Alert.alert('Профиль', 'Настройки аккаунта будут расширены в следующем UI-спринте.', [
      {text: 'Закрыть', style: 'cancel'},
      {text: 'Выйти', style: 'destructive', onPress: onLogout},
    ]);
  }

  return (
    <ScrollView testID="home-screen" contentContainerStyle={styles.container}>
      <View style={styles.topline}>
        <View>
          <Text style={styles.eyebrow}>AI FITNESS OS</Text>
          <Text accessibilityRole="header" style={styles.today}>Сегодня</Text>
        </View>
        <Pressable
          accessibilityRole="button"
          accessibilityLabel="Открыть профиль"
          testID="home-profile"
          hitSlop={4}
          onPress={openProfileMenu}
          style={({pressed}) => [styles.profileButton, pressed && styles.pressed]}>
          <Text style={styles.profileText}>Н</Text>
        </Pressable>
      </View>


      <Pressable
        accessibilityRole="button"
        accessibilityLabel="Открыть восстановление и готовность"
        testID="home-recovery"
        onPress={onRecovery}
        style={({pressed}) => [styles.recoveryCard, pressed && styles.pressed]}>
        <View style={styles.flexShrink}>
          <Text style={styles.programKicker}>READINESS</Text>
          <Text style={styles.recoveryTitle}>{readinessLoading ? 'Считаем…' : readinessError ? 'Нет данных' : readiness?.check_in_completed ? `${readiness.score}/100` : 'Отметить состояние'}</Text>
          <Text style={styles.programMeta}>{readinessLoading ? 'Загружаем показатели готовности' : readinessError ? 'Открой, чтобы повторить загрузку' : readiness?.check_in_completed ? `Объём ×${readiness.volume_multiplier.toFixed(2)} · интенсивность ×${readiness.intensity_multiplier.toFixed(2)}` : 'Сон · энергия · стресс · болезненность'}</Text>
        </View>
        <Text style={styles.programArrow}>→</Text>
      </Pressable>

      <Pressable
        accessibilityRole="button"
        accessibilityLabel="Открыть подключённые устройства"
        testID="home-devices"
        onPress={onDevices}
        style={({pressed}) => [styles.deviceCard, pressed && styles.pressed]}>
        <View style={styles.flexShrink}>
          <Text style={styles.programKicker}>XIAOMI / HEALTH CONNECT</Text>
          <Text style={styles.deviceTitle}>{readinessLoading ? 'Проверяем подключение…' : readiness?.wearable ? 'Часы синхронизированы' : 'Подключить часы'}</Text>
          <Text style={styles.programMeta}>{readiness?.wearable ? `${readiness.wearable.source_label} · сон ${Math.floor(readiness.wearable.sleep_minutes/60)} ч ${readiness.wearable.sleep_minutes%60} мин` : 'Xiaomi Watch S3 · сон · шаги · тренировки'}</Text>
        </View>
        <Text style={styles.programArrow}>→</Text>
      </Pressable>

      <Pressable
        accessibilityRole="button"
        accessibilityLabel="Спросить AI тренера"
        testID="home-ai-coach"
        onPress={onAI}
        style={({pressed}) => [styles.aiCard, pressed && styles.pressed]}>
        <View style={styles.flexShrink}>
          <Text style={styles.aiKicker}>AI COACH</Text>
          <Text style={styles.aiTitle}>Спросить персонального тренера</Text>
        </View>
        <Text style={styles.aiArrow}>→</Text>
      </Pressable>

      <Pressable
        accessibilityRole="button"
        accessibilityLabel="Открыть тренировочную программу"
        testID="home-program"
        onPress={onPrograms}
        style={({pressed}) => [styles.programCard, pressed && styles.pressed]}>
        <View style={styles.flexShrink}>
          <Text style={styles.programKicker}>ПРОГРАММА</Text>
          <Text style={styles.programTitle}>План на 4 / 8 / 12 недель</Text>
          <Text style={styles.programMeta}>Календарь · deload · adherence · автоперенос</Text>
        </View>
        <Text style={styles.programArrow}>→</Text>
      </Pressable>

      <Text accessibilityRole="header" style={styles.title}>Что тренируем сегодня?</Text>
      <Text style={styles.subtitle}>Выбери место тренировки, затем группу мышц. На следующем экране увидишь историю нагрузки и персональную тренировку.</Text>

      {profileLoading ? <View accessibilityRole="progressbar" accessibilityLabel="Загрузка мест тренировки" style={styles.inlineLoading}><ActivityIndicator size="small"/><Text style={styles.programMeta}>Загружаем доступные места…</Text></View> : null}
      <View style={styles.segment} accessibilityRole="tablist">
        {environments.filter(item => allowedEnvironments.includes(item.id)).map(item => {
          const selected = environment === item.id;
          return (
            <Pressable
              key={item.id}
              accessibilityRole="tab"
              accessibilityLabel={`Место тренировки: ${item.title}`}
              accessibilityState={{selected, disabled: profileLoading}}
              testID={`environment-${item.id}`}
              disabled={profileLoading}
              onPress={() => setEnvironment(item.id)}
              style={[styles.segmentItem, selected && styles.segmentItemActive, profileLoading && styles.disabled]}>
              <Text style={[styles.segmentText, selected && styles.segmentTextActive]}>{item.title}</Text>
            </Pressable>
          );
        })}
      </View>

      <BodyMap onSelect={muscle => onMuscle(muscle, environment)} />

      <Text style={styles.sectionTitle}>Все группы мышц</Text>
      <View style={styles.grid}>
        {muscleIds.map(id => (
          <Pressable
            key={id}
            accessibilityRole="button"
            accessibilityLabel={`Открыть тренировку: ${muscleMeta[id].title}`}
            testID={`muscle-${id}`}
            onPress={() => onMuscle(id, environment)}
            style={({pressed}) => [styles.card, pressed && styles.pressed]}>
            <Text style={styles.cardTitle}>{muscleMeta[id].title}</Text>
            <Text style={styles.cardMeta}>Открыть →</Text>
          </Pressable>
        ))}
      </View>

      <AppButton label="История тренировок" variant="secondary" testID="home-history" onPress={onHistory} />
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {padding: spacing.lg, paddingBottom: spacing.xl, gap: spacing.md, backgroundColor: colors.background},
  topline: {flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center'},
  eyebrow: {fontSize: 11, fontWeight: '900', letterSpacing: 1.5, color: colors.textMuted},
  today: {fontSize: 26, fontWeight: '900', color: colors.text, marginTop: 2},
  profileButton: {width: control.minTouch, height: control.minTouch, borderRadius: radius.pill, backgroundColor: colors.primary, alignItems: 'center', justifyContent: 'center'},
  profileText: {color: colors.inverse, fontWeight: '900', fontSize: 17},
  recoveryCard: {borderWidth: 1, borderColor: colors.border, borderRadius: radius.lg, padding: spacing.md, flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', minHeight: 88},
  deviceCard: {borderWidth: 1, borderColor: colors.border, borderRadius: radius.lg, padding: spacing.md, flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', minHeight: 84, backgroundColor: '#f7f7f7'},
  deviceTitle: {fontSize: 16, fontWeight: '900', color: colors.text, marginTop: 4},
  recoveryTitle: {fontSize: 24, fontWeight: '900', color: colors.text, marginTop: 3},
  aiCard: {backgroundColor: colors.primary, borderRadius: radius.lg, padding: spacing.md, flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', minHeight: 88},
  aiKicker: {color: colors.inverse, fontSize: 10, fontWeight: '900', letterSpacing: 1.3, opacity: 0.55},
  aiTitle: {color: colors.inverse, fontSize: 16, fontWeight: '900', marginTop: 4},
  aiArrow: {color: colors.inverse, fontSize: 28, fontWeight: '700'},
  programCard: {borderWidth: 1, borderColor: colors.border, borderRadius: radius.lg, padding: spacing.md, flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', minHeight: 94},
  programKicker: {fontSize: 10, fontWeight: '900', letterSpacing: 1.3, color: colors.textMuted},
  programTitle: {fontSize: 16, fontWeight: '900', marginTop: 4, color: colors.text},
  programMeta: {fontSize: 11, color: colors.textMuted, marginTop: 3},
  programArrow: {fontSize: 28, fontWeight: '700', color: colors.text},
  flexShrink: {flex: 1, paddingRight: spacing.sm},
  title: {fontSize: 30, lineHeight: 36, fontWeight: '900', color: colors.text},
  subtitle: {fontSize: 16, lineHeight: 22, color: colors.textMuted},
  segment: {flexDirection: 'row', padding: spacing.xs, borderWidth: 1, borderColor: colors.border, borderRadius: radius.md},
  segmentItem: {flex: 1, minHeight: control.minTouch, alignItems: 'center', justifyContent: 'center', borderRadius: 12},
  segmentItemActive: {backgroundColor: colors.primary},
  segmentText: {fontWeight: '700', color: colors.text},
  segmentTextActive: {color: colors.inverse},
  sectionTitle: {fontSize: 19, fontWeight: '900', marginTop: 2, color: colors.text},
  grid: {flexDirection: 'row', flexWrap: 'wrap', gap: 10},
  card: {width: '48%', borderWidth: 1, borderColor: colors.border, borderRadius: 18, padding: 15, minHeight: 88, justifyContent: 'space-between'},
  cardTitle: {fontSize: 16, fontWeight: '800', color: colors.text},
  cardMeta: {fontSize: 12, color: colors.textMuted, marginTop: 8},
  pressed: {opacity: 0.7},
  inlineLoading: {minHeight: control.minTouch, flexDirection: 'row', alignItems: 'center', gap: spacing.sm},
  disabled: {opacity: 0.45},
});
