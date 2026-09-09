# Sprint 2C — AI Layer

## Goal
Add the first bounded AI layer without moving workout or nutrition arithmetic into an LLM.

## Implemented vertical slices

### 1. Text food input
```text
"200 г творога, банан и 30 г протеина"
  -> AI extraction
  -> catalog matching
  -> editable draft
  -> explicit user confirmation
  -> deterministic nutrition service
  -> food diary
```

The AI returns names, estimated edible grams, confidence and notes. It never supplies trusted calorie/macro values. Calories and macros are calculated from the application's food catalog only after a catalog match.

### 2. Food photo analysis
The Android source contains a small `FoodPhotoPicker` native module. It picks an image from the gallery, caps the raw file at 5 MiB, converts it to a data URL and sends it to the AI endpoint. The HTTP endpoint has a dedicated 8 MiB request-body limit.

Vision output remains a draft. The UI clearly labels portion estimates as approximate and requires confirmation before logging.

### 3. AI Coach
`POST /api/v1/ai/chat` uses a provider abstraction. In production, the OpenAI provider uses the Responses API with read-only function tools:
- `get_profile`
- `get_today_nutrition`
- `get_recent_workouts`
- `get_progress_summary`
- `get_personal_records`

The model cannot directly mutate workouts, nutrition, profile or progress data.

### 4. Weekly AI report
The backend first calculates seven-day statistics deterministically:
- completed workouts;
- training volume;
- nutrition logging days;
- average calories/protein;
- nutrition targets;
- weight trend when available;
- new PR count.

Only that normalized structure is sent to the provider to produce `summary`, `wins`, `focus` and `next_actions`.

## Provider modes

### OpenAI
Enabled when `OPENAI_API_KEY` is present.

Environment:
```bash
OPENAI_API_KEY=...
OPENAI_MODEL=gpt-5.6-terra
```

### Deterministic local fallback
When no API key is configured, the API still starts. Text food parsing supports a small deterministic Russian alias set, weekly reports are generated locally and the Coach explains that the full provider is disabled. Photo analysis intentionally returns an error without a vision provider.

This mode exists for tests/offline development and never calls the network.

## Safety boundaries
- LLM is not a calorie/macro calculator.
- LLM is not the workout progression engine.
- Coach tools are read-only in Sprint 2C.
- Food mutations require a separate explicit confirm call.
- Image portion estimation is marked approximate.
- Coach prompt forbids diagnosis and directs medical questions to appropriate professional evaluation.
- Client history is capped; tools fetch fresh facts from backend.

## API added
- `GET /api/v1/ai/status`
- `POST /api/v1/ai/food/parse`
- `POST /api/v1/ai/food/photo`
- `POST /api/v1/ai/food/confirm`
- `POST /api/v1/ai/chat`
- `GET /api/v1/ai/reports/weekly`

OpenAPI: **0.8.0**.

## Mobile added
- `AIFoodInputScreen`
- `FoodPhotoScreen`
- `AICoachScreen`
- `WeeklyAIReportScreen`
- `AIFoodDraftView`
- Android `FoodPhotoPicker` native bridge source

## Tests
`TestSprint2CAIFlow` injects a fake provider and verifies:
1. onboarding;
2. nutrition setup;
3. AI text extraction;
4. explicit food confirmation;
5. photo-flow service wiring;
6. Coach tool execution;
7. weekly report;
8. provider status.

No external AI request is made by automated tests.

## Known limits / next slice
- The mobile client persists the last 30 Coach messages on-device and clears them on logout; persistent server-side cross-device conversation memory belongs in a later AI observability/memory slice.
- Production OpenAI E2E requires a real `OPENAI_API_KEY` and outbound network access.
- Android gallery module must be registered after the generated native shell exists.
- Camera capture/compression can replace gallery-only selection later.
