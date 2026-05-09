import 'package:flutter/foundation.dart';

/// Notifies listeners when the user navigates between tabs,
/// so screens can check if their data is stale and reload.
class NavigationRefreshNotifier extends ChangeNotifier {
  void triggerRefreshCheck() {
    notifyListeners();
  }
}
