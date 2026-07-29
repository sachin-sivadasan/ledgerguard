import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../core/utils/file_download.dart';
import '../../models/review_model.dart';
import '../../providers/apps_provider.dart';
import '../../providers/reviews_provider.dart';
import '../../services/reviews_service.dart';
import '../../theme/app_breakpoints.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_service_unavailable.dart';

const _starColor = LgColors.starRating;

class ReviewsReportScreen extends StatefulWidget {
  const ReviewsReportScreen({super.key});

  @override
  State<ReviewsReportScreen> createState() => _ReviewsReportScreenState();
}

class _ReviewsReportScreenState extends State<ReviewsReportScreen>
    with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<ReviewsProvider>().setSelectedApp(appId);
  }

  Future<void> _exportCsv() async {
    final messenger = ScaffoldMessenger.of(context);
    final provider = context.read<ReviewsProvider>();
    final appId = provider.selectedAppId;
    if (appId == null) return;

    try {
      final bytes = await provider.fetchCsvBytes();
      if (bytes == null || bytes.isEmpty) {
        if (!mounted) return;
        messenger.showSnackBar(
          const SnackBar(content: Text('CSV export returned no data.')),
        );
        return;
      }
      final filename =
          'reviews-${DateTime.now().toIso8601String().split('T').first}.csv';
      final ok = downloadBytes(bytes, filename, 'text/csv');
      if (!mounted) return;
      if (!ok) {
        messenger.showSnackBar(
          const SnackBar(
            content: Text('CSV export is only available on the web app.'),
          ),
        );
      }
    } catch (e) {
      // Don't swallow the cause — surface a 503 with the same "service
      // unavailable" copy the report body uses, and log everything else so
      // export failures stay diagnosable.
      debugPrint('reviews: CSV export failed: $e');
      if (!mounted) return;
      final isUnavailable = e is DioException && e.response?.statusCode == 503;
      messenger.showSnackBar(
        SnackBar(
          content: Text(
            isUnavailable
                ? 'Service temporarily unavailable. Please try again shortly.'
                : 'Could not export CSV. Please try again.',
          ),
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final appsProvider = context.watch<AppsProvider>();
    final hasApps = appsProvider.apps.isNotEmpty;

    if (!hasApps) {
      return LgPage(
        title: 'Reviews',
        breadcrumb: 'Reports › Retention & Risk',
        backAction: () => context.go('/reports'),
        child: LgEmptyState(
          icon: Icons.star_outline_rounded,
          heading: 'No data yet',
          description: 'Connect your Shopify app to see reviews.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<ReviewsProvider>();

    if (provider.isServiceUnavailable) {
      return LgPage(
        title: 'Reviews',
        breadcrumb: 'Reports › Retention & Risk',
        backAction: () => context.go('/reports'),
        child: LgServiceUnavailable(onRetry: retryLoad),
      );
    }

    if (provider.error != null) {
      return LgPage(
        title: 'Reviews',
        breadcrumb: 'Reports › Retention & Risk',
        backAction: () => context.go('/reports'),
        child: LgErrorState(message: provider.error!, onRetry: retryLoad),
      );
    }

    if (provider.isLoading && provider.report == null) {
      return LgPage(
        title: 'Reviews',
        breadcrumb: 'Reports › Retention & Risk',
        backAction: () => context.go('/reports'),
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    final report = provider.report ?? ReviewsReport.empty();
    final appsList = appsProvider.apps;
    final showAppFilter = appsList.isNotEmpty;
    final hasData = report.totalReviews > 0 || report.recent.isNotEmpty;

    return LgPage(
      title: 'Reviews',
      breadcrumb: 'Reports › Retention & Risk',
      subtitle:
          'App Store ratings — distribution, average, and most recent feedback',
      backAction: () => context.go('/reports'),
      onRefresh: refreshData,
      secondaryActions: [
        LgPageAction(label: 'Export CSV', onPressed: _exportCsv),
      ],
      // Fixed: app-selector + KPI hero. Scrollable: distribution + reviews.
      pinned: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (showAppFilter) ...[
            PopupMenuButton<String>(
              onSelected: provider.setSelectedApp,
              itemBuilder: (_) => appsList
                  .map(
                    (app) =>
                        PopupMenuItem(value: app.id, child: Text(app.name)),
                  )
                  .toList(),
              child: Chip(
                label: Text(
                  appsList
                      .firstWhere(
                        (a) => a.id == provider.selectedAppId,
                        orElse: () => appsList.first,
                      )
                      .name,
                ),
              ),
            ),
            if (hasData) const SizedBox(height: LgSpacing.s300),
          ],
          if (hasData) _Hero(report: report),
        ],
      ),
      child: !hasData
          ? const LgEmptyState(
              icon: Icons.star_outline_rounded,
              heading: 'No reviews yet',
              description:
                  'Once merchants leave App Store reviews for this app, the average rating, rating distribution, and most recent feedback will appear here.',
            )
          : Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                if (LgBreakpoints.isMobile(context)) ...[
                  _DistributionCard(report: report),
                  const SizedBox(height: LgSpacing.s600),
                  _RecentReviews(report: report),
                ] else
                  Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Expanded(child: _DistributionCard(report: report)),
                      const SizedBox(width: LgSpacing.s600),
                      Expanded(child: _RecentReviews(report: report)),
                    ],
                  ),
              ],
            ),
    );
  }
}

// ─── Hero: big average rating + total ───────────────────────────────
class _Hero extends StatelessWidget {
  final ReviewsReport report;
  const _Hero({required this.report});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final avg = report.avgRating;

    final content = Row(
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        Text(
          '${avg.toStringAsFixed(1)} ★',
          style: theme.textTheme.displaySmall?.copyWith(
            color: _starColor,
            fontWeight: FontWeight.w700,
          ),
        ),
        const SizedBox(width: LgSpacing.s400),
        Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            Row(
              children: List.generate(
                5,
                (i) => Icon(
                  i < avg.round()
                      ? Icons.star_rounded
                      : Icons.star_outline_rounded,
                  size: 22,
                  color: _starColor,
                ),
              ),
            ),
            const SizedBox(height: LgSpacing.s100),
            Text(
              '${report.totalReviews} review${report.totalReviews == 1 ? '' : 's'}',
              style: theme.textTheme.bodyMedium?.copyWith(
                color: LgColors.textSecondary,
              ),
            ),
          ],
        ),
      ],
    );

    return LgCard(child: content);
  }
}

