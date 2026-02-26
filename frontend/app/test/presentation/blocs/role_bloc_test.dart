import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:ledgerguard/domain/entities/user_profile.dart';
import 'package:ledgerguard/domain/repositories/user_profile_repository.dart';
import 'package:ledgerguard/presentation/blocs/role/role_bloc.dart';
import 'package:ledgerguard/presentation/blocs/role/role_event.dart';
import 'package:ledgerguard/presentation/blocs/role/role_state.dart';

class MockUserProfileRepository extends Mock implements UserProfileRepository {}

void main() {
  late MockUserProfileRepository mockRepository;

  const ownerProfile = UserProfile(
    id: 'user-1',
    email: 'owner@example.com',
    role: UserRole.owner,
    planTier: PlanTier.pro,
    displayName: 'Owner User',
  );

  const adminProfile = UserProfile(
    id: 'user-2',
    email: 'admin@example.com',
    role: UserRole.admin,
    planTier: PlanTier.starter,
    displayName: 'Admin User',
  );

  setUp(() {
    mockRepository = MockUserProfileRepository();
  });

  group('RoleBloc', () {
    test('initial state is RoleInitial', () {
      final bloc = RoleBloc(userProfileRepository: mockRepository);
      expect(bloc.state, equals(const RoleInitial()));
      bloc.close();
    });

    group('FetchRoleRequested', () {
      blocTest<RoleBloc, RoleState>(
        'emits [RoleLoading, RoleLoaded] with owner profile',
        setUp: () {
          when(() => mockRepository.fetchUserProfile(any()))
              .thenAnswer((_) async => ownerProfile);
        },
        build: () => RoleBloc(userProfileRepository: mockRepository),
        act: (bloc) => bloc.add(const FetchRoleRequested(authToken: 'token')),
        expect: () => [
          const RoleLoading(),
          const RoleLoaded(ownerProfile),
        ],
        verify: (_) {
          verify(() => mockRepository.fetchUserProfile('token')).called(1);
        },
      );

      blocTest<RoleBloc, RoleState>(
        'emits [RoleLoading, RoleLoaded] with admin profile',
        setUp: () {
          when(() => mockRepository.fetchUserProfile(any()))
              .thenAnswer((_) async => adminProfile);
        },
        build: () => RoleBloc(userProfileRepository: mockRepository),
        act: (bloc) => bloc.add(const FetchRoleRequested(authToken: 'token')),
        expect: () => [
          const RoleLoading(),
          const RoleLoaded(adminProfile),
        ],
      );

      blocTest<RoleBloc, RoleState>(
        'emits [RoleLoading, RoleError] when profile not found',
        setUp: () {
          when(() => mockRepository.fetchUserProfile(any()))
              .thenThrow(const ProfileNotFoundException());
        },
        build: () => RoleBloc(userProfileRepository: mockRepository),
        act: (bloc) => bloc.add(const FetchRoleRequested(authToken: 'token')),
        expect: () => [
          const RoleLoading(),
          const RoleError('Profile not found'),
        ],
      );

      blocTest<RoleBloc, RoleState>(
        'emits [RoleLoading, RoleError] when unauthorized',
        setUp: () {
          when(() => mockRepository.fetchUserProfile(any()))
              .thenThrow(const UnauthorizedException());
        },
        build: () => RoleBloc(userProfileRepository: mockRepository),
        act: (bloc) => bloc.add(const FetchRoleRequested(authToken: 'token')),
        expect: () => [
          const RoleLoading(),
          const RoleError('Unauthorized'),
        ],
      );
    });

    group('ClearRoleRequested', () {
      blocTest<RoleBloc, RoleState>(
        'emits [RoleInitial] and clears cache',
        setUp: () {
          when(() => mockRepository.clearCache()).thenReturn(null);
        },
        build: () => RoleBloc(userProfileRepository: mockRepository),
        seed: () => const RoleLoaded(ownerProfile),
        act: (bloc) => bloc.add(const ClearRoleRequested()),
        expect: () => [const RoleInitial()],
        verify: (_) {
          verify(() => mockRepository.clearCache()).called(1);
        },
      );
    });

    group('RoleLoaded state helpers', () {
      test('isOwner returns true for owner role', () {
        const state = RoleLoaded(ownerProfile);
        expect(state.isOwner, isTrue);
        expect(state.isAdmin, isTrue);
      });

      test('isOwner returns false for admin role', () {
        const state = RoleLoaded(adminProfile);
        expect(state.isOwner, isFalse);
        expect(state.isAdmin, isTrue);
      });

      test('isPro returns true for pro tier', () {
        const state = RoleLoaded(ownerProfile);
        expect(state.isPro, isTrue);
      });

      test('isPro returns false for starter tier', () {
        const state = RoleLoaded(adminProfile);
        expect(state.isPro, isFalse);
      });

      test('hasRole checks permission correctly', () {
        const ownerState = RoleLoaded(ownerProfile);
        const adminState = RoleLoaded(adminProfile);

        // Owner has all permissions
        expect(ownerState.hasRole(UserRole.owner), isTrue);
        expect(ownerState.hasRole(UserRole.admin), isTrue);

        // Admin only has admin permission
        expect(adminState.hasRole(UserRole.owner), isFalse);
        expect(adminState.hasRole(UserRole.admin), isTrue);
      });
    });
  });
}
