import React from 'react';
import { Alert } from 'react-native';
import { useNavigation, useRoute } from '@react-navigation/native';
import type { NativeStackNavigationProp } from '@react-navigation/native-stack';
import type { RouteProp } from '@react-navigation/native';
import * as SecureStore from 'expo-secure-store';
import SASVerification from '../components/SASVerification';
import { useMachineStore } from '../stores/machineStore';
import { useAuthStore } from '../stores/authStore';
import type { RootStackParamList } from '../lib/navigationRef';

type SASNavProp = NativeStackNavigationProp<RootStackParamList, 'SASVerification'>;
type SASRouteProp = RouteProp<RootStackParamList, 'SASVerification'>;

// Key used to persist peer public keys per machine
function peerKeyStorageKey(machineId: string): string {
  return `relix_peer_key_${machineId}`;
}

export async function savePeerKey(machineId: string, publicKeyHex: string): Promise<void> {
  await SecureStore.setItemAsync(peerKeyStorageKey(machineId), publicKeyHex);
}

export async function loadPeerKey(machineId: string): Promise<string | null> {
  return SecureStore.getItemAsync(peerKeyStorageKey(machineId));
}

export async function deletePeerKey(machineId: string): Promise<void> {
  await SecureStore.deleteItemAsync(peerKeyStorageKey(machineId));
}

export default function SASVerificationScreen() {
  const navigation = useNavigation<SASNavProp>();
  const route = useRoute<SASRouteProp>();
  const { peerPublicKey, machineId } = route.params;
  const { token } = useAuthStore();
  const { fetchMachines } = useMachineStore();

  // Use a placeholder own public key hex until real key management is wired in.
  // In production this would come from the locally generated X25519 key pair.
  const ownPublicKeyHex = 'ownkey_placeholder_00000000000000000000000000000000';

  const handleConfirm = async () => {
    try {
      await savePeerKey(machineId, peerPublicKey);
      if (token) {
        await fetchMachines(token).catch(() => {});
      }
      navigation.reset({ index: 0, routes: [{ name: 'MainTabs' }] });
    } catch (e: any) {
      Alert.alert('Error', e.message ?? 'Failed to save peer key');
    }
  };

  const handleReject = () => {
    // Pairing aborted — go back to start pairing or onboarding
    navigation.reset({ index: 0, routes: [{ name: 'Onboarding' }] });
  };

  return (
    <SASVerification
      ownPublicKeyHex={ownPublicKeyHex}
      peerPublicKeyHex={peerPublicKey}
      onConfirm={handleConfirm}
      onReject={handleReject}
    />
  );
}
