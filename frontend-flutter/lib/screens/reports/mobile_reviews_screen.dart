import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../providers/apps_provider.dart';
import '../../providers/mobile_reviews_provider.dart';
import '../../services/mobile_reviews_service.dart';
import '../../theme/app_breakpoints.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_service_unavailable.dart';

String _count(int n) => NumberFormat.compact().format(n);

class MobileReviewsScreen extends StatefulWidget {
  const MobileReviewsScreen({super.key});

  @override
  State<MobileReviewsScreen> createState() => _MobileReviewsScreenState();
}

class _MobileReviewsScreenState extends State<MobileReviewsScreen>
    with DataLoadingMixin {
  final _appleCtrl = TextEditingController();
  final _playCtrl = TextEditingController();
  bool _editing = false;

  @override
  void loadData(String appId) {
    context.read<MobileReviewsProvider>().setSelectedApp(appId);
  }

  @override
  void dispose() {
    _appleCtrl.dispose();
    _playCtrl.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    final messenger = ScaffoldMessenger.of(context);
    final provider = context.read<MobileReviewsProvider>();
    final ok = await provider.saveLinks(_appleCtrl.text, _playCtrl.text);
    if (!mounted) return;
    if (ok) setState(() => _editing = false);
    messenger.showSnackBar(SnackBar(
        content: Text(ok ? 'Store links saved.' : (provider.error ?? 'Could not save.'))));
  }

  LgPage _shell(Widget child, {List<LgPageAction> actions = const []}) => LgPage(
        title: 'Mobile Ratings & Reviews',
        breadcrumb: 'Reports › Mobile',
        subtitle: 'Your iOS / Android app ratings from the public stores (no login needed).',
        backAction: () => context.go('/reports'),
        secondaryActions: actions,
        child: child,
      );

  @override
  Widget build(BuildContext context) {
    final appsProvider = context.watch<AppsProvider>();
    if (appsProvider.apps.isEmpty) {
      return _shell(LgEmptyState(
        icon: Icons.smartphone_outlined,
        heading: 'No app yet',
        description: 'Connect a Shopify app, then add its mobile store links here.',
        actionLabel: 'Go to Apps',
        onAction: () => context.go('/apps'),
      ));
    }

    final provider = context.watch<MobileReviewsProvider>();
    if (provider.isServiceUnavailable) {
      return _shell(LgServiceUnavailable(onRetry: retryLoad));
    }
    if (provider.error != null && provider.data == null) {
      return _shell(LgErrorState(message: provider.error!, onRetry: retryLoad));
    }
    if (provider.isLoading && provider.data == null) {
      return _shell(const Center(child: CircularProgressIndicator()));
    }

    final data = provider.data ?? MobileReviewsData.empty();
    final apps = appsProvider.apps;
    final showForm = _editing || !data.hasAnyLink;

    final appChip = PopupMenuButton<String>(
      onSelected: provider.setSelectedApp,
      itemBuilder: (_) =>
          apps.map((a) => PopupMenuItem(value: a.id, child: Text(a.name))).toList(),
      child: Chip(
          label: Text(apps
              .firstWhere((a) => a.id == provider.selectedAppId, orElse: () => apps.first)
              .name)),
    );

    return _shell(
      actions: [
        if (!showForm)
          LgPageAction(
              label: 'Edit links',
              onPressed: () {
                _appleCtrl.text = data.iosAppId;
                _playCtrl.text = data.playPackage;
                setState(() => _editing = true);
              }),
      ],
      Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          appChip,
          const SizedBox(height: LgSpacing.s400),
          if (showForm)
            _LinkForm(
              appleCtrl: _appleCtrl,
              playCtrl: _playCtrl,
              saving: provider.isSaving,
              onSave: _save,
              onCancel: data.hasAnyLink ? () => setState(() => _editing = false) : null,
            )
          else
            _StoreCards(data: data),
        ],
      ),
    );
  }
}

class _LinkForm extends StatelessWidget {
  final TextEditingController appleCtrl;
  final TextEditingController playCtrl;
  final bool saving;
  final VoidCallback onSave;
  final VoidCallback? onCancel;
  const _LinkForm(
      {required this.appleCtrl,
      required this.playCtrl,
      required this.saving,
      required this.onSave,
      required this.onCancel});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return LgCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Add your mobile app links',
              style: theme.textTheme.titleSmall),
          const SizedBox(height: LgSpacing.s100),
          Text(
            'Paste the App Store and/or Google Play link (or the app id / package). '
            'We pull public ratings + reviews — downloads and revenue aren\'t public.',
            style: theme.textTheme.bodySmall?.copyWith(color: LgColors.textSecondary),
          ),
          const SizedBox(height: LgSpacing.s400),
          TextField(
            controller: appleCtrl,
            decoration: const InputDecoration(
              labelText: 'App Store',
              hintText: 'apps.apple.com/…/id310633997  or  310633997',
              isDense: true,
              border: OutlineInputBorder(),
            ),
          ),
          const SizedBox(height: LgSpacing.s300),
          TextField(
            controller: playCtrl,
            decoration: const InputDecoration(
              labelText: 'Google Play',
              hintText: 'play.google.com/…?id=com.your.app  or  com.your.app',
              isDense: true,
              border: OutlineInputBorder(),
            ),
          ),
          const SizedBox(height: LgSpacing.s400),
          Row(
            children: [
              FilledButton(
                onPressed: saving ? null : onSave,
                child: Text(saving ? 'Saving…' : 'Save'),
              ),
              if (onCancel != null) ...[
                const SizedBox(width: LgSpacing.s300),
                TextButton(onPressed: onCancel, child: const Text('Cancel')),
              ],
            ],
          ),
        ],
      ),
    );
  }
}

