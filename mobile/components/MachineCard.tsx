import React from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
} from 'react-native';
import type { Machine } from '../stores/machineStore';

interface Props {
  machine: Machine;
  onPress: (machine: Machine) => void;
}

const STATUS_COLOR: Record<Machine['status'], string> = {
  online: '#10b981',
  active: '#6366f1',
  offline: '#94a3b8',
};

const STATUS_LABEL: Record<Machine['status'], string> = {
  online: 'Online',
  active: 'Active',
  offline: 'Offline',
};

export default function MachineCard({ machine, onPress }: Props) {
  const activeSessions = machine.sessions.filter((s) => s.status === 'active').length;
  const dotColor = STATUS_COLOR[machine.status];

  return (
    <TouchableOpacity style={styles.card} onPress={() => onPress(machine)} activeOpacity={0.75}>
      <View style={styles.row}>
        <View style={[styles.statusDot, { backgroundColor: dotColor }]} />
        <View style={styles.info}>
          <Text style={styles.name}>{machine.name}</Text>
          <Text style={styles.status}>{STATUS_LABEL[machine.status]}</Text>
        </View>
        {activeSessions > 0 ? (
          <View style={styles.sessionBadge}>
            <Text style={styles.sessionBadgeText}>
              {activeSessions} session{activeSessions !== 1 ? 's' : ''}
            </Text>
          </View>
        ) : null}
        <Text style={styles.chevron}>›</Text>
      </View>
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    marginBottom: 10,
    shadowColor: '#000',
    shadowOpacity: 0.05,
    shadowRadius: 6,
    shadowOffset: { width: 0, height: 2 },
    elevation: 2,
  },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
  },
  statusDot: {
    width: 10,
    height: 10,
    borderRadius: 5,
  },
  info: {
    flex: 1,
  },
  name: {
    fontSize: 16,
    fontWeight: '600',
    color: '#0f172a',
  },
  status: {
    fontSize: 13,
    color: '#64748b',
    marginTop: 2,
  },
  sessionBadge: {
    backgroundColor: '#ede9fe',
    borderRadius: 6,
    paddingHorizontal: 8,
    paddingVertical: 3,
  },
  sessionBadgeText: {
    color: '#6366f1',
    fontSize: 12,
    fontWeight: '600',
  },
  chevron: {
    fontSize: 20,
    color: '#cbd5e1',
    fontWeight: '300',
  },
});
