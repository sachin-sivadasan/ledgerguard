import { Metadata } from 'next';
import Header from '@/components/Header';
import Footer from '@/components/Footer';
import About from '@/components/About';

export const metadata: Metadata = {
  title: 'About | LedgerSpear',
  description:
    'LedgerSpear is a revenue intelligence platform built for Shopify app developers. Learn about our mission, approach, and the team behind the product.',
  openGraph: {
    title: 'About | LedgerSpear',
    description:
      'LedgerSpear is a revenue intelligence platform built for Shopify app developers.',
    type: 'website',
  },
};

export default function AboutPage() {
  return (
    <>
      <Header />
      <main className="pt-16">
        <About />
      </main>
      <Footer />
    </>
  );
}
