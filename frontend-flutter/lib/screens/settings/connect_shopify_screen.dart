import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../core/network/api_client.dart';
import '../../providers/analytics_provider.dart';
import '../../providers/apps_provider.dart';
import '../../providers/dashboard_provider.dart';
import '../../providers/earnings_provider.dart';
import '../../providers/events_provider.dart';
import '../../providers/insights_provider.dart';
import '../../providers/risk_provider.dart';
import '../../providers/store_provider.dart';
import '../../providers/subscription_provider.dart';
import '../../providers/transaction_provider.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';
import '../../widgets/lg_page.dart';

class ConnectShopifyScreen extends StatefulWidget {
  const ConnectShopifyScreen({super.key});

  @override
  State<ConnectShopifyScreen> createState() => _ConnectShopifyScreenState();
}

class _ConnectShopifyScreenState extends State<ConnectShopifyScreen> {
  final _formKey = GlobalKey<FormState>();
  final _partnerIdCtrl = TextEditingController();
  final _tokenCtrl = TextEditingController();
  bool _loading = false;
  String? _error;
  String? _statusMsg;

  // Step 2: app selection
  List<Map<String, String>> _availableApps = [];
  final Set<String> _selectedAppIds = {};
  bool _connected = false;

  // Integration status
  bool _alreadyConnected = false;
  String? _connectedPartnerId;

  @override
  void initState() {
    super.initState();
    _checkStatus();
  }

  @override
  void dispose() {
    _partnerIdCtrl.dispose();
    _tokenCtrl.dispose();
    super.dispose();
  }

  Future<void> _checkStatus() async {
    try {
      final client = context.read<ApiClient>();
      final resp = await client.get('/api/v1/integrations/shopify/status');
      final data = resp.data;
      if (data['connected'] == true) {
        setState(() {
          _alreadyConnected = true;
          _connectedPartnerId = data['partner_id']?.toString();
        });
      }
    } catch (_) {
      // Not connected or error — show connect form
    }
  }

  Future<void> _connectOAuth() async {
    setState(() {
      _loading = true;
      _error = null;
      _statusMsg = 'Redirecting to Shopify...';
    });

    try {
      final client = context.read<ApiClient>();
      final resp = await client.get('/api/v1/integrations/shopify/oauth');
      final url = resp.data['url']?.toString();
      if (url != null && url.isNotEmpty) {
        await launchUrl(Uri.parse(url), mode: LaunchMode.externalApplication);
      }
      if (!mounted) return;
      setState(() {
        _loading = false;
        _statusMsg = null;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.toString();
        _loading = false;
        _statusMsg = null;
      });
    }
  }

  Future<void> _connectManual() async {
    if (!(_formKey.currentState?.validate() ?? false)) return;

    final partnerId = _partnerIdCtrl.text.trim();
    final token = _tokenCtrl.text.trim();

    setState(() {
      _loading = true;
      _error = null;
      _statusMsg = 'Connecting partner account...';
    });

    try {
      final client = context.read<ApiClient>();

      // Step 1: Save token
      await client.post('/api/v1/integrations/shopify/token', data: {
        'partner_id': partnerId,
        'token': token,
      });

      if (!mounted) return;
      setState(() => _statusMsg = 'Discovering apps...');

      // Step 2: Fetch available apps
      final resp = await client.get('/api/v1/apps/available');
      final appsList = resp.data['apps'] as List<dynamic>? ?? [];

      if (!mounted) return;

      if (appsList.isEmpty) {
        setState(() {
          _error = 'No apps found for this partner account. Check your org ID.';
          _loading = false;
          _statusMsg = null;
        });
        return;
      }

      _availableApps = appsList
          .map((a) => {
                'id': a['id']?.toString() ?? '',
                'name': a['name']?.toString() ?? 'Unknown',
              })
          .toList();
      _selectedAppIds.addAll(_availableApps.map((a) => a['id']!));

      setState(() {
        _connected = true;
        _loading = false;
        _statusMsg = null;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.toString();
        _loading = false;
        _statusMsg = null;
      });
    }
  }

