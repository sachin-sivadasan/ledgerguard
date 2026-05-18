import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../core/network/api_client.dart';
import '../models/organization_model.dart';

class OrganizationService {
  final ApiClient _client;

  OrganizationService(this._client);

  // --- Organization CRUD ---

  Future<Organization?> createOrganization(String name) async {
    try {
      final response =
          await _client.post('/api/v1/orgs', data: {'name': name});
      return Organization.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      debugPrint('[OrgService] createOrg error: ${e.response?.statusCode}');
      return null;
    }
  }

  Future<List<OrgMembership>> listOrganizations() async {
    try {
      final response = await _client.get('/api/v1/orgs');
      final list =
          response.data['organizations'] as List<dynamic>? ?? [];
      return list
          .map((json) => OrgMembership.fromJson(json as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      debugPrint('[OrgService] listOrgs error: ${e.response?.statusCode}');
      if (e.response?.statusCode == 503) rethrow;
      return [];
    }
  }

  Future<Organization?> getOrganization(String orgId) async {
    try {
      final response = await _client.get('/api/v1/orgs/$orgId');
      return Organization.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      debugPrint('[OrgService] getOrg error: ${e.response?.statusCode}');
      return null;
    }
  }

  Future<bool> updateOrganization(String orgId, String name) async {
    try {
      await _client.put('/api/v1/orgs/$orgId', data: {'name': name});
      return true;
    } on DioException catch (e) {
      debugPrint('[OrgService] updateOrg error: ${e.response?.statusCode}');
      return false;
    }
  }

  Future<bool> deleteOrganization(String orgId) async {
    try {
      await _client.delete('/api/v1/orgs/$orgId');
      return true;
    } on DioException catch (e) {
      debugPrint('[OrgService] deleteOrg error: ${e.response?.statusCode}');
      return false;
    }
  }

  // --- Members ---

  Future<List<OrgMember>> listMembers(String orgId) async {
    try {
      final response = await _client.get('/api/v1/orgs/$orgId/members');
      final list = response.data['members'] as List<dynamic>? ?? [];
      return list
          .map((json) => OrgMember.fromJson(json as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      debugPrint('[OrgService] listMembers error: ${e.response?.statusCode}');
      return [];
    }
  }

  Future<bool> removeMember(String orgId, String memberId) async {
    try {
      await _client.delete('/api/v1/orgs/$orgId/members/$memberId');
      return true;
    } on DioException catch (e) {
      debugPrint('[OrgService] removeMember error: ${e.response?.statusCode}');
      return false;
    }
  }

  Future<bool> changeRole(String orgId, String memberId, String role) async {
    try {
      await _client
          .put('/api/v1/orgs/$orgId/members/$memberId/role', data: {'role': role});
      return true;
    } on DioException catch (e) {
      debugPrint('[OrgService] changeRole error: ${e.response?.statusCode}');
      return false;
    }
  }

  Future<bool> suspendMember(String orgId, String memberId) async {
    try {
      await _client.put('/api/v1/orgs/$orgId/members/$memberId/suspend');
      return true;
    } on DioException catch (e) {
      debugPrint('[OrgService] suspend error: ${e.response?.statusCode}');
      return false;
    }
  }

  Future<bool> unsuspendMember(String orgId, String memberId) async {
    try {
      await _client.put('/api/v1/orgs/$orgId/members/$memberId/unsuspend');
      return true;
    } on DioException catch (e) {
      debugPrint('[OrgService] unsuspend error: ${e.response?.statusCode}');
      return false;
    }
  }

  // --- Invitations ---

  Future<OrgInvitation?> inviteMember(
      String orgId, String email, String role) async {
    try {
      final response = await _client.post('/api/v1/orgs/$orgId/invitations',
          data: {'email': email, 'role': role});
      return OrgInvitation.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      debugPrint('[OrgService] invite error: ${e.response?.statusCode}');
      return null;
    }
  }

  Future<bool> revokeInvitation(String orgId, String invitationId) async {
    try {
      await _client.delete('/api/v1/orgs/$orgId/invitations/$invitationId');
      return true;
    } on DioException catch (e) {
      debugPrint('[OrgService] revoke error: ${e.response?.statusCode}');
      return false;
    }
  }

  Future<bool> acceptInvitation(String token) async {
    try {
      await _client.post('/api/v1/invitations/$token/accept');
      return true;
    } on DioException catch (e) {
      debugPrint('[OrgService] accept error: ${e.response?.statusCode}');
      return false;
    }
  }

  // --- Webhooks ---

  Future<bool> configureWebhook(String orgId, String url) async {
    try {
      await _client.put('/api/v1/orgs/$orgId/webhooks', data: {'url': url});
      return true;
    } on DioException catch (e) {
      debugPrint('[OrgService] webhook error: ${e.response?.statusCode}');
      return false;
    }
  }

  // --- Audit Log ---

  Future<List<OrgAuditEntry>> getAuditLog(String orgId,
      {int limit = 50, int offset = 0}) async {
    try {
      final response = await _client.get('/api/v1/orgs/$orgId/audit-log',
          queryParameters: {'limit': limit, 'offset': offset});
      final list = response.data['entries'] as List<dynamic>? ?? [];
      return list
          .map(
              (json) => OrgAuditEntry.fromJson(json as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      debugPrint('[OrgService] auditLog error: ${e.response?.statusCode}');
      return [];
    }
  }
}
