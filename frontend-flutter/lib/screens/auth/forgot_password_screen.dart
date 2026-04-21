import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import '../../providers/auth_provider.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';

class ForgotPasswordScreen extends StatefulWidget {
  const ForgotPasswordScreen({super.key});

  @override
  State<ForgotPasswordScreen> createState() => _ForgotPasswordScreenState();
}

class _ForgotPasswordScreenState extends State<ForgotPasswordScreen> {
  final _emailController = TextEditingController();
  final _formKey = GlobalKey<FormState>();
  bool _sent = false;

  @override
  void dispose() {
    _emailController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    final auth = context.read<AuthProvider>();
    auth.clearError();
    final success = await auth.resetPassword(_emailController.text.trim());
    if (success && mounted) {
      setState(() => _sent = true);
    }
  }

  @override
  Widget build(BuildContext context) {
    final auth = context.watch<AuthProvider>();
    final theme = Theme.of(context);

    return Scaffold(
      backgroundColor: LgColors.backdrop,
      body: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(LgSpacing.s600),
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 420),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Image.asset('assets/images/logo.jpeg', height: 72),
                const SizedBox(height: LgSpacing.s300),
                Text(
                  'LedgerGuard',
                  style: theme.textTheme.headlineMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                    color: LgColors.primary,
                  ),
                ),
                const SizedBox(height: LgSpacing.s100),
                Text(
                  'Revenue Intelligence Platform',
                  style: theme.textTheme.bodyMedium?.copyWith(
                    color: LgColors.textSecondary,
                  ),
                ),
                const SizedBox(height: LgSpacing.s800),
                LgCard(
                  child: Form(
                    key: _formKey,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Reset your password',
                          style: theme.textTheme.titleMedium,
                        ),
                        const SizedBox(height: LgSpacing.s200),
                        Text(
                          'Enter your email and we\'ll send you a reset link.',
                          style: theme.textTheme.bodySmall?.copyWith(
                            color: LgColors.textSecondary,
                          ),
                        ),
                        const SizedBox(height: LgSpacing.s600),
                        TextFormField(
                          controller: _emailController,
                          decoration: const InputDecoration(
                            labelText: 'Email',
                            hintText: 'you@example.com',
                          ),
                          keyboardType: TextInputType.emailAddress,
                          textInputAction: TextInputAction.done,
                          onFieldSubmitted: (_) => _submit(),
                          validator: (v) {
                            if (v == null || v.trim().isEmpty) {
                              return 'Email is required';
                            }
                            if (!RegExp(r'^[^@\s]+@[^@\s]+\.[^@\s]+$').hasMatch(v.trim())) {
                              return 'Enter a valid email address';
                            }
                            return null;
                          },
                        ),
                        if (auth.error != null) ...[
                          const SizedBox(height: LgSpacing.s300),
                          Container(
                            width: double.infinity,
                            padding: const EdgeInsets.all(LgSpacing.s300),
                            decoration: BoxDecoration(
                              color: LgColors.critical.withValues(alpha: 0.1),
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: Row(
                              children: [
                                const Icon(Icons.error_outline, size: 16, color: LgColors.critical),
                                const SizedBox(width: LgSpacing.s200),
                                Expanded(
                                  child: Text(
                                    auth.error!,
                                    style: theme.textTheme.bodySmall?.copyWith(color: LgColors.critical),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ],
                        if (_sent) ...[
                          const SizedBox(height: LgSpacing.s300),
                          Text(
                            'Check your email for a reset link',
                            style: theme.textTheme.bodySmall?.copyWith(
                              color: LgColors.success,
                            ),
                          ),
                        ],
                        const SizedBox(height: LgSpacing.s600),
                        SizedBox(
                          width: double.infinity,
                          child: FilledButton(
                            onPressed: auth.isLoading ? null : _submit,
                            child: auth.isLoading
                                ? const SizedBox(
                                    height: 20,
                                    width: 20,
                                    child: CircularProgressIndicator(
                                      strokeWidth: 2,
                                      color: Colors.white,
                                    ),
                                  )
                                : const Text('Send Reset Link'),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: LgSpacing.s600),
                TextButton(
                  onPressed: () => context.go('/login'),
                  child: const Text('Back to Sign in'),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
