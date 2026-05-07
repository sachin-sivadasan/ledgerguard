class Organization {
  final String id;
  final String name;
  final String slug;
  final String planTier;
  final DateTime createdAt;

  const Organization({
    required this.id,
    required this.name,
    required this.slug,
    required this.planTier,
    required this.createdAt,
  });

  factory Organization.fromJson(Map<String, dynamic> json) {
    return Organization(
      id: json['id'] as String? ?? '',
      name: json['name'] as String? ?? '',
      slug: json['slug'] as String? ?? '',
      planTier: json['plan_tier'] as String? ?? 'FREE',
      createdAt: DateTime.parse(
          json['created_at'] as String? ?? DateTime.now().toIso8601String()),
    );
  }

  int get maxMembers {
    switch (planTier) {
      case 'FREE':
        return 1;
      case 'STARTER':
        return 3;
      case 'PRO':
        return 10;
      default:
        return 1;
    }
  }
}

class OrgMembership {
  final String orgId;
  final String role;
  final String status;

  const OrgMembership({
    required this.orgId,
    required this.role,
    required this.status,
  });

  factory OrgMembership.fromJson(Map<String, dynamic> json) {
    return OrgMembership(
      orgId: json['org_id'] as String? ?? '',
      role: json['role'] as String? ?? 'VIEWER',
      status: json['status'] as String? ?? 'ACTIVE',
    );
  }

  bool get isOwner => role == 'OWNER';
  bool get isAdmin => role == 'ADMIN' || role == 'OWNER';
  bool get isActive => status == 'ACTIVE';
}

class OrgMember {
  final String id;
  final String userId;
  final String role;
  final String status;
  final DateTime joinedAt;
  final DateTime? suspendedAt;

  const OrgMember({
    required this.id,
    required this.userId,
    required this.role,
    required this.status,
    required this.joinedAt,
    this.suspendedAt,
  });

  factory OrgMember.fromJson(Map<String, dynamic> json) {
    return OrgMember(
      id: json['id'] as String? ?? '',
      userId: json['user_id'] as String? ?? '',
      role: json['role'] as String? ?? 'VIEWER',
      status: json['status'] as String? ?? 'ACTIVE',
      joinedAt: DateTime.parse(
          json['joined_at'] as String? ?? DateTime.now().toIso8601String()),
      suspendedAt: json['suspended_at'] != null
          ? DateTime.parse(json['suspended_at'] as String)
          : null,
    );
  }

  bool get isOwner => role == 'OWNER';
  bool get isAdmin => role == 'ADMIN' || role == 'OWNER';
  bool get isActive => status == 'ACTIVE';
  bool get isSuspended => status == 'SUSPENDED';

  String get roleLabel {
    switch (role) {
      case 'OWNER':
        return 'Owner';
      case 'ADMIN':
        return 'Admin';
      case 'VIEWER':
        return 'Viewer';
      default:
        return role;
    }
  }
}

class OrgInvitation {
  final String id;
  final String email;
  final String role;
  final String token;
  final DateTime expiresAt;

  const OrgInvitation({
    required this.id,
    required this.email,
    required this.role,
    required this.token,
    required this.expiresAt,
  });

  factory OrgInvitation.fromJson(Map<String, dynamic> json) {
    return OrgInvitation(
      id: json['id'] as String? ?? '',
      email: json['email'] as String? ?? '',
      role: json['role'] as String? ?? 'VIEWER',
      token: json['token'] as String? ?? '',
      expiresAt: DateTime.parse(
          json['expires_at'] as String? ?? DateTime.now().toIso8601String()),
    );
  }

  bool get isExpired => DateTime.now().isAfter(expiresAt);
}

class OrgAuditEntry {
  final String id;
  final String actorId;
  final String action;
  final String? targetType;
  final String? targetId;
  final Map<String, dynamic>? metadata;
  final DateTime createdAt;

  const OrgAuditEntry({
    required this.id,
    required this.actorId,
    required this.action,
    this.targetType,
    this.targetId,
    this.metadata,
    required this.createdAt,
  });

  factory OrgAuditEntry.fromJson(Map<String, dynamic> json) {
    return OrgAuditEntry(
      id: json['id'] as String? ?? '',
      actorId: json['actor_id'] as String? ?? '',
      action: json['action'] as String? ?? '',
      targetType: json['target_type'] as String?,
      targetId: json['target_id'] as String?,
      metadata: json['metadata'] as Map<String, dynamic>?,
      createdAt: DateTime.parse(
          json['created_at'] as String? ?? DateTime.now().toIso8601String()),
    );
  }

  String get actionLabel {
    switch (action) {
      case 'org.created':
        return 'Created organization';
      case 'org.updated':
        return 'Updated organization';
      case 'org.deleted':
        return 'Deleted organization';
      case 'member.invited':
        return 'Invited member';
      case 'member.joined':
        return 'Member joined';
      case 'member.removed':
        return 'Removed member';
      case 'member.suspended':
        return 'Suspended member';
      case 'member.unsuspended':
        return 'Unsuspended member';
      case 'role.changed':
        return 'Changed role';
      case 'invitation.revoked':
        return 'Revoked invitation';
      case 'webhook.configured':
        return 'Configured webhook';
      default:
        return action.replaceAll('.', ' ').replaceAll('_', ' ');
    }
  }
}