class _StoreCards extends StatelessWidget {
  final MobileReviewsData data;
  const _StoreCards({required this.data});

  @override
  Widget build(BuildContext context) {
    final cards = <Widget>[
      if (data.appStore != null)
        _StoreCard(title: 'App Store', icon: Icons.apple, block: data.appStore!),
      if (data.googlePlay != null)
        _StoreCard(title: 'Google Play', icon: Icons.shop, block: data.googlePlay!),
    ];
    if (cards.isEmpty) {
      return const LgEmptyState(
          icon: Icons.smartphone_outlined,
          heading: 'No store linked',
          description: 'Add an App Store or Google Play link to see ratings.');
    }
    if (LgBreakpoints.isMobile(context) || cards.length == 1) {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          for (var i = 0; i < cards.length; i++) ...[
            if (i > 0) const SizedBox(height: LgSpacing.s400),
            cards[i],
          ],
        ],
      );
    }
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Expanded(child: cards[0]),
        const SizedBox(width: LgSpacing.s400),
        Expanded(child: cards[1]),
      ],
    );
  }
}

class _StoreCard extends StatelessWidget {
  final String title;
  final IconData icon;
  final StoreBlock block;
  const _StoreCard({required this.title, required this.icon, required this.block});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return LgCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(children: [
            Icon(icon, size: 20),
            const SizedBox(width: LgSpacing.s200),
            Text(title, style: theme.textTheme.titleSmall),
          ]),
          const SizedBox(height: LgSpacing.s400),
          if (block.error.isNotEmpty)
            Text(block.error,
                style: theme.textTheme.bodySmall?.copyWith(color: LgColors.warning))
          else ...[
            Row(crossAxisAlignment: CrossAxisAlignment.end, children: [
              Text(block.ratingValue.toStringAsFixed(2),
                  style: theme.textTheme.headlineMedium
                      ?.copyWith(fontWeight: FontWeight.w700)),
              const SizedBox(width: LgSpacing.s200),
              Padding(
                padding: const EdgeInsets.only(bottom: 6),
                child: _Stars(value: block.ratingValue),
              ),
            ]),
            Text('${_count(block.ratingCount)} ratings',
                style: theme.textTheme.bodySmall?.copyWith(color: LgColors.textSecondary)),
            const SizedBox(height: LgSpacing.s400),
            if (block.reviewsAvailable) ...[
              if (block.positive + block.neutral + block.negative > 0)
                _SentimentBar(
                    positive: block.positive, neutral: block.neutral, negative: block.negative),
              const SizedBox(height: LgSpacing.s300),
              for (final r in block.reviews.take(6)) _ReviewRow(review: r),
              if (block.reviews.isEmpty)
                Text('No recent reviews.',
                    style: theme.textTheme.bodySmall
                        ?.copyWith(color: LgColors.textSecondary)),
            ] else
              Text('Review text isn\'t public on Google Play — rating shown above.',
                  style:
                      theme.textTheme.bodySmall?.copyWith(color: LgColors.textSecondary)),
          ],
        ],
      ),
    );
  }
}

class _Stars extends StatelessWidget {
  final double value;
  const _Stars({required this.value});
  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        for (var i = 1; i <= 5; i++)
          Icon(
            value >= i
                ? Icons.star
                : (value >= i - 0.5 ? Icons.star_half : Icons.star_border),
            size: 16,
            color: LgColors.warning,
          ),
      ],
    );
  }
}

class _SentimentBar extends StatelessWidget {
  final int positive;
  final int neutral;
  final int negative;
  const _SentimentBar(
      {required this.positive, required this.neutral, required this.negative});
  @override
  Widget build(BuildContext context) {
    final total = (positive + neutral + negative).clamp(1, 1 << 30);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        ClipRRect(
          borderRadius: BorderRadius.circular(4),
          child: Row(children: [
            Expanded(flex: positive, child: Container(height: 8, color: LgColors.success)),
            Expanded(flex: neutral, child: Container(height: 8, color: LgColors.textSecondary)),
            Expanded(flex: negative, child: Container(height: 8, color: LgColors.critical)),
            if (total == 1) const Expanded(child: SizedBox(height: 8)),
          ]),
        ),
        const SizedBox(height: LgSpacing.s100),
        Text('$positive positive · $neutral neutral · $negative negative (recent)',
            style: Theme.of(context)
                .textTheme
                .bodySmall
                ?.copyWith(color: LgColors.textSecondary)),
      ],
    );
  }
}

class _ReviewRow extends StatelessWidget {
  final MobileReview review;
  const _ReviewRow({required this.review});
  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.only(bottom: LgSpacing.s300),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(children: [
            _Stars(value: review.rating.toDouble()),
            const SizedBox(width: LgSpacing.s200),
            Expanded(
              child: Text(review.title,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: theme.textTheme.bodyMedium
                      ?.copyWith(fontWeight: FontWeight.w600)),
            ),
          ]),
          if (review.body.isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(top: 2),
              child: Text(review.body,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: theme.textTheme.bodySmall),
            ),
          Text(
            '${review.author}${review.version.isNotEmpty ? ' · v${review.version}' : ''}',
            style: theme.textTheme.bodySmall?.copyWith(color: LgColors.textSecondary),
          ),
        ],
      ),
    );
  }
}
