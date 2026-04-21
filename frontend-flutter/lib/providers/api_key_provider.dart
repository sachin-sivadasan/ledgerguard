import 'package:flutter/foundation.dart';
import '../mock_data/mock_api_keys.dart';
import '../models/api_key_model.dart';

class ApiKeyProvider extends ChangeNotifier {
  final List<ApiKey> _keys = List.of(mockApiKeys);
  int _nextId = 5;

  List<ApiKey> get keys =>
      List.unmodifiable(_keys..sort((a, b) => b.createdAt.compareTo(a.createdAt)));

  List<ApiKey> get activeKeys =>
      keys.where((k) => k.status == ApiKeyStatus.active).toList();

  void createKey(String name, List<String> permissions) {
    final key = ApiKey(
      id: 'key-${_nextId++}',
      name: name,
      keyPrefix: 'lg_live_${DateTime.now().millisecondsSinceEpoch.toRadixString(16).substring(0, 4)}...',
      createdAt: DateTime.now(),
      status: ApiKeyStatus.active,
      permissions: permissions,
    );
    _keys.add(key);
    notifyListeners();
  }

  void revokeKey(String id) {
    final idx = _keys.indexWhere((k) => k.id == id);
    if (idx != -1) {
      _keys[idx] = _keys[idx].copyWith(status: ApiKeyStatus.revoked);
      notifyListeners();
    }
  }

  String copyKey(String id) {
    final key = _keys.firstWhere((k) => k.id == id);
    return 'lg_live_${key.id}_${DateTime.now().millisecondsSinceEpoch}';
  }
}
