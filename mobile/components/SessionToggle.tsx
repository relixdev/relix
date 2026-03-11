import React, { useEffect } from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import * as SecureStore from 'expo-secure-store';

const STORAGE_KEY = 'relix_session_mode';

interface SessionToggleProps {
  mode: 'chat' | 'terminal';
  onToggle: (mode: 'chat' | 'terminal') => void;
}

export default function SessionToggle({ mode, onToggle }: SessionToggleProps) {
  useEffect(() => {
    SecureStore.getItemAsync(STORAGE_KEY)
      .then((stored) => {
        if (stored === 'chat' || stored === 'terminal') {
          onToggle(stored);
        }
      })
      .catch(() => {});
  }, []);

  function handlePress(next: 'chat' | 'terminal') {
    if (next === mode) return;
    SecureStore.setItemAsync(STORAGE_KEY, next).catch(() => {});
    onToggle(next);
  }

  return (
    <View style={styles.container}>
      <TouchableOpacity
        style={[styles.segment, mode === 'chat' && styles.segmentActive]}
        onPress={() => handlePress('chat')}
        activeOpacity={0.8}
      >
        <Text style={[styles.segmentText, mode === 'chat' && styles.segmentTextActive]}>
          Chat
        </Text>
      </TouchableOpacity>
      <TouchableOpacity
        style={[styles.segment, mode === 'terminal' && styles.segmentActive]}
        onPress={() => handlePress('terminal')}
        activeOpacity={0.8}
      >
        <Text style={[styles.segmentText, mode === 'terminal' && styles.segmentTextActive]}>
          Terminal
        </Text>
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    backgroundColor: '#E5E5EA',
    borderRadius: 8,
    padding: 2,
  },
  segment: {
    paddingHorizontal: 16,
    paddingVertical: 5,
    borderRadius: 6,
    minWidth: 64,
    alignItems: 'center',
  },
  segmentActive: {
    backgroundColor: '#FFFFFF',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.12,
    shadowRadius: 2,
    elevation: 2,
  },
  segmentText: {
    fontSize: 13,
    fontWeight: '500',
    color: '#8E8E93',
  },
  segmentTextActive: {
    color: '#1C1C1E',
    fontWeight: '600',
  },
});
