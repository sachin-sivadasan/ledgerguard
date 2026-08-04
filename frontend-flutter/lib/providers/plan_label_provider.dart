import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../services/plan_label_service.dart';

/// Manages the plan-label editor: loads the detected price tiers, holds the developer's
/// in-progress edits, and saves the set.
class PlanLabelProvider extends ChangeNotifier {
  final PlanLabelService _service;

  PlanLabelProvider(this._service);

  bool _isLoading = false;
  bool _isSaving = false;
  String? _error;
  bool _isServiceUnavailable = false;
  String? _selectedAppId;
  List<PlanTier> _tiers = const [];
  final Map<String, String> _edits = {}; // tier key → edited label
  bool _saved = false;

  bool get isLoading => _isLoading;
  bool get isSaving => _isSaving;
  String? get error => _error;
  bool get isServiceUnavailable => _isServiceUnavailable;
  String? get selectedAppId => _selectedAppId;
  List<PlanTier> get tiers => _tiers;
  bool get justSaved => _saved;

  /// The current label for a tier: the developer's unsaved edit if present, else the saved
  /// value.
  String labelFor(PlanTier t) => _edits[t.key] ?? t.label;

  bool get isDirty =>
      _edits.entries.any((e) => e.value != _savedLabelForKey(e.key));

  String _savedLabelForKey(String key) =>
      _tiers.firstWhere((t) => t.key == key,
          orElse: () => const PlanTier(
              billingInterval: '',
              priceCents: 0,
              key: '',
              pseudoLabel: '',
              label: '',
              customers: 0)).label;

  void setSelectedApp(String? appId) {
    if (appId == _selectedAppId) return;
    _selectedAppId = appId;
    _edits.clear();
    notifyListeners();
    if (appId != null) load(appId);
  }

  void editLabel(String tierKey, String label) {
    _edits[tierKey] = label;
    _saved = false;
    notifyListeners();
  }

  Future<void> load(String appId) async {
    _isLoading = true;
    _error = null;
    _isServiceUnavailable = false;
    _saved = false;
    notifyListeners();
    try {
      _tiers = await _service.fetchTiers(appId);
      _edits.clear();
    } on DioException catch (e) {
      if (e.response?.statusCode == 503) {
        _isServiceUnavailable = true;
        _error = 'Service temporarily unavailable.';
      } else {
        _error = e.message ?? e.toString();
      }
    } catch (e) {
      _error = e.toString();
    }
    _isLoading = false;
    notifyListeners();
  }

  Future<bool> save() async {
    final appId = _selectedAppId;
    if (appId == null) return false;
    _isSaving = true;
    _error = null;
    notifyListeners();
    // Apply the edits onto the tiers to send the full set.
    final payload = _tiers
        .map((t) => PlanTier(
              billingInterval: t.billingInterval,
              priceCents: t.priceCents,
              key: t.key,
              pseudoLabel: t.pseudoLabel,
              label: labelFor(t).trim(),
              customers: t.customers,
            ))
        .toList();
    var ok = false;
    try {
      await _service.saveTiers(appId, payload);
      _saved = true;
      ok = true;
      await load(appId); // refresh saved state
    } on DioException catch (e) {
      _error = e.response?.statusCode == 503
          ? 'Service temporarily unavailable.'
          : (e.message ?? e.toString());
    } catch (e) {
      _error = e.toString();
    }
    _isSaving = false;
    notifyListeners();
    return ok;
  }
}
