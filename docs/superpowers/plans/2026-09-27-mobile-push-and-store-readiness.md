# Mobile Push Client & Store Readiness Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make LedgerGuard's live mobile app (`frontend-flutter`) receive FCM push and deep-link on tap, and make the backend push carry the data needed to route — plus prep the app for store release.

**Architecture:** Backend already registers device tokens (`POST /api/v1/devices`) and sends FCM (`SendPush`); we extend `SendPush` with an additive `data` map and prune dead tokens. The Flutter app gains a `PushNotificationService` that registers the FCM token, re-registers on refresh, unregisters on logout, and routes notification taps through the existing `go_router`. Store-readiness is config (bundle IDs, signing, APNs, manifest).

**Tech Stack:** Go (backend, `firebase.google.com/go/v4/messaging`), Flutter/Dart (`frontend-flutter`, Provider + go_router + firebase_messaging), Firebase (Auth + FCM/APNs).

**Spec:** `docs/designs/18-mobile-app.md` (design) + `docs/prds/18-mobile-app.md` (PRD). Read both before starting.

## Global Constraints

- **Backend `data` change is additive** — the FCM `notification` block stays; old clients ignore `data`. No breaking change to `POST /api/v1/devices` (`{device_token, platform}`).
- **TDD, one feature per commit** (repo CLAUDE.md): write failing test → minimal code → pass → commit. Run `go test ./...` (backend) / `flutter test` (frontend) before each commit.
- **Domain layer keeps zero infra deps** (CLAUDE.md §3) — new push code lives in `application/service` + `infrastructure/external` (backend) and `lib/services` (Flutter), never `internal/domain`.
- **Platform values** are exactly `ios|android|web` (`entity/device_token.go:13`).
- **Live app is `frontend-flutter`** (Provider). Do not touch `frontend/app` (dead Bloc).
- **firebase_messaging pairs with firebase_core ^3.x** → use `^15.x`; let `flutter pub get` resolve the exact patch.
- **Bundle IDs are permanent once published** — the ID chosen in Task C1 is final; confirm with the owner before merging that task.

## Review Focus

Inputs the spec implies but tasks must be made to cover (each pinned to a task below):
1. **Push to a dead/expired token** — must delete the row, not retry forever (Task A3).
2. **Notification tap with missing/malformed `data`** — must fall back to Dashboard, never crash (Task B3).
3. **Cross-org `app_id` in a tap payload** — routing must not leak another org's data; unknown id → Dashboard (Task B3).
4. **Logout on a shared device** — token must be unregistered so the next user gets no prior alerts (Task B2).
5. **Register call fails / offline at login** — app keeps working; retry on next launch or token refresh (Task B1).

---

## Part A — Backend: data payload + dead-token pruning (Go)

### Task A1: `SendPush` gains a `data` map

**Files:**
- Modify: `backend/internal/application/service/notification_service.go` (interface `PushNotificationProvider`, ~line 22-27)
- Modify: `backend/internal/infrastructure/external/firebase_messaging.go:35` (`SendPush`)
- Test: `backend/internal/infrastructure/external/firebase_messaging_test.go` (new or existing)

**Interfaces:**
- Produces: `SendPush(ctx, deviceToken string, platform entity.Platform, title, body string, data map[string]string) error` — consumed by Task A2/A3 and by any `PushNotificationProvider` implementer.

- [ ] **Step 1: Update the interface signature**

In `notification_service.go`, change the interface:

```go
type PushNotificationProvider interface {
	// SendPush sends a push notification to a device. data is an optional
	// key/value payload delivered alongside the notification for client-side
	// deep-linking (e.g. {"type","app_id","subscription_id"}); nil is fine.
	SendPush(ctx context.Context, deviceToken string, platform entity.Platform, title string, body string, data map[string]string) error
}
```

- [ ] **Step 2: Write the failing test** (FCM message carries Data)

Add to `firebase_messaging_test.go` — assert the built `*messaging.Message` includes `Data`. Extract message-building into a testable helper:

