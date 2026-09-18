import {Platform} from 'react-native';
import NativeFoodPhotoPicker from '../../specs/NativeFoodPhotoPicker';

type NativePickerError = Error & {code?: string};

export function isFoodPhotoPickCancelled(error: unknown) {
  return (error as NativePickerError | null)?.code === 'PHOTO_PICK_CANCELLED';
}

export function foodPhotoErrorMessage(error: unknown) {
  const code = (error as NativePickerError | null)?.code;
  switch (code) {
    case 'PHOTO_PICK_TOO_LARGE': return 'Фото больше 5 МБ. Выберите снимок меньшего размера.';
    case 'PHOTO_PICK_UNSUPPORTED': return 'Формат фото не поддерживается. Выберите JPEG, PNG или WebP.';
    case 'PHOTO_PICK_EMPTY': return 'Не удалось прочитать выбранное фото. Выберите другой снимок.';
    case 'PHOTO_PICK_NO_ACTIVITY': return 'Галерея временно недоступна. Вернитесь в приложение и попробуйте снова.';
    default: return error instanceof Error && error.message ? error.message : 'Не удалось обработать фото. Попробуйте снова.';
  }
}

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
