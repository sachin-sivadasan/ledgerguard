import 'package:flutter/foundation.dart';
import '../core/network/api_client.dart';
import '../models/organization_model.dart';
import '../services/organization_service.dart';

class OrganizationProvider extends ChangeNotifier {
  final OrganizationService _orgService;
  final ApiClient? _apiClient;

  bool _isLoading = false;
  String? _error;

  // Current org context
  Organization? _currentOrg;
  List<OrgMembership> _memberships = [];

  // Members of current org
  List<OrgMember> _members = [];
  bool _isMembersLoading = false;

  // Audit log
  List<OrgAuditEntry> _auditEntries = [];
  bool _isAuditLoading = false;
  final int _auditTotal = 0;

  OrganizationProvider(this._orgService, {ApiClient? apiClient})
      : _apiClient = apiClient;

  // --- Getters ---
  bool get isLoading => _isLoading;
  String? get error => _error;
  Organization? get currentOrg => _currentOrg;
  List<OrgMembership> get memberships => _memberships;
  List<OrgMember> get members => _members;
  bool get isMembersLoading => _isMembersLoading;
  List<OrgAuditEntry> get auditEntries => _auditEntries;
  bool get isAuditLoading => _isAuditLoading;
  int get auditTotal => _auditTotal;

  bool get hasMultipleOrgs => _memberships.length > 1;
  String? get currentOrgId => _currentOrg?.id;

  OrgMembership? get currentMembership {
    if (_currentOrg == null) return null;
    try {
      return _memberships.firstWhere((m) => m.orgId == _currentOrg!.id);
    } catch (_) {
      return null;
    }
  }

  bool get isOwner => currentMembership?.isOwner ?? false;
  bool get isAdmin => currentMembership?.isAdmin ?? false;

  // --- Load Organizations ---

  Future<void> loadOrganizations() async {
    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      _memberships = await _orgService.listOrganizations();

      // Auto-select first org if none selected
      if (_currentOrg == null && _memberships.isNotEmpty) {
        await selectOrganization(_memberships.first.orgId);
        return; // selectOrganization calls notifyListeners
      }
    } catch (e) {
      _error = e.toString();
    }

    _isLoading = false;
    notifyListeners();
  }

  // --- Select / Switch Org ---

  Future<void> selectOrganization(String orgId) async {
    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      _currentOrg = await _orgService.getOrganization(orgId);
      _apiClient?.setOrgId(_currentOrg?.id);
    } catch (e) {
      _error = e.toString();
    }

    _isLoading = false;
    notifyListeners();
  }

  // --- Create Org ---

  Future<Organization?> createOrganization(String name) async {
    final org = await _orgService.createOrganization(name);
    if (org != null) {
      _currentOrg = org;
      await loadOrganizations();
    }
    return org;
  }

  // --- Members ---

  Future<void> loadMembers() async {
    if (_currentOrg == null) return;
    _isMembersLoading = true;
    notifyListeners();

    try {
      _members = await _orgService.listMembers(_currentOrg!.id);
    } catch (e) {
      _error = e.toString();
    }

    _isMembersLoading = false;
    notifyListeners();
  }

  Future<OrgInvitation?> inviteMember(String email, String role) async {
    if (_currentOrg == null) return null;
    final inv = await _orgService.inviteMember(_currentOrg!.id, email, role);
    if (inv != null) {
      await loadMembers();
    }
    return inv;
  }

  Future<bool> removeMember(String memberId) async {
    if (_currentOrg == null) return false;
    final ok = await _orgService.removeMember(_currentOrg!.id, memberId);
    if (ok) await loadMembers();
    return ok;
  }

  Future<bool> suspendMember(String memberId) async {
    if (_currentOrg == null) return false;
    final ok = await _orgService.suspendMember(_currentOrg!.id, memberId);
    if (ok) await loadMembers();
    return ok;
  }

  Future<bool> unsuspendMember(String memberId) async {
    if (_currentOrg == null) return false;
    final ok = await _orgService.unsuspendMember(_currentOrg!.id, memberId);
    if (ok) await loadMembers();
    return ok;
  }

  Future<bool> changeRole(String memberId, String role) async {
    if (_currentOrg == null) return false;
    final ok =
        await _orgService.changeRole(_currentOrg!.id, memberId, role);
    if (ok) await loadMembers();
    return ok;
  }

  // --- Audit Log ---

  Future<void> loadAuditLog({int limit = 50, int offset = 0}) async {
    if (_currentOrg == null) return;
    _isAuditLoading = true;
    notifyListeners();

    try {
      _auditEntries = await _orgService.getAuditLog(_currentOrg!.id,
          limit: limit, offset: offset);
    } catch (e) {
      _error = e.toString();
    }

    _isAuditLoading = false;
    notifyListeners();
  }

  // --- Webhooks ---

  Future<bool> configureWebhook(String url) async {
    if (_currentOrg == null) return false;
    return _orgService.configureWebhook(_currentOrg!.id, url);
  }
}
