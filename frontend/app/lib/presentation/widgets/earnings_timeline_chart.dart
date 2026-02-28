import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../core/theme/app_theme.dart';
import '../../domain/entities/earnings_timeline.dart';
import '../../domain/entities/time_range.dart';
import '../blocs/earnings/earnings.dart';

/// Earnings Timeline chart widget displaying daily earnings as a bar chart
class EarningsTimelineChart extends StatefulWidget {
  final TimeRange timeRange;

  const EarningsTimelineChart({
    super.key,
    required this.timeRange,
  });

  @override
  State<EarningsTimelineChart> createState() => _EarningsTimelineChartState();
}

class _EarningsTimelineChartState extends State<EarningsTimelineChart> {
  TimeRange? _lastLoadedRange;
  EarningsLoaded? _lastLoadedState;

  @override
  void initState() {
    super.initState();
    _loadDataIfNeeded();
  }

  @override
  void didUpdateWidget(EarningsTimelineChart oldWidget) {
    super.didUpdateWidget(oldWidget);
    _loadDataIfNeeded();
  }

  void _loadDataIfNeeded() {
    // Check if timeRange changed since last load
    if (_lastLoadedRange != widget.timeRange) {
      final isFirstLoad = _lastLoadedRange == null;
      _lastLoadedRange = widget.timeRange;
      if (isFirstLoad) {
        context.read<EarningsBloc>().add(LoadEarningsRequested(widget.timeRange));
      } else {
        context.read<EarningsBloc>().add(EarningsTimeRangeChanged(widget.timeRange));
      }
    }
  }

  void _loadData() {
    _lastLoadedRange = widget.timeRange;
    context.read<EarningsBloc>().add(LoadEarningsRequested(widget.timeRange));
  }

