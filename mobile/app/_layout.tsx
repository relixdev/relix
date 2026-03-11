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
import {
  configureNotificationHandler,
  registerForPushNotifications,
  subscribeToNotificationResponses,
} from '../lib/notifications';
import type { Envelope } from '../lib/protocol';
import LoginScreen from './(auth)/login';
import OnboardingScreen from './(auth)/onboarding';
import PairingScreen from './pairing';
import SASVerificationScreen from './sas-verification';
import TabsLayout from './(tabs)/_layout';

const Stack = createNativeStackNavigator<RootStackParamList>();

const RELAY_URL = 'wss://relay.relix.sh/v1/ws';

// How long (ms) the app can be backgrounded before requiring biometric re-auth
const LOCK_TIMEOUT_MS = 5 * 60 * 1000; // 5 minutes

// Configure notification display before the component mounts
configureNotificationHandler();

export default function RootLayout() {
  const { token, isLoading, loadToken } = useAuthStore();
  const { machines, fetchMachines, updateMachineStatus, addApproval } = useMachineStore();
  const { addEvent, clearSession } = useSessionStore();
  const relayRef = useRef<RelayClient | null>(null);

  // Biometric / lock state
  const [isLocked, setIsLocked] = useState(false);
  const [biometricAvailable, setBiometricAvailable] = useState(false);
  const backgroundedAtRef = useRef<number | null>(null);

  // Load persisted token on mount
  useEffect(() => {
    loadToken();
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
            // Clear sensitive session data and require re-auth
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
      // If auth fails or is cancelled, stay locked (user can retry via button)
    } catch {
      // Device may not support it; unlock anyway
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

    // Set up relay
    const client = new RelayClient(RELAY_URL, handleEnvelope);
    relayRef.current = client;
    client.connect(token);

    return () => {
      client.disconnect();
      relayRef.current = null;
    };
  }, [token]);

  function handleEnvelope(env: Envelope) {
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
        addEvent({ kind: env.type, seq: 0, data: env });
        break;
      }
      case 'approval_response':
      case 'user_input': {
        break;
      }
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
