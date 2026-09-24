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
requireSource(nutrition.includes('nutrition-return-today') && nutrition.includes('api.nutritionToday(accessToken, date, zone)') && nutrition.includes('api.nutritionHistory(accessToken, 7, zone)') && nutrition.includes('const date = selectedDateRef.current') && nutrition.includes('const zone = localTimeZone()'), 'past nutrition day and local time zone flow is incomplete');
requireSource(devices.includes('health-connection-progress') && devices.includes('ConnectionStep number={4}'), 'Health Connect setup progress is incomplete');
requireSource(devices.includes("origin='home'") && devices.includes("origin==='progress'?'\u0410\u043d\u0430\u043b\u0438\u0442\u0438\u043a\u0430':'\u0413\u043b\u0430\u0432\u043d\u0430\u044f'") && app.includes('origin={devicesOrigin}'), 'Connected Devices back destination label is not origin-aware');
requireSource(devices.includes("AppState.addEventListener('change'") && devices.includes('missing_permissions.map(permissionName)'), 'Health Connect settings recovery is incomplete');
requireSource(devices.includes('нет данных') && devices.includes('healthError(e'), 'Health Connect sync recovery feedback is missing');
requireSource(analytics.includes('analytics-tab-${id}'), 'analytics tab identifiers are missing');
requireSource(analytics.includes('Promise.allSettled') && analytics.includes('api.workoutProgressSummary') && analytics.includes('api.healthInsights'), 'analytics sources are not isolated');
requireSource(analytics.includes('analytics-connect-device') && app.includes("onDevices={() => setStep('devices')}"), 'analytics-to-device navigation is missing');

console.log('mobile nutrition/devices/analytics screen contract: PASS');
