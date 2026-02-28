import 'dart:async';

import 'package:flutter/material.dart';

import '../../core/theme/app_theme.dart';
import '../../domain/entities/subscription.dart';
import '../../domain/entities/subscription_filter.dart';

/// Filter bar for subscription list with multi-select status, price, interval, and search
class SubscriptionFilterBar extends StatefulWidget {
  final SubscriptionFilters filters;
  final PriceStats? priceStats;
  final ValueChanged<SubscriptionFilters> onFiltersChanged;
  final VoidCallback onClearFilters;
  final bool isLoading;

  const SubscriptionFilterBar({
    super.key,
    required this.filters,
    this.priceStats,
    required this.onFiltersChanged,
    required this.onClearFilters,
    this.isLoading = false,
  });

  @override
  State<SubscriptionFilterBar> createState() => _SubscriptionFilterBarState();
}

class _SubscriptionFilterBarState extends State<SubscriptionFilterBar> {
  late TextEditingController _searchController;
  Timer? _debounceTimer;

  @override
  void initState() {
    super.initState();
    _searchController = TextEditingController(text: widget.filters.searchQuery ?? '');
  }

  @override
  void didUpdateWidget(covariant SubscriptionFilterBar oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.filters.searchQuery != oldWidget.filters.searchQuery) {
      final newQuery = widget.filters.searchQuery ?? '';
      if (_searchController.text != newQuery) {
        _searchController.text = newQuery;
      }
    }
  }

  @override
  void dispose() {
    _debounceTimer?.cancel();
    _searchController.dispose();
    super.dispose();
  }

  void _onSearchChanged(String value) {
    _debounceTimer?.cancel();
    _debounceTimer = Timer(const Duration(milliseconds: 300), () {
      widget.onFiltersChanged(
        widget.filters.copyWith(
          searchQuery: value.isEmpty ? null : value,
          clearSearchQuery: value.isEmpty,
        ),
      );
    });
  }

  void _toggleRiskState(RiskState state) {
    final newStates = Set<RiskState>.from(widget.filters.riskStates);
    if (newStates.contains(state)) {
      newStates.remove(state);
    } else {
      newStates.add(state);
    }
    widget.onFiltersChanged(widget.filters.copyWith(riskStates: newStates));
  }

  void _setPriceFilter(int? priceCents) {
    if (priceCents == null) {
      widget.onFiltersChanged(
        widget.filters.copyWith(clearPriceFilter: true),
      );
    } else {
      // Filter for exact price match
      widget.onFiltersChanged(
        widget.filters.copyWith(
          priceMinCents: priceCents,
          priceMaxCents: priceCents,
          clearPriceFilter: false,
        ),
      );
    }
  }

  int? _getCurrentPriceFilter() {
    final minCents = widget.filters.priceMinCents;
    final maxCents = widget.filters.priceMaxCents;

    // If min == max, it's an exact price filter
    if (minCents != null && minCents == maxCents) {
      return minCents;
    }
    return null;
  }

  void _setBillingInterval(BillingInterval? interval) {
    widget.onFiltersChanged(
      widget.filters.copyWith(
        billingInterval: interval,
        clearBillingInterval: interval == null,
      ),
    );
  }

  Color _getRiskStateColor(RiskState state) {
    switch (state) {
      case RiskState.safe:
        return AppTheme.success;
      case RiskState.oneCycleMissed:
        return AppTheme.warning;
      case RiskState.twoCyclesMissed:
        return Colors.orange[700]!;
      case RiskState.churned:
        return AppTheme.danger;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border(
          bottom: BorderSide(color: Colors.grey[200]!),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Search row
          Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _searchController,
                  enabled: !widget.isLoading,
                  decoration: InputDecoration(
                    hintText: 'Search by shop name or domain...',
                    prefixIcon: const Icon(Icons.search, size: 20),
                    suffixIcon: _searchController.text.isNotEmpty
                        ? IconButton(
                            icon: const Icon(Icons.clear, size: 18),
                            onPressed: () {
                              _searchController.clear();
                              _onSearchChanged('');
                            },
                          )
                        : null,
                    isDense: true,
                    contentPadding: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 10,
                    ),
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(8),
                      borderSide: BorderSide(color: Colors.grey[300]!),
                    ),
                    enabledBorder: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(8),
                      borderSide: BorderSide(color: Colors.grey[300]!),
                    ),
                  ),
                  onChanged: _onSearchChanged,
                ),
              ),
              if (widget.filters.hasActiveFilters) ...[
                const SizedBox(width: 12),
                _ClearFiltersButton(
                  filterCount: widget.filters.activeFilterCount,
                  onPressed: widget.onClearFilters,
                ),
              ],
            ],
          ),
          const SizedBox(height: 12),
          // Filter chips row
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: Row(
              children: [
                // Risk state chips
                ...RiskState.values.map((state) {
                  final isSelected = widget.filters.riskStates.contains(state);
                  return Padding(
                    padding: const EdgeInsets.only(right: 8),
                    child: FilterChip(
                      label: Text(state.displayName),
                      selected: isSelected,
                      onSelected: widget.isLoading
                          ? null
                          : (_) => _toggleRiskState(state),
                      selectedColor: _getRiskStateColor(state).withOpacity(0.2),
                      checkmarkColor: _getRiskStateColor(state),
                      labelStyle: TextStyle(
                        color: isSelected
                            ? _getRiskStateColor(state)
                            : Colors.grey[700],
                        fontWeight:
                            isSelected ? FontWeight.w600 : FontWeight.normal,
                      ),
                      side: BorderSide(
                        color: isSelected
                            ? _getRiskStateColor(state)
                            : Colors.grey[300]!,
                      ),
                    ),
                  );
                }),
                // Divider
                Container(
                  height: 24,
                  width: 1,
                  color: Colors.grey[300],
                  margin: const EdgeInsets.symmetric(horizontal: 8),
                ),
                // Price dropdown with all distinct prices
                if (widget.priceStats != null && widget.priceStats!.prices.isNotEmpty)
                  _FilterDropdown<int?>(
                    label: 'Price',
                    value: _getCurrentPriceFilter(),
                    items: [
                      const DropdownMenuItem(
                        value: null,
                        child: Text('All prices'),
                      ),
                      ...widget.priceStats!.prices.map((price) => DropdownMenuItem(
                            value: price.priceCents,
                            child: Text('${price.formatted} (${price.count})'),
                          )),
                    ],
                    onChanged: widget.isLoading ? null : _setPriceFilter,
                    isActive: widget.filters.priceMinCents != null ||
                        widget.filters.priceMaxCents != null,
                  ),
                const SizedBox(width: 8),
                // Billing interval dropdown
                _FilterDropdown<BillingInterval?>(
                  label: 'Interval',
                  value: widget.filters.billingInterval,
                  items: [
                    const DropdownMenuItem(
                      value: null,
                      child: Text('All intervals'),
                    ),
                    ...BillingInterval.values.map((interval) =>
                        DropdownMenuItem(
                          value: interval,
                          child: Text(interval.displayName),
                        )),
                  ],
                  onChanged: widget.isLoading ? null : _setBillingInterval,
                  isActive: widget.filters.billingInterval != null,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _FilterDropdown<T> extends StatelessWidget {
  final String label;
  final T value;
  final List<DropdownMenuItem<T>> items;
  final ValueChanged<T?>? onChanged;
  final bool isActive;

  const _FilterDropdown({
    required this.label,
    required this.value,
    required this.items,
    required this.onChanged,
    this.isActive = false,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        border: Border.all(
          color: isActive ? AppTheme.primary : Colors.grey[300]!,
        ),
        borderRadius: BorderRadius.circular(8),
        color: isActive ? AppTheme.primary.withOpacity(0.05) : null,
      ),
      child: DropdownButtonHideUnderline(
        child: DropdownButton<T>(
          value: value,
          items: items,
          onChanged: onChanged,
          isDense: true,
          style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                color: isActive ? AppTheme.primary : Colors.grey[700],
                fontWeight: isActive ? FontWeight.w600 : FontWeight.normal,
              ),
          icon: Icon(
            Icons.keyboard_arrow_down,
            size: 20,
            color: isActive ? AppTheme.primary : Colors.grey[500],
          ),
        ),
      ),
    );
  }
}

class _ClearFiltersButton extends StatelessWidget {
  final int filterCount;
  final VoidCallback onPressed;

  const _ClearFiltersButton({
    required this.filterCount,
    required this.onPressed,
  });

  @override
  Widget build(BuildContext context) {
    return TextButton.icon(
      onPressed: onPressed,
      style: TextButton.styleFrom(
        foregroundColor: AppTheme.danger,
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      ),
      icon: const Icon(Icons.clear_all, size: 18),
      label: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Text('Clear'),
          if (filterCount > 0) ...[
            const SizedBox(width: 4),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
              decoration: BoxDecoration(
                color: AppTheme.danger.withOpacity(0.1),
                borderRadius: BorderRadius.circular(10),
              ),
              child: Text(
                filterCount.toString(),
                style: const TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }
}
