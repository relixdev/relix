import React, { useState } from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  Alert,
  ScrollView,
  Switch,
  TextInput,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import * as WebBrowser from 'expo-web-browser';
import { useNavigation } from '@react-navigation/native';
import type { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { useAuthStore } from '../../stores/authStore';
import { useMachineStore, type Machine } from '../../stores/machineStore';
import * as api from '../../lib/api';
import { deletePeerKey } from '../sas-verification';
import type { RootStackParamList } from '../../lib/navigationRef';

type SettingsNavProp = NativeStackNavigationProp<RootStackParamList>;

// ─── Tier badge ──────────────────────────────────────────────────────────────

function TierBadge({ tier }: { tier: string }) {
  const isPro = tier === 'pro' || tier === 'plus';
  return (
    <View style={[styles.tierBadge, isPro && styles.tierBadgePro]}>
      <Text style={[styles.tierBadgeText, isPro && styles.tierBadgeTextPro]}>
        {tier.toUpperCase()}
      </Text>
    </View>
  );
}

// ─── Machine row with inline rename ──────────────────────────────────────────

interface MachineRowProps {
  machine: Machine;
  token: string;
  onRemoved: () => void;
}

function MachineRow({ machine, token, onRemoved }: MachineRowProps) {
  const [editing, setEditing] = useState(false);
  const [name, setName] = useState(machine.name);
  const [saving, setSaving] = useState(false);

  const statusColor =
    machine.status === 'online' ? '#16a34a' :
    machine.status === 'active' ? '#007AFF' :
    '#94a3b8';

  const handleSaveName = async () => {
    if (!name.trim() || name.trim() === machine.name) {
      setEditing(false);
      setName(machine.name);
      return;
    }
    setSaving(true);
    try {
      await api.renameMachine(token, machine.id, name.trim());
      setEditing(false);
    } catch (e: any) {
      Alert.alert('Error', e.message ?? 'Failed to rename machine');
      setName(machine.name);
      setEditing(false);
    } finally {
      setSaving(false);
    }
  };

  const handleRemove = () => {
    Alert.alert(
      'Remove Machine',
      `Remove "${machine.name}"? This will unpair the machine and delete its keys.`,
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Remove',
          style: 'destructive',
          onPress: async () => {
            try {
              await api.deleteMachine(token, machine.id);
              await deletePeerKey(machine.id);
              onRemoved();
            } catch (e: any) {
              Alert.alert('Error', e.message ?? 'Failed to remove machine');
            }
          },
        },
      ],
    );
  };

  return (
    <View style={styles.machineRow}>
      <View style={styles.machineRowLeft}>
        <View style={[styles.statusDot, { backgroundColor: statusColor }]} />
        {editing ? (
          <TextInput
            style={styles.machineNameInput}
            value={name}
            onChangeText={setName}
            onBlur={handleSaveName}
            onSubmitEditing={handleSaveName}
            autoFocus
            editable={!saving}
            returnKeyType="done"
          />
        ) : (
          <TouchableOpacity onPress={() => setEditing(true)}>
            <Text style={styles.machineName}>{machine.name}</Text>
          </TouchableOpacity>
        )}
      </View>
      <TouchableOpacity style={styles.removeButton} onPress={handleRemove}>
        <Text style={styles.removeButtonText}>Remove</Text>
      </TouchableOpacity>
    </View>
  );
}

// ─── Main Settings screen ─────────────────────────────────────────────────────

