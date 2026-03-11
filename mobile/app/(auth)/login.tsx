import React, { useState } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
  KeyboardAvoidingView,
  Platform,
  Alert,
} from 'react-native';
import * as WebBrowser from 'expo-web-browser';
import { useAuthNavigation } from '../../lib/navigationRef';
import { useAuthStore } from '../../stores/authStore';

WebBrowser.maybeCompleteAuthSession();

export default function LoginScreen() {
  const [mode, setMode] = useState<'signin' | 'signup'>('signin');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showEmail, setShowEmail] = useState(false);

  const { login, isLoading, error } = useAuthStore();
  const navigation = useAuthNavigation();

  const handleEmailSubmit = async () => {
    if (!email.trim() || !password.trim()) {
      Alert.alert('Missing fields', 'Please enter your email and password.');
      return;
    }
    try {
      await login('email', { email: email.trim(), password, mode });
      // navigation is handled by root layout reacting to token state
    } catch {
      // error is set in store
    }
  };

  const handleGitHub = async () => {
    try {
      // Open GitHub OAuth in browser — the redirect will call maybeCompleteAuthSession
      const result = await WebBrowser.openAuthSessionAsync(
        'https://api.relix.sh/v1/auth/github',
        'relix://auth',
      );
      if (result.type === 'success' && result.url) {
        const url = new URL(result.url);
        const token = url.searchParams.get('token');
        if (token) {
          await login('github', { token });
        }
      }
    } catch (e: any) {
      Alert.alert('GitHub login failed', e.message ?? 'Unknown error');
    }
  };

  return (
    <KeyboardAvoidingView
      style={styles.root}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
    >
      <View style={styles.header}>
        <Text style={styles.logo}>Relix</Text>
        <Text style={styles.tagline}>Your AI sessions, in your pocket</Text>
      </View>

      <View style={styles.card}>
        {!showEmail ? (
          <>
            <TouchableOpacity
              style={styles.githubButton}
              onPress={handleGitHub}
              disabled={isLoading}
            >
              {isLoading ? (
                <ActivityIndicator color="#fff" />
              ) : (
                <Text style={styles.githubButtonText}>Continue with GitHub</Text>
              )}
            </TouchableOpacity>

            <View style={styles.dividerRow}>
              <View style={styles.divider} />
              <Text style={styles.dividerText}>or</Text>
              <View style={styles.divider} />
            </View>

            <TouchableOpacity
              style={styles.emailButton}
              onPress={() => setShowEmail(true)}
            >
              <Text style={styles.emailButtonText}>Sign in with Email</Text>
            </TouchableOpacity>
          </>
        ) : (
          <>
            <View style={styles.toggleRow}>
              <TouchableOpacity
                style={[styles.toggleTab, mode === 'signin' && styles.toggleTabActive]}
                onPress={() => setMode('signin')}
              >
                <Text style={[styles.toggleTabText, mode === 'signin' && styles.toggleTabTextActive]}>
                  Sign In
                </Text>
              </TouchableOpacity>
              <TouchableOpacity
                style={[styles.toggleTab, mode === 'signup' && styles.toggleTabActive]}
                onPress={() => setMode('signup')}
              >
                <Text style={[styles.toggleTabText, mode === 'signup' && styles.toggleTabTextActive]}>
                  Sign Up
                </Text>
              </TouchableOpacity>
            </View>

            <TextInput
              style={styles.input}
              placeholder="Email"
              placeholderTextColor="#aaa"
              keyboardType="email-address"
              autoCapitalize="none"
              autoCorrect={false}
              value={email}
              onChangeText={setEmail}
            />
            <TextInput
              style={styles.input}
              placeholder="Password"
              placeholderTextColor="#aaa"
              secureTextEntry
              value={password}
              onChangeText={setPassword}
            />

            {error ? <Text style={styles.errorText}>{error}</Text> : null}

            <TouchableOpacity
              style={styles.submitButton}
              onPress={handleEmailSubmit}
              disabled={isLoading}
            >
              {isLoading ? (
                <ActivityIndicator color="#fff" />
              ) : (
                <Text style={styles.submitButtonText}>
                  {mode === 'signin' ? 'Sign In' : 'Create Account'}
                </Text>
              )}
            </TouchableOpacity>

            <TouchableOpacity onPress={() => setShowEmail(false)} style={styles.backLink}>
              <Text style={styles.backLinkText}>← Back</Text>
            </TouchableOpacity>
          </>
        )}
      </View>
    </KeyboardAvoidingView>
  );
}

const DARK = '#0f172a';
const ACCENT = '#6366f1';

const styles = StyleSheet.create({
  root: {
    flex: 1,
    backgroundColor: DARK,
    justifyContent: 'center',
    paddingHorizontal: 24,
  },
  header: {
    alignItems: 'center',
    marginBottom: 40,
  },
  logo: {
    fontSize: 42,
    fontWeight: '800',
    color: '#fff',
    letterSpacing: -1,
  },
  tagline: {
    fontSize: 14,
    color: '#94a3b8',
    marginTop: 6,
  },
  card: {
    backgroundColor: '#fff',
    borderRadius: 16,
    padding: 24,
    shadowColor: '#000',
    shadowOpacity: 0.15,
    shadowRadius: 20,
    shadowOffset: { width: 0, height: 8 },
    elevation: 8,
  },
  githubButton: {
    backgroundColor: '#24292e',
    borderRadius: 10,
    paddingVertical: 14,
    alignItems: 'center',
  },
  githubButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  dividerRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginVertical: 16,
  },
  divider: {
    flex: 1,
    height: 1,
    backgroundColor: '#e2e8f0',
  },
  dividerText: {
    marginHorizontal: 12,
    color: '#94a3b8',
    fontSize: 13,
  },
  emailButton: {
    borderWidth: 1.5,
    borderColor: '#e2e8f0',
    borderRadius: 10,
    paddingVertical: 14,
    alignItems: 'center',
  },
  emailButtonText: {
    color: '#0f172a',
    fontSize: 16,
    fontWeight: '600',
  },
  toggleRow: {
    flexDirection: 'row',
    backgroundColor: '#f1f5f9',
    borderRadius: 8,
    padding: 3,
    marginBottom: 20,
  },
  toggleTab: {
    flex: 1,
    paddingVertical: 8,
    alignItems: 'center',
    borderRadius: 6,
  },
  toggleTabActive: {
    backgroundColor: '#fff',
    shadowColor: '#000',
    shadowOpacity: 0.08,
    shadowRadius: 4,
    shadowOffset: { width: 0, height: 1 },
    elevation: 2,
  },
  toggleTabText: {
    fontSize: 14,
    fontWeight: '500',
    color: '#94a3b8',
  },
  toggleTabTextActive: {
    color: '#0f172a',
    fontWeight: '600',
  },
  input: {
    borderWidth: 1.5,
    borderColor: '#e2e8f0',
    borderRadius: 10,
    paddingHorizontal: 14,
    paddingVertical: 12,
    fontSize: 15,
    color: '#0f172a',
    marginBottom: 12,
  },
  errorText: {
    color: '#ef4444',
    fontSize: 13,
    marginBottom: 8,
  },
  submitButton: {
    backgroundColor: ACCENT,
    borderRadius: 10,
    paddingVertical: 14,
    alignItems: 'center',
    marginTop: 4,
  },
  submitButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  backLink: {
    marginTop: 16,
    alignItems: 'center',
  },
  backLinkText: {
    color: '#6366f1',
    fontSize: 14,
  },
});