  @override
  Widget build(BuildContext context) {
    // Also check in build in case widget was recreated
    if (_lastLoadedRange != widget.timeRange) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted) _loadDataIfNeeded();
      });
    }

    return BlocBuilder<EarningsBloc, EarningsState>(
      builder: (context, state) {
        // Cache the last loaded state for smooth transitions
        if (state is EarningsLoaded) {
          _lastLoadedState = state;
        }

        if (state is EarningsInitial) {
          return _buildLoadingState(state);
        }

        // Show previous chart with loading overlay during loading
        if (state is EarningsLoading) {
          if (_lastLoadedState != null) {
            return _buildChartWithLoading(context, _lastLoadedState!, state);
          }
          return _buildLoadingState(state);
        }

        if (state is EarningsEmpty) {
          return _buildEmptyState(context, state);
        }

        if (state is EarningsError) {
          return _buildErrorState(context, state);
        }

        if (state is EarningsLoaded) {
          return _buildChart(context, state, isLoading: false);
        }

        return const SizedBox.shrink();
      },
    );
  }

  Widget _buildChartWithLoading(BuildContext context, EarningsLoaded loadedState, EarningsLoading loadingState) {
    return _buildChart(
      context,
      loadedState.copyWith(
        year: loadingState.year,
        month: loadingState.month,
        canGoNext: false,
        canGoPrevious: false,
      ),
      isLoading: true,
    );
  }

  Widget _buildLoadingState(EarningsState state) {
    int year = DateTime.now().year;
    int month = DateTime.now().month;
    if (state is EarningsLoading) {
      year = state.year;
      month = state.month;
    }

    return Card(
      key: const ValueKey('earnings-card'),
      child: Container(
        height: 300,
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            _buildHeader(context, year, month, false, false, EarningsMode.combined),
            const Expanded(
              child: Center(child: CircularProgressIndicator()),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildEmptyState(BuildContext context, EarningsEmpty state) {
    return Card(
      key: const ValueKey('earnings-card'),
      child: Container(
        height: 300,
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            _buildHeader(context, state.year, state.month, state.canGoNext, state.canGoPrevious, EarningsMode.combined),
            const Expanded(
              child: Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(Icons.bar_chart, size: 48, color: Colors.grey),
                    SizedBox(height: 8),
                    Text(
                      'No earnings data for this month',
                      style: TextStyle(color: Colors.grey),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildErrorState(BuildContext context, EarningsError state) {
    return Card(
      key: const ValueKey('earnings-card'),
      child: Container(
        height: 300,
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            _buildHeader(context, state.year, state.month, false, true, EarningsMode.combined),
            Expanded(
              child: Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    const Icon(Icons.error_outline, size: 48, color: Colors.red),
                    const SizedBox(height: 8),
                    Text(
                      state.message,
                      style: const TextStyle(color: Colors.red),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 16),
                    TextButton.icon(
                      onPressed: _loadData,
                      icon: const Icon(Icons.refresh),
                      label: const Text('Retry'),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildChart(BuildContext context, EarningsLoaded state, {bool isLoading = false}) {
    final timeline = state.timeline;
    final isSplitMode = state.mode == EarningsMode.split;

    return Card(
      key: const ValueKey('earnings-card'),
      child: Container(
        height: 300,
        padding: const EdgeInsets.all(16),
        child: Stack(
          children: [
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _buildHeader(context, state.year, state.month, state.canGoNext, state.canGoPrevious, state.mode),
                const SizedBox(height: 12),
                _buildTotalSummary(context, timeline, isSplitMode),
                const SizedBox(height: 16),
                Expanded(
                  child: timeline.earnings.isEmpty
                      ? const Center(child: Text('No data'))
                      : _buildBarChart(context, timeline, isSplitMode),
                ),
              ],
            ),
            if (isLoading)
              Positioned.fill(
                child: Container(
                  color: Colors.white.withOpacity(0.7),
                  child: const Center(
                    child: CircularProgressIndicator(),
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildHeader(
    BuildContext context,
    int year,
    int month,
    bool canGoNext,
    bool canGoPrevious,
    EarningsMode currentMode,
  ) {
    const months = [
      'January', 'February', 'March', 'April', 'May', 'June',
      'July', 'August', 'September', 'October', 'November', 'December',
    ];
    final monthName = months[month - 1];

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Expanded(
              child: Text(
                '$monthName $year',
                style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
              ),
            ),
            IconButton(
              icon: const Icon(Icons.chevron_left),
              onPressed: canGoPrevious
                  ? () => context.read<EarningsBloc>().add(const PreviousMonthRequested())
                  : null,
              tooltip: 'Previous month',
            ),
            IconButton(
              icon: const Icon(Icons.chevron_right),
              onPressed: canGoNext
                  ? () => context.read<EarningsBloc>().add(const NextMonthRequested())
                  : null,
              tooltip: 'Next month',
            ),
          ],
        ),
        const SizedBox(height: 8),
        _buildModeToggle(context, currentMode),
      ],
    );
  }

  Widget _buildModeToggle(BuildContext context, EarningsMode currentMode) {
    return Row(
      children: [
        const Text('View:', style: TextStyle(fontSize: 12)),
        const SizedBox(width: 8),
        ChoiceChip(
          label: const Text('Combined'),
          selected: currentMode == EarningsMode.combined,
          onSelected: (selected) {
            if (selected) {
              context.read<EarningsBloc>().add(const EarningsModeChanged(EarningsMode.combined));
            }
          },
          visualDensity: VisualDensity.compact,
        ),
        const SizedBox(width: 8),
        ChoiceChip(
          label: const Text('Split'),
          selected: currentMode == EarningsMode.split,
          onSelected: (selected) {
            if (selected) {
              context.read<EarningsBloc>().add(const EarningsModeChanged(EarningsMode.split));
            }
          },
          visualDensity: VisualDensity.compact,
        ),
      ],
    );
  }

  Widget _buildTotalSummary(
    BuildContext context,
    EarningsTimeline timeline,
    bool isSplitMode,
  ) {
    return Row(
      children: [
        Expanded(
          child: _buildSummaryItem(
            context,
            'Total',
            timeline.formattedTotal,
            AppTheme.primary,
          ),
        ),
        if (isSplitMode) ...[
          const SizedBox(width: 16),
          Expanded(
            child: _buildSummaryItem(
              context,
              'Subscription',
              _formatCurrency(timeline.totalSubscription),
              AppTheme.success,
            ),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: _buildSummaryItem(
              context,
              'Usage',
              _formatCurrency(timeline.totalUsage),
              AppTheme.secondary,
            ),
          ),
        ],
      ],
    );
  }

  Widget _buildSummaryItem(
    BuildContext context,
    String label,
    String value,
    Color color,
  ) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Container(
              width: 8,
              height: 8,
              decoration: BoxDecoration(
                color: color,
                shape: BoxShape.circle,
              ),
            ),
            const SizedBox(width: 4),
            Text(
              label,
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                    color: Colors.grey[600],
                  ),
            ),
          ],
        ),
        const SizedBox(height: 2),
        Text(
          value,
          style: Theme.of(context).textTheme.titleSmall?.copyWith(
                fontWeight: FontWeight.bold,
              ),
        ),
      ],
    );
  }

  Widget _buildBarChart(
    BuildContext context,
    EarningsTimeline timeline,
    bool isSplitMode,
  ) {
    final maxY = (timeline.maxDailyTotal / 100).ceilToDouble() + 10;

    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: SizedBox(
        width: timeline.earnings.length * 20.0 + 40,
        child: BarChart(
          BarChartData(
            alignment: BarChartAlignment.spaceAround,
            maxY: maxY,
            minY: 0,
            barTouchData: BarTouchData(
              touchTooltipData: BarTouchTooltipData(
                fitInsideHorizontally: true,
                fitInsideVertically: true,
                getTooltipItem: (group, groupIndex, rod, rodIndex) {
                  final entry = timeline.earnings[groupIndex];
                  String text;
                  if (isSplitMode) {
                    if (rodIndex == 0) {
                      text = 'Subscription: ${entry.formattedSubscription}';
                    } else {
                      text = 'Usage: ${entry.formattedUsage}';
                    }
                  } else {
                    text = 'Total: ${entry.formattedTotal}';
                  }
                  return BarTooltipItem(
                    '${entry.date}\n$text',
                    const TextStyle(
                      color: Colors.white,
                      fontWeight: FontWeight.bold,
                      fontSize: 12,
                    ),
                  );
                },
              ),
            ),
            titlesData: FlTitlesData(
              show: true,
              rightTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
              topTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
              bottomTitles: AxisTitles(
                sideTitles: SideTitles(
                  showTitles: true,
                  getTitlesWidget: (value, meta) {
                    final index = value.toInt();
                    if (index >= 0 && index < timeline.earnings.length) {
                      final day = timeline.earnings[index].dayOfMonth;
                      if (day == 1 || day % 5 == 0 || index == timeline.earnings.length - 1) {
                        return Padding(
                          padding: const EdgeInsets.only(top: 8),
                          child: Text(
                            '$day',
                            style: const TextStyle(fontSize: 10, color: Colors.grey),
                          ),
                        );
                      }
                    }
                    return const SizedBox.shrink();
                  },
                  reservedSize: 28,
                ),
              ),
              leftTitles: AxisTitles(
                sideTitles: SideTitles(
                  showTitles: true,
                  reservedSize: 40,
                  getTitlesWidget: (value, meta) {
                    if (value == 0) return const SizedBox.shrink();
                    return Text(
                      '\$${value.toInt()}',
                      style: const TextStyle(fontSize: 10, color: Colors.grey),
                    );
                  },
                ),
              ),
            ),
            borderData: FlBorderData(show: false),
            gridData: FlGridData(
              show: true,
              drawVerticalLine: false,
              horizontalInterval: maxY / 4,
              getDrawingHorizontalLine: (value) {
                return FlLine(color: Colors.grey[200]!, strokeWidth: 1);
              },
            ),
            barGroups: _buildBarGroups(timeline, isSplitMode),
          ),
          swapAnimationDuration: const Duration(milliseconds: 250),
          swapAnimationCurve: Curves.easeInOut,
        ),
      ),
    );
  }

  List<BarChartGroupData> _buildBarGroups(
    EarningsTimeline timeline,
    bool isSplitMode,
  ) {
    return timeline.earnings.asMap().entries.map((entry) {
      final index = entry.key;
      final data = entry.value;

      if (isSplitMode) {
        return BarChartGroupData(
          x: index,
          barRods: [
            BarChartRodData(
              toY: data.subscriptionAmountCents / 100,
              color: AppTheme.success,
              width: 12,
              borderRadius: const BorderRadius.only(
                topLeft: Radius.circular(4),
                topRight: Radius.circular(4),
              ),
            ),
            BarChartRodData(
              toY: data.usageAmountCents / 100,
              color: AppTheme.secondary,
              width: 12,
              borderRadius: const BorderRadius.only(
                topLeft: Radius.circular(4),
                topRight: Radius.circular(4),
              ),
            ),
          ],
        );
      } else {
        return BarChartGroupData(
          x: index,
          barRods: [
            BarChartRodData(
              toY: data.totalAmountCents / 100,
              color: AppTheme.primary,
              width: 12,
              borderRadius: const BorderRadius.only(
                topLeft: Radius.circular(4),
                topRight: Radius.circular(4),
              ),
            ),
          ],
        );
      }
    }).toList();
  }

  String _formatCurrency(int cents) {
    final dollars = cents / 100;
    if (dollars >= 1000000) {
      return '\$${(dollars / 1000000).toStringAsFixed(2)}M';
    } else if (dollars >= 1000) {
      return '\$${(dollars / 1000).toStringAsFixed(1)}K';
    }
    return '\$${dollars.toStringAsFixed(2)}';
  }
}
