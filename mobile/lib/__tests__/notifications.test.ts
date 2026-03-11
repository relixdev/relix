// Minimal smoke tests for notification utilities.
// The actual expo-notifications API is not available in the Node test environment,
// so we mock it and verify the integration logic.

jest.mock('expo-notifications', () => ({
  setNotificationHandler: jest.fn(),
  getPermissionsAsync: jest.fn().mockResolvedValue({ status: 'granted' }),
  requestPermissionsAsync: jest.fn().mockResolvedValue({ status: 'granted' }),
  getExpoPushTokenAsync: jest.fn().mockResolvedValue({ data: 'ExponentPushToken[test]' }),
  setNotificationCategoryAsync: jest.fn().mockResolvedValue(undefined),
  addNotificationResponseReceivedListener: jest.fn().mockReturnValue({ remove: jest.fn() }),
}));

jest.mock('react-native', () => ({
  Platform: { OS: 'ios' },
}));

jest.mock('../api', () => ({
  registerPushToken: jest.fn().mockResolvedValue(undefined),
}));

jest.mock('../navigationRef', () => ({
  navigate: jest.fn(),
}));

import * as Notifications from 'expo-notifications';
import * as api from '../api';
import {
  configureNotificationHandler,
  registerForPushNotifications,
  subscribeToNotificationResponses,
} from '../notifications';

describe('configureNotificationHandler', () => {
  it('calls setNotificationHandler', () => {
    configureNotificationHandler();
    expect(Notifications.setNotificationHandler).toHaveBeenCalledTimes(1);
  });
});

describe('registerForPushNotifications', () => {
  it('requests permissions and registers token', async () => {
    await registerForPushNotifications('test-auth-token');
    expect(Notifications.getPermissionsAsync).toHaveBeenCalled();
    expect(api.registerPushToken).toHaveBeenCalledWith(
      'test-auth-token',
      'ExponentPushToken[test]',
      'ios',
    );
  });

  it('does not throw when permission is denied', async () => {
    (Notifications.getPermissionsAsync as jest.Mock).mockResolvedValueOnce({
      status: 'denied',
    });
    (Notifications.requestPermissionsAsync as jest.Mock).mockResolvedValueOnce({
      status: 'denied',
    });
    await expect(registerForPushNotifications('tok')).resolves.toBeUndefined();
    // registerPushToken should NOT have been called again in this test
  });

  it('swallows API errors gracefully', async () => {
    (api.registerPushToken as jest.Mock).mockRejectedValueOnce(new Error('network error'));
    await expect(registerForPushNotifications('tok')).resolves.toBeUndefined();
  });
});

describe('subscribeToNotificationResponses', () => {
  it('returns an unsubscribe function', () => {
    const unsub = subscribeToNotificationResponses();
    expect(typeof unsub).toBe('function');
    expect(() => unsub()).not.toThrow();
  });
});
