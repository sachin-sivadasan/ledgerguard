import { Metadata } from 'next';
import Header from '@/components/Header';
import Footer from '@/components/Footer';
import APIIntegrationGuide from '@/components/APIIntegrationGuide';

export const metadata: Metadata = {
  title: 'API Integration Guide - LedgerGuard',
  description: 'Learn how to integrate LedgerGuard API into your Shopify app for real-time subscription status',
  openGraph: {
    title: 'API Integration Guide - LedgerGuard',
    description: 'Learn how to integrate LedgerGuard API into your Shopify app for real-time subscription status',
    type: 'website',
  },
};

export default function APIGuidePage() {
  return (
    <>
      <Header />
      <main className="min-h-screen bg-gradient-to-b from-slate-950 via-indigo-950 to-slate-950 pt-24 pb-12 px-4">
        <div className="max-w-6xl mx-auto">
          <a
            href="/"
            className="text-indigo-400 hover:text-indigo-300 text-sm transition-colors inline-flex items-center gap-2 mb-6"
          >
            &larr; Back to Home
          </a>
          <APIIntegrationGuide />
        </div>
      </main>
      <Footer />
    </>
  );
}