  Future<void> _selectApps() async {
    if (_selectedAppIds.isEmpty) {
      setState(() => _error = 'Select at least one app.');
      return;
    }

    setState(() {
      _loading = true;
      _error = null;
      _statusMsg = 'Adding apps...';
    });

    try {
      final client = context.read<ApiClient>();
      for (final app in _availableApps) {
        if (!_selectedAppIds.contains(app['id'])) continue;
        await client.post('/api/v1/apps/select', data: {
          'partner_app_id': app['id'],
          'name': app['name'],
        });
      }

      if (!mounted) return;
      setState(() => _statusMsg = 'Switching to live mode...');

      // Switch all providers from demo to live mode
      final appsProvider = context.read<AppsProvider>();
      appsProvider.setDemoMode(false);
      context.read<DashboardProvider>().setDemoMode(false);
      context.read<SubscriptionProvider>().setDemoMode(false);
      context.read<StoreProvider>().setDemoMode(false);
      context.read<TransactionProvider>().setDemoMode(false);
      context.read<EventsProvider>().setDemoMode(false);
      context.read<RiskProvider>().setDemoMode(false);
      context.read<AnalyticsProvider>().setDemoMode(false);
      context.read<EarningsProvider>().setDemoMode(false);
      context.read<InsightsProvider>().setDemoMode(false);

      setState(() => _statusMsg = 'Loading apps...');
      await appsProvider.loadApps();

      // Load data for all providers with first app
      if (mounted && appsProvider.apps.isNotEmpty) {
        final appId = appsProvider.apps.first.id;
        context.read<DashboardProvider>().loadMetrics(appId);
        context.read<SubscriptionProvider>().loadSubscriptions(appId);
        context.read<StoreProvider>().loadStores(appId);
        context.read<TransactionProvider>().loadTransactions(appId);
        context.read<EventsProvider>().loadEvents(appId);
        context.read<RiskProvider>().loadRiskSummary(appId);
        context.read<AnalyticsProvider>().loadAnalytics(appId);
        context.read<EarningsProvider>().loadEarnings(appId);
        context.read<InsightsProvider>().loadInsights(appId);
      }

      if (mounted) context.go('/apps');
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.toString();
        _loading = false;
        _statusMsg = null;
      });
    }
  }

  Future<void> _disconnect() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final client = context.read<ApiClient>();
      await client.delete('/api/v1/integrations/shopify/token');
      if (!mounted) return;

      // Switch all providers back to demo mode
      context.read<AppsProvider>().setDemoMode(true);
      context.read<DashboardProvider>().setDemoMode(true);
      context.read<SubscriptionProvider>().setDemoMode(true);
      context.read<StoreProvider>().setDemoMode(true);
      context.read<TransactionProvider>().setDemoMode(true);
      context.read<EventsProvider>().setDemoMode(true);
      context.read<RiskProvider>().setDemoMode(true);
      context.read<AnalyticsProvider>().setDemoMode(true);
      context.read<EarningsProvider>().setDemoMode(true);
      context.read<InsightsProvider>().setDemoMode(true);

      setState(() {
        _alreadyConnected = false;
        _connectedPartnerId = null;
        _loading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.toString();
        _loading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return LgPage(
      title: 'Connect Shopify',
      subtitle: _connected
          ? 'Select apps to track'
          : 'Link your Shopify Partner account',
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 600),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (_error != null) ...[
              _buildErrorBanner(),
              const SizedBox(height: LgSpacing.s600),
            ],
            if (_alreadyConnected && !_connected) _buildConnectedCard(),
            if (!_alreadyConnected && !_connected) ...[
              _buildOAuthSection(),
              const SizedBox(height: LgSpacing.s600),
              _buildManualTokenSection(),
            ],
            if (_connected) _buildAppSelectionSection(),
          ],
        ),
      ),
    );
  }

  Widget _buildErrorBanner() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: LgColors.critical.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: LgColors.critical.withValues(alpha: 0.3)),
      ),
      child: Row(
        children: [
          const Icon(Icons.error_outline, color: LgColors.critical),
          const SizedBox(width: 12),
          Expanded(
            child: Text(_error!, style: const TextStyle(color: LgColors.critical)),
          ),
          IconButton(
            icon: const Icon(Icons.close, size: 18),
            onPressed: () => setState(() => _error = null),
          ),
        ],
      ),
    );
  }

  Widget _buildConnectedCard() {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: LgColors.success.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: LgColors.success.withValues(alpha: 0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: LgColors.success,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(Icons.check, color: Colors.white, size: 20),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('Connected',
                        style: Theme.of(context).textTheme.titleMedium?.copyWith(
                              fontWeight: FontWeight.bold,
                              color: LgColors.success,
                            )),
                    if (_connectedPartnerId != null)
                      Text('Partner ID: $_connectedPartnerId',
                          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                color: LgColors.textSecondary,
                              )),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          OutlinedButton.icon(
            onPressed: _loading ? null : _disconnect,
            icon: const Icon(Icons.link_off, size: 18),
            label: const Text('Disconnect'),
            style: OutlinedButton.styleFrom(
              foregroundColor: LgColors.critical,
              side: const BorderSide(color: LgColors.critical),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildOAuthSection() {
    return LgCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(10),
                decoration: BoxDecoration(
                  color: LgColors.primary.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(Icons.storefront, color: LgColors.primary, size: 24),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('Connect with OAuth',
                        style: Theme.of(context)
                            .textTheme
                            .titleMedium
                            ?.copyWith(fontWeight: FontWeight.bold)),
                    Text('Recommended',
                        style: Theme.of(context)
                            .textTheme
                            .bodySmall
                            ?.copyWith(color: LgColors.success)),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          const Text(
            'Securely connect your Shopify Partner account using OAuth. '
            'You will be redirected to Shopify to authorize access.',
            style: TextStyle(color: LgColors.textSecondary, fontSize: 14),
          ),
          const SizedBox(height: 20),
          SizedBox(
            width: double.infinity,
            height: 48,
            child: FilledButton.icon(
              onPressed: _loading ? null : _connectOAuth,
              icon: const Icon(Icons.link),
              label: const Text('Connect Shopify Partner'),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildManualTokenSection() {
    return LgCard(
      child: Form(
        key: _formKey,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: LgColors.warning.withValues(alpha: 0.1),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: const Icon(Icons.key, color: LgColors.warning, size: 24),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('Manual Token Entry',
                          style: Theme.of(context)
                              .textTheme
                              .titleMedium
                              ?.copyWith(fontWeight: FontWeight.bold)),
                      Text('Admin / Dev',
                          style: Theme.of(context)
                              .textTheme
                              .bodySmall
                              ?.copyWith(color: LgColors.warning)),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            const Text(
              'Manually enter your Partner API credentials. '
              'Use this for testing or when OAuth is not available.',
              style: TextStyle(color: LgColors.textSecondary, fontSize: 14),
            ),
            const SizedBox(height: 20),
            TextFormField(
              controller: _partnerIdCtrl,
              decoration: const InputDecoration(
                labelText: 'Partner ID',
                hintText: 'Enter your Partner ID (e.g. 1002)',
                prefixIcon: Icon(Icons.badge_outlined),
                border: OutlineInputBorder(),
              ),
              validator: (v) =>
                  (v == null || v.trim().isEmpty) ? 'Partner ID is required' : null,
            ),
            const SizedBox(height: 16),
            TextFormField(
              controller: _tokenCtrl,
              obscureText: true,
              decoration: const InputDecoration(
                labelText: 'API Token',
                hintText: 'Enter your API token',
                prefixIcon: Icon(Icons.vpn_key_outlined),
                border: OutlineInputBorder(),
              ),
              validator: (v) =>
                  (v == null || v.trim().isEmpty) ? 'API Token is required' : null,
            ),
            const SizedBox(height: 20),
            SizedBox(
              width: double.infinity,
              height: 48,
              child: FilledButton(
                onPressed: _loading ? null : _connectManual,
                style: FilledButton.styleFrom(
                  backgroundColor: LgColors.warning,
                ),
                child: _loading
                    ? Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          const SizedBox(
                            width: 18, height: 18,
                            child: CircularProgressIndicator(
                                strokeWidth: 2, color: Colors.white),
                          ),
                          const SizedBox(width: 12),
                          Text(_statusMsg ?? 'Connecting...',
                              style: const TextStyle(color: Colors.white)),
                        ],
                      )
                    : const Text('Save Token'),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildAppSelectionSection() {
    return LgCard(
      title: 'Select Apps (${_selectedAppIds.length}/${_availableApps.length})',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Found ${_availableApps.length} app${_availableApps.length == 1 ? '' : 's'}. '
            'Select which apps to track in LedgerGuard.',
            style: const TextStyle(color: LgColors.textSecondary, fontSize: 14),
          ),
          const SizedBox(height: LgSpacing.s400),
          ..._availableApps.map((app) {
            final id = app['id']!;
            return CheckboxListTile(
              value: _selectedAppIds.contains(id),
              onChanged: (v) {
                setState(() {
                  if (v == true) {
                    _selectedAppIds.add(id);
                  } else {
                    _selectedAppIds.remove(id);
                  }
                });
              },
              title: Text(app['name']!,
                  style: const TextStyle(fontWeight: FontWeight.w600)),
              subtitle: Text(id,
                  style: const TextStyle(
                      fontSize: 12, color: LgColors.textSecondary)),
              controlAffinity: ListTileControlAffinity.leading,
              contentPadding: EdgeInsets.zero,
            );
          }),
          const SizedBox(height: LgSpacing.s400),
          SizedBox(
            width: double.infinity,
            height: 48,
            child: FilledButton(
              onPressed: _loading ? null : _selectApps,
              child: _loading
                  ? Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        const SizedBox(
                          width: 18, height: 18,
                          child: CircularProgressIndicator(
                              strokeWidth: 2, color: Colors.white),
                        ),
                        const SizedBox(width: 12),
                        Text(_statusMsg ?? 'Adding...'),
                      ],
                    )
                  : Text(
                      'Add ${_selectedAppIds.length} App${_selectedAppIds.length == 1 ? '' : 's'}'),
            ),
          ),
        ],
      ),
    );
  }
}
