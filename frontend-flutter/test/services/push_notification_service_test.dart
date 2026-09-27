import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:ledgerguard_flutter/core/network/api_client.dart';
import 'package:ledgerguard_flutter/services/push_notification_service.dart';

/// Records post/delete calls; other verbs are unused by PushNotificationService.
class FakeApiClient implements ApiClient {
  final List<(String, dynamic)> posts = [];
  final List<(String, dynamic)> deletes = [];

  Response<T> _ok<T>(String path) =>
      Response<T>(requestOptions: RequestOptions(path: path), statusCode: 200);

  @override
  Future<Response<T>> post<T>(String path,
      {dynamic data,
      Map<String, dynamic>? queryParameters,
      Options? options,
      CancelToken? cancelToken}) async {
    posts.add((path, data));
    return _ok<T>(path);
  }

  @override
  Future<Response<T>> delete<T>(String path,
      {dynamic data,
      Map<String, dynamic>? queryParameters,
      Options? options,
      CancelToken? cancelToken}) async {
    deletes.add((path, data));
    return _ok<T>(path);
  }

  @override
  dynamic noSuchMethod(Invocation invocation) =>
      throw UnimplementedError('${invocation.memberName} not used in tests');
}

void main() {
  test('registerToken posts device_token + platform to /api/v1/devices', () async {
    final api = FakeApiClient();
    final svc = PushNotificationService(
      api: api,
      requestPermission: () async => true,
      getToken: () async => 'fcm-tok-1',
      platform: 'android',
    );

    await svc.registerToken();

    expect(api.posts.length, 1);
    expect(api.posts.first.$1, '/api/v1/devices');
    expect(api.posts.first.$2, {'device_token': 'fcm-tok-1', 'platform': 'android'});
  });

  test('permission denied -> no registration (Review Focus #5)', () async {
    final api = FakeApiClient();
    final svc = PushNotificationService(
      api: api,
      requestPermission: () async => false,
      getToken: () async => 'fcm-tok-1',
      platform: 'ios',
    );

    await svc.registerToken();

    expect(api.posts, isEmpty);
  });

  test('null/empty token -> no registration', () async {
    final api = FakeApiClient();
    final svc = PushNotificationService(
      api: api,
      requestPermission: () async => true,
      getToken: () async => null,
      platform: 'ios',
    );

    await svc.registerToken();

    expect(api.posts, isEmpty);
  });

  test('token refresh re-registers the new token', () async {
    final api = FakeApiClient();
    final refresh = StreamController<String>();
    final svc = PushNotificationService(
      api: api,
      requestPermission: () async => true,
      getToken: () async => 'tok-1',
      onTokenRefresh: () => refresh.stream,
      platform: 'android',
    );

    await svc.registerToken();
    refresh.add('tok-2');
    await Future<void>.delayed(Duration.zero); // let the stream listener run

    expect(api.posts.map((p) => (p.$2 as Map)['device_token']).toList(),
        ['tok-1', 'tok-2']);
    await refresh.close();
  });

  test('unregister deletes the current token (Review Focus #4 mechanism)', () async {
    final api = FakeApiClient();
    final svc = PushNotificationService(
      api: api,
      requestPermission: () async => true,
      getToken: () async => 'tok-1',
      platform: 'ios',
    );

    await svc.registerToken();
    await svc.unregister();

    expect(api.deletes.length, 1);
    expect(api.deletes.first.$1, '/api/v1/devices');
    expect(api.deletes.first.$2, {'device_token': 'tok-1'});
  });

  test('unregister with no token is a no-op', () async {
    final api = FakeApiClient();
    final svc = PushNotificationService(
      api: api,
      requestPermission: () async => true,
      getToken: () async => null,
      platform: 'ios',
    );

    await svc.unregister();

    expect(api.deletes, isEmpty);
  });
}
