import React, { useRef, useState, useCallback, useEffect } from 'react';
import {
  View,
  FlatList,
  TextInput,
  TouchableOpacity,
  Text,
  KeyboardAvoidingView,
  Platform,
  StyleSheet,
  SafeAreaView,
} from 'react-native';
import { useRoute, useNavigation } from '@react-navigation/native';
import { useSessionStore } from '../../stores/sessionStore';
import { useMachineStore } from '../../stores/machineStore';
import type { Approval } from '../../stores/machineStore';
import ChatMessage from '../../components/ChatMessage';
import ToolCallCard from '../../components/ToolCallCard';
import ApprovalCard from '../../components/ApprovalCard';
import SessionToggle from '../../components/SessionToggle';
import TerminalView from '../../components/TerminalView';
import type { Payload } from '../../lib/protocol';

function payloadToTerminalLine(payload: Payload): string {
  switch (payload.kind) {
    case 'user_message':
      return `> ${payload.data?.text ?? payload.data ?? ''}`;
    case 'assistant_message':
      return payload.data?.text ?? payload.data?.content ?? JSON.stringify(payload.data);
    case 'tool_use':
      return `[tool] ${payload.data?.name ?? 'unknown'}: ${JSON.stringify(payload.data?.input ?? {})}`;
    case 'tool_result':
      return `[result] ${JSON.stringify(payload.data?.content ?? payload.data)}`;
    case 'approval':
      return `[approval needed] ${payload.data?.tool ?? ''}: ${payload.data?.description ?? ''}`;
    case 'error':
      return `[error] ${payload.data?.message ?? JSON.stringify(payload.data)}`;
    default:
      return JSON.stringify(payload.data);
  }
}

