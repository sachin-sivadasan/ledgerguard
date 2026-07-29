import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import '../../core/mixins/data_loading_mixin.dart';
import '../../core/utils/file_download.dart';
import '../../providers/apps_provider.dart';
import '../../providers/cohorts_provider.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_cohort_heatmap.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_error_state.dart';
import '../../widgets/lg_page.dart';
import '../../widgets/lg_service_unavailable.dart';

class CohortsReportScreen extends StatefulWidget {
  const CohortsReportScreen({super.key});

  @override
  State<CohortsReportScreen> createState() => _CohortsReportScreenState();
}

class _CohortsReportScreenState extends State<CohortsReportScreen>
    with DataLoadingMixin {
  @override
  void loadData(String appId) {
    context.read<CohortsProvider>().setSelectedApp(appId);
  }

  Future<void> _exportCsv() async {
    final messenger = ScaffoldMessenger.of(context);
    final provider = context.read<CohortsProvider>();
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
          'cohorts-${DateTime.now().toIso8601String().split('T').first}.csv';
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
      debugPrint('cohorts: CSV export failed: $e');
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
        title: 'Retention Cohorts',
        breadcrumb: 'Reports › Retention & Risk',
        backAction: () => context.go('/reports'),
        child: LgEmptyState(
          icon: Icons.group_work,
          heading: 'No data yet',
          description: 'Connect your Shopify app to see cohort retention.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<CohortsProvider>();

    if (provider.isServiceUnavailable) {
      return LgPage(
        title: 'Retention Cohorts',
        breadcrumb: 'Reports › Retention & Risk',
        backAction: () => context.go('/reports'),
        child: LgServiceUnavailable(onRetry: retryLoad),
      );
    }

    if (provider.error != null) {
      return LgPage(
        title: 'Retention Cohorts',
        breadcrumb: 'Reports › Retention & Risk',
        backAction: () => context.go('/reports'),
        child: LgErrorState(message: provider.error!, onRetry: retryLoad),
      );
    }

    if (provider.isLoading && provider.cohorts.isEmpty) {
      return LgPage(
        title: 'Retention Cohorts',
        breadcrumb: 'Reports › Retention & Risk',
        backAction: () => context.go('/reports'),
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    final cohorts = provider.cohorts;
    final appsList = appsProvider.apps;
    final showAppFilter = appsList.isNotEmpty;

    return LgPage(
      title: 'Retention Cohorts',
      breadcrumb: 'Reports › Retention & Risk',
      subtitle: '% of each signup-month cohort still retained N months later',
      backAction: () => context.go('/reports'),
      onRefresh: refreshData,
      secondaryActions: [
        LgPageAction(label: 'Export CSV', onPressed: _exportCsv),
      ],
      child: Column(
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
            const SizedBox(height: LgSpacing.s300),
          ],
          if (cohorts.isEmpty)
            const LgEmptyState(
              icon: Icons.group_work,
              heading: 'Cohort data not yet available',
              description:
                  'Cohort retention analysis requires at least two months of subscription data.',
            )
          else
            LgCard(
              title: 'Cohort Retention Heatmap',
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'rows = signup month · columns = months since signup · cell = % retained',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: LgColors.textSecondary,
                    ),
                  ),
                  const SizedBox(height: LgSpacing.s300),
                  CohortHeatmap(cohorts: cohorts),
                ],
              ),
            ),
        ],
      ),
    );
  }
}
