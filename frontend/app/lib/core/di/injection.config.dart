// GENERATED CODE - DO NOT MODIFY BY HAND

// **************************************************************************
// InjectableConfigGenerator
// **************************************************************************

// Run: dart run build_runner build

import 'package:get_it/get_it.dart' as _i1;
import 'package:injectable/injectable.dart' as _i2;

import '../../data/repositories/api_api_key_repository.dart';
import '../../data/repositories/api_dashboard_repository.dart';
import '../../data/repositories/api_dashboard_preferences_repository.dart';
import '../../data/repositories/api_insight_repository.dart';
import '../../data/repositories/api_notification_preferences_repository.dart';
import '../../data/repositories/api_risk_repository.dart';
import '../../data/repositories/api_subscription_repository.dart';
import '../../data/repositories/api_user_profile_repository.dart';
import '../../data/repositories/firebase_auth_repository.dart';
import '../../data/repositories/api_app_repository.dart';
import '../../data/repositories/api_partner_integration_repository.dart';
import '../../domain/repositories/api_key_repository.dart';
import '../../domain/repositories/app_repository.dart';
import '../../domain/repositories/auth_repository.dart';
import '../../domain/repositories/dashboard_repository.dart';
import '../../domain/repositories/dashboard_preferences_repository.dart';
import '../../domain/repositories/insight_repository.dart';
import '../../domain/repositories/notification_preferences_repository.dart';
import '../../domain/repositories/partner_integration_repository.dart';
import '../../domain/repositories/risk_repository.dart';
import '../../domain/repositories/subscription_repository.dart';
import '../../domain/repositories/user_profile_repository.dart';
import '../../presentation/blocs/api_key/api_key_bloc.dart';
import '../../presentation/blocs/app_selection/app_selection_bloc.dart';
import '../../presentation/blocs/auth/auth_bloc.dart';
import '../../presentation/blocs/dashboard/dashboard_bloc.dart';
import '../../presentation/blocs/insight/insight_bloc.dart';
import '../../presentation/blocs/notification_preferences/notification_preferences_bloc.dart';
import '../../presentation/blocs/partner_integration/partner_integration_bloc.dart';
import '../../presentation/blocs/preferences/preferences_bloc.dart';
import '../../presentation/blocs/risk/risk_bloc.dart';
import '../../presentation/blocs/role/role_bloc.dart';
import '../../presentation/blocs/subscription_detail/subscription_detail_bloc.dart';
import '../../presentation/blocs/subscription_list/subscription_list_bloc.dart';
import '../network/api_client.dart';
import '../services/snackbar_service.dart';

extension GetItInjectableX on _i1.GetIt {
  _i1.GetIt init({
    String? environment,
    _i2.EnvironmentFilter? environmentFilter,
  }) {
    // Services
    registerLazySingleton<SnackbarService>(() => SnackbarService());

    // Repositories
    registerLazySingleton<AuthRepository>(() => FirebaseAuthRepository());

    // API Client (depends on AuthRepository and SnackbarService)
    registerLazySingleton<ApiClient>(() => ApiClient(
          authRepository: get<AuthRepository>(),
          snackbarService: get<SnackbarService>(),
        ));
    registerLazySingleton<UserProfileRepository>(() => ApiUserProfileRepository());
    registerLazySingleton<PartnerIntegrationRepository>(() => ApiPartnerIntegrationRepository(
          apiClient: get<ApiClient>(),
        ));
    registerLazySingleton<AppRepository>(() => ApiAppRepository(
          apiClient: get<ApiClient>(),
        ));
    registerLazySingleton<DashboardRepository>(() => ApiDashboardRepository(
          authRepository: get<AuthRepository>(),
          appRepository: get<AppRepository>(),
        ));
    registerLazySingleton<DashboardPreferencesRepository>(() => ApiDashboardPreferencesRepository(
          authRepository: get<AuthRepository>(),
        ));
    registerLazySingleton<RiskRepository>(() => ApiRiskRepository(
          authRepository: get<AuthRepository>(),
          appRepository: get<AppRepository>(),
        ));
    registerLazySingleton<InsightRepository>(() => ApiInsightRepository(
          authRepository: get<AuthRepository>(),
          appRepository: get<AppRepository>(),
        ));
    registerLazySingleton<NotificationPreferencesRepository>(() => ApiNotificationPreferencesRepository(
          authRepository: get<AuthRepository>(),
        ));
    registerLazySingleton<SubscriptionRepository>(() => ApiSubscriptionRepository(
          apiClient: get<ApiClient>(),
        ));
    registerLazySingleton<ApiKeyRepository>(() => ApiApiKeyRepository(
          authRepository: get<AuthRepository>(),
        ));

    // Blocs
    registerFactory<AuthBloc>(() => AuthBloc(authRepository: get<AuthRepository>()));
    registerFactory<RoleBloc>(() => RoleBloc(userProfileRepository: get<UserProfileRepository>()));
    registerFactory<PartnerIntegrationBloc>(() => PartnerIntegrationBloc(repository: get<PartnerIntegrationRepository>()));
    registerFactory<AppSelectionBloc>(() => AppSelectionBloc(appRepository: get<AppRepository>()));
    registerFactory<DashboardBloc>(() => DashboardBloc(repository: get<DashboardRepository>()));
    registerFactory<InsightBloc>(() => InsightBloc(repository: get<InsightRepository>()));
    registerFactory<NotificationPreferencesBloc>(() => NotificationPreferencesBloc(repository: get<NotificationPreferencesRepository>()));
    registerFactory<PreferencesBloc>(() => PreferencesBloc(repository: get<DashboardPreferencesRepository>()));
    registerFactory<RiskBloc>(() => RiskBloc(repository: get<RiskRepository>()));
    registerFactory<SubscriptionListBloc>(() => SubscriptionListBloc(repository: get<SubscriptionRepository>()));
    registerFactory<SubscriptionDetailBloc>(() => SubscriptionDetailBloc(repository: get<SubscriptionRepository>()));
    registerFactory<ApiKeyBloc>(() => ApiKeyBloc(repository: get<ApiKeyRepository>()));

    return this;
  }
}
