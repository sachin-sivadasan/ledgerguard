import 'package:firebase_auth/firebase_auth.dart';
import 'package:firebase_core/firebase_core.dart';
import 'package:google_sign_in/google_sign_in.dart';

import '../../domain/entities/user_entity.dart';
import '../../domain/repositories/auth_repository.dart';

/// Firebase implementation of AuthRepository
class FirebaseAuthRepository implements AuthRepository {
  FirebaseAuth? _firebaseAuth;
  GoogleSignIn? _googleSignIn;

  FirebaseAuthRepository({
    FirebaseAuth? firebaseAuth,
    GoogleSignIn? googleSignIn,
  })  : _firebaseAuth = firebaseAuth,
        _googleSignIn = googleSignIn;

  /// Check if Firebase is initialized and available
  bool get _isFirebaseAvailable => Firebase.apps.isNotEmpty;

  /// Get FirebaseAuth instance, returns null if Firebase not initialized
  FirebaseAuth? get _auth {
    if (!_isFirebaseAvailable) return null;
    _firebaseAuth ??= FirebaseAuth.instance;
    return _firebaseAuth;
  }

  /// Get GoogleSignIn instance
  GoogleSignIn get _google {
    _googleSignIn ??= GoogleSignIn();
    return _googleSignIn!;
  }

  @override
  Stream<UserEntity?> get authStateChanges {
    final auth = _auth;
    if (auth == null) {
      // Firebase not initialized - return stream with null (unauthenticated)
      return Stream.value(null);
    }
    try {
      return auth.authStateChanges().map(_mapFirebaseUser);
    } catch (e) {
      // Return empty stream if Firebase has issues
      return Stream.value(null);
    }
  }

  @override
  UserEntity? get currentUser {
    final auth = _auth;
    if (auth == null) return null;
    try {
      return _mapFirebaseUser(auth.currentUser);
    } catch (e) {
      return null;
    }
  }

  @override
  Future<UserEntity> signInWithEmailAndPassword({
    required String email,
    required String password,
  }) async {
    final auth = _auth;
    if (auth == null) {
      throw const AuthException(
        'Firebase not configured. Run: flutterfire configure',
        code: 'firebase-not-configured',
      );
    }
    try {
      final credential = await auth.signInWithEmailAndPassword(
        email: email,
        password: password,
      );
      final user = _mapFirebaseUser(credential.user);
      if (user == null) {
        throw const AuthException('Failed to sign in');
      }
      return user;
    } on FirebaseAuthException catch (e) {
      throw _mapFirebaseException(e);
    } on FirebaseException catch (e) {
      throw _mapFirebaseException(e);
    } catch (e) {
      // Handle JavaScript interop errors on web
      throw AuthException('Authentication failed: ${e.runtimeType}');
    }
  }

  @override
  Future<UserEntity> signInWithGoogle() async {
    final auth = _auth;
    if (auth == null) {
      throw const AuthException(
        'Firebase not configured. Run: flutterfire configure',
        code: 'firebase-not-configured',
      );
    }
    try {
      final googleUser = await _google.signIn();
      if (googleUser == null) {
        throw const GoogleSignInCancelledException();
      }

      final googleAuth = await googleUser.authentication;
      final credential = GoogleAuthProvider.credential(
        accessToken: googleAuth.accessToken,
        idToken: googleAuth.idToken,
      );

      final userCredential = await auth.signInWithCredential(credential);
      final user = _mapFirebaseUser(userCredential.user);
      if (user == null) {
        throw const AuthException('Failed to sign in with Google');
      }
      return user;
    } on FirebaseAuthException catch (e) {
      throw _mapFirebaseException(e);
    } on FirebaseException catch (e) {
      throw _mapFirebaseException(e);
    } catch (e) {
      // Handle JavaScript interop errors on web
      throw AuthException('Google sign-in failed: ${e.runtimeType}');
    }
  }

  @override
  Future<void> signOut() async {
    final auth = _auth;
    if (auth == null) return; // Nothing to sign out from
    try {
      await Future.wait([
        auth.signOut(),
        _google.signOut(),
      ]);
    } on FirebaseAuthException catch (e) {
      throw _mapFirebaseException(e);
    } on FirebaseException catch (e) {
      throw _mapFirebaseException(e);
    } catch (e) {
      // Handle JavaScript interop errors on web
      throw AuthException('Sign out failed: ${e.runtimeType}');
    }
  }

  @override
  Future<String?> getIdToken() async {
    final auth = _auth;
    if (auth == null) return null;
    try {
      final user = auth.currentUser;
      if (user == null) return null;
      return user.getIdToken();
    } catch (e) {
      return null;
    }
  }

  UserEntity? _mapFirebaseUser(User? firebaseUser) {
    if (firebaseUser == null) return null;

    return UserEntity(
      id: firebaseUser.uid,
      email: firebaseUser.email ?? '',
      displayName: firebaseUser.displayName,
      photoUrl: firebaseUser.photoURL,
    );
  }

  AuthException _mapFirebaseException(FirebaseException e) {
    final code = e.code;
    final message = e.message;

    switch (code) {
      case 'user-not-found':
        return const UserNotFoundException();
      case 'wrong-password':
      case 'invalid-credential':
        return const InvalidCredentialsException();
      case 'weak-password':
        return const WeakPasswordException();
      case 'email-already-in-use':
        return const EmailAlreadyInUseException();
      default:
        return AuthException(message ?? 'Authentication failed', code: code);
    }
  }
}
