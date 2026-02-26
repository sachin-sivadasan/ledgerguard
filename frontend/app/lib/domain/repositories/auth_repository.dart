import '../entities/user_entity.dart';

/// Authentication repository interface
/// Defines contract for authentication operations
abstract class AuthRepository {
  /// Stream of authentication state changes
  Stream<UserEntity?> get authStateChanges;

  /// Get currently authenticated user
  UserEntity? get currentUser;

  /// Sign in with email and password
  Future<UserEntity> signInWithEmailAndPassword({
    required String email,
    required String password,
  });

  /// Sign in with Google
  Future<UserEntity> signInWithGoogle();

  /// Sign out
  Future<void> signOut();

  /// Get Firebase ID token for API calls
  Future<String?> getIdToken();
}

/// Authentication exceptions
class AuthException implements Exception {
  final String message;
  final String? code;

  const AuthException(this.message, {this.code});

  @override
  String toString() => 'AuthException: $message';
}

class InvalidCredentialsException extends AuthException {
  const InvalidCredentialsException([String message = 'Invalid email or password'])
      : super(message, code: 'invalid-credentials');
}

class UserNotFoundException extends AuthException {
  const UserNotFoundException([String message = 'User not found'])
      : super(message, code: 'user-not-found');
}

class WeakPasswordException extends AuthException {
  const WeakPasswordException([String message = 'Password is too weak'])
      : super(message, code: 'weak-password');
}

class EmailAlreadyInUseException extends AuthException {
  const EmailAlreadyInUseException([String message = 'Email already in use'])
      : super(message, code: 'email-already-in-use');
}

class GoogleSignInCancelledException extends AuthException {
  const GoogleSignInCancelledException([String message = 'Google sign in cancelled'])
      : super(message, code: 'google-sign-in-cancelled');
}
