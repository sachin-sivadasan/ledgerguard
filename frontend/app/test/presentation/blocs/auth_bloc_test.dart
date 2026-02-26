import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/domain/entities/user_entity.dart';
import 'package:ledgerguard/domain/repositories/auth_repository.dart';
import 'package:ledgerguard/presentation/blocs/auth/auth_bloc.dart';
import 'package:ledgerguard/presentation/blocs/auth/auth_event.dart';
import 'package:ledgerguard/presentation/blocs/auth/auth_state.dart';

class MockAuthRepository extends Mock implements AuthRepository {}

void main() {
  late MockAuthRepository mockAuthRepository;

  const testUser = UserEntity(
    id: 'test-uid',
    email: 'test@example.com',
    displayName: 'Test User',
  );

  setUp(() {
    mockAuthRepository = MockAuthRepository();
  });

  group('AuthBloc', () {
    test('initial state is AuthInitial', () {
      when(() => mockAuthRepository.authStateChanges)
          .thenAnswer((_) => const Stream.empty());

      final bloc = AuthBloc(authRepository: mockAuthRepository);
      expect(bloc.state, equals(const AuthInitial()));
      bloc.close();
    });

    group('AuthCheckRequested', () {
      blocTest<AuthBloc, AuthState>(
        'emits [AuthLoading, Authenticated] when user is logged in',
        setUp: () {
          when(() => mockAuthRepository.authStateChanges)
              .thenAnswer((_) => Stream.value(testUser));
        },
        build: () => AuthBloc(authRepository: mockAuthRepository),
        act: (bloc) => bloc.add(const AuthCheckRequested()),
        expect: () => [
          const AuthLoading(),
          const Authenticated(testUser),
        ],
      );

      blocTest<AuthBloc, AuthState>(
        'emits [AuthLoading, Unauthenticated] when user is not logged in',
        setUp: () {
          when(() => mockAuthRepository.authStateChanges)
              .thenAnswer((_) => Stream.value(null));
        },
        build: () => AuthBloc(authRepository: mockAuthRepository),
        act: (bloc) => bloc.add(const AuthCheckRequested()),
        expect: () => [
          const AuthLoading(),
          const Unauthenticated(),
        ],
      );
    });

    group('SignInWithEmailRequested', () {
      blocTest<AuthBloc, AuthState>(
        'emits [AuthLoading, Authenticated] on successful email sign in',
        setUp: () {
          when(() => mockAuthRepository.authStateChanges)
              .thenAnswer((_) => const Stream.empty());
          when(() => mockAuthRepository.signInWithEmailAndPassword(
                email: any(named: 'email'),
                password: any(named: 'password'),
              )).thenAnswer((_) async => testUser);
        },
        build: () => AuthBloc(authRepository: mockAuthRepository),
        act: (bloc) => bloc.add(const SignInWithEmailRequested(
          email: 'test@example.com',
          password: 'password123',
        )),
        expect: () => [
          const AuthLoading(),
          const Authenticated(testUser),
        ],
        verify: (_) {
          verify(() => mockAuthRepository.signInWithEmailAndPassword(
                email: 'test@example.com',
                password: 'password123',
              )).called(1);
        },
      );

      blocTest<AuthBloc, AuthState>(
        'emits [AuthLoading, AuthError] on invalid credentials',
        setUp: () {
          when(() => mockAuthRepository.authStateChanges)
              .thenAnswer((_) => const Stream.empty());
          when(() => mockAuthRepository.signInWithEmailAndPassword(
                email: any(named: 'email'),
                password: any(named: 'password'),
              )).thenThrow(const InvalidCredentialsException());
        },
        build: () => AuthBloc(authRepository: mockAuthRepository),
        act: (bloc) => bloc.add(const SignInWithEmailRequested(
          email: 'test@example.com',
          password: 'wrongpassword',
        )),
        expect: () => [
          const AuthLoading(),
          const AuthError('Invalid email or password'),
        ],
      );

      blocTest<AuthBloc, AuthState>(
        'emits [AuthLoading, AuthError] on user not found',
        setUp: () {
          when(() => mockAuthRepository.authStateChanges)
              .thenAnswer((_) => const Stream.empty());
          when(() => mockAuthRepository.signInWithEmailAndPassword(
                email: any(named: 'email'),
                password: any(named: 'password'),
              )).thenThrow(const UserNotFoundException());
        },
        build: () => AuthBloc(authRepository: mockAuthRepository),
        act: (bloc) => bloc.add(const SignInWithEmailRequested(
          email: 'notfound@example.com',
          password: 'password123',
        )),
        expect: () => [
          const AuthLoading(),
          const AuthError('User not found'),
        ],
      );
    });

    group('SignInWithGoogleRequested', () {
      blocTest<AuthBloc, AuthState>(
        'emits [AuthLoading, Authenticated] on successful Google sign in',
        setUp: () {
          when(() => mockAuthRepository.authStateChanges)
              .thenAnswer((_) => const Stream.empty());
          when(() => mockAuthRepository.signInWithGoogle())
              .thenAnswer((_) async => testUser);
        },
        build: () => AuthBloc(authRepository: mockAuthRepository),
        act: (bloc) => bloc.add(const SignInWithGoogleRequested()),
        expect: () => [
          const AuthLoading(),
          const Authenticated(testUser),
        ],
        verify: (_) {
          verify(() => mockAuthRepository.signInWithGoogle()).called(1);
        },
      );

      blocTest<AuthBloc, AuthState>(
        'emits [AuthLoading, Unauthenticated] when Google sign in is cancelled',
        setUp: () {
          when(() => mockAuthRepository.authStateChanges)
              .thenAnswer((_) => const Stream.empty());
          when(() => mockAuthRepository.signInWithGoogle())
              .thenThrow(const GoogleSignInCancelledException());
        },
        build: () => AuthBloc(authRepository: mockAuthRepository),
        act: (bloc) => bloc.add(const SignInWithGoogleRequested()),
        expect: () => [
          const AuthLoading(),
          const Unauthenticated(),
        ],
      );

      blocTest<AuthBloc, AuthState>(
        'emits [AuthLoading, AuthError] on Google sign in failure',
        setUp: () {
          when(() => mockAuthRepository.authStateChanges)
              .thenAnswer((_) => const Stream.empty());
          when(() => mockAuthRepository.signInWithGoogle())
              .thenThrow(const AuthException('Google sign in failed'));
        },
        build: () => AuthBloc(authRepository: mockAuthRepository),
        act: (bloc) => bloc.add(const SignInWithGoogleRequested()),
        expect: () => [
          const AuthLoading(),
          const AuthError('Google sign in failed'),
        ],
      );
    });

    group('SignOutRequested', () {
      blocTest<AuthBloc, AuthState>(
        'emits [AuthLoading, Unauthenticated] on successful sign out',
        setUp: () {
          when(() => mockAuthRepository.authStateChanges)
              .thenAnswer((_) => const Stream.empty());
          when(() => mockAuthRepository.signOut()).thenAnswer((_) async {});
        },
        build: () => AuthBloc(authRepository: mockAuthRepository),
        act: (bloc) => bloc.add(const SignOutRequested()),
        expect: () => [
          const AuthLoading(),
          const Unauthenticated(),
        ],
        verify: (_) {
          verify(() => mockAuthRepository.signOut()).called(1);
        },
      );

      blocTest<AuthBloc, AuthState>(
        'emits [AuthLoading, AuthError] on sign out failure',
        setUp: () {
          when(() => mockAuthRepository.authStateChanges)
              .thenAnswer((_) => const Stream.empty());
          when(() => mockAuthRepository.signOut())
              .thenThrow(const AuthException('Sign out failed'));
        },
        build: () => AuthBloc(authRepository: mockAuthRepository),
        act: (bloc) => bloc.add(const SignOutRequested()),
        expect: () => [
          const AuthLoading(),
          const AuthError('Sign out failed'),
        ],
      );
    });
  });
}
