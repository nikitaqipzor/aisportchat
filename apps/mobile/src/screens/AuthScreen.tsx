import React, {useRef, useState} from 'react';
import {
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  type TextInputInstance,
  View,
} from 'react-native';
import {AppButton} from '../components/AppButton';
import {colors, control, radius, spacing} from '../theme/tokens';

type Props = {
  onSubmit: (mode: 'register' | 'login', email: string, password: string) => Promise<void>;
};

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export function AuthScreen({onSubmit}: Props) {
  const passwordRef = useRef<TextInputInstance>(null);
  const [mode, setMode] = useState<'register' | 'login'>('register');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [emailTouched, setEmailTouched] = useState(false);
  const [passwordTouched, setPasswordTouched] = useState(false);
  const [attempted, setAttempted] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const normalizedEmail = email.trim();
  const emailError = !normalizedEmail
    ? 'Введите email'
    : !emailPattern.test(normalizedEmail)
      ? 'Проверьте формат email'
      : '';
  const passwordError = !password
    ? 'Введите пароль'
    : password.length < 8
      ? 'Пароль должен содержать минимум 8 символов'
      : '';
  const valid = !emailError && !passwordError;

  const submit = async () => {
    setAttempted(true);
    if (!valid || loading) {
      return;
    }
    setLoading(true);
    setError('');
    try {
      await onSubmit(mode, normalizedEmail, password);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось выполнить вход. Попробуйте ещё раз.');
    } finally {
      setLoading(false);
    }
  };

  function switchMode() {
    setMode(current => current === 'register' ? 'login' : 'register');
    setAttempted(false);
    setError('');
  }

  return (
    <KeyboardAvoidingView
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
      style={styles.screen}>
      <ScrollView
        testID="auth-screen"
        contentContainerStyle={styles.container}
        keyboardShouldPersistTaps="handled">
        <View style={styles.content}>
          <Text style={styles.eyebrow}>AI FITNESS OS</Text>
          <Text accessibilityRole="header" style={styles.title}>
            {mode === 'register' ? 'Создай профиль' : 'С возвращением'}
          </Text>
          <Text style={styles.subtitle}>
            {mode === 'register'
              ? 'Персональные тренировки с учётом оборудования, самочувствия и прогресса.'
              : 'Войди, чтобы продолжить тренировки и сохранить прогресс.'}
          </Text>

          <View style={styles.field}>
            <Text nativeID="auth-email-label" style={styles.label}>Email</Text>
            <TextInput
              accessibilityLabelledBy="auth-email-label"
              accessibilityHint="Введите адрес электронной почты"
              testID="auth-email"
              autoCapitalize="none"
              autoComplete="email"
              autoCorrect={false}
              keyboardType="email-address"
              returnKeyType="next"
              textContentType="emailAddress"
              placeholder="name@example.com"
              placeholderTextColor={colors.textMuted}
              value={email}
              editable={!loading}
              maxLength={254}
              onBlur={() => setEmailTouched(true)}
              onChangeText={value => {
                setEmail(value);
                setError('');
              }}
              onSubmitEditing={() => passwordRef.current?.focus()}
              style={[styles.input, (attempted || emailTouched) && emailError ? styles.inputError : null]}
            />
            {(attempted || emailTouched) && emailError ? (
              <Text accessibilityRole="alert" style={styles.fieldError}>{emailError}</Text>
            ) : null}
          </View>

          <View style={styles.field}>
            <Text nativeID="auth-password-label" style={styles.label}>Пароль</Text>
            <TextInput
              ref={passwordRef}
              accessibilityLabelledBy="auth-password-label"
              accessibilityHint="Минимум 8 символов"
              testID="auth-password"
              secureTextEntry
              autoCapitalize="none"
              autoCorrect={false}
              autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
              textContentType={mode === 'login' ? 'password' : 'newPassword'}
              returnKeyType="done"
              placeholder="Минимум 8 символов"
              placeholderTextColor={colors.textMuted}
              value={password}
              editable={!loading}
              maxLength={128}
              onBlur={() => setPasswordTouched(true)}
              onChangeText={value => {
                setPassword(value);
                setError('');
              }}
              onSubmitEditing={() => void submit()}
              style={[styles.input, (attempted || passwordTouched) && passwordError ? styles.inputError : null]}
            />
            {(attempted || passwordTouched) && passwordError ? (
              <Text accessibilityRole="alert" style={styles.fieldError}>{passwordError}</Text>
            ) : null}
          </View>

          {error ? (
            <Text accessibilityRole="alert" accessibilityLiveRegion="assertive" style={styles.submitError}>
              {error}
            </Text>
          ) : null}

          <View style={styles.actions}>
            <AppButton
              label={mode === 'register' ? 'Продолжить' : 'Войти'}
              accessibilityLabel={mode === 'register' ? 'Создать аккаунт и продолжить' : 'Войти в аккаунт'}
              testID="auth-submit"
              loading={loading}
              disabled={!valid}
              onPress={() => void submit()}
            />
            <AppButton
              label={mode === 'register' ? 'Уже есть аккаунт? Войти' : 'Нет аккаунта? Создать'}
              accessibilityLabel={mode === 'register' ? 'Перейти ко входу' : 'Перейти к созданию аккаунта'}
              testID="auth-switch-mode"
              variant="text"
              disabled={loading}
              onPress={switchMode}
            />
          </View>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  screen: {flex: 1, backgroundColor: colors.background},
  container: {flexGrow: 1, justifyContent: 'center', padding: spacing.xl},
  content: {width: '100%', maxWidth: 520, alignSelf: 'center'},
  eyebrow: {fontSize: 12, fontWeight: '800', letterSpacing: 1.6, color: colors.textMuted},
  title: {fontSize: 34, lineHeight: 40, fontWeight: '900', color: colors.text, marginTop: spacing.sm},
  subtitle: {fontSize: 16, lineHeight: 23, color: colors.textMuted, marginTop: spacing.sm, marginBottom: spacing.lg},
  field: {marginBottom: spacing.md},
  label: {fontSize: 14, lineHeight: 20, fontWeight: '800', color: colors.text, marginBottom: spacing.sm},
  input: {
    minHeight: control.minTouch,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: radius.md,
    paddingHorizontal: spacing.md,
    paddingVertical: 13,
    fontSize: 16,
    color: colors.text,
    backgroundColor: colors.surface,
  },
  inputError: {borderColor: colors.danger},
  fieldError: {fontSize: 13, lineHeight: 18, color: colors.danger, marginTop: spacing.xs},
  submitError: {fontSize: 14, lineHeight: 20, color: colors.danger, fontWeight: '700', marginBottom: spacing.md},
  actions: {gap: spacing.sm},
});
