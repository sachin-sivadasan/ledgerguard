// GENERATED CODE - DO NOT MODIFY BY HAND

// **************************************************************************
// InjectableConfigGenerator
// **************************************************************************

// Run: dart run build_runner build

import 'package:get_it/get_it.dart' as _i1;
import 'package:injectable/injectable.dart' as _i2;

import '../../data/repositories/api_dashboard_repository.dart';
import '../../data/repositories/api_dashboard_preferences_repository.dart';
import '../../data/repositories/api_user_profile_repository.dart';
import '../../data/repositories/firebase_auth_repository.dart';
import '../../data/repositories/mock_app_repository.dart';
import '../../data/repositories/mock_partner_integration_repository.dart';
import '../../domain/repositories/app_repository.dart';
import '../../domain/repositories/auth_repository.dart';
import '../../domain/repositories/dashboard_repository.dart';
import '../../domain/repositories/dashboard_preferences_repository.dart';
import '../../domain/repositories/partner_integration_repository.dart';
import '../../domain/repositories/user_profile_repository.dart';
import '../../presentation/blocs/app_selection/app_selection_bloc.dart';
import '../../presentation/blocs/auth/auth_bloc.dart';
import '../../presentation/blocs/dashboard/dashboard_bloc.dart';
import '../../presentation/blocs/partner_integration/partner_integration_bloc.dart';
import '../../presentation/blocs/preferences/preferences_bloc.dart';
import '../../presentation/blocs/role/role_bloc.dart';

extension GetItInjectableX on _i1.GetIt {
  _i1.GetIt init({
    String? environment,
    _i2.EnvironmentFilter? environmentFilter,
  }) {
    // Repositories
    registerLazySingleton<AuthRepository>(() => FirebaseAuthRepository());
    registerLazySingleton<UserProfileRepository>(() => ApiUserProfileRepository());
    registerLazySingleton<PartnerIntegrationRepository>(() => MockPartnerIntegrationRepository());
    registerLazySingleton<AppRepository>(() => MockAppRepository());
    registerLazySingleton<DashboardRepository>(() => ApiDashboardRepository(
          authRepository: get<AuthRepository>(),
          appRepository: get<AppRepository>(),
        ));
    registerLazySingleton<DashboardPreferencesRepository>(() => ApiDashboardPreferencesRepository(
          authRepository: get<AuthRepository>(),
        ));

    // Blocs
    registerFactory<AuthBloc>(() => AuthBloc(authRepository: get<AuthRepository>()));
    registerFactory<RoleBloc>(() => RoleBloc(userProfileRepository: get<UserProfileRepository>()));
    registerFactory<PartnerIntegrationBloc>(() => PartnerIntegrationBloc(repository: get<PartnerIntegrationRepository>()));
    registerFactory<AppSelectionBloc>(() => AppSelectionBloc(appRepository: get<AppRepository>()));
    registerFactory<DashboardBloc>(() => DashboardBloc(repository: get<DashboardRepository>()));
    registerFactory<PreferencesBloc>(() => PreferencesBloc(repository: get<DashboardPreferencesRepository>()));

    return this;
  }
}
