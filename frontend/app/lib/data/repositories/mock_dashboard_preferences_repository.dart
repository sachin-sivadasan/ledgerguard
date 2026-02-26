import '../../domain/entities/dashboard_preferences.dart';
import '../../domain/repositories/dashboard_preferences_repository.dart';

/// Mock implementation of DashboardPreferencesRepository for testing
class MockDashboardPreferencesRepository
    implements DashboardPreferencesRepository {
  DashboardPreferences _preferences = DashboardPreferences.defaults();

  /// Flag to simulate fetch error
  bool shouldFailFetch = false;

  /// Flag to simulate save error
  bool shouldFailSave = false;

  /// Simulated network delay in milliseconds
  int delayMs = 100;

  @override
  Future<DashboardPreferences> fetchPreferences() async {
    await Future.delayed(Duration(milliseconds: delayMs));

    if (shouldFailFetch) {
      throw const FetchPreferencesException();
    }

    return _preferences;
  }

  @override
  Future<void> savePreferences(DashboardPreferences preferences) async {
    await Future.delayed(Duration(milliseconds: delayMs));

    if (shouldFailSave) {
      throw const SavePreferencesException();
    }

    _preferences = preferences;
  }

  /// Reset to default preferences (for testing)
  void reset() {
    _preferences = DashboardPreferences.defaults();
    shouldFailFetch = false;
    shouldFailSave = false;
  }

  /// Get current preferences without async (for testing)
  DashboardPreferences get currentPreferences => _preferences;
}
