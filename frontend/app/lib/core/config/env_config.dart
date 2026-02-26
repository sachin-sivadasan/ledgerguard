/// Environment configuration for LedgerGuard
enum Environment { dev, prod }

class EnvConfig {
  final Environment environment;
  final String apiBaseUrl;
  final String firebaseProjectId;

  const EnvConfig._({
    required this.environment,
    required this.apiBaseUrl,
    required this.firebaseProjectId,
  });

  static const EnvConfig dev = EnvConfig._(
    environment: Environment.dev,
    apiBaseUrl: 'http://localhost:8080',
    firebaseProjectId: 'ledgerguard-dev',
  );

  static const EnvConfig prod = EnvConfig._(
    environment: Environment.prod,
    apiBaseUrl: 'https://api.ledgerguard.com',
    firebaseProjectId: 'ledgerguard-prod',
  );

  bool get isDev => environment == Environment.dev;
  bool get isProd => environment == Environment.prod;
}
