import React, { useEffect, useRef, useState } from 'react';
import { ActivityIndicator, AppState, AppStateStatus, View, Text, StyleSheet, TouchableOpacity } from 'react-native';
import { NavigationContainer } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import * as LocalAuthentication from 'expo-local-authentication';
import { navigationRef, type RootStackParamList } from '../lib/navigationRef';
import { useAuthStore } from '../stores/authStore';
import { useMachineStore } from '../stores/machineStore';
import { useSessionStore } from '../stores/sessionStore';
import { RelayClient } from '../lib/relay';
import { getOrCreateKeyPair, fromBase64, initCrypto } from '../lib/crypto';
import { getRelayWsUrl } from '../lib/api';
import { loadPeerKey } from './sas-verification';
import {
  configureNotificationHandler,
  registerForPushNotifications,
  subscribeToNotificationResponses,
} from '../lib/notifications';
import { RelayMessageBus } from '../lib/relayBus';
import type { Envelope, Payload } from '../lib/protocol';
import LoginScreen from './(auth)/login';
import SessionScreen from './session/[id]';
import OnboardingScreen from './(auth)/onboarding';
import PairingScreen from './pairing';
import SASVerificationScreen from './sas-verification';
import TabsLayout from './(tabs)/_layout';

const Stack = createNativeStackNavigator<RootStackParamList>();

const RELAY_URL = getRelayWsUrl();

// How long (ms) the app can be backgrounded before requiring biometric re-auth
const LOCK_TIMEOUT_MS = 5 * 60 * 1000; // 5 minutes

// Configure notification display before the component mounts
configureNotificationHandler();

