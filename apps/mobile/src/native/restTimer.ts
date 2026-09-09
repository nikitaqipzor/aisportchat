import {PermissionsAndroid, Platform} from 'react-native';
import NativeRestTimerNotifications from '../../specs/NativeRestTimerNotifications';

export const restTimerNotifications = {
  get available() {
    return Platform.OS === 'android' && Boolean(NativeRestTimerNotifications);
  },
  async schedule(seconds: number, exerciseName: string) {
    if (!NativeRestTimerNotifications || Platform.OS !== 'android' || seconds <= 0) return false;
    try {
      if (Number(Platform.Version) >= 33) {
        const permission = PermissionsAndroid.PERMISSIONS.POST_NOTIFICATIONS;
        const alreadyGranted = await PermissionsAndroid.check(permission);
        if (!alreadyGranted) {
          const result = await PermissionsAndroid.request(permission);
          if (result !== PermissionsAndroid.RESULTS.GRANTED) return false;
        }
      }
      return await NativeRestTimerNotifications.schedule(seconds, exerciseName);
    } catch {
      return false;
    }
  },
  async cancel() {
    if (!NativeRestTimerNotifications || Platform.OS !== 'android') return false;
    try {
      return await NativeRestTimerNotifications.cancel();
    } catch {
      return false;
    }
  },
};
