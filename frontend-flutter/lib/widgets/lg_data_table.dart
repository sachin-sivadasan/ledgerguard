import 'package:flutter/material.dart';
import '../theme/app_colors.dart';

class LgColumn {
  final String title;
  final bool numeric;
  const LgColumn({required this.title, this.numeric = false});
}

class LgDataTable extends StatelessWidget {
  final List<LgColumn> columns;
  final List<DataRow> rows;
  final bool showCheckboxColumn;

  const LgDataTable({
    super.key,
    required this.columns,
    required this.rows,
    this.showCheckboxColumn = false,
  });

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: double.infinity,
      child: DataTable(
        headingRowColor: WidgetStateProperty.all(LgColors.surfaceSecondary),
        headingTextStyle: const TextStyle(
            fontSize: 12,
            fontWeight: FontWeight.w600,
            color: LgColors.textSecondary),
        dataTextStyle:
            const TextStyle(fontSize: 13, color: LgColors.textPrimary),
        columnSpacing: 24,
        horizontalMargin: 16,
        showCheckboxColumn: showCheckboxColumn,
        columns: columns
            .map((c) => DataColumn(
                  label: Text(c.title),
                  numeric: c.numeric,
                ))
            .toList(),
        rows: rows,
      ),
    );
  }
}