export default function SessionScreen() {
  const route = useRoute<{ key: string; name: string; params: { id: string; machineId?: string } }>();
  const navigation = useNavigation();
  const sessionId = route.params?.id;
  const machineId = route.params?.machineId;

  const events = useSessionStore((s) => s.events);
  const addEvent = useSessionStore((s) => s.addEvent);
  const setSession = useSessionStore((s) => s.setSession);
  const machines = useMachineStore((s) => s.machines);
  const resolveApproval = useMachineStore((s) => s.resolveApproval);

  const [mode, setMode] = useState<'chat' | 'terminal'>('chat');
  const [inputText, setInputText] = useState('');
  const flatListRef = useRef<FlatList>(null);

  // Set the active session on mount
  useEffect(() => {
    if (sessionId) {
      setSession(sessionId);
    }
  }, [sessionId]);

  // Find project name for this session
  const session = machines
    .flatMap((m) => m.sessions)
    .find((s) => s.id === sessionId);
  const projectName = session?.project ?? sessionId ?? 'Session';

  const handleSend = useCallback(() => {
    const text = inputText.trim();
    if (!text || !sessionId) return;
    setInputText('');

    // Optimistic: add user message to local events immediately
    const optimistic: Payload = {
      kind: 'user_message',
      seq: Date.now(),
      data: { text, timestamp: Date.now() },
    };
    addEvent(optimistic);

    // Send via relay — the _layout relay client handles encryption
    // We dispatch a custom event that the layout picks up
    if (machineId) {
      const { RelayMessageBus } = require('../../lib/relayBus');
      RelayMessageBus.emit('send', {
        machineId,
        sessionId,
        type: 'user_input' as const,
        payload: { kind: 'user_message', seq: Date.now(), data: { text } },
      });
    }
  }, [inputText, addEvent, sessionId, machineId]);

  const handleApprovalAllow = useCallback(
    (approvalId: string) => {
      const response: Payload = {
        kind: 'approval',
        seq: Date.now(),
        data: { approval_id: approvalId, allowed: true, timestamp: Date.now() },
      };
      addEvent(response);
      resolveApproval(approvalId, true);

      if (machineId && sessionId) {
        const { RelayMessageBus } = require('../../lib/relayBus');
        RelayMessageBus.emit('send', {
          machineId,
          sessionId,
          type: 'approval_response' as const,
          payload: { kind: 'approval', seq: Date.now(), data: { approval_id: approvalId, allowed: true } },
        });
      }
    },
    [addEvent, resolveApproval, machineId, sessionId],
  );

  const handleApprovalDeny = useCallback(
    (approvalId: string) => {
      const response: Payload = {
        kind: 'approval',
        seq: Date.now(),
        data: { approval_id: approvalId, allowed: false, timestamp: Date.now() },
      };
      addEvent(response);
      resolveApproval(approvalId, false);

      if (machineId && sessionId) {
        const { RelayMessageBus } = require('../../lib/relayBus');
        RelayMessageBus.emit('send', {
          machineId,
          sessionId,
          type: 'approval_response' as const,
          payload: { kind: 'approval', seq: Date.now(), data: { approval_id: approvalId, allowed: false } },
        });
      }
    },
    [addEvent, resolveApproval, machineId, sessionId],
  );

  const renderItem = useCallback(
    ({ item }: { item: Payload }) => {
      if (item.kind === 'approval' && item.data?.tool && item.data?.allowed === undefined) {
        const approval: Approval = {
          id: item.data.approval_id ?? String(item.seq),
          machine_id: item.data.machine_id ?? machineId ?? '',
          session_id: item.data.session_id ?? sessionId ?? '',
          tool: item.data.tool,
          description: item.data.description ?? '',
          timestamp: item.data.timestamp ?? item.seq,
        };
        return (
          <ApprovalCard
            approval={approval}
            onAllow={handleApprovalAllow}
            onDeny={handleApprovalDeny}
          />
        );
      }
      if (item.kind === 'tool_use' || item.kind === 'tool_result') {
        return (
          <ToolCallCard
            toolName={item.data?.name ?? item.data?.tool_use_id ?? 'tool'}
            args={item.data?.input ?? item.data}
            status={item.data?.status ?? 'pending'}
            result={item.kind === 'tool_result' ? item.data?.content : undefined}
          />
        );
      }
      if (item.kind === 'user_message' || item.kind === 'assistant_message') {
        return <ChatMessage payload={item} />;
      }
      return null;
    },
    [handleApprovalAllow, handleApprovalDeny, sessionId, machineId],
  );

  const terminalLines = events.map(payloadToTerminalLine);

  return (
    <SafeAreaView style={styles.safeArea}>
      {/* Header */}
      <View style={styles.header}>
        <TouchableOpacity onPress={() => navigation.goBack()} style={styles.backButton}>
          <Text style={styles.backText}>←</Text>
        </TouchableOpacity>
        <Text style={styles.projectName} numberOfLines={1}>
          {projectName}
        </Text>
        <SessionToggle mode={mode} onToggle={setMode} />
      </View>

      <KeyboardAvoidingView
        style={styles.flex}
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
        keyboardVerticalOffset={Platform.OS === 'ios' ? 0 : 20}
      >
        {mode === 'chat' ? (
          <FlatList
            ref={flatListRef}
            style={styles.flex}
            data={events}
            keyExtractor={(_, i) => String(i)}
            renderItem={renderItem}
            contentContainerStyle={styles.listContent}
            onContentSizeChange={() =>
              flatListRef.current?.scrollToEnd({ animated: true })
            }
            onLayout={() => flatListRef.current?.scrollToEnd({ animated: false })}
          />
        ) : (
          <TerminalView lines={terminalLines} />
        )}

        {/* Input bar */}
        <View
          style={[
            styles.inputBar,
            mode === 'terminal' && styles.inputBarTerminal,
          ]}
        >
          {mode === 'terminal' && (
            <Text style={styles.terminalPrompt}>{'> '}</Text>
          )}
          <TextInput
            style={[styles.input, mode === 'terminal' && styles.inputTerminal]}
            value={inputText}
            onChangeText={setInputText}
            placeholder={mode === 'chat' ? 'Message...' : ''}
            placeholderTextColor={mode === 'terminal' ? '#636366' : '#C7C7CC'}
            returnKeyType="send"
            onSubmitEditing={handleSend}
            blurOnSubmit={false}
            multiline={false}
          />
          <TouchableOpacity
            style={[
              styles.sendButton,
              mode === 'terminal' && styles.sendButtonTerminal,
              !inputText.trim() && styles.sendButtonDisabled,
            ]}
            onPress={handleSend}
            disabled={!inputText.trim()}
            activeOpacity={0.8}
          >
            <Text
              style={[
                styles.sendButtonText,
                mode === 'terminal' && styles.sendButtonTextTerminal,
              ]}
            >
              {mode === 'terminal' ? 'Run' : 'Send'}
            </Text>
          </TouchableOpacity>
        </View>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    backgroundColor: '#FFFFFF',
  },
  flex: {
    flex: 1,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: '#C6C6C8',
    backgroundColor: '#FFFFFF',
  },
  backButton: {
    paddingRight: 12,
  },
  backText: {
    fontSize: 20,
    color: '#007AFF',
    fontWeight: '600',
  },
  projectName: {
    fontSize: 16,
    fontWeight: '600',
    color: '#1C1C1E',
    flex: 1,
    marginRight: 12,
  },
  listContent: {
    paddingVertical: 12,
  },
  inputBar: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 10,
    borderTopWidth: StyleSheet.hairlineWidth,
    borderTopColor: '#C6C6C8',
    backgroundColor: '#FFFFFF',
    gap: 8,
  },
  inputBarTerminal: {
    backgroundColor: '#0d0d1a',
    borderTopColor: '#2a2a4a',
  },
  terminalPrompt: {
    fontFamily: Platform.select({ ios: 'Courier New', android: 'monospace', default: 'monospace' }),
    fontSize: 14,
    color: '#34C759',
  },
  input: {
    flex: 1,
    height: 40,
    backgroundColor: '#F2F2F7',
    borderRadius: 20,
    paddingHorizontal: 14,
    fontSize: 15,
    color: '#1C1C1E',
  },
  inputTerminal: {
    backgroundColor: '#1a1a2e',
    borderRadius: 4,
    color: '#E0E0E0',
    fontFamily: Platform.select({ ios: 'Courier New', android: 'monospace', default: 'monospace' }),
    fontSize: 13,
  },
  sendButton: {
    backgroundColor: '#007AFF',
    borderRadius: 20,
    paddingHorizontal: 16,
    height: 40,
    justifyContent: 'center',
    alignItems: 'center',
  },
  sendButtonTerminal: {
    backgroundColor: '#34C759',
    borderRadius: 4,
  },
  sendButtonDisabled: {
    opacity: 0.4,
  },
  sendButtonText: {
    color: '#FFFFFF',
    fontWeight: '600',
    fontSize: 14,
  },
  sendButtonTextTerminal: {
    fontFamily: Platform.select({ ios: 'Courier New', android: 'monospace', default: 'monospace' }),
    fontSize: 13,
  },
});
