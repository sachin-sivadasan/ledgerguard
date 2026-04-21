import 'package:flutter/foundation.dart';

class AuthUser {
  final String email;
  final String name;

  AuthUser({required this.email, required this.name});
}

class AuthProvider extends ChangeNotifier {
  AuthUser? _user;
  bool _isLoading = false;
  String? _error;

  AuthUser? get user => _user;
  bool get isAuthenticated => _user != null;
  bool get isLoading => _isLoading;
  String? get error => _error;

  static final _emailRegex = RegExp(r'^[^@\s]+@[^@\s]+\.[^@\s]+$');

  Future<void> signIn(String email, String password) async {
    _error = null;
    _isLoading = true;
    notifyListeners();

    await Future.delayed(const Duration(milliseconds: 500));

    if (!_emailRegex.hasMatch(email)) {
      _error = 'Please enter a valid email address';
      _isLoading = false;
      notifyListeners();
      return;
    }

    if (password.length < 6) {
      _error = 'Password must be at least 6 characters';
      _isLoading = false;
      notifyListeners();
      return;
    }

    _user = AuthUser(email: email, name: email.split('@').first);
    _isLoading = false;
    notifyListeners();
  }

  Future<void> signUp(String name, String email, String password) async {
    _error = null;
    _isLoading = true;
    notifyListeners();

    await Future.delayed(const Duration(milliseconds: 500));

    if (name.trim().isEmpty) {
      _error = 'Please enter your name';
      _isLoading = false;
      notifyListeners();
      return;
    }

    if (!_emailRegex.hasMatch(email)) {
      _error = 'Please enter a valid email address';
      _isLoading = false;
      notifyListeners();
      return;
    }

    if (password.length < 6) {
      _error = 'Password must be at least 6 characters';
      _isLoading = false;
      notifyListeners();
      return;
    }

    _user = AuthUser(email: email, name: name.trim());
    _isLoading = false;
    notifyListeners();
  }

  Future<bool> resetPassword(String email) async {
    _error = null;
    _isLoading = true;
    notifyListeners();

    await Future.delayed(const Duration(milliseconds: 500));

    if (!_emailRegex.hasMatch(email)) {
      _error = 'Please enter a valid email address';
      _isLoading = false;
      notifyListeners();
      return false;
    }

    _isLoading = false;
    notifyListeners();
    return true;
  }

  void signOut() {
    _user = null;
    _error = null;
    notifyListeners();
  }

  void clearError() {
    _error = null;
    notifyListeners();
  }
}
