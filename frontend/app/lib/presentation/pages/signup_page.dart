import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../core/theme/app_theme.dart';
import '../blocs/auth/auth.dart';

class SignupPage extends StatefulWidget {
  const SignupPage({super.key});

  @override
  State<SignupPage> createState() => _SignupPageState();
}

class _SignupPageState extends State<SignupPage> {
  final _formKey = GlobalKey<FormState>();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  void _onCreateAccount() {
    if (_formKey.currentState?.validate() ?? false) {
      // TODO: Implement signup with email
      // For now, we'll use Google sign in as the primary method
    }
  }

  void _onGoogleSignIn() {
    context.read<AuthBloc>().add(const SignInWithGoogleRequested());
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 400),
              child: BlocConsumer<AuthBloc, AuthState>(
                listener: (context, state) {
                  if (state is Authenticated) {
                    context.go('/dashboard');
                  }
                },
                builder: (context, state) {
                  final isLoading = state is AuthLoading;

                  return Form(
                    key: _formKey,
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        // Logo / Title
                        Text(
                          'LedgerGuard',
                          style: Theme.of(context).textTheme.headlineLarge?.copyWith(
                                fontWeight: FontWeight.bold,
                                color: AppTheme.primary,
                              ),
                          textAlign: TextAlign.center,
                        ),
                        const SizedBox(height: 8),
                        Text(
                          'Create your account',
                          style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                                color: AppTheme.secondary,
                              ),
                          textAlign: TextAlign.center,
                        ),
                        const SizedBox(height: 48),

                        // Error message
                        if (state is AuthError) ...[
                          Container(
                            padding: const EdgeInsets.all(12),
                            decoration: BoxDecoration(
                              color: AppTheme.danger.withOpacity(0.1),
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: Text(
                              state.message,
                              style: const TextStyle(color: AppTheme.danger),
                              textAlign: TextAlign.center,
                            ),
                          ),
                          const SizedBox(height: 16),
                        ],

                        // Email field
                        TextFormField(
                          controller: _emailController,
                          keyboardType: TextInputType.emailAddress,
                          textInputAction: TextInputAction.next,
                          decoration: const InputDecoration(
                            labelText: 'Email',
                            prefixIcon: Icon(Icons.email_outlined),
                          ),
                          validator: (value) {
                            if (value == null || value.isEmpty) {
                              return 'Please enter your email';
                            }
                            return null;
                          },
                        ),
                        const SizedBox(height: 16),

                        // Password field
                        TextFormField(
                          controller: _passwordController,
                          obscureText: true,
                          textInputAction: TextInputAction.done,
                          decoration: const InputDecoration(
                            labelText: 'Password',
                            prefixIcon: Icon(Icons.lock_outlined),
                            helperText: 'At least 8 characters',
                          ),
                          validator: (value) {
                            if (value == null || value.isEmpty) {
                              return 'Please enter a password';
                            }
                            if (value.length < 8) {
                              return 'Password must be at least 8 characters';
                            }
                            return null;
                          },
                          onFieldSubmitted: (_) => _onCreateAccount(),
                        ),
                        const SizedBox(height: 24),

                        // Create Account button
                        SizedBox(
                          height: 48,
                          child: ElevatedButton(
                            onPressed: isLoading ? null : _onCreateAccount,
                            child: isLoading
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
                        const SizedBox(height: 16),

                        // Divider
                        Row(
                          children: [
                            const Expanded(child: Divider()),
                            Padding(
                              padding: const EdgeInsets.symmetric(horizontal: 16),
                              child: Text(
                                'or',
                                style: TextStyle(color: Colors.grey[600]),
                              ),
                            ),
                            const Expanded(child: Divider()),
                          ],
                        ),
                        const SizedBox(height: 16),

                        // Google Sign In button
                        SizedBox(
                          height: 48,
                          child: OutlinedButton.icon(
                            onPressed: isLoading ? null : _onGoogleSignIn,
                            icon: Image.network(
                              'https://www.google.com/favicon.ico',
                              height: 18,
                              width: 18,
                              errorBuilder: (_, __, ___) => const Icon(Icons.g_mobiledata),
                            ),
                            label: const Text('Continue with Google'),
                          ),
                        ),
                        const SizedBox(height: 32),

                        // Sign in link
                        Wrap(
                          alignment: WrapAlignment.center,
                          crossAxisAlignment: WrapCrossAlignment.center,
                          children: [
                            const Text('Already have an account?'),
                            TextButton(
                              onPressed: () => context.go('/login'),
                              child: const Text('Sign In'),
                            ),
                          ],
                        ),
                      ],
                    ),
                  );
                },
              ),
            ),
          ),
        ),
      ),
    );
  }
}
