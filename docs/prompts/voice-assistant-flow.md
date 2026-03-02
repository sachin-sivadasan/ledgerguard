# Voice AI Assistant Flow - Interactive Visualization

## Context
You are a senior frontend + visualization engineer building an interactive animated guide showing how LedgerGuard's voice AI assistant will work. This is a future feature specification that will be updated until implementation starts.

Build an educational visualization that helps developers understand:
1. How voice input is captured and processed
2. How AI classifies intent and extracts entities
3. How the app navigates to the correct screen
4. Supported voice commands and their outcomes

---

## Design Philosophy

### Target Audience
LedgerGuard mobile app users who:
- Want hands-free access to subscription data
- Need quick answers about store health and risk
- Prefer voice interaction over manual navigation
- Are busy app developers checking metrics on the go

### Key Principles
1. **Natural language** - Users speak naturally, AI understands context
2. **Quick navigation** - Voice → Screen in under 2 seconds
3. **Smart entity extraction** - Recognize store names, filters, metrics
4. **Graceful fallback** - Show suggestions if intent unclear

---

## Flow Stages

### Stage 1: Voice Capture
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           VOICE CAPTURE                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  🎤 Microphone    →    📊 Waveform    →    📝 Transcript                    │
│  (Listening)          (Processing)        (Recognized)                      │
│                                                                              │
│  Flutter: speech_to_text package                                            │
│  Permissions: Microphone access                                             │
│  Output: Raw text string                                                    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Stage 2: Intent Classification
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        INTENT CLASSIFICATION                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  📝 Transcript    →    🤖 AI Model    →    🎯 Intent                        │
│  "show store                              STORE_DETAILS                     │
│   acme health"                            confidence: 0.95                  │
│                                                                              │
│  Model Options:                                                             │
│  - Claude API (cloud, high accuracy)                                        │
│  - Local LLM (offline, faster)                                              │
│  - Rule-based fallback (keywords)                                           │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Stage 3: Entity Extraction
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         ENTITY EXTRACTION                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Input: "show store acme health"                                            │
│                                                                              │
│  Extracted Entities:                                                        │
│  ┌──────────────┬─────────────┬─────────────────────────────────┐          │
│  │ Entity       │ Value       │ Matched From                    │          │
│  ├──────────────┼─────────────┼─────────────────────────────────┤          │
│  │ store_name   │ "acme"      │ Fuzzy match against shop list   │          │
│  │ view_type    │ "health"    │ Keyword detection               │          │
│  └──────────────┴─────────────┴─────────────────────────────────┘          │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Stage 4: Navigation
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            NAVIGATION                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Intent: STORE_DETAILS                                                       │
│  Entities: { store_name: "acme", view_type: "health" }                      │
│                                                                              │
│  Route Builder:                                                             │
│  /subscriptions/sub_123/health                                              │
│                                                                              │
│  GoRouter.go(context, '/subscriptions/$subId/health')                       │
│                                                                              │
│  📱 Screen Transition → Subscription Health Page                            │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Supported Voice Commands

| Voice Command | Intent | Entities | Target Screen |
|---------------|--------|----------|---------------|
| "Show details of store [name]" | STORE_DETAILS | store_name | /subscriptions/:id |
| "Store [name] health" | STORE_HEALTH | store_name | /subscriptions/:id/health |
| "List subscriptions at risk" | LIST_FILTER | filter: at_risk | /subscriptions?tab=at_risk |
| "Show churned customers" | LIST_FILTER | filter: churned | /subscriptions?tab=churned |
| "What's my MRR?" | METRIC_QUERY | metric: mrr | /dashboard (highlight MRR) |
| "Show revenue trends" | METRIC_QUERY | metric: trends | /dashboard (trends section) |
| "Any billing failures?" | ALERT_QUERY | alert_type: billing | /alerts?type=billing |
| "Go to dashboard" | NAVIGATE | screen: dashboard | /dashboard |
| "Open settings" | NAVIGATE | screen: settings | /settings |

