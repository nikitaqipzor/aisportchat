import {Platform} from 'react-native';
import NativeFoodPhotoPicker from '../../specs/NativeFoodPhotoPicker';

export const foodPhotoPicker = {
  get available() {
    return Platform.OS === 'android' && Boolean(NativeFoodPhotoPicker);
  },
  async pickImage() {
    if (!NativeFoodPhotoPicker || Platform.OS !== 'android') {
      throw new Error('Нативный выбор фото пока недоступен на этом устройстве.');
    }
    return NativeFoodPhotoPicker.pickImage();
  },
};
