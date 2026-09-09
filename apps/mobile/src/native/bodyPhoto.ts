import {Platform} from 'react-native';
import NativeBodyPhotoCapture from '../../specs/NativeBodyPhotoCapture';

export const bodyPhotoCapture = {
  get available() {
    return Platform.OS === 'android' && Boolean(NativeBodyPhotoCapture);
  },
  async captureImage() {
    if (!NativeBodyPhotoCapture || Platform.OS !== 'android') {
      throw new Error('Камера Body Scan пока недоступна на этом устройстве.');
    }
    return NativeBodyPhotoCapture.captureImage();
  },
};