```go
func TestBuildMessage_IncludesData(t *testing.T) {
	data := map[string]string{"type": "risk_alert", "app_id": "app-123", "subscription_id": "sub-9"}
	msg := buildMessage("tok", entity.PlatformAndroid, "T", "B", data)
	if msg.Data["type"] != "risk_alert" || msg.Data["app_id"] != "app-123" {
		t.Fatalf("expected data payload on message, got %#v", msg.Data)
	}
	if msg.Notification == nil || msg.Notification.Title != "T" {
		t.Fatal("notification block must remain")
	}
}
```

- [ ] **Step 3: Run it — expect FAIL** (`buildMessage` undefined)

Run: `cd backend && go test ./internal/infrastructure/external/ -run TestBuildMessage_IncludesData`
Expected: FAIL (undefined: buildMessage).

- [ ] **Step 4: Implement** — extract `buildMessage` and set `Data`

In `firebase_messaging.go`, refactor `SendPush` to build via a pure helper and add the `data` param:

```go
func (s *FirebaseMessagingService) SendPush(ctx context.Context, deviceToken string, platform entity.Platform, title, body string, data map[string]string) error {
	_, err := s.client.Send(ctx, buildMessage(deviceToken, platform, title, body, data))
	if err != nil {
		return fmt.Errorf("failed to send push notification: %w", err)
	}
	return nil
}

func buildMessage(deviceToken string, platform entity.Platform, title, body string, data map[string]string) *messaging.Message {
	message := &messaging.Message{
		Token:        deviceToken,
		Notification: &messaging.Notification{Title: title, Body: body},
		Data:         data,
	}
	switch platform {
	case entity.PlatformIOS:
		message.APNS = &messaging.APNSConfig{Payload: &messaging.APNSPayload{Aps: &messaging.Aps{Sound: "default", Badge: intPtr(1)}}}
	case entity.PlatformAndroid:
		message.Android = &messaging.AndroidConfig{Priority: "high", Notification: &messaging.AndroidNotification{Sound: "default", ClickAction: "FLUTTER_NOTIFICATION_CLICK"}}
	}
	return message
}
```

- [ ] **Step 5: Run test — expect PASS.** `go test ./internal/infrastructure/external/ -run TestBuildMessage_IncludesData`

- [ ] **Step 6: Fix compile fallout** — every `SendPush(...)` call now needs the 6th arg. Temporarily pass `nil` at the two call sites in `notification_service.go` (Task A2 fills them) and any test doubles. Run `go build ./...`.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/application/service/notification_service.go backend/internal/infrastructure/external/firebase_messaging.go backend/internal/infrastructure/external/firebase_messaging_test.go
git commit -m "feat(push): SendPush carries a data payload for deep-linking"
```

### Task A2: populate `data` at the risk-alert & daily-summary call sites

**Files:**
- Modify: `backend/internal/application/service/notification_service.go` (`SendCriticalAlert` ~141, `SendDailySummary` ~190)
- Modify: `backend/internal/application/service/webhook_service.go:448` (caller passes `app.ID`, `sub.ID`)
- Test: `backend/internal/application/service/notification_service_test.go`

**Interfaces:**
- Consumes: `SendPush(..., data)` from Task A1.
- Produces: `SendCriticalAlert(ctx, userID, appID uuid.UUID, subscriptionID uuid.UUID, appName, storeDomain string, oldState, newState valueobject.RiskState) error` (adds `appID`, `subscriptionID`).

- [ ] **Step 1: Write the failing test** — a fake push provider captures `data`.

```go
type capturingPush struct{ lastData map[string]string }
func (c *capturingPush) SendPush(_ context.Context, _ string, _ entity.Platform, _, _ string, data map[string]string) error {
	c.lastData = data; return nil
}

