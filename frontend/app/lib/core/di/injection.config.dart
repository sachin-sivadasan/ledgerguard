// GENERATED CODE - DO NOT MODIFY BY HAND

// **************************************************************************
// InjectableConfigGenerator
// **************************************************************************

// Run: dart run build_runner build

import 'package:get_it/get_it.dart' as _i1;
import 'package:injectable/injectable.dart' as _i2;

import '../../data/repositories/api_user_profile_repository.dart';
import '../../data/repositories/firebase_auth_repository.dart';
import '../../domain/repositories/auth_repository.dart';
import '../../domain/repositories/user_profile_repository.dart';
import '../../presentation/blocs/auth/auth_bloc.dart';
import '../../presentation/blocs/role/role_bloc.dart';

extension GetItInjectableX on _i1.GetIt {
  _i1.GetIt init({
    String? environment,
    _i2.EnvironmentFilter? environmentFilter,
  }) {
    // Repositories
    registerLazySingleton<AuthRepository>(() => FirebaseAuthRepository());
    registerLazySingleton<UserProfileRepository>(() => ApiUserProfileRepository());

    // Blocs
    registerFactory<AuthBloc>(() => AuthBloc(authRepository: get<AuthRepository>()));
    registerFactory<RoleBloc>(() => RoleBloc(userProfileRepository: get<UserProfileRepository>()));

    return this;
  }
}
