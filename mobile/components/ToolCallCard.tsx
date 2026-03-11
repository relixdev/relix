import React, { useState } from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';

interface ToolCallCardProps {
  toolName: string;
  args: any;
  status: string;
  result?: any;
}

const STATUS_COLORS: Record<string, string> = {
  pending: '#FF9500',
  approved: '#34C759',
  denied: '#FF3B30',
};

const STATUS_LABELS: Record<string, string> = {
  pending: 'Pending',
  approved: 'Approved',
  denied: 'Denied',
};

function truncate(str: string, maxLen: number): string {
  if (str.length <= maxLen) return str;
  return str.slice(0, maxLen) + '…';
}

function formatArgs(args: any): string {
  try {
    return JSON.stringify(args, null, 2);
  } catch {
    return String(args);
  }
}

function hasDiff(args: any): boolean {
  if (!args || typeof args !== 'object') return false;
  return 'old_str' in args || 'new_str' in args || 'diff' in args || 'patch' in args;
}

function renderDiff(args: any): React.ReactNode {
  if (!hasDiff(args)) return null;
  const oldStr: string = args.old_str ?? args.diff ?? '';
  const newStr: string = args.new_str ?? args.patch ?? '';
  return (
    <View style={styles.diffContainer}>
      {oldStr ? (
        <Text style={styles.diffRemoved}>{'- ' + oldStr.split('\n').join('\n- ')}</Text>
      ) : null}
      {newStr ? (
        <Text style={styles.diffAdded}>{'+ ' + newStr.split('\n').join('\n+ ')}</Text>
      ) : null}
    </View>
  );
}

export default function ToolCallCard({ toolName, args, status, result }: ToolCallCardProps) {
  const [expanded, setExpanded] = useState(false);

  const statusColor = STATUS_COLORS[status] ?? '#8E8E93';
  const statusLabel = STATUS_LABELS[status] ?? status;
  const argsStr = formatArgs(args);

  return (
    <TouchableOpacity
      style={styles.card}
      onPress={() => setExpanded((v) => !v)}
      activeOpacity={0.85}
    >
      <View style={styles.header}>
        <View style={styles.headerLeft}>
          <Text style={styles.toolIcon}>⚙</Text>
          <Text style={styles.toolName}>{toolName}</Text>
        </View>
        <View style={[styles.statusBadge, { backgroundColor: statusColor + '22' }]}>
          <View style={[styles.statusDot, { backgroundColor: statusColor }]} />
          <Text style={[styles.statusText, { color: statusColor }]}>{statusLabel}</Text>
        </View>
      </View>

      {!expanded && (
        <Text style={styles.argsPreview} numberOfLines={2}>
          {truncate(argsStr, 120)}
        </Text>
      )}

      {expanded && (
        <View style={styles.expandedContent}>
          <Text style={styles.sectionLabel}>Arguments</Text>
          <View style={styles.codeBlock}>
            <Text style={styles.codeText}>{argsStr}</Text>
          </View>
          {hasDiff(args) && (
            <>
              <Text style={styles.sectionLabel}>Diff</Text>
              {renderDiff(args)}
            </>
          )}
          {result !== undefined && (
            <>
              <Text style={styles.sectionLabel}>Result</Text>
              <View style={styles.codeBlock}>
                <Text style={styles.codeText}>{formatArgs(result)}</Text>
              </View>
            </>
          )}
        </View>
      )}

      <Text style={styles.expandHint}>{expanded ? 'Tap to collapse' : 'Tap to expand'}</Text>
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: '#FFFFFF',
    borderWidth: 1,
    borderColor: '#E5E5EA',
    borderRadius: 10,
    padding: 14,
    marginHorizontal: 16,
    marginVertical: 6,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.06,
    shadowRadius: 3,
    elevation: 1,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginBottom: 8,
  },
  headerLeft: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
  },
  toolIcon: {
    fontSize: 14,
  },
  toolName: {
    fontSize: 14,
    fontWeight: '600',
    color: '#1C1C1E',
  },
  statusBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 5,
    paddingHorizontal: 8,
    paddingVertical: 3,
    borderRadius: 12,
  },
  statusDot: {
    width: 6,
    height: 6,
    borderRadius: 3,
  },
  statusText: {
    fontSize: 12,
    fontWeight: '600',
  },
  argsPreview: {
    fontSize: 12,
    color: '#8E8E93',
    fontFamily: 'Courier',
    lineHeight: 17,
  },
  expandedContent: {
    marginTop: 6,
  },
  sectionLabel: {
    fontSize: 11,
    fontWeight: '600',
    color: '#8E8E93',
    textTransform: 'uppercase',
    letterSpacing: 0.5,
    marginTop: 10,
    marginBottom: 4,
  },
  codeBlock: {
    backgroundColor: '#F2F2F7',
    borderRadius: 6,
    padding: 10,
  },
  codeText: {
    fontSize: 12,
    fontFamily: 'Courier',
    color: '#1C1C1E',
    lineHeight: 17,
  },
  diffContainer: {
    borderRadius: 6,
    overflow: 'hidden',
  },
  diffRemoved: {
    backgroundColor: '#FFF0F0',
    color: '#FF3B30',
    fontSize: 12,
    fontFamily: 'Courier',
    padding: 8,
    lineHeight: 17,
  },
  diffAdded: {
    backgroundColor: '#F0FFF4',
    color: '#34C759',
    fontSize: 12,
    fontFamily: 'Courier',
    padding: 8,
    lineHeight: 17,
  },
  expandHint: {
    fontSize: 11,
    color: '#8E8E93',
    marginTop: 8,
    textAlign: 'right',
  },
});
