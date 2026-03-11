import React, { useEffect, useRef, useState } from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  ScrollView,
  Clipboard,
  Alert,
  Platform,
} from 'react-native';
import { useAuthStore } from '../../stores/authStore';
import { useMachineStore } from '../../stores/machineStore';
import { useOnboardingNavigation } from '../../lib/navigationRef';

const STEPS = [
  {
    number: 1,
    title: 'Install the agent',
    description: 'Run this command on your Mac or Linux machine:',
    command: 'curl -fsSL relix.sh/install | sh',
  },
  {
    number: 2,
    title: 'Login on your machine',
    description: 'Authenticate the agent with your Relix account:',
    command: 'relixctl login',
  },
  {
    number: 3,
    title: 'Pair your device',
    description: "Once the agent is running, tap below to pair this phone.",
    command: null,
  },
];

export default function OnboardingScreen() {
  const { token } = useAuthStore();
  const { machines, fetchMachines } = useMachineStore();
  const navigation = useOnboardingNavigation();
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const [didCopy, setDidCopy] = useState<number | null>(null);

  useEffect(() => {
    if (!token) return;

    // Poll every 3 seconds for first machine
    pollRef.current = setInterval(async () => {
      try {
        await fetchMachines(token);
      } catch {
        // ignore poll errors
      }
    }, 3000);

    return () => {
      if (pollRef.current) clearInterval(pollRef.current);
    };
  }, [token]);

  useEffect(() => {
    if (machines.length > 0) {
      if (pollRef.current) clearInterval(pollRef.current);
      navigation.navigateToDashboard();
    }
  }, [machines.length]);

  const copyToClipboard = (text: string, stepNum: number) => {
    Clipboard.setString(text);
    setDidCopy(stepNum);
    setTimeout(() => setDidCopy(null), 2000);
  };

  const handlePair = () => {
    Alert.alert(
      'Pairing',
      'Open Relix on your machine and run: relixctl pair\n\nThis phone will be detected automatically.',
    );
  };

  return (
    <ScrollView style={styles.root} contentContainerStyle={styles.content}>
      <Text style={styles.title}>Get Started</Text>
      <Text style={styles.subtitle}>
        Connect your first machine to start monitoring AI sessions.
      </Text>

      {STEPS.map((step) => (
        <View key={step.number} style={styles.stepCard}>
          <View style={styles.stepHeader}>
            <View style={styles.stepBadge}>
              <Text style={styles.stepBadgeText}>{step.number}</Text>
            </View>
            <Text style={styles.stepTitle}>{step.title}</Text>
          </View>
          <Text style={styles.stepDescription}>{step.description}</Text>

          {step.command ? (
            <TouchableOpacity
              style={styles.commandBox}
              onPress={() => copyToClipboard(step.command!, step.number)}
              activeOpacity={0.7}
            >
              <Text style={styles.commandText}>{step.command}</Text>
              <Text style={styles.copyHint}>
                {didCopy === step.number ? 'Copied!' : 'Tap to copy'}
              </Text>
            </TouchableOpacity>
          ) : (
            <TouchableOpacity style={styles.pairButton} onPress={handlePair}>
              <Text style={styles.pairButtonText}>Pair this device</Text>
            </TouchableOpacity>
          )}
        </View>
      ))}

      <View style={styles.waitingRow}>
        <View style={styles.pulseDot} />
        <Text style={styles.waitingText}>Waiting for machine to connect…</Text>
      </View>
    </ScrollView>
  );
}

const DARK = '#0f172a';
const ACCENT = '#6366f1';

const styles = StyleSheet.create({
  root: {
    flex: 1,
    backgroundColor: '#f8fafc',
  },
  content: {
    padding: 24,
    paddingTop: 60,
  },
  title: {
    fontSize: 30,
    fontWeight: '800',
    color: DARK,
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 15,
    color: '#64748b',
    marginBottom: 32,
    lineHeight: 22,
  },
  stepCard: {
    backgroundColor: '#fff',
    borderRadius: 14,
    padding: 20,
    marginBottom: 16,
    shadowColor: '#000',
    shadowOpacity: 0.05,
    shadowRadius: 8,
    shadowOffset: { width: 0, height: 2 },
    elevation: 2,
  },
  stepHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 10,
  },
  stepBadge: {
    width: 28,
    height: 28,
    borderRadius: 14,
    backgroundColor: ACCENT,
    alignItems: 'center',
    justifyContent: 'center',
    marginRight: 10,
  },
  stepBadgeText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '700',
  },
  stepTitle: {
    fontSize: 17,
    fontWeight: '700',
    color: DARK,
  },
  stepDescription: {
    fontSize: 14,
    color: '#64748b',
    marginBottom: 12,
    lineHeight: 20,
  },
  commandBox: {
    backgroundColor: '#0f172a',
    borderRadius: 8,
    padding: 14,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  commandText: {
    color: '#a5f3fc',
    fontFamily: Platform.OS === 'ios' ? 'Menlo' : 'monospace',
    fontSize: 13,
    flex: 1,
  },
  copyHint: {
    color: '#64748b',
    fontSize: 12,
    marginLeft: 10,
  },
  pairButton: {
    backgroundColor: ACCENT,
    borderRadius: 10,
    paddingVertical: 14,
    alignItems: 'center',
  },
  pairButtonText: {
    color: '#fff',
    fontSize: 15,
    fontWeight: '600',
  },
  waitingRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    marginTop: 24,
    gap: 8,
  },
  pulseDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
    backgroundColor: '#10b981',
  },
  waitingText: {
    fontSize: 13,
    color: '#94a3b8',
  },
});