// ─── Rating distribution (5★..1★ bars) ──────────────────────────────
class _DistributionCard extends StatelessWidget {
  final ReviewsReport report;
  const _DistributionCard({required this.report});

  @override
  Widget build(BuildContext context) {
    // Normalize into a 5..1 ordered map, defaulting missing buckets to 0.
    final counts = <int, int>{for (var star = 1; star <= 5; star++) star: 0};
    for (final b in report.distribution) {
      if (b.rating >= 1 && b.rating <= 5) counts[b.rating] = b.count;
    }
    final maxCount = counts.values.fold<int>(0, (a, b) => a > b ? a : b);

    return LgCard(
      title: 'Rating distribution',
      child: Column(
        children: [
          for (var star = 5; star >= 1; star--) ...[
            if (star < 5) const SizedBox(height: LgSpacing.s200),
            _DistRow(star: star, count: counts[star] ?? 0, maxCount: maxCount),
          ],
        ],
      ),
    );
  }
}

class _DistRow extends StatelessWidget {
  final int star;
  final int count;
  final int maxCount;
  const _DistRow({
    required this.star,
    required this.count,
    required this.maxCount,
  });

  // Severity color per the wireframe: 4–5★ green, 3★ amber, 1–2★ red.
  Color get _barColor => star >= 4
      ? LgColors.success
      : (star == 3 ? LgColors.warning : LgColors.critical);

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final pct = maxCount > 0 ? count / maxCount : 0.0;

    return Row(
      children: [
        SizedBox(
          width: 32,
          child: Row(
            children: [
              Text('$star', style: theme.textTheme.bodySmall),
              const Icon(Icons.star_rounded, size: 12, color: _starColor),
            ],
          ),
        ),
        const SizedBox(width: LgSpacing.s200),
        Expanded(
          child: ClipRRect(
            borderRadius: BorderRadius.circular(3),
            child: LinearProgressIndicator(
              value: pct,
              minHeight: 10,
              backgroundColor: LgColors.surfaceSecondary,
              color: _barColor,
            ),
          ),
        ),
        const SizedBox(width: LgSpacing.s300),
        SizedBox(
          width: 40,
          child: Text(
            '$count',
            textAlign: TextAlign.right,
            style: theme.textTheme.bodySmall?.copyWith(
              color: LgColors.textSecondary,
            ),
          ),
        ),
      ],
    );
  }
}

// ─── Recent reviews list ────────────────────────────────────────────
class _RecentReviews extends StatelessWidget {
  final ReviewsReport report;
  const _RecentReviews({required this.report});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final reviews = report.recent;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Recent Reviews', style: theme.textTheme.titleMedium),
        const SizedBox(height: LgSpacing.s300),
        if (reviews.isEmpty)
          Text(
            'No recent reviews to show.',
            style: theme.textTheme.bodySmall?.copyWith(
              color: LgColors.textSecondary,
            ),
          )
        else
          ...reviews.map(
            (r) => Padding(
              padding: const EdgeInsets.only(bottom: LgSpacing.s300),
              child: _ReviewRow(review: r),
            ),
          ),
      ],
    );
  }
}

class _ReviewRow extends StatelessWidget {
  final AppReview review;
  const _ReviewRow({required this.review});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final dateFmt = DateFormat('MMM d, y');
    final maxLines = LgBreakpoints.isMobile(context) ? 4 : 3;

    return LgCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              ...List.generate(
                5,
                (i) => Icon(
                  i < review.rating
                      ? Icons.star_rounded
                      : Icons.star_outline_rounded,
                  size: 16,
                  color: _starColor,
                ),
              ),
              const Spacer(),
              Text(
                dateFmt.format(review.date),
                style: theme.textTheme.bodySmall?.copyWith(
                  color: LgColors.textSecondary,
                ),
              ),
            ],
          ),
          const SizedBox(height: LgSpacing.s200),
          Text(
            review.text,
            style: theme.textTheme.bodyMedium,
            maxLines: maxLines,
            overflow: TextOverflow.ellipsis,
          ),
          const SizedBox(height: LgSpacing.s200),
          Row(
            children: [
              Text(
                review.author,
                style: theme.textTheme.bodySmall?.copyWith(
                  fontWeight: FontWeight.w600,
                  color: LgColors.textPrimary,
                ),
              ),
              if (review.location.isNotEmpty) ...[
                const SizedBox(width: LgSpacing.s200),
                const Icon(
                  Icons.location_on_outlined,
                  size: 13,
                  color: LgColors.textDisabled,
                ),
                const SizedBox(width: 2),
                Flexible(
                  child: Text(
                    review.location,
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: LgColors.textSecondary,
                    ),
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ],
            ],
          ),
        ],
      ),
    );
  }
}
