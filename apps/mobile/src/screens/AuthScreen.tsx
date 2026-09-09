import React, {useState} from 'react';
import {StyleSheet, Text, TextInput, View} from 'react-native';
import {AppButton} from '../components/AppButton';
import {colors, radius, spacing} from '../theme/tokens';

type Props = {
  onSubmit: (mode: 'register' | 'login', email: string, password: string) => Promise<void>;
};

export function AuthScreen({onSubmit}: Props) {
  const [mode, setMode] = useState<'register' | 'login'>('register');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const submit = async () => {
    setLoading(true);
    setError('');
    try {
      await onSubmit(mode, email.trim(), password);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось выполнить вход');
    } finally {
      setLoading(false);
    }
  };

  const submitDisabled = !email.trim() || password.length < 8;
  return (
    <View style={styles.container} testID="auth-screen">
      <Text style={styles.eyebrow}>AI FITNESS OS</Text>
      <Text accessibilityRole="header" style={styles.title}>{mode === 'register' ? 'Создай профиль' : 'С возвращением'}</Text>
      <Text style={styles.subtitle}>Тренировки будут подстраиваться под тебя, оборудование и прогресс.</Text>

      <TextInput
        accessibilityLabel="Email"
        testID="auth-email"
        autoCapitalize="none"
        autoComplete="email"
        keyboardType="email-address"
        placeholder="Email"
        value={email}
        onChangeText={setEmail}
        style={styles.input}
      />
      <TextInput
        accessibilityLabel="Пароль"
        testID="auth-password"
        secureTextEntry
        autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
        placeholder="Пароль — минимум 8 символов"
        value={password}
        onChangeText={setPassword}
        style={styles.input}
      />

      {!!error && <Text accessibilityRole="alert" style={styles.error}>{error}</Text>}

      <AppButton
        label={mode === 'register' ? 'Продолжить' : 'Войти'}
        accessibilityLabel={mode === 'register' ? 'Создать аккаунт и продолжить' : 'Войти в аккаунт'}
        testID="auth-submit"
        loading={loading}
        disabled={submitDisabled}
        onPress={() => void submit()}
      />
      <AppButton
        label={mode === 'register' ? 'Уже есть аккаунт? Войти' : 'Нет аккаунта? Создать'}
        testID="auth-switch-mode"
        variant="text"
        disabled={loading}
        onPress={() => setMode(mode === 'register' ? 'login' : 'register')}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {flex: 1, padding: spacing.xl, justifyContent: 'center', gap: spacing.md, backgroundColor: colors.background},
  eyebrow: {fontSize: 12, fontWeight: '800', letterSpacing: 1.6, color: colors.textMuted},
  title: {fontSize: 34, lineHeight: 40, fontWeight: '900', color: colors.text},
  subtitle: {fontSize: 16, lineHeight: 23, color: colors.textMuted, marginBottom: spacing.sm},
  input: {borderWidth: 1, borderColor: colors.border, borderRadius: radius.md, paddingHorizontal: 16, paddingVertical: 14, fontSize: 16, color: colors.text, backgroundColor: colors.surface},
  error: {color: colors.danger, fontWeight: '700'},
});
