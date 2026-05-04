import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../../models/insight_model.dart';
import '../../providers/apps_provider.dart';
import '../../providers/insights_provider.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../theme/app_breakpoints.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_empty_state.dart';
import '../../widgets/lg_page.dart';

class InsightsScreen extends StatelessWidget {
  const InsightsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final hasApps = context.watch<AppsProvider>().apps.isNotEmpty;
    if (!hasApps) {
      return LgPage(
        title: 'AI Insights',
        child: LgEmptyState(
          icon: Icons.lightbulb_outlined,
          heading: 'No insights yet',
          description:
              'Connect your Shopify app to get AI-powered insights.',
          actionLabel: 'Go to Apps',
          onAction: () => context.go('/apps'),
        ),
      );
    }

    final provider = context.watch<InsightsProvider>();
    final theme = Theme.of(context);
    final dateFmt = DateFormat('MMM d, h:mm a');

    return LgPage(
      title: 'AI Insights',
      subtitle: 'Daily briefs and revenue chat',
      child: LgResponsive(
        mobile: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _BriefsSection(provider: provider, theme: theme, dateFmt: dateFmt, severityIcon: _severityIcon, severityColor: _severityColor),
            const SizedBox(height: LgSpacing.s600),
            _ChatSection(provider: provider),
          ],
        ),
        desktop: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              child: _BriefsSection(provider: provider, theme: theme, dateFmt: dateFmt, severityIcon: _severityIcon, severityColor: _severityColor),
            ),
            const SizedBox(width: LgSpacing.s600),
            SizedBox(
              width: 400,
              child: _ChatSection(provider: provider),
            ),
          ],
        ),
      ),
    );
  }

  IconData _severityIcon(InsightSeverity s) => switch (s) {
        InsightSeverity.info => Icons.info_outline,
        InsightSeverity.warning => Icons.warning_amber,
        InsightSeverity.critical => Icons.error_outline,
      };

  Color _severityColor(InsightSeverity s) => switch (s) {
        InsightSeverity.info => LgColors.info,
        InsightSeverity.warning => LgColors.warning,
        InsightSeverity.critical => LgColors.critical,
      };
}

class _BriefsSection extends StatelessWidget {
  final InsightsProvider provider;
  final ThemeData theme;
  final DateFormat dateFmt;
  final IconData Function(InsightSeverity) severityIcon;
  final Color Function(InsightSeverity) severityColor;

  const _BriefsSection({
    required this.provider,
    required this.theme,
    required this.dateFmt,
    required this.severityIcon,
    required this.severityColor,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Daily Briefs', style: theme.textTheme.titleMedium),
        const SizedBox(height: LgSpacing.s300),
        ...provider.insights.map((insight) {
          return Padding(
            padding: const EdgeInsets.only(bottom: LgSpacing.s300),
            child: LgCard(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(severityIcon(insight.severity), size: 16, color: severityColor(insight.severity)),
                      const SizedBox(width: LgSpacing.s200),
                      Expanded(child: Text(insight.title, style: theme.textTheme.titleSmall)),
                      Text(dateFmt.format(insight.date), style: theme.textTheme.bodySmall),
                    ],
                  ),
                  const SizedBox(height: LgSpacing.s200),
                  Text(insight.summary, style: theme.textTheme.bodyMedium),
                ],
              ),
            ),
          );
        }),
      ],
    );
  }
}

class _ChatSection extends StatelessWidget {
  final InsightsProvider provider;
  const _ChatSection({required this.provider});

  @override
  Widget build(BuildContext context) {
    return LgCard(
      title: 'Ask about your revenue',
      child: SizedBox(
        height: 400,
        child: Column(
          children: [
            Expanded(
              child: ListView(
                children: provider.messages.map((msg) {
                  return Align(
                    alignment: msg.isUser ? Alignment.centerRight : Alignment.centerLeft,
                    child: Container(
                      margin: const EdgeInsets.only(bottom: LgSpacing.s200),
                      padding: const EdgeInsets.all(LgSpacing.s300),
                      constraints: const BoxConstraints(maxWidth: 320),
                      decoration: BoxDecoration(
                        color: msg.isUser ? LgColors.primary : LgColors.surfaceSecondary,
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Text(
                        msg.text,
                        style: TextStyle(
                          fontSize: 13,
                          color: msg.isUser ? Colors.white : LgColors.textPrimary,
                        ),
                      ),
                    ),
                  );
                }).toList(),
              ),
            ),
            const SizedBox(height: LgSpacing.s200),
            _ChatInput(onSend: provider.sendMessage),
          ],
        ),
      ),
    );
  }
}

class _ChatInput extends StatefulWidget {
  final ValueChanged<String> onSend;
  const _ChatInput({required this.onSend});

  @override
  State<_ChatInput> createState() => _ChatInputState();
}

class _ChatInputState extends State<_ChatInput> {
  final _controller = TextEditingController();

  void _send() {
    final text = _controller.text.trim();
    if (text.isEmpty) return;
    widget.onSend(text);
    _controller.clear();
  }

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Expanded(
          child: TextField(
            controller: _controller,
            decoration: const InputDecoration(hintText: 'Ask about revenue...'),
            onSubmitted: (_) => _send(),
          ),
        ),
        const SizedBox(width: LgSpacing.s200),
        IconButton(
          onPressed: _send,
          icon: const Icon(Icons.send, color: LgColors.primary),
        ),
      ],
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }
}
