import 'package:flutter/material.dart';

import '../../core/theme/app_theme.dart';

/// Pagination controls with page navigation and page size selector
class PaginationControls extends StatelessWidget {
  final int page;
  final int pageSize;
  final int totalPages;
  final int total;
  final List<int> pageSizeOptions;
  final ValueChanged<int> onPageChanged;
  final ValueChanged<int> onPageSizeChanged;
  final bool isLoading;

  const PaginationControls({
    super.key,
    required this.page,
    required this.pageSize,
    required this.totalPages,
    required this.total,
    this.pageSizeOptions = const [10, 25, 50],
    required this.onPageChanged,
    required this.onPageSizeChanged,
    this.isLoading = false,
  });

  String get _rangeText {
    if (total == 0) return '0 items';
    final start = (page - 1) * pageSize + 1;
    final end = (start + pageSize - 1).clamp(1, total);
    return 'Showing $start-$end of $total';
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border(
          top: BorderSide(color: Colors.grey[200]!),
        ),
      ),
      child: Row(
        children: [
          // Range text
          Text(
            _rangeText,
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: Colors.grey[600],
                ),
          ),
          const Spacer(),
          // Page size selector
          _PageSizeSelector(
            pageSize: pageSize,
            options: pageSizeOptions,
            onChanged: isLoading ? null : onPageSizeChanged,
          ),
          const SizedBox(width: 16),
          // Page navigation
          _PageNavigation(
            page: page,
            totalPages: totalPages,
            onPageChanged: isLoading ? null : onPageChanged,
          ),
        ],
      ),
    );
  }
}

class _PageSizeSelector extends StatelessWidget {
  final int pageSize;
  final List<int> options;
  final ValueChanged<int>? onChanged;

  const _PageSizeSelector({
    required this.pageSize,
    required this.options,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(
          'Rows:',
          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: Colors.grey[600],
              ),
        ),
        const SizedBox(width: 8),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 8),
          decoration: BoxDecoration(
            border: Border.all(color: Colors.grey[300]!),
            borderRadius: BorderRadius.circular(6),
          ),
          child: DropdownButtonHideUnderline(
            child: DropdownButton<int>(
              value: pageSize,
              items: options
                  .map((size) => DropdownMenuItem(
                        value: size,
                        child: Text(
                          size.toString(),
                          style: Theme.of(context).textTheme.bodySmall,
                        ),
                      ))
                  .toList(),
              onChanged: onChanged == null ? null : (value) {
                if (value != null) onChanged!(value);
              },
              isDense: true,
              icon: Icon(
                Icons.keyboard_arrow_down,
                size: 18,
                color: Colors.grey[500],
              ),
            ),
          ),
        ),
      ],
    );
  }
}

class _PageNavigation extends StatelessWidget {
  final int page;
  final int totalPages;
  final ValueChanged<int>? onPageChanged;

  const _PageNavigation({
    required this.page,
    required this.totalPages,
    required this.onPageChanged,
  });

  List<int> get _visiblePages {
    if (totalPages <= 7) {
      return List.generate(totalPages, (i) => i + 1);
    }

    final pages = <int>[];

    // Always show first page
    pages.add(1);

    // Calculate middle pages around current page
    int start = (page - 1).clamp(2, totalPages - 4);
    int end = (page + 1).clamp(4, totalPages - 1);

    // Adjust if near the beginning
    if (page <= 3) {
      start = 2;
      end = 5;
    }

    // Adjust if near the end
    if (page >= totalPages - 2) {
      start = totalPages - 4;
      end = totalPages - 1;
    }

    // Add ellipsis indicator (-1) if needed
    if (start > 2) {
      pages.add(-1); // Ellipsis
    }

    // Add middle pages
    for (int i = start; i <= end; i++) {
      pages.add(i);
    }

    // Add ellipsis indicator if needed
    if (end < totalPages - 1) {
      pages.add(-1); // Ellipsis
    }

    // Always show last page
    if (totalPages > 1) {
      pages.add(totalPages);
    }

    return pages;
  }

  @override
  Widget build(BuildContext context) {
    if (totalPages <= 1) {
      return const SizedBox.shrink();
    }

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        // Previous button
        _NavButton(
          icon: Icons.chevron_left,
          onPressed: onPageChanged != null && page > 1
              ? () => onPageChanged!(page - 1)
              : null,
        ),
        const SizedBox(width: 4),
        // Page numbers
        ..._visiblePages.map((pageNum) {
          if (pageNum == -1) {
            return const Padding(
              padding: EdgeInsets.symmetric(horizontal: 4),
              child: Text('...'),
            );
          }
          return _PageButton(
            page: pageNum,
            isActive: pageNum == page,
            onPressed: onPageChanged != null && pageNum != page
                ? () => onPageChanged!(pageNum)
                : null,
          );
        }),
        const SizedBox(width: 4),
        // Next button
        _NavButton(
          icon: Icons.chevron_right,
          onPressed: onPageChanged != null && page < totalPages
              ? () => onPageChanged!(page + 1)
              : null,
        ),
      ],
    );
  }
}

class _NavButton extends StatelessWidget {
  final IconData icon;
  final VoidCallback? onPressed;

  const _NavButton({
    required this.icon,
    required this.onPressed,
  });

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 32,
      height: 32,
      child: IconButton(
        icon: Icon(icon, size: 18),
        onPressed: onPressed,
        padding: EdgeInsets.zero,
        style: IconButton.styleFrom(
          foregroundColor: onPressed != null ? Colors.grey[700] : Colors.grey[400],
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(6),
            side: BorderSide(color: Colors.grey[300]!),
          ),
        ),
      ),
    );
  }
}

class _PageButton extends StatelessWidget {
  final int page;
  final bool isActive;
  final VoidCallback? onPressed;

  const _PageButton({
    required this.page,
    required this.isActive,
    required this.onPressed,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 2),
      child: SizedBox(
        width: 32,
        height: 32,
        child: TextButton(
          onPressed: onPressed,
          style: TextButton.styleFrom(
            backgroundColor: isActive ? AppTheme.primary : null,
            foregroundColor: isActive ? Colors.white : Colors.grey[700],
            padding: EdgeInsets.zero,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(6),
              side: BorderSide(
                color: isActive ? AppTheme.primary : Colors.grey[300]!,
              ),
            ),
          ),
          child: Text(
            page.toString(),
            style: const TextStyle(fontSize: 13),
          ),
        ),
      ),
    );
  }
}
