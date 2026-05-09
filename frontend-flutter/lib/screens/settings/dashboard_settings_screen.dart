import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import '../../core/dashboard_registry.dart';
import '../../providers/dashboard_provider.dart';
import '../../widgets/lg_page.dart';

class DashboardSettingsScreen extends StatelessWidget {
  const DashboardSettingsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final dp = context.watch<DashboardProvider>();

    return LgPage(
      title: 'Dashboard',
      subtitle: 'Customize KPIs and widgets',
      backAction: () => context.go('/settings'),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const ListTile(
            title: Text('Primary KPIs'),
            subtitle: Text('Choose up to 4 metrics'),
          ),
          for (final kpi in kAllKpis)
            CheckboxListTile(
              title: Text(kpi.label),
              value: dp.primaryKpis.contains(kpi.id),
              onChanged: (v) {
                final current = List<String>.from(dp.primaryKpis);
                if (v == true) {
                  if (current.length >= 4) {
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(content: Text('Maximum 4 KPIs')),
                    );
                    return;
                  }
                  current.add(kpi.id);
                } else {
                  current.remove(kpi.id);
                }
                dp.saveDashboardPreferences(current, dp.secondaryWidgets);
              },
            ),
          const Divider(),
          const ListTile(
            title: Text('Secondary Widgets'),
            subtitle: Text('Toggle dashboard widgets'),
          ),
          for (final w in kAllWidgets)
            CheckboxListTile(
              title: Text(w.label),
              value: dp.secondaryWidgets.contains(w.id),
              onChanged: (v) {
                final current = List<String>.from(dp.secondaryWidgets);
                if (v == true) {
                  current.add(w.id);
                } else {
                  current.remove(w.id);
                }
                dp.saveDashboardPreferences(dp.primaryKpis, current);
              },
            ),
        ],
      ),
    );
  }
}
