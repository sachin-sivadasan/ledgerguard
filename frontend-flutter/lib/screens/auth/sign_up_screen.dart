import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import '../../providers/auth_provider.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_spacing.dart';
import '../../widgets/lg_card.dart';

class SignUpScreen extends StatefulWidget {
  const SignUpScreen({super.key});

  @override
  State<SignUpScreen> createState() => _SignUpScreenState();
}

class _SignUpScreenState extends State<SignUpScreen> {
  final _nameController = TextEditingController();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  final _confirmController = TextEditingController();
  final _formKey = GlobalKey<FormState>();
  bool _obscurePassword = true;
  bool _obscureConfirm = true;

  @override
  void dispose() {
    _nameController.dispose();
    _emailController.dispose();
    _passwordController.dispose();
    _confirmController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    final auth = context.read<AuthProvider>();
    auth.clearError();
    await auth.signUp(
      _nameController.text.trim(),
      _emailController.text.trim(),
      _passwordController.text,
    );
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
                          'Create your account',
                          style: theme.textTheme.titleMedium,
                        ),
                        const SizedBox(height: LgSpacing.s600),
                        TextFormField(
                          controller: _nameController,
                          decoration: const InputDecoration(
                            labelText: 'Full Name',
                          ),
                          textInputAction: TextInputAction.next,
                          validator: (v) {
                            if (v == null || v.trim().isEmpty) {
                              return 'Name is required';
                            }
                            return null;
                          },
                        ),
                        const SizedBox(height: LgSpacing.s400),
                        TextFormField(
                          controller: _emailController,
                          decoration: const InputDecoration(
                            labelText: 'Email',
                            hintText: 'you@example.com',
                          ),
                          keyboardType: TextInputType.emailAddress,
                          textInputAction: TextInputAction.next,
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
                        const SizedBox(height: LgSpacing.s400),
                        TextFormField(
                          controller: _passwordController,
                          decoration: InputDecoration(
                            labelText: 'Password',
                            suffixIcon: IconButton(
                              icon: Icon(_obscurePassword ? Icons.visibility_off : Icons.visibility, size: 20),
                              onPressed: () => setState(() => _obscurePassword = !_obscurePassword),
                            ),
                          ),
                          obscureText: _obscurePassword,
                          textInputAction: TextInputAction.next,
                          validator: (v) {
                            if (v == null || v.length < 6) {
                              return 'Password must be at least 6 characters';
                            }
                            return null;
                          },
                        ),
                        const SizedBox(height: LgSpacing.s400),
                        TextFormField(
                          controller: _confirmController,
                          decoration: InputDecoration(
                            labelText: 'Confirm Password',
                            suffixIcon: IconButton(
                              icon: Icon(_obscureConfirm ? Icons.visibility_off : Icons.visibility, size: 20),
                              onPressed: () => setState(() => _obscureConfirm = !_obscureConfirm),
                            ),
                          ),
                          obscureText: _obscureConfirm,
                          textInputAction: TextInputAction.done,
                          onFieldSubmitted: (_) => _submit(),
                          validator: (v) {
                            if (v != _passwordController.text) {
                              return 'Passwords do not match';
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
                                : const Text('Create Account'),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: LgSpacing.s600),
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(
                      'Already have an account? ',
                      style: theme.textTheme.bodyMedium?.copyWith(
                        color: LgColors.textSecondary,
                      ),
                    ),
                    TextButton(
                      onPressed: () => context.go('/login'),
                      child: const Text('Sign in'),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
