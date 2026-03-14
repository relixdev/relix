import React, { useCallback } from 'react';
import {
  View,
  Text,
  FlatList,
  StyleSheet,
  RefreshControl,
  Alert,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useNavigation } from '@react-navigation/native';
import type { NativeStackNavigationProp } from '@react-navigation/native-stack';
import ApprovalCard from '../../components/ApprovalCard';
import MachineCard from '../../components/MachineCard';
import { useAuthStore } from '../../stores/authStore';
import { useSessionStore } from '../../stores/sessionStore';
import { useMachineStore, type Machine, type Approval } from '../../stores/machineStore';

type DashboardNavProp = NativeStackNavigationProp<any>;

function MachineSkeletonRow() {
  return (
    <View style={skeletonStyles.row}>
      <View style={skeletonStyles.dot} />
      <View style={skeletonStyles.textBlock} />
      <View style={skeletonStyles.statusBlock} />
    </View>
  );
}

const skeletonStyles = StyleSheet.create({
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    marginBottom: 10,
    gap: 12,
  },
  dot: {
    width: 10,
    height: 10,
    borderRadius: 5,
    backgroundColor: '#e2e8f0',
  },
  textBlock: {
    flex: 1,
    height: 14,
    borderRadius: 7,
    backgroundColor: '#e2e8f0',
  },
  statusBlock: {
    width: 56,
    height: 22,
    borderRadius: 11,
    backgroundColor: '#e2e8f0',
  },
});

export default function DashboardScreen() {
  const { token } = useAuthStore();
  const navigation = useNavigation<DashboardNavProp>();
  const setSession = useSessionStore((s) => s.setSession);
  const {
    machines,
    pendingApprovals,
    fetchMachines,
    isLoading,
    error,
    resolveApproval,
  } = useMachineStore();

  const onRefresh = useCallback(async () => {
    if (!token) return;
    try {
      await fetchMachines(token);
    } catch (e: any) {
      Alert.alert('Error', e.message ?? 'Failed to refresh machines');
    }
  }, [token]);

  const handleAllow = useCallback(
    (approvalId: string) => {
      resolveApproval(approvalId, true);
    },
    [resolveApproval],
  );

  const handleDeny = useCallback(
    (approvalId: string) => {
      resolveApproval(approvalId, false);
    },
    [resolveApproval],
  );

  const handleMachinePress = useCallback((machine: Machine) => {
    if (machine.sessions.length === 0) {
      Alert.alert(machine.name, `Status: ${machine.status}\nNo active sessions`);
      return;
    }
    if (machine.sessions.length === 1) {
      // Go directly to the single session
      const session = machine.sessions[0];
      setSession(session.id);
      navigation.navigate('Session', { id: session.id, machineId: machine.id });
      return;
    }
    // Multiple sessions — let user pick
    Alert.alert(
      machine.name,
      'Select a session:',
      [
        ...machine.sessions.map((s) => ({
          text: `${s.project} (${s.status})`,
          onPress: () => {
            setSession(s.id);
            navigation.navigate('Session', { id: s.id, machineId: machine.id });
          },
        })),
        { text: 'Cancel', style: 'cancel' as const },
      ],
    );
  }, [navigation, setSession]);

  const ListHeader = (
    <>
      <View style={styles.header}>
        <Text style={styles.headerTitle}>Relix</Text>
      </View>

      {error && (
        <View style={styles.errorBanner}>
          <Text style={styles.errorText}>⚠️  {error}</Text>
        </View>
      )}

      {pendingApprovals.length > 0 && (
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>
            Pending Approvals ({pendingApprovals.length})
          </Text>
          {pendingApprovals.map((approval: Approval) => (
            <ApprovalCard
              key={approval.id}
              approval={approval}
              onAllow={handleAllow}
              onDeny={handleDeny}
            />
          ))}
        </View>
      )}

      <Text style={styles.sectionTitle}>Machines</Text>

      {isLoading && machines.length === 0 && (
        <>
          <MachineSkeletonRow />
          <MachineSkeletonRow />
          <MachineSkeletonRow />
        </>
      )}
    </>
  );

  const ListEmpty = !isLoading ? (
    <View style={styles.emptyState}>
      <Text style={styles.emptyIcon}>💻</Text>
      <Text style={styles.emptyTitle}>No machines connected</Text>
      <Text style={styles.emptySubtitle}>
        Go to Settings to add your first machine.
      </Text>
    </View>
  ) : null;

  return (
    <SafeAreaView style={styles.root} edges={['top']}>
      <FlatList
        data={machines}
        keyExtractor={(m) => m.id}
        renderItem={({ item }) => (
          <MachineCard machine={item} onPress={handleMachinePress} />
        )}
        ListHeaderComponent={ListHeader}
        ListEmptyComponent={ListEmpty}
        contentContainerStyle={styles.list}
        refreshControl={
          <RefreshControl
            refreshing={isLoading && machines.length > 0}
            onRefresh={onRefresh}
            tintColor="#007AFF"
          />
        }
      />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  root: {
    flex: 1,
    backgroundColor: '#f8fafc',
  },
  header: {
    paddingHorizontal: 20,
    paddingTop: 8,
    paddingBottom: 16,
  },
  headerTitle: {
    fontSize: 28,
    fontWeight: '800',
    color: '#0f172a',
    letterSpacing: -0.5,
  },
  list: {
    paddingHorizontal: 16,
    paddingBottom: 32,
  },
  section: {
    marginBottom: 8,
  },
  sectionTitle: {
    fontSize: 13,
    fontWeight: '600',
    color: '#64748b',
    textTransform: 'uppercase',
    letterSpacing: 0.8,
    marginBottom: 10,
    paddingHorizontal: 4,
  },
  emptyState: {
    alignItems: 'center',
    paddingTop: 60,
    paddingHorizontal: 32,
  },
  emptyIcon: {
    fontSize: 48,
    marginBottom: 16,
  },
  emptyTitle: {
    fontSize: 18,
    fontWeight: '700',
    color: '#0f172a',
    marginBottom: 6,
    textAlign: 'center',
  },
  emptySubtitle: {
    fontSize: 14,
    color: '#64748b',
    textAlign: 'center',
  },
  errorBanner: {
    backgroundColor: '#fee2e2',
    borderRadius: 10,
    padding: 12,
    marginBottom: 12,
    borderWidth: 1,
    borderColor: '#fca5a5',
  },
  errorText: {
    fontSize: 13,
    color: '#dc2626',
    lineHeight: 18,
  },
});
