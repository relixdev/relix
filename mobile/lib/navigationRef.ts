import { createNavigationContainerRef, CommonActions } from '@react-navigation/native';

export type RootStackParamList = {
  Login: undefined;
  Onboarding: undefined;
  Pairing: undefined;
  SASVerification: { pairingCode: string; peerPublicKey: string; machineId: string };
  MainTabs: undefined;
  Session: { id: string; machineId: string };
};

export const navigationRef = createNavigationContainerRef<RootStackParamList>();

export function navigate(name: keyof RootStackParamList, params?: any) {
  if (navigationRef.isReady()) {
    navigationRef.navigate(name as any, params);
  }
}

export function reset(name: keyof RootStackParamList) {
  if (navigationRef.isReady()) {
    navigationRef.dispatch(
      CommonActions.reset({
        index: 0,
        routes: [{ name }],
      }),
    );
  }
}

// Thin hooks used in screens to avoid importing navigationRef directly
export function useAuthNavigation() {
  return {
    navigateToOnboarding: () => reset('Onboarding'),
    navigateToDashboard: () => reset('MainTabs'),
  };
}

export function useOnboardingNavigation() {
  return {
    navigateToDashboard: () => reset('MainTabs'),
  };
}
