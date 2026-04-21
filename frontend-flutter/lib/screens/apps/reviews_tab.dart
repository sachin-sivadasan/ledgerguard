import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../models/review_model.dart';
import '../../providers/apps_provider.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';

const _starColor = LgColors.starRating;

enum _SortMode { newest, oldest, lowestRating }

class ReviewsTab extends StatefulWidget {
  const ReviewsTab({super.key});

  @override
  State<ReviewsTab> createState() => _ReviewsTabState();
}

class _ReviewsTabState extends State<ReviewsTab> {
  int? _ratingFilter; // null = All
  String? _appFilter; // null = All apps
  _SortMode _sort = _SortMode.newest;
  final Set<String> _expandedReviews = {};
  String? _hoveredCardId;

  bool get _hasActiveFilters => _appFilter != null || _ratingFilter != null;

  void _clearFilters() {
    setState(() {
      _appFilter = null;
      _ratingFilter = null;
    });
  }

  List<AppReview> _sortReviews(List<AppReview> list) {
    final sorted = List<AppReview>.from(list);
    switch (_sort) {
      case _SortMode.newest:
        sorted.sort((a, b) => b.date.compareTo(a.date));
      case _SortMode.oldest:
        sorted.sort((a, b) => a.date.compareTo(b.date));
      case _SortMode.lowestRating:
        sorted.sort((a, b) => a.rating != b.rating
            ? a.rating.compareTo(b.rating)
            : b.date.compareTo(a.date));
    }
    return sorted;
  }

