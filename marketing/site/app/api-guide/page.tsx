import APIIntegrationGuide from '../../components/APIIntegrationGuide';

export const metadata = {
  title: 'API Integration Guide - LedgerGuard',
  description: 'Learn how to integrate LedgerGuard API into your Shopify app for real-time subscription status',
};

export default function APIGuidePage() {
  return (
    <main style={{
      minHeight: '100vh',
      background: 'linear-gradient(180deg, #0f0f1a 0%, #1a1a2e 50%, #0f0f1a 100%)',
      padding: '40px 20px',
    }}>
      <APIIntegrationGuide />
    </main>
  );
}
