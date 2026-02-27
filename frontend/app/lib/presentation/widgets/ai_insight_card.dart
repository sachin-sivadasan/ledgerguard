import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../core/theme/app_theme.dart';
import '../../domain/entities/daily_insight.dart';
import '../blocs/insight/insight.dart';

/// Collapsible card displaying AI-generated daily insights
class AiInsightCard extends StatefulWidget {
  const AiInsightCard({super.key});

  @override
  State<AiInsightCard> createState() => _AiInsightCardState();
}

class _AiInsightCardState extends State<AiInsightCard> {
  bool _isExpanded = true;

  @override
  void initState() {
    super.initState();
    // Load insight on init
    context.read<InsightBloc>().add(const LoadInsightRequested());
  }

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<InsightBloc, InsightState>(
      builder: (context, state) {
        // Don't show anything for empty or error states
        if (state is InsightEmpty || state is InsightError) {
          return const SizedBox.shrink();
        }

        if (state is InsightLoading) {
          return _buildLoadingState(context);
        }

        if (state is InsightLoaded) {
          return _buildCard(context, state.insight, state.isRefreshing);
        }

        // Initial state - show nothing until loaded
        return const SizedBox.shrink();
      },
    );
  }

  Widget _buildLoadingState(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(bottom: 20),
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [
            AppTheme.primary.withOpacity(0.05),
            AppTheme.secondary.withOpacity(0.05),
          ],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: AppTheme.primary.withOpacity(0.2),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: AppTheme.primary.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(
                  Icons.auto_awesome,
                  color: AppTheme.primary,
                  size: 20,
                ),
              ),
              const SizedBox(width: 12),
              Text(
                'Daily Insight',
                style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          _buildShimmer(width: double.infinity, height: 16),
          const SizedBox(height: 8),
          _buildShimmer(width: double.infinity, height: 16),
          const SizedBox(height: 8),
          _buildShimmer(width: 200, height: 16),
        ],
      ),
    );
  }

  Widget _buildShimmer({required double width, required double height}) {
    return Container(
      width: width,
      height: height,
      decoration: BoxDecoration(
        color: Colors.grey[300],
        borderRadius: BorderRadius.circular(4),
      ),
    );
  }

  Widget _buildCard(
      BuildContext context, DailyInsight insight, bool isRefreshing) {
    return Container(
      margin: const EdgeInsets.only(bottom: 20),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [
            AppTheme.primary.withOpacity(0.05),
            AppTheme.secondary.withOpacity(0.05),
          ],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: AppTheme.primary.withOpacity(0.2),
        ),
      ),
      child: Column(
        children: [
          _buildHeader(context, insight, isRefreshing),
          AnimatedCrossFade(
            firstChild: _buildContent(context, insight),
            secondChild: const SizedBox(height: 0),
            crossFadeState: _isExpanded
                ? CrossFadeState.showFirst
                : CrossFadeState.showSecond,
            duration: const Duration(milliseconds: 200),
          ),
        ],
      ),
    );
  }

  Widget _buildHeader(
      BuildContext context, DailyInsight insight, bool isRefreshing) {
    return InkWell(
      onTap: () => setState(() => _isExpanded = !_isExpanded),
      borderRadius: const BorderRadius.vertical(top: Radius.circular(16)),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: AppTheme.primary.withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
              ),
              child: const Icon(
                Icons.auto_awesome,
                color: AppTheme.primary,
                size: 20,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Daily Insight',
                    style: Theme.of(context).textTheme.titleMedium?.copyWith(
                          fontWeight: FontWeight.bold,
                        ),
                  ),
                  Text(
                    'Generated ${insight.formattedGeneratedAt}',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: Colors.grey[600],
                        ),
                  ),
                ],
              ),
            ),
            if (isRefreshing)
              const SizedBox(
                width: 20,
                height: 20,
                child: CircularProgressIndicator(strokeWidth: 2),
              )
            else
              IconButton(
                icon: const Icon(Icons.refresh, size: 20),
                onPressed: () => context
                    .read<InsightBloc>()
                    .add(const RefreshInsightRequested()),
                tooltip: 'Refresh insight',
              ),
            AnimatedRotation(
              turns: _isExpanded ? 0 : -0.25,
              duration: const Duration(milliseconds: 200),
              child: Icon(
                Icons.keyboard_arrow_down,
                color: Colors.grey[600],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildContent(BuildContext context, DailyInsight insight) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Divider(height: 1),
          const SizedBox(height: 16),
          Text(
            insight.summary,
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  height: 1.5,
                ),
          ),
          if (insight.hasKeyPoints) ...[
            const SizedBox(height: 16),
            Text(
              'Key Takeaways',
              style: Theme.of(context).textTheme.titleSmall?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const SizedBox(height: 8),
            ...insight.keyPoints.map((point) => Padding(
                  padding: const EdgeInsets.only(bottom: 6),
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Container(
                        margin: const EdgeInsets.only(top: 6),
                        width: 6,
                        height: 6,
                        decoration: const BoxDecoration(
                          color: AppTheme.primary,
                          shape: BoxShape.circle,
                        ),
                      ),
                      const SizedBox(width: 10),
                      Expanded(
                        child: Text(
                          point,
                          style: Theme.of(context).textTheme.bodyMedium,
                        ),
                      ),
                    ],
                  ),
                )),
          ],
        ],
      ),
    );
  }
}