  @override
  Widget build(BuildContext context) {
    final provider = context.watch<AppsProvider>();
    final theme = Theme.of(context);
    final dateFmt = DateFormat('MMM d, y');

    final appsWithReviews = provider.apps.where((app) {
      return provider.getReviewsForApp(app.id).isNotEmpty;
    }).toList();

    // Base reviews (app-filtered)
    var baseReviews = _appFilter != null
        ? provider.getReviewsForApp(_appFilter!)
        : provider.allReviews;

    // Then rating filter
    var filtered = _ratingFilter != null
        ? baseReviews.where((r) => r.rating == _ratingFilter).toList()
        : baseReviews;

    // Sort
    final reviews = _sortReviews(filtered);

    return ListView(
      padding: const EdgeInsets.only(bottom: LgSpacing.s800),
      children: [
        // ── Compact summary cards ──
        if (appsWithReviews.isNotEmpty)
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: LgSpacing.s200),
            child: Row(
              children: appsWithReviews.map((app) {
                return Expanded(
                  child: Padding(
                    padding: const EdgeInsets.symmetric(horizontal: LgSpacing.s200),
                    child: _buildCompactCard(context, provider, app.id, app.name),
                  ),
                );
              }).toList(),
            ),
          ),

        const SizedBox(height: LgSpacing.s400),

        // ── Rating filter chips ──
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: LgSpacing.s400),
          child: Wrap(
            spacing: LgSpacing.s200,
            children: [
              _filterChip(null, 'All', baseReviews.length),
              _filterChip(5, '5 stars', baseReviews.where((r) => r.rating == 5).length),
              _filterChip(4, '4 stars', baseReviews.where((r) => r.rating == 4).length),
              _filterChip(3, '3 stars', baseReviews.where((r) => r.rating == 3).length),
              _filterChip(2, '2 stars', baseReviews.where((r) => r.rating == 2).length),
              _filterChip(1, '1 star', baseReviews.where((r) => r.rating == 1).length),
            ],
          ),
        ),

        const SizedBox(height: LgSpacing.s300),

        // ── Review count + sort + clear filters ──
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: LgSpacing.s400),
          child: Row(
            children: [
              Text(
                '${reviews.length} review${reviews.length == 1 ? '' : 's'}'
                '${_appFilter != null ? ' for ${appsWithReviews.where((a) => a.id == _appFilter).firstOrNull?.name ?? ''}' : ''}',
                style: theme.textTheme.bodySmall?.copyWith(color: LgColors.textSecondary),
              ),
              if (_hasActiveFilters) ...[
                const SizedBox(width: LgSpacing.s200),
                GestureDetector(
                  onTap: _clearFilters,
                  child: MouseRegion(
                    cursor: SystemMouseCursors.click,
                    child: Text('Clear filters',
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: LgColors.primary,
                          fontWeight: FontWeight.w600,
                        )),
                  ),
                ),
              ],
              const Spacer(),
              // Sort dropdown
              SizedBox(
                height: 28,
                child: DropdownButtonHideUnderline(
                  child: DropdownButton<_SortMode>(
                    value: _sort,
                    isDense: true,
                    icon: const Icon(Icons.sort, size: 14),
                    style: theme.textTheme.bodySmall?.copyWith(color: LgColors.textSecondary),
                    items: const [
                      DropdownMenuItem(value: _SortMode.newest, child: Text('Newest')),
                      DropdownMenuItem(value: _SortMode.oldest, child: Text('Oldest')),
                      DropdownMenuItem(value: _SortMode.lowestRating, child: Text('Lowest rated')),
                    ],
                    onChanged: (v) => setState(() => _sort = v!),
                  ),
                ),
              ),
            ],
          ),
        ),

        const SizedBox(height: LgSpacing.s200),

        // ── Empty filter state ──
        if (reviews.isEmpty)
          Padding(
            padding: const EdgeInsets.symmetric(vertical: LgSpacing.s800),
            child: Center(
              child: Column(
                children: [
                  Icon(Icons.filter_list_off, size: 40, color: LgColors.textDisabled),
                  const SizedBox(height: LgSpacing.s300),
                  Text('No reviews match this filter',
                      style: theme.textTheme.bodyMedium?.copyWith(color: LgColors.textSecondary)),
                ],
              ),
            ),
          ),

        // ── Review list (max-width constrained) ──
        ...reviews.map((review) {
          final appName = _appFilter == null
              ? provider.apps.where((a) => a.id == review.appId).firstOrNull?.name
              : null;
          final isExpanded = _expandedReviews.contains(review.id);
          final needsClamp = review.text.length > 200;

          return ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 800),
              child: Padding(
                padding: const EdgeInsets.only(
                  left: LgSpacing.s200,
                  right: LgSpacing.s200,
                  bottom: LgSpacing.s300,
                ),
                child: LgCard(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          ...List.generate(5, (i) => Icon(
                                i < review.rating ? Icons.star_rounded : Icons.star_outline_rounded,
                                size: 16, color: _starColor,
                              )),
                          if (appName != null) ...[
                            const SizedBox(width: LgSpacing.s200),
                            Container(
                              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 1),
                              decoration: BoxDecoration(
                                color: LgColors.surfaceSecondary,
                                borderRadius: BorderRadius.circular(4),
                              ),
                              child: Text(appName,
                                  style: theme.textTheme.bodySmall?.copyWith(
                                    fontSize: 10,
                                    color: LgColors.textSecondary,
                                  )),
                            ),
                          ],
                          const Spacer(),
                          Text(dateFmt.format(review.date),
                              style: theme.textTheme.bodySmall?.copyWith(
                                color: LgColors.textSecondary,
                              )),
                        ],
                      ),
                      const SizedBox(height: LgSpacing.s300),
                      // Review text with clamp
                      Text(
                        review.text,
                        style: theme.textTheme.bodyMedium,
                        maxLines: isExpanded || !needsClamp ? null : 3,
                        overflow: isExpanded || !needsClamp ? null : TextOverflow.ellipsis,
                      ),
                      if (needsClamp)
                        GestureDetector(
                          onTap: () => setState(() {
                            if (isExpanded) {
                              _expandedReviews.remove(review.id);
                            } else {
                              _expandedReviews.add(review.id);
                            }
                          }),
                          child: MouseRegion(
                            cursor: SystemMouseCursors.click,
                            child: Padding(
                              padding: const EdgeInsets.only(top: 4),
                              child: Text(
                                isExpanded ? 'Show less' : 'Read more',
                                style: theme.textTheme.bodySmall?.copyWith(
                                  color: LgColors.primary,
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                            ),
                          ),
                        ),
                      const SizedBox(height: LgSpacing.s300),
                      Row(
                        children: [
                          Text(review.author,
                              style: theme.textTheme.bodySmall?.copyWith(
                                fontWeight: FontWeight.w600,
                                color: LgColors.textPrimary,
                              )),
                          if (review.location.isNotEmpty) ...[
                            const SizedBox(width: LgSpacing.s200),
                            Icon(Icons.location_on_outlined, size: 13, color: LgColors.textDisabled),
                            const SizedBox(width: 2),
                            Flexible(
                              child: Text(review.location,
                                  style: theme.textTheme.bodySmall?.copyWith(
                                    color: LgColors.textSecondary,
                                  ),
                                  overflow: TextOverflow.ellipsis),
                            ),
                          ],
                        ],
                      ),
                      if (review.timeUsing.isNotEmpty)
                        Padding(
                          padding: const EdgeInsets.only(top: 2),
                          child: Text(review.timeUsing,
                              style: theme.textTheme.bodySmall?.copyWith(
                                color: LgColors.textDisabled,
                                fontSize: 11,
                              )),
                        ),
                    ],
                  ),
                ),
              ),
          );
        }),
      ],
    );
  }

  Widget _buildCompactCard(BuildContext context, AppsProvider provider, String appId, String appName) {
    final dist = provider.ratingDistribution(appId);
    final avg = provider.avgRatingForApp(appId);
    final total = dist.values.fold<int>(0, (a, b) => a + b);
    final theme = Theme.of(context);
    final isSelected = _appFilter == appId;

    final isHovered = _hoveredCardId == appId;

    return MouseRegion(
      cursor: SystemMouseCursors.click,
      onEnter: (_) => setState(() => _hoveredCardId = appId),
      onExit: (_) => setState(() => _hoveredCardId = null),
      child: GestureDetector(
        onTap: () => setState(() {
          _appFilter = isSelected ? null : appId;
          _ratingFilter = null;
        }),
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 200),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(12),
            border: isSelected
                ? Border.all(color: LgColors.primary, width: 2)
                : Border.all(color: isHovered ? LgColors.border : Colors.transparent, width: 2),
          ),
          child: LgCard(
            title: appName,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  crossAxisAlignment: CrossAxisAlignment.baseline,
                  textBaseline: TextBaseline.alphabetic,
                  children: [
                    Text(avg.toStringAsFixed(1),
                        style: const TextStyle(fontSize: 28, fontWeight: FontWeight.w700)),
                    const SizedBox(width: LgSpacing.s200),
                    ...List.generate(5, (i) => Icon(
                          i < avg.round() ? Icons.star_rounded : Icons.star_outline_rounded,
                          size: 18, color: _starColor,
                        )),
                    const SizedBox(width: LgSpacing.s200),
                    Text('$total reviews', style: theme.textTheme.bodySmall?.copyWith(
                      color: LgColors.textSecondary,
                    )),
                  ],
                ),
                const SizedBox(height: LgSpacing.s300),
                ...List.generate(5, (i) {
                  final star = 5 - i;
                  final count = dist[star] ?? 0;
                  final pct = total > 0 ? count / total : 0.0;
                  return Padding(
                    padding: const EdgeInsets.only(bottom: 3),
                    child: Row(
                      children: [
                        SizedBox(width: 12, child: Text('$star', style: const TextStyle(fontSize: 12, color: LgColors.textSecondary))),
                        const SizedBox(width: 6),
                        Expanded(
                          child: ClipRRect(
                            borderRadius: BorderRadius.circular(2),
                            child: LinearProgressIndicator(
                              value: pct,
                              minHeight: 8,
                              backgroundColor: LgColors.surfaceSecondary,
                              color: _starColor,
                            ),
                          ),
                        ),
                        const SizedBox(width: 8),
                        SizedBox(width: 24, child: Text('$count',
                            textAlign: TextAlign.right,
                            style: const TextStyle(fontSize: 12, color: LgColors.textSecondary))),
                      ],
                    ),
                  );
                }),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _filterChip(int? rating, String label, int count) {
    final isActive = _ratingFilter == rating;
    if (count == 0 && rating != null) return const SizedBox.shrink();

    return FilterChip(
      selected: isActive,
      label: Text('$label ($count)'),
      labelStyle: TextStyle(
        fontSize: 12,
        fontWeight: isActive ? FontWeight.w600 : FontWeight.w400,
        color: isActive ? Colors.white : LgColors.textPrimary,
      ),
      selectedColor: LgColors.primary,
      backgroundColor: LgColors.surfaceSecondary,
      side: BorderSide.none,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
      showCheckmark: false,
      padding: const EdgeInsets.symmetric(horizontal: 4),
      onSelected: (_) => setState(() => _ratingFilter = isActive ? null : rating),
    );
  }
}
