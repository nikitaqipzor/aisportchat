import type {AIFoodDraftItem, NutritionDay} from '../api/client';

export const MAX_FOOD_QUANTITY_G = 5000;

export type FoodQuantityResult =
  | {value: number; error: ''}
  | {value: null; error: string};

export function parseFoodQuantity(value: string): FoodQuantityResult {
  if (!value.trim()) {
    return {value: null, error: 'Укажите вес'};
  }
  const quantity = Number(value.replace(',', '.'));
  if (!Number.isFinite(quantity)) {
    return {value: null, error: 'Введите число'};
  }
  if (quantity <= 0 || quantity > MAX_FOOD_QUANTITY_G) {
    return {value: null, error: `Вес должен быть от 0 до ${MAX_FOOD_QUANTITY_G} г`};
  }
  return {value: quantity, error: ''};
}

export function catalogMacrosForQuantity(item: AIFoodDraftItem, quantityG: number) {
  const food = item.matched_food;
  if (!food) {
    return {calories: 0, proteinG: 0, fatG: 0, carbsG: 0};
  }
  const factor = quantityG / 100;
  return {
    calories: food.kcal_per_100g * factor,
    proteinG: food.protein_per_100g * factor,
    fatG: food.fat_per_100g * factor,
    carbsG: food.carbs_per_100g * factor,
  };
}

export function repeatedNutritionState(repeatedDay: NutritionDay) {
  return {selectedDate: repeatedDay.date, day: repeatedDay};
}
