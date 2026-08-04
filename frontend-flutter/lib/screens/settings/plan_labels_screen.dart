import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../providers/apps_provider.dart';
import '../../providers/plan_label_provider.dart';
import '../../services/plan_label_service.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_service_unavailable.dart';

/// Settings › Plan labels — name your price tiers so the plan-based reports show
/// "Starter" / "Starter (old)" instead of derived "$29.00/mo" labels.
class PlanLabelsScreen extends StatefulWidget {
  const PlanLabelsScreen({super.key});

  @override
  State<PlanLabelsScreen> createState() => _PlanLabelsScreenState();
}

class _PlanLabelsScreenState extends State<PlanLabelsScreen>
    with DataLoadingMixin {
  final Map<String, TextEditingController> _controllers = {};
  String? _controllersAppId; // the app the current controllers were seeded for

  @override
  void loadData(String appId) {
    context.read<PlanLabelProvider>().setSelectedApp(appId);
  }

  @override
  void dispose() {
    for (final c in _controllers.values) {
      c.dispose();
    }
    super.dispose();
  }

  // Keep one controller per tier key, seeded with the current label. When the selected app
  // changes, discard all controllers first — otherwise a tier key shared by two apps (e.g.
  // both priced at $29 → "MONTHLY:2900") would keep the previous app's text.
  void _syncControllers(PlanLabelProvider provider) {
    if (_controllersAppId != provider.selectedAppId) {
      for (final c in _controllers.values) {
        c.dispose();
      }
      _controllers.clear();
      _controllersAppId = provider.selectedAppId;
    }
    final keys = provider.tiers.map((t) => t.key).toSet();
    _controllers.removeWhere((k, c) {
      if (!keys.contains(k)) {
        c.dispose();
        return true;
      }
      return false;
    });
    for (final t in provider.tiers) {
      _controllers.putIfAbsent(t.key, () => TextEditingController(text: provider.labelFor(t)));
    }
  }

  Future<void> _save() async {
    final messenger = ScaffoldMessenger.of(context);
    final provider = context.read<PlanLabelProvider>();
    final ok = await provider.save();
    if (!mounted) return;
    messenger.showSnackBar(SnackBar(
      content: Text(ok ? 'Plan labels saved.' : (provider.error ?? 'Could not save.')),
    ));
  }

  LgPage _shell(Widget child, {List<LgPageAction> actions = const []}) => LgPage(
        title: 'Plan labels',
        breadcrumb: 'Settings › Plan labels',
        subtitle: 'Name your price tiers — these labels show across your plan reports.',
        backAction: () => context.go('/settings'),
        secondaryActions: actions,
        child: child,
      );

  @override
  Widget build(BuildContext context) {
    final apps = context.watch<AppsProvider>().apps;
    if (apps.isEmpty) {
      return _shell(LgEmptyState(
        icon: Icons.sell_outlined,
        heading: 'No app yet',
        description: 'Connect a Shopify app to name its plans.',
        actionLabel: 'Go to Apps',
        onAction: () => context.go('/apps'),
      ));
    }

    final provider = context.watch<PlanLabelProvider>();
    if (provider.isServiceUnavailable) {
      return _shell(LgServiceUnavailable(
          onRetry: () =>
              provider.selectedAppId != null ? provider.load(provider.selectedAppId!) : null));
    }
    if (provider.error != null && provider.tiers.isEmpty) {
      return _shell(LgErrorState(
          message: provider.error!,
          onRetry: () =>
              provider.selectedAppId != null ? provider.load(provider.selectedAppId!) : null));
    }
    if (provider.isLoading && provider.tiers.isEmpty) {
      return _shell(const Center(child: CircularProgressIndicator()));
    }

    _syncControllers(provider);

    final appChip = PopupMenuButton<String>(
      onSelected: provider.setSelectedApp,
      itemBuilder: (_) =>
          apps.map((a) => PopupMenuItem(value: a.id, child: Text(a.name))).toList(),
      child: Chip(
        label: Text(apps
            .firstWhere((a) => a.id == provider.selectedAppId, orElse: () => apps.first)
            .name),
      ),
    );

    if (provider.tiers.isEmpty) {
      return _shell(Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          appChip,
          const SizedBox(height: LgSpacing.s400),
          const LgEmptyState(
            icon: Icons.sell_outlined,
            heading: 'No unnamed tiers',
            description:
                'Every plan already has a name, or no paying subscriptions have synced yet.',
          ),
        ],
      ));
    }

    return _shell(
      actions: [
        LgPageAction(
          label: provider.isSaving ? 'Saving…' : 'Save',
          onPressed: provider.isSaving ? () {} : _save,
        ),
      ],
      Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          appChip,
          const SizedBox(height: LgSpacing.s400),
          Text(
            'A price change creates a new tier, so you can name "Starter" and "Starter (old)" separately.'
            '${provider.hiddenTiers > 0 ? ' ${provider.hiddenTiers} minor tiers (prorations / one-off charges) are not shown.' : ''}',
            style: Theme.of(context)
                .textTheme
                .bodySmall
                ?.copyWith(color: LgColors.textSecondary),
          ),
          const SizedBox(height: LgSpacing.s400),
          LgCard(
            child: Column(
              children: [
                for (var i = 0; i < provider.tiers.length; i++) ...[
                  if (i > 0) const Divider(height: LgSpacing.s500),
                  _TierRow(
                    tier: provider.tiers[i],
                    controller: _controllers[provider.tiers[i].key]!,
                    onChanged: (v) => provider.editLabel(provider.tiers[i].key, v),
                  ),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _TierRow extends StatelessWidget {
  final PlanTier tier;
  final TextEditingController controller;
  final ValueChanged<String> onChanged;
  const _TierRow(
      {required this.tier, required this.controller, required this.onChanged});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: LgSpacing.s200),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          SizedBox(
            width: 130,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(tier.pseudoLabel,
                    style: theme.textTheme.bodyMedium
                        ?.copyWith(fontWeight: FontWeight.w600)),
                Text('${tier.customers} customers',
                    style: theme.textTheme.bodySmall
                        ?.copyWith(color: LgColors.textSecondary)),
              ],
            ),
          ),
          const SizedBox(width: LgSpacing.s400),
          Expanded(
            child: TextField(
              controller: controller,
              onChanged: onChanged,
              maxLength: 120,
              decoration: const InputDecoration(
                hintText: 'Plan name (e.g. Starter)',
                isDense: true,
                counterText: '',
                border: OutlineInputBorder(),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
