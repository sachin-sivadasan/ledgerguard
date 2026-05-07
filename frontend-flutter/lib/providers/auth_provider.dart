import 'dart:async';
import 'package:firebase_auth/firebase_auth.dart';
import 'package:flutter/foundation.dart';

import '../services/mixpanel_service.dart';

class AuthUser {
  final String uid;
  final String email;
  final String name;

  AuthUser({required this.uid, required this.email, required this.name});

  factory AuthUser.fromFirebase(User user) {
    return AuthUser(
      uid: user.uid,
      email: user.email ?? '',
      name: user.displayName ?? user.email?.split('@').first ?? '',
    );
  }
}

class AuthProvider extends ChangeNotifier {
  MixpanelService? _mixpanel;
  StreamSubscription<User?>? _authSubscription;

  AuthUser? _user;
  bool _isLoading = false;
  String? _error;

  AuthProvider() {
    _authSubscription =
        FirebaseAuth.instance.authStateChanges().listen(_onAuthStateChanged);
  }

  void setMixpanel(MixpanelService mixpanel) => _mixpanel = mixpanel;

  AuthUser? get user => _user;
  bool get isAuthenticated => _user != null;
  bool get isLoading => _isLoading;
  String? get error => _error;

  void _onAuthStateChanged(User? firebaseUser) {
    if (firebaseUser != null) {
      _user = AuthUser.fromFirebase(firebaseUser);
      _mixpanel?.identify(firebaseUser.uid, email: firebaseUser.email ?? '');
    } else {
      _user = null;
    }
    notifyListeners();
  }

  Future<void> signIn(String email, String password) async {
    _error = null;
    _isLoading = true;
    notifyListeners();

    try {
      await FirebaseAuth.instance.signInWithEmailAndPassword(
        email: email.trim(),
        password: password,
      );
      _mixpanel?.trackLogin('email');
    } on FirebaseAuthException catch (e) {
      _error = _mapAuthError(e.code);
    } catch (e) {
      _error = 'An unexpected error occurred. Please try again.';
    }

    _isLoading = false;
    notifyListeners();
  }

  Future<void> signUp(String name, String email, String password) async {
    _error = null;
    _isLoading = true;
    notifyListeners();

    try {
      final credential =
          await FirebaseAuth.instance.createUserWithEmailAndPassword(
        email: email.trim(),
        password: password,
      );
      await credential.user?.updateDisplayName(name.trim());
      // Reload to pick up display name
      await credential.user?.reload();
      _user = AuthUser.fromFirebase(FirebaseAuth.instance.currentUser!);
      _mixpanel?.trackSignup('email');
    } on FirebaseAuthException catch (e) {
      _error = _mapAuthError(e.code);
    } catch (e) {
      _error = 'An unexpected error occurred. Please try again.';
    }

    _isLoading = false;
    notifyListeners();
  }

  Future<bool> resetPassword(String email) async {
    _error = null;
    _isLoading = true;
    notifyListeners();

    try {
      await FirebaseAuth.instance.sendPasswordResetEmail(email: email.trim());
      _isLoading = false;
      notifyListeners();
      return true;
    } on FirebaseAuthException catch (e) {
      _error = _mapAuthError(e.code);
      _isLoading = false;
      notifyListeners();
      return false;
    } catch (e) {
      _error = 'An unexpected error occurred. Please try again.';
      _isLoading = false;
      notifyListeners();
      return false;
    }
  }

  Future<void> signOut() async {
    _mixpanel?.trackLogout();
    _mixpanel?.reset();
    await FirebaseAuth.instance.signOut();
    _error = null;
  }

  void clearError() {
    _error = null;
    notifyListeners();
  }

  String _mapAuthError(String code) {
    switch (code) {
      case 'invalid-email':
        return 'Please enter a valid email address.';
      case 'user-disabled':
        return 'This account has been disabled.';
      case 'user-not-found':
        return 'No account found with this email.';
      case 'wrong-password':
      case 'invalid-credential':
        return 'Incorrect email or password.';
      case 'email-already-in-use':
        return 'An account already exists with this email.';
      case 'weak-password':
        return 'Password must be at least 6 characters.';
      case 'too-many-requests':
        return 'Too many attempts. Please try again later.';
      case 'network-request-failed':
        return 'Network error. Please check your connection.';
      default:
        return 'Authentication failed. Please try again.';
    }
  }

  @override
  void dispose() {
    _authSubscription?.cancel();
    super.dispose();
  }
}
