import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../domain/repositories/user_profile_repository.dart';
import 'role_event.dart';
import 'role_state.dart';

/// Bloc for managing user role and plan tier
class RoleBloc extends Bloc<RoleEvent, RoleState> {
  final UserProfileRepository _userProfileRepository;

  RoleBloc({required UserProfileRepository userProfileRepository})
      : _userProfileRepository = userProfileRepository,
        super(const RoleInitial()) {
    on<FetchRoleRequested>(_onFetchRoleRequested);
    on<ClearRoleRequested>(_onClearRoleRequested);
  }

  Future<void> _onFetchRoleRequested(
    FetchRoleRequested event,
    Emitter<RoleState> emit,
  ) async {
    emit(const RoleLoading());

    try {
      final profile = await _userProfileRepository.fetchUserProfile(event.authToken);
      if (profile != null) {
        emit(RoleLoaded(profile));
      } else {
        emit(const RoleError('Profile not found'));
      }
    } on UserProfileException catch (e) {
      emit(RoleError(e.message));
    }
  }

  void _onClearRoleRequested(
    ClearRoleRequested event,
    Emitter<RoleState> emit,
  ) {
    _userProfileRepository.clearCache();
    emit(const RoleInitial());
  }
}