---

## Intent Classification Model

### Prompt Template (Claude API)
```
You are a voice command classifier for a subscription management app.

Given the user's voice input, classify the intent and extract entities.

Intents:
- STORE_DETAILS: User wants to see a specific store's subscription details
- STORE_HEALTH: User wants to see a store's health/risk score
- LIST_FILTER: User wants to see a filtered list of subscriptions
- METRIC_QUERY: User wants to know a specific metric value
- ALERT_QUERY: User wants to see alerts or notifications
- NAVIGATE: User wants to go to a specific screen

Voice input: "{transcript}"

Respond in JSON:
{
  "intent": "INTENT_NAME",
  "confidence": 0.0-1.0,
  "entities": {
    "store_name": "extracted store name or null",
    "filter": "at_risk|churned|safe|all or null",
    "metric": "mrr|arr|churn_rate|health or null",
    "screen": "dashboard|settings|subscriptions or null"
  }
}
```

### Fallback: Keyword Matching
```dart
Intent classifyByKeywords(String transcript) {
  final lower = transcript.toLowerCase();

  if (lower.contains('health') && lower.contains('store')) {
    return Intent.storeHealth;
  }
  if (lower.contains('at risk') || lower.contains('risk')) {
    return Intent.listFilter(filter: 'at_risk');
  }
  if (lower.contains('churned') || lower.contains('churn')) {
    return Intent.listFilter(filter: 'churned');
  }
  if (lower.contains('mrr') || lower.contains('revenue')) {
    return Intent.metricQuery(metric: 'mrr');
  }
  if (lower.contains('dashboard')) {
    return Intent.navigate(screen: 'dashboard');
  }

  return Intent.unknown;
}
```

---

## Fallback: Show Suggestions

When the AI cannot determine the user's intent (confidence < 0.7 or Intent.unknown), display a text response with relevant suggestions.

### Trigger Conditions
- Intent classification confidence below threshold (0.7)
- Keyword matching returns `Intent.unknown`
- No entities could be extracted

### Suggestion Display
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                              │
│  🎤 "I didn't understand that"                                              │
│                                                                              │
│  Try saying:                                                                │
│  • "Show store [name]" - View subscription details                         │
│  • "List at-risk subscriptions" - See subscriptions needing attention      │
│  • "What's my MRR?" - Check your monthly recurring revenue                 │
│                                                                              │
│                         [🎤 Try Again]                                      │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Flutter Implementation
```dart
class SuggestionSnackBar {
  static void show(BuildContext context, VoidCallback onTryAgain) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              "I didn't understand that",
              style: TextStyle(fontWeight: FontWeight.bold),
            ),
            SizedBox(height: 8),
            Text("Try saying:"),
            Text('• "Show store [name]"'),
            Text('• "List at-risk subscriptions"'),
            Text('• "What\'s my MRR?"'),
          ],
        ),
        action: SnackBarAction(
          label: "Try Again",
          onPressed: onTryAgain,
        ),
        duration: Duration(seconds: 6),
      ),
    );
  }
}
```

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         VOICE ASSISTANT ARCHITECTURE                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐                   │
│  │   Flutter    │    │   Voice      │    │   Intent     │                   │
│  │   App UI     │───▶│   Service    │───▶│   Classifier │                   │
│  │              │    │              │    │              │                   │
│  └──────────────┘    └──────────────┘    └──────────────┘                   │
│         │                   │                   │                           │
│         │                   ▼                   ▼                           │
│         │            ┌──────────────┐    ┌──────────────┐                   │
│         │            │ speech_to_   │    │  Claude API  │                   │
│         │            │ text Plugin  │    │  (optional)  │                   │
│         │            └──────────────┘    └──────────────┘                   │
│         │                                       │                           │
│         │                   ┌───────────────────┘                           │
│         │                   ▼                                               │
│         │            ┌──────────────┐                                       │
│         │            │   Entity     │                                       │
│         │            │   Resolver   │                                       │
│         │            │ (fuzzy match)│                                       │
│         │            └──────────────┘                                       │
│         │                   │                                               │
│         │                   ▼                                               │
│         │            ┌──────────────┐                                       │
│         └───────────▶│  Navigation  │                                       │
│                      │  Dispatcher  │                                       │
│                      │  (GoRouter)  │                                       │
│                      └──────────────┘                                       │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Flutter Implementation Outline

