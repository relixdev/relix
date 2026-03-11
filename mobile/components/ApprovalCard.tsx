import React from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
} from 'react-native';
import type { Approval } from '../stores/machineStore';

interface Props {
  approval: Approval;
  onAllow: (approvalId: string) => void;
  onDeny: (approvalId: string) => void;
}

export default function ApprovalCard({ approval, onAllow, onDeny }: Props) {
  return (
    <View style={styles.card}>
      <View style={styles.header}>
        <View style={styles.badge}>
          <Text style={styles.badgeText}>Approval needed</Text>
        </View>
        <Text style={styles.toolName}>{approval.tool}</Text>
      </View>
      {approval.description ? (
        <Text style={styles.description}>{approval.description}</Text>
      ) : null}
      <View style={styles.actions}>
        <TouchableOpacity
          style={[styles.button, styles.denyButton]}
          onPress={() => onDeny(approval.id)}
          activeOpacity={0.75}
        >
          <Text style={styles.denyText}>Deny</Text>
        </TouchableOpacity>
        <TouchableOpacity
          style={[styles.button, styles.allowButton]}
          onPress={() => onAllow(approval.id)}
          activeOpacity={0.75}
        >
          <Text style={styles.allowText}>Allow</Text>
        </TouchableOpacity>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: '#fffbeb',
    borderWidth: 1.5,
    borderColor: '#f59e0b',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
    gap: 8,
  },
  badge: {
    backgroundColor: '#f59e0b',
    borderRadius: 6,
    paddingHorizontal: 8,
    paddingVertical: 3,
  },
  badgeText: {
    color: '#fff',
    fontSize: 11,
    fontWeight: '700',
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  toolName: {
    fontSize: 15,
    fontWeight: '700',
    color: '#92400e',
    flex: 1,
  },
  description: {
    fontSize: 13,
    color: '#78350f',
    marginBottom: 12,
    lineHeight: 19,
  },
  actions: {
    flexDirection: 'row',
    gap: 10,
  },
  button: {
    flex: 1,
    paddingVertical: 10,
    borderRadius: 8,
    alignItems: 'center',
  },
  denyButton: {
    backgroundColor: '#fee2e2',
    borderWidth: 1,
    borderColor: '#fca5a5',
  },
  denyText: {
    color: '#dc2626',
    fontWeight: '600',
    fontSize: 14,
  },
  allowButton: {
    backgroundColor: '#10b981',
  },
  allowText: {
    color: '#fff',
    fontWeight: '600',
    fontSize: 14,
  },
});
