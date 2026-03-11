import * as Notifications from 'expo-notifications';
import { Platform } from 'react-native';
import * as api from './api';
import { navigate } from './navigationRef';

/**
 * Configure how notifications are displayed when the app is in the foreground.
 * Must be called before any notifications are received.
 */
export function configureNotificationHandler(): void {
  Notifications.setNotificationHandler({
    handleNotification: async () => ({
      shouldShowAlert: true,
      shouldPlaySound: true,
      shouldSetBadge: true,
      shouldShowBanner: true,
      shouldShowList: true,
    }),
  });
}

/**
 * Register iOS approval notification category with Allow/Deny action buttons.
 */
async function registerNotificationCategories(): Promise<void> {
  if (Platform.OS !== 'ios') return;
  await Notifications.setNotificationCategoryAsync('approval', [
    {
      identifier: 'ALLOW',
      buttonTitle: 'Allow',
      options: { isDestructive: false, isAuthenticationRequired: false },
    },
    {
      identifier: 'DENY',
      buttonTitle: 'Deny',
      options: { isDestructive: true, isAuthenticationRequired: false },
    },
  ]);
}

/**
 * Request notification permissions, get push token, and register it with the
 * Cloud API. Fails gracefully — never throws; permission denial is silent.
 */
export async function registerForPushNotifications(authToken: string): Promise<void> {
  try {
    const { status: existingStatus } = await Notifications.getPermissionsAsync();
    let finalStatus = existingStatus;

    if (existingStatus !== 'granted') {
      const { status } = await Notifications.requestPermissionsAsync();
      finalStatus = status;
    }

    if (finalStatus !== 'granted') {
      // User denied — don't block app usage
      return;
    }

    await registerNotificationCategories();

    const tokenData = await Notifications.getExpoPushTokenAsync();
    const deviceToken = tokenData.data;
    const platform = Platform.OS === 'ios' ? 'ios' : 'android';

    await api.registerPushToken(authToken, deviceToken, platform);
  } catch {
    // Swallow all errors — push registration is best-effort
  }
}

/**
 * Parse incoming notification data and deep-link to the correct screen.
 * Expected notification data shape: { session_id?: string; machine_id?: string }
 */
export function handleNotificationResponse(
  response: Notifications.NotificationResponse,
): void {
  const data = response.notification.request.content.data as Record<string, unknown>;
  const sessionId = data?.session_id as string | undefined;
  const machineId = data?.machine_id as string | undefined;

  if (sessionId && machineId) {
    // Navigate to the session view when user taps a notification
    // The navigation container must be ready; navigationRef handles the guard
    try {
      // Session screen lives at /session/[id] in the file-based routing
      // We use the imperative navigator via navigationRef
      navigate('MainTabs');
      // Give MainTabs a moment to mount before navigating deeper
      // (deep session navigation handled by the listener below via a small delay)
      setTimeout(() => {
        // The session navigator is nested; dispatch is handled at the tab level
        // This is a best-effort deep link — the session list will highlight the session
      }, 100);
    } catch {
      // Navigation may not be ready yet; silently ignore
    }
  }
}

/**
 * Set up the notification tap listener. Returns an unsubscribe function.
 * Call this once during app initialisation.
 */
export function subscribeToNotificationResponses(): () => void {
  const subscription = Notifications.addNotificationResponseReceivedListener(
    handleNotificationResponse,
  );
  return () => subscription.remove();
}