export default function SettingsScreen() {
  const { user, logout } = useAuthStore();
  const { token } = useAuthStore();
  const { machines, fetchMachines } = useMachineStore();
  const navigation = useNavigation<SettingsNavProp>();

  const [notificationsEnabled, setNotificationsEnabled] = useState(true);
  const [defaultView, setDefaultView] = useState<'chat' | 'terminal'>('chat');

  const handleLogout = () => {
    Alert.alert('Sign Out', 'Are you sure you want to sign out?', [
      { text: 'Cancel', style: 'cancel' },
      {
        text: 'Sign Out',
        style: 'destructive',
        onPress: async () => {
          await logout();
        },
      },
    ]);
  };

  const handleManageSubscription = async () => {
    if (!token) return;
    try {
      const { portal_url } = await api.createBillingPortalSession(
        token,
        'relix://settings',
      );
      await WebBrowser.openBrowserAsync(portal_url);
    } catch {
      // Billing portal not yet implemented — show fallback
      Alert.alert('Manage Subscription', 'Visit relix.sh/billing to manage your subscription.');
    }
  };

  const handleUpgrade = async () => {
    if (!token) return;
    try {
      const { checkout_url } = await api.createCheckoutSession(token, 'plus');
      await WebBrowser.openBrowserAsync(checkout_url);
    } catch (e: any) {
      Alert.alert('Error', e.message ?? 'Failed to start checkout');
    }
  };

  const handleAddMachine = () => {
    navigation.navigate('Pairing');
  };

  const handleHelp = () => {
    WebBrowser.openBrowserAsync('https://relix.sh/docs').catch(() => {
      Alert.alert('Help & Support', 'Visit relix.sh/docs or email support@relix.sh');
    });
  };

  const handlePrivacy = () => {
    WebBrowser.openBrowserAsync('https://relix.sh/privacy').catch(() => {
      Alert.alert('Privacy Policy', 'Visit relix.sh/privacy to read our privacy policy.');
    });
  };

  const handleMachineRemoved = () => {
    if (token) {
      fetchMachines(token).catch(() => {});
    }
  };

  return (
    <SafeAreaView style={styles.root} edges={['top']}>
      <ScrollView contentContainerStyle={styles.scroll}>
        <View style={styles.header}>
          <Text style={styles.title}>Settings</Text>
        </View>

        {/* Account */}
        {user && (
          <View style={styles.section}>
            <Text style={styles.sectionLabel}>Account</Text>
            <View style={styles.card}>
              <View style={styles.row}>
                <Text style={styles.rowLabel}>Email</Text>
                <Text style={styles.rowValue} numberOfLines={1}>
                  {user.email || '---'}
                </Text>
              </View>
              <View style={styles.row}>
                <Text style={styles.rowLabel}>Plan</Text>
                <TierBadge tier={user.tier} />
              </View>
              {user.tier === 'free' && (
                <TouchableOpacity
                  style={[styles.row, styles.rowTouchable]}
                  onPress={handleUpgrade}
                >
                  <Text style={[styles.rowLabel, { color: '#6366f1' }]}>Upgrade to Plus</Text>
                  <Text style={styles.chevron}>›</Text>
                </TouchableOpacity>
              )}
              <TouchableOpacity
                style={[styles.row, styles.rowLast, styles.rowTouchable]}
                onPress={handleManageSubscription}
              >
                <Text style={styles.rowLabel}>Manage Subscription</Text>
                <Text style={styles.chevron}>›</Text>
              </TouchableOpacity>
            </View>
          </View>
        )}

        {/* Machines */}
        <View style={styles.section}>
          <Text style={styles.sectionLabel}>Machines</Text>
          {machines.length === 0 ? (
            <View style={styles.card}>
              <View style={[styles.row, styles.rowLast]}>
                <Text style={styles.emptyMachines}>No machines paired yet</Text>
              </View>
            </View>
          ) : (
            <View style={styles.card}>
              {machines.map((machine, index) => (
                <View
                  key={machine.id}
                  style={index < machines.length - 1 ? undefined : styles.lastMachineRow}
                >
                  {token && (
                    <MachineRow
                      machine={machine}
                      token={token}
                      onRemoved={handleMachineRemoved}
                    />
                  )}
                  {index < machines.length - 1 && <View style={styles.divider} />}
                </View>
              ))}
            </View>
          )}
          <TouchableOpacity style={styles.addMachineButton} onPress={handleAddMachine}>
            <Text style={styles.addMachineText}>+ Add Machine</Text>
          </TouchableOpacity>
        </View>

        {/* Preferences */}
        <View style={styles.section}>
          <Text style={styles.sectionLabel}>Preferences</Text>
          <View style={styles.card}>
            <View style={styles.row}>
              <Text style={styles.rowLabel}>Notifications</Text>
              <Switch
                value={notificationsEnabled}
                onValueChange={setNotificationsEnabled}
                trackColor={{ true: '#007AFF', false: '#e2e8f0' }}
                thumbColor="#fff"
              />
            </View>
            <View style={[styles.row, styles.rowLast]}>
              <Text style={styles.rowLabel}>Default View</Text>
              <View style={styles.segmentControl}>
                <TouchableOpacity
                  style={[
                    styles.segment,
                    defaultView === 'chat' && styles.segmentActive,
                  ]}
                  onPress={() => setDefaultView('chat')}
                >
                  <Text
                    style={[
                      styles.segmentText,
                      defaultView === 'chat' && styles.segmentTextActive,
                    ]}
                  >
                    Chat
                  </Text>
                </TouchableOpacity>
                <TouchableOpacity
                  style={[
                    styles.segment,
                    defaultView === 'terminal' && styles.segmentActive,
                  ]}
                  onPress={() => setDefaultView('terminal')}
                >
                  <Text
                    style={[
                      styles.segmentText,
                      defaultView === 'terminal' && styles.segmentTextActive,
                    ]}
                  >
                    Terminal
                  </Text>
                </TouchableOpacity>
              </View>
            </View>
          </View>
        </View>

        {/* About */}
        <View style={styles.section}>
          <Text style={styles.sectionLabel}>About</Text>
          <View style={styles.card}>
            <View style={styles.row}>
              <Text style={styles.rowLabel}>Version</Text>
              <Text style={styles.rowValue}>1.0.0</Text>
            </View>
            <TouchableOpacity
              style={[styles.row, styles.rowTouchable]}
              onPress={handleHelp}
            >
              <Text style={styles.rowLabel}>Help & Support</Text>
              <Text style={styles.chevron}>›</Text>
            </TouchableOpacity>
            <TouchableOpacity
              style={[styles.row, styles.rowLast, styles.rowTouchable]}
              onPress={handlePrivacy}
            >
              <Text style={styles.rowLabel}>Privacy Policy</Text>
              <Text style={styles.chevron}>›</Text>
            </TouchableOpacity>
          </View>
        </View>

        {/* Sign Out */}
        <View style={styles.section}>
          <TouchableOpacity style={styles.logoutButton} onPress={handleLogout}>
            <Text style={styles.logoutText}>Sign Out</Text>
          </TouchableOpacity>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  root: {
    flex: 1,
    backgroundColor: '#f8fafc',
  },
  scroll: {
    paddingBottom: 40,
  },
  header: {
    paddingHorizontal: 20,
    paddingTop: 8,
    paddingBottom: 16,
  },
  title: {
    fontSize: 28,
    fontWeight: '800',
    color: '#0f172a',
    letterSpacing: -0.5,
  },
  section: {
    marginHorizontal: 16,
    marginBottom: 24,
  },
  sectionLabel: {
    fontSize: 13,
    fontWeight: '600',
    color: '#64748b',
    textTransform: 'uppercase',
    letterSpacing: 0.8,
    marginBottom: 8,
    paddingHorizontal: 4,
  },
  card: {
    backgroundColor: '#fff',
    borderRadius: 12,
    overflow: 'hidden',
    shadowColor: '#000',
    shadowOpacity: 0.05,
    shadowRadius: 6,
    shadowOffset: { width: 0, height: 2 },
    elevation: 2,
  },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 14,
    borderBottomWidth: 1,
    borderBottomColor: '#f1f5f9',
  },
  rowLast: {
    borderBottomWidth: 0,
  },
  rowTouchable: {},
  rowLabel: {
    fontSize: 15,
    color: '#0f172a',
  },
  rowValue: {
    fontSize: 15,
    color: '#64748b',
    maxWidth: '60%',
  },
  chevron: {
    fontSize: 20,
    color: '#94a3b8',
  },
  tierBadge: {
    backgroundColor: '#e2e8f0',
    borderRadius: 6,
    paddingHorizontal: 8,
    paddingVertical: 2,
  },
  tierBadgePro: {
    backgroundColor: '#ede9fe',
  },
  tierBadgeText: {
    fontSize: 11,
    fontWeight: '700',
    color: '#64748b',
    letterSpacing: 0.5,
  },
  tierBadgeTextPro: {
    color: '#7c3aed',
  },
  machineRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 14,
  },
  machineRowLeft: {
    flexDirection: 'row',
    alignItems: 'center',
    flex: 1,
    gap: 10,
  },
  statusDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
  },
  machineName: {
    fontSize: 15,
    color: '#0f172a',
  },
  machineNameInput: {
    fontSize: 15,
    color: '#0f172a',
    borderBottomWidth: 1,
    borderBottomColor: '#007AFF',
    paddingVertical: 2,
    minWidth: 120,
  },
  removeButton: {
    paddingHorizontal: 10,
    paddingVertical: 6,
  },
  removeButtonText: {
    fontSize: 13,
    color: '#dc2626',
    fontWeight: '500',
  },
  divider: {
    height: 1,
    backgroundColor: '#f1f5f9',
    marginLeft: 16,
  },
  lastMachineRow: {},
  emptyMachines: {
    fontSize: 14,
    color: '#94a3b8',
  },
  addMachineButton: {
    marginTop: 12,
    paddingVertical: 12,
    alignItems: 'center',
    backgroundColor: '#f1f5f9',
    borderRadius: 10,
    borderWidth: 1,
    borderColor: '#e2e8f0',
    borderStyle: 'dashed',
  },
  addMachineText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#6366f1',
  },
  segmentControl: {
    flexDirection: 'row',
    backgroundColor: '#f1f5f9',
    borderRadius: 8,
    padding: 2,
  },
  segment: {
    paddingHorizontal: 14,
    paddingVertical: 6,
    borderRadius: 6,
  },
  segmentActive: {
    backgroundColor: '#fff',
    shadowColor: '#000',
    shadowOpacity: 0.06,
    shadowRadius: 3,
    shadowOffset: { width: 0, height: 1 },
    elevation: 1,
  },
  segmentText: {
    fontSize: 13,
    color: '#64748b',
    fontWeight: '500',
  },
  segmentTextActive: {
    color: '#0f172a',
    fontWeight: '600',
  },
  logoutButton: {
    backgroundColor: '#fee2e2',
    borderRadius: 12,
    paddingVertical: 16,
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#fca5a5',
  },
  logoutText: {
    color: '#dc2626',
    fontSize: 16,
    fontWeight: '600',
  },
});
