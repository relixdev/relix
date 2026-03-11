import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import type { Payload } from '../lib/protocol';

interface ChatMessageProps {
  payload: Payload;
}

function formatTime(ts: number): string {
  const d = new Date(ts);
  const h = d.getHours().toString().padStart(2, '0');
  const m = d.getMinutes().toString().padStart(2, '0');
  return `${h}:${m}`;
}

export default function ChatMessage({ payload }: ChatMessageProps) {
  const isUser = payload.kind === 'user_message';
  const text: string =
    typeof payload.data === 'string'
      ? payload.data
      : payload.data?.text ?? payload.data?.content ?? JSON.stringify(payload.data);
  const timestamp: number = payload.data?.timestamp ?? Date.now();

  return (
    <View style={[styles.row, isUser ? styles.rowRight : styles.rowLeft]}>
      <View style={[styles.bubble, isUser ? styles.bubbleUser : styles.bubbleAssistant]}>
        <Text style={[styles.text, isUser ? styles.textUser : styles.textAssistant]}>{text}</Text>
      </View>
      <Text style={[styles.timestamp, isUser ? styles.timestampRight : styles.timestampLeft]}>
        {formatTime(timestamp)}
      </Text>
    </View>
  );
}

const styles = StyleSheet.create({
  row: {
    marginHorizontal: 16,
    marginVertical: 4,
    maxWidth: '80%',
  },
  rowRight: {
    alignSelf: 'flex-end',
    alignItems: 'flex-end',
  },
  rowLeft: {
    alignSelf: 'flex-start',
    alignItems: 'flex-start',
  },
  bubble: {
    borderRadius: 18,
    paddingHorizontal: 14,
    paddingVertical: 10,
  },
  bubbleUser: {
    backgroundColor: '#007AFF',
    borderBottomRightRadius: 4,
  },
  bubbleAssistant: {
    backgroundColor: '#F2F2F7',
    borderBottomLeftRadius: 4,
  },
  text: {
    fontSize: 15,
    lineHeight: 20,
  },
  textUser: {
    color: '#FFFFFF',
  },
  textAssistant: {
    color: '#1C1C1E',
  },
  timestamp: {
    fontSize: 11,
    color: '#8E8E93',
    marginTop: 3,
  },
  timestampRight: {
    marginRight: 2,
  },
  timestampLeft: {
    marginLeft: 2,
  },
});