func TestSendCriticalAlert_IncludesDeepLinkData(t *testing.T) {
	cp := &capturingPush{}
	svc := newTestNotificationService(t, cp) // helper wiring repos with one device token + default prefs
	appID, subID := uuid.New(), uuid.New()
	err := svc.SendCriticalAlert(context.Background(), testUserID, appID, subID, "Acme", "acme.myshopify.com", valueobject.RiskStateSafe, valueobject.RiskStateOneCycleMissed)
	if err != nil { t.Fatalf("unexpected: %v", err) }
	if cp.lastData["type"] != "risk_alert" || cp.lastData["app_id"] != appID.String() || cp.lastData["subscription_id"] != subID.String() {
		t.Fatalf("bad data payload: %#v", cp.lastData)
	}
}
```

> If `newTestNotificationService` doesn't exist, add it in this step: construct `NotificationService` with in-memory fakes for `DeviceTokenRepository` (returns one token for `testUserID`) and `NotificationPreferencesRepository` (returns defaults), and the passed push provider.

- [ ] **Step 2: Run it — expect FAIL** (signature mismatch / data nil).
Run: `cd backend && go test ./internal/application/service/ -run TestSendCriticalAlert_IncludesDeepLinkData`

- [ ] **Step 3: Implement** — add params + build data.

Change `SendCriticalAlert` signature to add `appID, subscriptionID uuid.UUID`, and at the push loop:

```go
data := map[string]string{
	"type":            "risk_alert",
	"app_id":          appID.String(),
	"subscription_id": subscriptionID.String(),
}
for _, token := range tokens {
	if err := s.pushProvider.SendPush(ctx, token.DeviceToken, token.Platform, title, body, data); err != nil {
		lastErr = err
	}
}
```

In `SendDailySummary`, use `data := map[string]string{"type": "daily_summary", "app_id": appID.String()}` (add `appID uuid.UUID` param there too) and pass it to `SendPush`.

- [ ] **Step 4: Update the caller** — `webhook_service.go:448` `sendRiskChangeNotification` has `app *entity.App` and `sub *entity.Subscription` in scope; pass `app.ID, sub.ID`:

```go
if err := s.notificationSvc.SendCriticalAlert(ctx, userID, app.ID, sub.ID, app.Name, sub.MyshopifyDomain, oldRiskState, newRiskState); err != nil {
```

Find the `SendDailySummary` caller (scheduler) and thread the app's ID the same way. Run `go build ./...`.

- [ ] **Step 5: Run tests — expect PASS.** `go test ./internal/application/service/...`

- [ ] **Step 6: Commit**

```bash
git add backend/internal/application/service/
git commit -m "feat(push): deep-link data (type/app_id/subscription_id) on risk + daily push"
```

### Task A3: prune dead device tokens on FCM send failure

**Files:**
- Modify: `backend/internal/application/service/notification_service.go` (new sentinel + prune in both push loops)
- Modify: `backend/internal/infrastructure/external/firebase_messaging.go` (`SendPush` returns the sentinel when FCM says unregistered)
- Test: `backend/internal/application/service/notification_service_test.go`

**Interfaces:**
- Produces: `var ErrPushTokenUnregistered = errors.New("push token unregistered")` in `service` package; returned by `SendPush` impls, checked by the service to prune via existing `DeviceTokenRepository.DeleteByToken`.

- [ ] **Step 1: Write the failing test** — a push provider that returns `ErrPushTokenUnregistered` causes `DeleteByToken`.

```go
type unregPush struct{}
func (unregPush) SendPush(_ context.Context, _ string, _ entity.Platform, _, _ string, _ map[string]string) error {
	return service.ErrPushTokenUnregistered
}

func TestSendCriticalAlert_PrunesDeadToken(t *testing.T) {
	repo := newFakeDeviceRepo(testUserID, "dead-token", entity.PlatformIOS)
	svc := newTestNotificationServiceWithRepo(t, unregPush{}, repo)
	_ = svc.SendCriticalAlert(context.Background(), testUserID, uuid.New(), uuid.New(), "Acme", "acme.myshopify.com", valueobject.RiskStateSafe, valueobject.RiskStateOneCycleMissed)
	if !repo.deletedByToken["dead-token"] {
		t.Fatal("expected dead token to be pruned via DeleteByToken")
	}
}
```

- [ ] **Step 2: Run it — expect FAIL** (`ErrPushTokenUnregistered` undefined / not pruned).
Run: `cd backend && go test ./internal/application/service/ -run TestSendCriticalAlert_PrunesDeadToken`

- [ ] **Step 3: Implement the sentinel + prune.**

In `notification_service.go`:

```go
// ErrPushTokenUnregistered signals the device token is no longer valid (uninstalled
// app / expired) and should be pruned. Push providers return it; the service deletes.
var ErrPushTokenUnregistered = errors.New("push token unregistered")
```

In both push loops, replace the error handling with:

```go
for _, token := range tokens {
	if err := s.pushProvider.SendPush(ctx, token.DeviceToken, token.Platform, title, body, data); err != nil {
		if errors.Is(err, ErrPushTokenUnregistered) {
			_ = s.deviceTokenRepo.DeleteByToken(ctx, token.DeviceToken) // prune; not a send failure
			continue
		}
		lastErr = err
	}
}
```

- [ ] **Step 4: Make the FCM impl return the sentinel.** In `firebase_messaging.go` `SendPush`:

```go
_, err := s.client.Send(ctx, buildMessage(deviceToken, platform, title, body, data))
if err != nil {
	if messaging.IsUnregistered(err) || messaging.IsRegistrationTokenNotRegistered(err) {
		return service.ErrPushTokenUnregistered
	}
	return fmt.Errorf("failed to send push notification: %w", err)
}
return nil
```

> If importing `service` into `external` creates an import cycle, instead define the sentinel in a small leaf package (e.g. `internal/domain/service` already imported by both) or return a typed `external.ErrTokenUnregistered` and map it in the service. Check `go build ./...`; pick whichever compiles without a cycle and note it in the commit.

- [ ] **Step 5: Run tests — expect PASS.** `go test ./internal/application/service/... ./internal/infrastructure/external/...`

- [ ] **Step 6: Commit**

```bash
git add backend/internal/application/service/ backend/internal/infrastructure/external/
git commit -m "feat(push): prune dead device tokens on FCM unregistered errors"
```

---

## Part B — Flutter push client (`frontend-flutter`)

### Task B1: `PushNotificationService` — register + refresh

**Files:**
- Modify: `frontend-flutter/pubspec.yaml` (add `firebase_messaging: ^15.1.0`)
- Create: `frontend-flutter/lib/services/push_notification_service.dart`
- Test: `frontend-flutter/test/services/push_notification_service_test.dart`

**Interfaces:**
- Produces: `class PushNotificationService` with
  `Future<void> registerToken()`, `Future<void> unregister()`, and a constructor taking injected callables so it's unit-testable without the FCM plugin:
  `PushNotificationService({required ApiClient api, required Future<String?> Function() getToken, required Future<bool> Function() requestPermission, Stream<String> Function()? onTokenRefresh, String platform})`.

- [ ] **Step 1: Add the dependency**

Edit `pubspec.yaml` under `dependencies:` (next to `firebase_core`): `firebase_messaging: ^15.1.0`. Run `cd frontend-flutter && flutter pub get`.

- [ ] **Step 2: Write the failing test** (register posts the token)

```dart
import 'package:flutter_test/flutter_test.dart';
// import your ApiClient + a fake; PushNotificationService under test

void main() {
  test('registerToken posts device_token + platform to /api/v1/devices', () async {
    final fakeApi = FakeApiClient();
    final svc = PushNotificationService(
      api: fakeApi,
      requestPermission: () async => true,
      getToken: () async => 'fcm-tok-1',
      platform: 'android',
    );
    await svc.registerToken();
    expect(fakeApi.lastPath, '/api/v1/devices');
    expect(fakeApi.lastBody, {'device_token': 'fcm-tok-1', 'platform': 'android'});
  });

  test('permission denied -> no registration', () async {
    final fakeApi = FakeApiClient();
    final svc = PushNotificationService(
      api: fakeApi, requestPermission: () async => false,
      getToken: () async => 'fcm-tok-1', platform: 'ios');
    await svc.registerToken();
    expect(fakeApi.calls, 0); // Review Focus #5-adjacent
  });
}
```

> Add a minimal `FakeApiClient` in the test recording `lastPath`/`lastBody`/`calls`, matching your `ApiClient.post` shape (`post<T>(path, {data})`).

- [ ] **Step 3: Run it — expect FAIL.** `cd frontend-flutter && flutter test test/services/push_notification_service_test.dart`

- [ ] **Step 4: Implement `PushNotificationService`**

```dart
import '../core/network/api_client.dart';

class PushNotificationService {
  PushNotificationService({
    required this.api,
    required this.getToken,
    required this.requestPermission,
    this.onTokenRefresh,
    required this.platform,
  });

  final ApiClient api;
  final Future<String?> Function() getToken;
  final Future<bool> Function() requestPermission;
  final Stream<String> Function()? onTokenRefresh;
  final String platform;
  String? _token;

  Future<void> registerToken() async {
    if (!await requestPermission()) return;         // Review Focus #5
    final token = await getToken();
    if (token == null || token.isEmpty) return;
    _token = token;
    try {
      await api.post('/api/v1/devices', data: {'device_token': token, 'platform': platform});
    } catch (_) {
      // offline / backend down: leave unregistered; retried on next launch/refresh
    }
    onTokenRefresh?.call().listen((t) {
      _token = t;
      api.post('/api/v1/devices', data: {'device_token': t, 'platform': platform});
    });
  }

  Future<void> unregister() async {
    final t = _token;
    if (t == null) return;
    try { await api.delete('/api/v1/devices', data: {'device_token': t}); } catch (_) {}
    _token = null;
  }
}
```

- [ ] **Step 5: Run tests — expect PASS.** `flutter test test/services/push_notification_service_test.dart`

- [ ] **Step 6: Commit**

```bash
git add frontend-flutter/pubspec.yaml frontend-flutter/pubspec.lock frontend-flutter/lib/services/push_notification_service.dart frontend-flutter/test/services/push_notification_service_test.dart
git commit -m "feat(push): PushNotificationService registers/refreshes FCM token"
```

### Task B2: wire into app lifecycle (init after login, unregister on logout)

**Files:**
- Modify: `frontend-flutter/lib/main.dart` (provide `PushNotificationService`, bind FCM plugin callables)
- Modify: `frontend-flutter/lib/providers/auth_provider.dart` (`signOut` unregisters first; call `registerToken()` after successful sign-in)
- Test: `frontend-flutter/test/providers/auth_provider_logout_test.dart`

**Interfaces:**
- Consumes: `PushNotificationService` from B1.

- [ ] **Step 1: Write the failing test** — signOut unregisters before Firebase sign-out (Review Focus #4).

```dart
test('signOut unregisters the push token', () async {
  final push = FakePushService();
  final auth = AuthProvider(pushService: push);      // inject
  await auth.signOut();
  expect(push.unregistered, isTrue);
});
```

- [ ] **Step 2: Run it — expect FAIL** (`pushService` param missing).
Run: `flutter test test/providers/auth_provider_logout_test.dart`

- [ ] **Step 3: Implement** — inject an optional `PushNotificationService` into `AuthProvider`; in `signOut()`:

```dart
Future<void> signOut() async {
  await _pushService?.unregister();   // before Firebase sign-out (Review Focus #4)
  _mixpanel?.trackLogout();
  _mixpanel?.reset();
  await FirebaseAuth.instance.signOut();
  _error = null;
}
```

In `_onAuthStateChanged` (or after a successful sign-in), call `_pushService?.registerToken()` (fire-and-forget) once authenticated.

- [ ] **Step 4: Wire real FCM in `main.dart`** — construct with plugin-backed callables:

```dart
import 'package:firebase_messaging/firebase_messaging.dart';
// after Firebase.initializeApp(...):
final push = PushNotificationService(
  api: apiClient,
  platform: kIsWeb ? 'web' : (Platform.isIOS ? 'ios' : 'android'),
  requestPermission: () async {
    final s = await FirebaseMessaging.instance.requestPermission();
    return s.authorizationStatus == AuthorizationStatus.authorized ||
           s.authorizationStatus == AuthorizationStatus.provisional;
  },
  getToken: () => FirebaseMessaging.instance.getToken(),
  onTokenRefresh: () => FirebaseMessaging.instance.onTokenRefresh,
);
```

Provide `push` to the widget tree (add to the existing `MultiProvider`) and hand it to `AuthProvider`.

- [ ] **Step 5: Run tests + analyze — expect PASS/clean.** `flutter test && flutter analyze`

- [ ] **Step 6: Commit**

```bash
git add frontend-flutter/lib/main.dart frontend-flutter/lib/providers/auth_provider.dart frontend-flutter/test/providers/auth_provider_logout_test.dart
git commit -m "feat(push): register on login, unregister on logout"
```

### Task B3: deep-link notification taps via go_router

**Files:**
- Create: `frontend-flutter/lib/services/push_deep_link.dart` (pure `routeForData` mapper)
- Modify: `frontend-flutter/lib/main.dart` (wire `onMessageOpenedApp` + `getInitialMessage` → `router.go(routeForData(data))`)
- Test: `frontend-flutter/test/services/push_deep_link_test.dart`

**Interfaces:**
- Consumes: the `data` map shape from Task A2 (`type`, `app_id`, `subscription_id`).
- Produces: `String routeForData(Map<String, dynamic> data)` — returns a go_router path, defaulting to `/`.

- [ ] **Step 1: Write the failing test** (Review Focus #2 + #3)

```dart
void main() {
  test('risk_alert with subscription_id -> subscription detail route', () {
    expect(routeForData({'type': 'risk_alert', 'subscription_id': 'sub-9'}), '/subscriptions/sub-9');
  });
  test('daily_summary -> dashboard', () {
    expect(routeForData({'type': 'daily_summary', 'app_id': 'a1'}), '/');
  });
  test('missing/unknown data -> dashboard fallback (no crash)', () {
    expect(routeForData({}), '/');
    expect(routeForData({'type': 'nonsense'}), '/');
    expect(routeForData({'type': 'risk_alert'}), '/'); // no subscription_id
  });
}
```

- [ ] **Step 2: Run it — expect FAIL** (`routeForData` undefined).
Run: `flutter test test/services/push_deep_link_test.dart`

- [ ] **Step 3: Implement `routeForData`**

```dart
String routeForData(Map<String, dynamic> data) {
  final type = data['type'];
  if (type == 'risk_alert') {
    final id = data['subscription_id'];
    if (id is String && id.isNotEmpty) return '/subscriptions/$id';
  }
  return '/'; // daily_summary, unknown, or missing ids -> Dashboard (Review Focus #2/#3)
}
```

- [ ] **Step 4: Run test — expect PASS.** `flutter test test/services/push_deep_link_test.dart`

- [ ] **Step 5: Wire the tap handlers in `main.dart`** (after the router is built):

```dart
FirebaseMessaging.instance.getInitialMessage().then((m) {
  if (m != null) router.go(routeForData(m.data));
});
FirebaseMessaging.onMessageOpenedApp.listen((m) => router.go(routeForData(m.data)));
```

> Cross-org safety (Review Focus #3): the detail screen already loads via the app's own org-scoped providers; an `app_id`/`subscription_id` the user can't access resolves to an empty/uninstalled state, not another org's data. Note this in the commit; no extra client check needed beyond the Dashboard fallback for malformed data.

- [ ] **Step 6: Run analyze — expect clean.** `flutter analyze`

- [ ] **Step 7: Commit**

```bash
git add frontend-flutter/lib/services/push_deep_link.dart frontend-flutter/lib/main.dart frontend-flutter/test/services/push_deep_link_test.dart
git commit -m "feat(push): deep-link notification taps via go_router with dashboard fallback"
```

---

## Part C — Store readiness (config; non-TDD checklist)

> **Adaptation (stated per skill rule):** these tasks are configuration/signing, not
> unit-testable code, so they use a verification step instead of a failing-test step.
> They are gated by human/owner confirmation (bundle IDs are irreversible once published).

### Task C1: real bundle identifiers

**Files:** `frontend-flutter/android/app/build.gradle.kts:9,24` (`namespace`, `applicationId`), `frontend-flutter/ios/Runner.xcodeproj/project.pbxproj` (`PRODUCT_BUNDLE_IDENTIFIER`, all configs).

- [ ] **Step 1:** Confirm the final ID with the owner (e.g. `com.ledgerspear.app`) — **irreversible after publish**.
- [ ] **Step 2:** Replace `com.example.ledgerguard_flutter` (Android `namespace` + `applicationId`) and `com.example.ledgerguardFlutter` (iOS bundle id, Debug/Release/Profile + drop the `.RunnerTests` suffix pattern accordingly).
- [ ] **Step 3:** Regenerate `firebase_options.dart` / re-download `google-services.json` + `GoogleService-Info.plist` for the new IDs (FCM is keyed on the bundle id) via `flutterfire configure`.
- [ ] **Step 4: Verify:** `flutter build apk --release` and `flutter build ios --no-codesign` succeed with the new IDs.
- [ ] **Step 5: Commit** `chore(mobile): set production bundle identifiers`.

### Task C2: app signing

- [ ] **Step 1:** Android: create an upload keystore; add `key.properties` (git-ignored) + `signingConfigs` in `build.gradle.kts`; ensure `release` uses it.
- [ ] **Step 2:** iOS: set the Apple Team + provisioning in Xcode (or `ExportOptions.plist` for CI).
- [ ] **Step 3: Verify:** a signed `flutter build appbundle --release` produces a signed `.aab`.
- [ ] **Step 4: Commit** `chore(mobile): release signing config (keystore git-ignored)`.

### Task C3: iOS APNs for FCM

- [ ] **Step 1:** Create an APNs Auth Key (.p8) in the Apple Developer portal; upload it to the Firebase iOS app (Cloud Messaging settings).
- [ ] **Step 2:** Add the **Push Notifications** capability + `aps-environment` entitlement in Xcode; enable Background Modes → Remote notifications.
- [ ] **Step 3: Verify (real device):** a test push from Firebase console arrives on a physical iOS device (APNs can't be emulated).
- [ ] **Step 4: Commit** `chore(ios): push capability + APNs entitlement`.

### Task C4: icons, splash, privacy manifest

- [ ] **Step 1:** Generate launcher icons + splash (e.g. `flutter_launcher_icons`, `flutter_native_splash`) from the LedgerGuard mark; run their generators.
- [ ] **Step 2:** Add the iOS `PrivacyInfo.xcprivacy` manifest declaring data use (device token, Mixpanel analytics) + any required-reason APIs; add Play Data-safety notes for the listing.
- [ ] **Step 3: Verify:** `flutter build ios --no-codesign` and `flutter build apk --release` succeed; manifest present in the built app.
- [ ] **Step 4: Commit** `chore(mobile): launcher icons, splash, iOS privacy manifest`.

---

## Notes for the executor

- After Part A: `cd backend && go build ./... && go test ./... && golangci-lint run` (CI gates on all three).
- After Part B: `cd frontend-flutter && flutter analyze && flutter test`.
- Part C's real-device push test (C3 Step 3) is the one thing CI cannot cover — do it manually before calling push "shipped."
- Keep the backend `data` change (Part A) mergeable on its own: old clients ignore `data`, so A can ship before B.
