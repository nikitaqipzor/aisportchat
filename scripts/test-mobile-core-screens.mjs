import fs from 'node:fs';

function source(path) {
  return fs.readFileSync(path, 'utf8');
}

function requireSource(condition, message) {
  if (!condition) throw new Error(message);
}

const nutrition = source('apps/mobile/src/screens/NutritionScreen.tsx');
const devices = source('apps/mobile/src/screens/ConnectedDevicesScreen.tsx');
const analytics = source('apps/mobile/src/screens/ProgressScreen.tsx');
const app = source('apps/mobile/App.tsx');

requireSource(nutrition.includes('nutrition-previous-day') && nutrition.includes('nutrition-next-day'), 'nutrition date navigation is missing');
requireSource(nutrition.includes('nutrition-return-today') && nutrition.includes('api.nutritionToday(accessToken, selectedDate)'), 'past nutrition day flow is incomplete');
requireSource(devices.includes('health-connection-progress') && devices.includes('ConnectionStep number={4}'), 'Health Connect setup progress is incomplete');
requireSource(analytics.includes('analytics-tab-${id}'), 'analytics tab identifiers are missing');
requireSource(analytics.includes('Promise.allSettled') && analytics.includes('api.workoutProgressSummary') && analytics.includes('api.healthInsights'), 'analytics sources are not isolated');
requireSource(analytics.includes('analytics-connect-device') && app.includes("onDevices={() => setStep('devices')}"), 'analytics-to-device navigation is missing');

console.log('mobile nutrition/devices/analytics screen contract: PASS');