### VoiceService
```dart
class VoiceService {
  final SpeechToText _speech = SpeechToText();
  final IntentClassifier _classifier;
  final EntityResolver _resolver;
  final GoRouter _router;

  Future<void> startListening() async {
    await _speech.listen(onResult: _onSpeechResult);
  }

  void _onSpeechResult(SpeechRecognitionResult result) async {
    if (result.finalResult) {
      final transcript = result.recognizedWords;
      final intent = await _classifier.classify(transcript);
      final entities = await _resolver.resolve(intent.entities);
      final route = _buildRoute(intent, entities);
      _router.go(route);
    }
  }

  String _buildRoute(Intent intent, Map<String, dynamic> entities) {
    switch (intent.type) {
      case IntentType.storeDetails:
        return '/subscriptions/${entities['subscription_id']}';
      case IntentType.storeHealth:
        return '/subscriptions/${entities['subscription_id']}/health';
      case IntentType.listFilter:
        return '/subscriptions?tab=${entities['filter']}';
      case IntentType.metricQuery:
        return '/dashboard?highlight=${entities['metric']}';
      default:
        return '/dashboard';
    }
  }
}
```

### VoiceButton Widget
```dart
class VoiceButton extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return BlocBuilder<VoiceBloc, VoiceState>(
      builder: (context, state) {
        return FloatingActionButton(
          onPressed: () => context.read<VoiceBloc>().add(StartListening()),
          child: AnimatedSwitcher(
            child: state is VoiceListening
              ? PulsingMicIcon()
              : Icon(Icons.mic),
          ),
        );
      },
    );
  }
}
```

---

## Technical Requirements

### Framework
- Next.js 14+ with App Router
- TailwindCSS for styling
- React hooks for state and animation

### Animation Approach
- Simulated voice waveform using SVG
- Step-by-step flow progression
- Typing effect for transcript display
- Screen mockup transitions

### Visual Style
- Dark theme (slate-950 background)
- Purple to cyan gradient accents (voice/AI theme)
- Glowing microphone icon when "listening"
- Card-based intent/entity display

### Interactions
- "Speak" button to trigger demo flow
- Example command selector
- Play/pause animation
- Show/hide technical details

---

## Component Structure

```
marketing/site/
├── app/voice-assistant/page.tsx    # Page wrapper
└── components/
    └── VoiceAssistantVisualization.tsx
        ├── VoiceWaveform           # Animated waveform SVG
        ├── TranscriptDisplay       # Typing effect transcript
        ├── IntentCard              # Shows classified intent
        ├── EntityTable             # Extracted entities
        ├── NavigationPreview       # Target screen mockup
        └── CommandExamples         # Supported commands list
```

---

## Demo Flow

1. **Idle State**: Show microphone button with "Tap to speak" hint
2. **Listening**: Animated waveform, pulsing mic icon
3. **Processing**: Show transcript appearing with typing effect
4. **Classification**: Intent card slides in with confidence score
5. **Extraction**: Entity table populates with matched values
6. **Navigation**: Screen mockup animates to show target page
7. **Complete**: Reset button to try another command

---

## Future Enhancements (Post-MVP)

- [ ] Wake word detection ("Hey LedgerGuard")
- [ ] Multi-language support
- [ ] Voice response/feedback (TTS)
- [ ] Conversation context (follow-up questions)
- [ ] Custom command training
- [ ] Offline intent classification