export default function RootLayout() {
  const { token, isLoading, loadToken } = useAuthStore();
  const { machines, fetchMachines, updateMachineStatus, addApproval } = useMachineStore();
  const { addSessionEvent, clearSession } = useSessionStore();
  const relayRef = useRef<RelayClient | null>(null);

  // Biometric / lock state
  const [isLocked, setIsLocked] = useState(false);
  const [biometricAvailable, setBiometricAvailable] = useState(false);
  const backgroundedAtRef = useRef<number | null>(null);

  // Load persisted token on mount
  useEffect(() => {
    loadToken();
    initCrypto().catch(() => {});
  }, []);

  // Check biometric availability once
  useEffect(() => {
    LocalAuthentication.hasHardwareAsync().then((hasHw) => {
      if (!hasHw) return;
      LocalAuthentication.isEnrolledAsync().then((enrolled) => {
        setBiometricAvailable(enrolled);
      });
    });
  }, []);

  // Register push notifications after auth
  useEffect(() => {
    if (!token) return;
    registerForPushNotifications(token);
  }, [token]);

  // Subscribe to notification taps
  useEffect(() => {
    return subscribeToNotificationResponses();
  }, []);

  // App lock: track background/foreground transitions
  useEffect(() => {
    if (!biometricAvailable) return;

    const handleAppStateChange = (nextState: AppStateStatus) => {
      if (nextState === 'background' || nextState === 'inactive') {
        backgroundedAtRef.current = Date.now();
      } else if (nextState === 'active') {
        const backgroundedAt = backgroundedAtRef.current;
        if (backgroundedAt !== null) {
          const elapsed = Date.now() - backgroundedAt;
          if (elapsed >= LOCK_TIMEOUT_MS) {
            clearSession();
            setIsLocked(true);
          }
          backgroundedAtRef.current = null;
        }
      }
    };

    const sub = AppState.addEventListener('change', handleAppStateChange);
    return () => sub.remove();
  }, [biometricAvailable]);

  // Trigger biometric auth whenever locked
  useEffect(() => {
    if (!isLocked) return;
    authenticate();
  }, [isLocked]);

  async function authenticate() {
    try {
      const result = await LocalAuthentication.authenticateAsync({
        promptMessage: 'Unlock Relix',
        disableDeviceFallback: false,
      });
      if (result.success) {
        setIsLocked(false);
      }
    } catch {
      setIsLocked(false);
    }
  }

  // Connect/disconnect relay when auth state changes
  useEffect(() => {
    if (!token) {
      relayRef.current?.disconnect();
      relayRef.current = null;
      return;
    }

    // Fetch machines on auth
    fetchMachines(token).catch(() => {});

    // Set up relay with encryption
    const client = new RelayClient(RELAY_URL, handleEnvelope);
    relayRef.current = client;

    // Load own key pair and peer keys, then connect
    (async () => {
      try {
        const ownKp = await getOrCreateKeyPair();
        client.setOwnKeyPair(ownKp);

        // Load peer keys for all known machines
        const currentMachines = useMachineStore.getState().machines;
        for (const m of currentMachines) {
          const peerKeyHex = await loadPeerKey(m.id);
          if (peerKeyHex) {
            try {
              client.setPeerPublicKey(m.id, fromBase64(peerKeyHex));
            } catch {
              // Key may be hex-encoded instead of base64; try raw
            }
          }
        }
      } catch {
        // Proceed without encryption keys — unencrypted messages still work
      }

      client.connect(token);
    })();

    // Listen for outbound messages from session screens
    const unsubBus = RelayMessageBus.on((msg) => {
      try {
        client.sendEncrypted(msg.machineId, msg.sessionId, msg.type, msg.payload);
      } catch {
        // relay not connected or missing keys — message dropped
      }
    });

    return () => {
      unsubBus();
      client.disconnect();
      relayRef.current = null;
    };
  }, [token]);

  // Re-load peer keys when machines list changes
  useEffect(() => {
    const client = relayRef.current;
    if (!client) return;

    (async () => {
      for (const m of machines) {
        const peerKeyHex = await loadPeerKey(m.id);
        if (peerKeyHex) {
          try {
            client.setPeerPublicKey(m.id, fromBase64(peerKeyHex));
          } catch {
            // ignore
          }
        }
      }
    })();
  }, [machines]);

  function handleEnvelope(env: Envelope) {
    const client = relayRef.current;

    switch (env.type) {
      case 'machine_status': {
        try {
          const data = JSON.parse(atob(env.payload)) as { status: string };
          updateMachineStatus(env.machine_id, data.status as any);
        } catch {
          // ignore
        }
        break;
      }

      case 'session_event': {
        // Try to decrypt the payload
        let payload: Payload | null = null;
        if (client) {
          payload = client.decryptPayload(env);
        }
        if (!payload) {
          // Fallback: treat as opaque event
          payload = { kind: env.type, seq: 0, data: env };
        }

        // Dispatch to the correct session
        const sessionId = env.session_id || 'unknown';
        addSessionEvent(sessionId, payload);

        // If it's an approval request, also add to pending approvals
        if (payload.kind === 'approval' && payload.data?.tool) {
          addApproval({
            id: payload.data.approval_id ?? String(payload.seq),
            machine_id: env.machine_id,
            session_id: sessionId,
            tool: payload.data.tool,
            description: payload.data.description ?? '',
            timestamp: payload.data.timestamp ?? Date.now(),
          });
        }
        break;
      }

      case 'session_list': {
        // Agent sent its list of active sessions
        let payload: Payload | null = null;
        if (client) {
          payload = client.decryptPayload(env);
        }
        if (payload?.data && Array.isArray(payload.data)) {
          useMachineStore.getState().setSessions(env.machine_id, payload.data);
        }
        break;
      }

      case 'approval_response':
      case 'user_input':
      case 'ping':
      case 'pong':
        break;

      default:
        break;
    }
  }

  if (isLoading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" color="#007AFF" />
      </View>
    );
  }

  // Lock screen overlay
  if (isLocked && token) {
    return (
      <View style={styles.lockScreen}>
        <Text style={styles.lockIcon}>🔒</Text>
        <Text style={styles.lockTitle}>Relix is locked</Text>
        <Text style={styles.lockSubtitle}>Authenticate to continue</Text>
        <TouchableOpacity style={styles.unlockButton} onPress={authenticate}>
          <Text style={styles.unlockText}>Unlock</Text>
        </TouchableOpacity>
      </View>
    );
  }

  const isAuthenticated = !!token;
  const hasNoMachines = isAuthenticated && machines.length === 0;

  return (
    <SafeAreaProvider>
      <NavigationContainer ref={navigationRef}>
        <Stack.Navigator screenOptions={{ headerShown: false }}>
          {!isAuthenticated ? (
            <Stack.Screen name="Login" component={LoginScreen} />
          ) : hasNoMachines ? (
            <>
              <Stack.Screen name="Onboarding" component={OnboardingScreen} />
              <Stack.Screen name="Pairing" component={PairingScreen} />
              <Stack.Screen name="SASVerification" component={SASVerificationScreen} />
            </>
          ) : (
            <>
              <Stack.Screen name="MainTabs" component={TabsLayout} />
              <Stack.Screen name="Session" component={SessionScreen} />
              <Stack.Screen name="Pairing" component={PairingScreen} />
              <Stack.Screen name="SASVerification" component={SASVerificationScreen} />
            </>
          )}
        </Stack.Navigator>
      </NavigationContainer>
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  center: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  lockScreen: {
    flex: 1,
    backgroundColor: '#0f172a',
    justifyContent: 'center',
    alignItems: 'center',
    gap: 12,
  },
  lockIcon: {
    fontSize: 56,
    marginBottom: 8,
  },
  lockTitle: {
    fontSize: 24,
    fontWeight: '700',
    color: '#f8fafc',
  },
  lockSubtitle: {
    fontSize: 15,
    color: '#94a3b8',
    marginBottom: 24,
  },
  unlockButton: {
    backgroundColor: '#007AFF',
    borderRadius: 14,
    paddingVertical: 14,
    paddingHorizontal: 40,
  },
  unlockText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
});
