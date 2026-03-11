import React from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  Alert,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { computeSASEmojis } from '../lib/sas';

interface SASVerificationProps {
  ownPublicKeyHex: string;
  peerPublicKeyHex: string;
  onConfirm: () => void;
  onReject: () => void;
}

export default function SASVerification({
  ownPublicKeyHex,
  peerPublicKeyHex,
  onConfirm,
  onReject,
}: SASVerificationProps) {
  const [emojis, setEmojis] = React.useState<string[] | null>(null);

  React.useEffect(() => {
    let cancelled = false;
    computeSASEmojis(ownPublicKeyHex, peerPublicKeyHex)
      .then((result) => {
        if (!cancelled) setEmojis(result);
      })
      .catch(() => {
        if (!cancelled) setEmojis(['?', '?', '?', '?']);
      });
    return () => {
      cancelled = true;
    };
  }, [ownPublicKeyHex, peerPublicKeyHex]);

  const handleReject = () => {
    Alert.alert(
      'Pairing Rejected',
      'The emojis did not match. Pairing has been aborted. Please try again.',
      [{ text: 'OK', onPress: onReject }],
    );
  };

  return (
    <SafeAreaView style={styles.root} edges={['top', 'bottom']}>
      <View style={styles.container}>
        <Text style={styles.title}>Verify Pairing</Text>
        <Text style={styles.subtitle}>
          Do you see these same emojis on your machine?
        </Text>

        <View style={styles.emojiContainer}>
          {emojis ? (
            emojis.map((emoji, i) => (
              <Text key={i} style={styles.emoji}>
                {emoji}
              </Text>
            ))
          ) : (
            <Text style={styles.loadingEmoji}>…</Text>
          )}
        </View>

        <Text style={styles.instruction}>
          Compare these emojis with the output shown in your terminal.
          Only confirm if they match exactly.
        </Text>

        <TouchableOpacity style={styles.confirmButton} onPress={onConfirm}>
          <Text style={styles.confirmText}>Yes, they match</Text>
        </TouchableOpacity>

        <TouchableOpacity style={styles.rejectButton} onPress={handleReject}>
          <Text style={styles.rejectText}>No, they don't match</Text>
        </TouchableOpacity>
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
    alignItems: 'center',
    justifyContent: 'center',
  },
  title: {
    fontSize: 28,
    fontWeight: '800',
    color: '#0f172a',
    letterSpacing: -0.5,
    marginBottom: 12,
    textAlign: 'center',
  },
  subtitle: {
    fontSize: 16,
    color: '#64748b',
    textAlign: 'center',
    lineHeight: 24,
    marginBottom: 40,
  },
  emojiContainer: {
    flexDirection: 'row',
    gap: 16,
    backgroundColor: '#fff',
    borderRadius: 20,
    paddingVertical: 28,
    paddingHorizontal: 32,
    shadowColor: '#000',
    shadowOpacity: 0.06,
    shadowRadius: 10,
    shadowOffset: { width: 0, height: 4 },
    elevation: 3,
    marginBottom: 32,
  },
  emoji: {
    fontSize: 48,
  },
  loadingEmoji: {
    fontSize: 32,
    color: '#94a3b8',
  },
  instruction: {
    fontSize: 13,
    color: '#94a3b8',
    textAlign: 'center',
    lineHeight: 20,
    marginBottom: 40,
    paddingHorizontal: 8,
  },
  confirmButton: {
    backgroundColor: '#16a34a',
    borderRadius: 14,
    paddingVertical: 16,
    paddingHorizontal: 40,
    width: '100%',
    alignItems: 'center',
    marginBottom: 12,
  },
  confirmText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '700',
  },
  rejectButton: {
    backgroundColor: '#fee2e2',
    borderRadius: 14,
    paddingVertical: 16,
    paddingHorizontal: 40,
    width: '100%',
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#fca5a5',
  },
  rejectText: {
    color: '#dc2626',
    fontSize: 16,
    fontWeight: '600',
  },
});
