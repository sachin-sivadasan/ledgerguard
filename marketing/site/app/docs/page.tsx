import { Metadata } from 'next';
import Header from '@/components/Header';
import Footer from '@/components/Footer';
import APIDocs from '@/components/APIDocs';

export const metadata: Metadata = {
  title: 'API Documentation | LedgerSpear',
  description:
    'LedgerSpear API documentation — endpoints, authentication, risk classification, and rate limits for Shopify revenue intelligence.',
  openGraph: {
    title: 'API Documentation | LedgerSpear',
    description:
      'LedgerSpear API documentation for Shopify revenue intelligence.',
    type: 'website',
  },
};

export default function DocsPage() {
  return (
    <>
      <Header />
      <main className="pt-16">
        <APIDocs />
      </main>
      <Footer />
    </>
  );
}
