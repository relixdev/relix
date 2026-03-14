import React, { useEffect, useRef, useState } from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
  Platform,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useNavigation } from '@react-navigation/native';
import type { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { useAuthStore } from '../stores/authStore';
import * as api from '../lib/api';
import { getOrCreateKeyPair, toBase64, toHex } from '../lib/crypto';
import type { RootStackParamList } from '../lib/navigationRef';

type PairingNavProp = NativeStackNavigationProp<RootStackParamList, 'Pairing'>;

const POLL_INTERVAL_MS = 2000;
const TIMEOUT_MS = 5 * 60 * 1000; // 5 minutes

type PairingState =
  | { phase: 'loading' }
  | { phase: 'waiting'; code: string; expiresAt: number }
  | { phase: 'expired' }
  | { phase: 'error'; message: string };

export default function PairingScreen() {
  const { token, user } = useAuthStore();
  const navigation = useNavigation<PairingNavProp>();
  const [state, setState] = useState<PairingState>({ phase: 'loading' });
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const startPairing = async () => {
    if (!token) return;
    setState({ phase: 'loading' });
    clearTimers();

    try {
      // Generate mobile key pair for E2E encryption
      const ownKp = await getOrCreateKeyPair();
      const mobilePublicKey = toBase64(ownKp.publicKey);
      const userId = user?.id ?? '';

      const { code, expires_at } = await api.generatePairingCode(
        token,
        userId,
        mobilePublicKey,
      );
      setState({ phase: 'waiting', code, expiresAt: expires_at });

      // Poll every 2 seconds for pairing completion
      pollRef.current = setInterval(async () => {
        if (!token) return;
        try {
          const result = await api.checkPairingStatus(token, code);
          if (result.status === 'completed') {
            clearTimers();
            const peerPublicKey = result.agent_public_key ?? result.peer_public_key ?? '';
            const machineId = result.machine_id ?? '';
            navigation.replace('SASVerification', {
              pairingCode: code,
              peerPublicKey,
              machineId,
            });
          } else if (result.status === 'expired' || result.status === 'failed') {
            clearTimers();
            setState({ phase: 'expired' });
          }
        } catch {
          // ignore transient poll errors
        }
      }, POLL_INTERVAL_MS);

      // Expire after 5 minutes
      timeoutRef.current = setTimeout(() => {
        clearTimers();
        setState({ phase: 'expired' });
      }, TIMEOUT_MS);
    } catch (e: any) {
      setState({ phase: 'error', message: e.message ?? 'Failed to generate pairing code' });
    }
  };

  const clearTimers = () => {
    if (pollRef.current) {
      clearInterval(pollRef.current);
      pollRef.current = null;
    }
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
  };

  useEffect(() => {
    startPairing();
    return clearTimers;
  }, []);

  const formatCode = (code: string) => {
    if (code.length === 6) {
      return `${code.slice(0, 3)} ${code.slice(3)}`;
    }
    return code;
  };

  return (
    <SafeAreaView style={styles.root} edges={['top', 'bottom']}>
      <View style={styles.container}>
        <TouchableOpacity style={styles.backButton} onPress={() => navigation.goBack()}>
          <Text style={styles.backText}>← Back</Text>
        </TouchableOpacity>

        <Text style={styles.title}>Pair Your Machine</Text>
        <Text style={styles.subtitle}>
          Run this command on the machine you want to pair:
        </Text>

        {state.phase === 'loading' && (
          <View style={styles.centerContent}>
            <ActivityIndicator size="large" color="#007AFF" />
            <Text style={styles.loadingText}>Generating pairing code...</Text>
          </View>
        )}

        {state.phase === 'waiting' && (
          <>
            <View style={styles.codeContainer}>
              <Text style={styles.code}>{formatCode(state.code)}</Text>
            </View>

            <View style={styles.commandBox}>
              <Text style={styles.commandLabel}>Command to run:</Text>
              <Text style={styles.commandText} selectable>
                {`relixctl pair ${state.code}`}
              </Text>
            </View>

            <View style={styles.waitingRow}>
              <ActivityIndicator size="small" color="#007AFF" />
              <Text style={styles.waitingText}>Waiting for pairing to complete...</Text>
            </View>

            <Text style={styles.expireNote}>Code expires in 5 minutes</Text>
          </>
        )}

        {state.phase === 'expired' && (
          <View style={styles.centerContent}>
            <Text style={styles.expiredIcon}>⏱</Text>
            <Text style={styles.expiredTitle}>Code Expired</Text>
            <Text style={styles.expiredSubtitle}>
              The pairing code has expired. Generate a new one to try again.
            </Text>
            <TouchableOpacity style={styles.retryButton} onPress={startPairing}>
              <Text style={styles.retryText}>Generate New Code</Text>
            </TouchableOpacity>
          </View>
        )}

        {state.phase === 'error' && (
          <View style={styles.centerContent}>
            <Text style={styles.expiredIcon}>⚠️</Text>
            <Text style={styles.expiredTitle}>Something Went Wrong</Text>
            <Text style={styles.expiredSubtitle}>{state.message}</Text>
            <TouchableOpacity style={styles.retryButton} onPress={startPairing}>
              <Text style={styles.retryText}>Try Again</Text>
            </TouchableOpacity>
          </View>
        )}
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  root: {
    flex: 1,
    backgroundColor: '#f8fafc',
  },
  container: {
    flex: 1,
    padding: 24,
  },
  backButton: {
    marginBottom: 24,
  },
  backText: {
    fontSize: 16,
    color: '#007AFF',
  },
  title: {
    fontSize: 28,
    fontWeight: '800',
    color: '#0f172a',
    letterSpacing: -0.5,
    marginBottom: 10,
  },
  subtitle: {
    fontSize: 15,
    color: '#64748b',
    lineHeight: 22,
    marginBottom: 32,
  },
  codeContainer: {
    backgroundColor: '#fff',
    borderRadius: 16,
    paddingVertical: 40,
    alignItems: 'center',
    shadowColor: '#000',
    shadowOpacity: 0.06,
    shadowRadius: 10,
    shadowOffset: { width: 0, height: 4 },
    elevation: 3,
    marginBottom: 24,
  },
  code: {
    fontSize: 52,
    fontWeight: '800',
    color: '#007AFF',
    letterSpacing: 8,
    fontFamily: Platform.OS === 'ios' ? 'Menlo' : 'monospace',
  },
  commandBox: {
    backgroundColor: '#0f172a',
    borderRadius: 12,
    padding: 16,
    marginBottom: 24,
  },
  commandLabel: {
    fontSize: 11,
    color: '#64748b',
    textTransform: 'uppercase',
    letterSpacing: 0.8,
    marginBottom: 6,
  },
  commandText: {
    fontSize: 15,
    color: '#a5f3fc',
    fontFamily: Platform.OS === 'ios' ? 'Menlo' : 'monospace',
  },
  waitingRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 10,
    marginBottom: 12,
  },
  waitingText: {
    fontSize: 14,
    color: '#64748b',
  },
  expireNote: {
    fontSize: 12,
    color: '#94a3b8',
  },
  centerContent: {
    alignItems: 'center',
    paddingTop: 40,
  },
  loadingText: {
    marginTop: 16,
    fontSize: 15,
    color: '#64748b',
  },
  expiredIcon: {
    fontSize: 48,
    marginBottom: 16,
  },
  expiredTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: '#0f172a',
    marginBottom: 8,
  },
  expiredSubtitle: {
    fontSize: 14,
    color: '#64748b',
    textAlign: 'center',
    lineHeight: 20,
    marginBottom: 28,
    paddingHorizontal: 16,
  },
  retryButton: {
    backgroundColor: '#007AFF',
    borderRadius: 12,
    paddingVertical: 14,
    paddingHorizontal: 32,
  },
  retryText: {
    color: '#fff',
    fontSize: 15,
    fontWeight: '600',
  },
});
