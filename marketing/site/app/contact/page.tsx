import { Metadata } from 'next';
import Header from '@/components/Header';
import Footer from '@/components/Footer';
import Contact from '@/components/Contact';

export const metadata: Metadata = {
  title: 'Contact | LedgerSpear',
  description:
    'Get in touch with the LedgerSpear team. Email us for support, partnerships, or general inquiries.',
  openGraph: {
    title: 'Contact | LedgerSpear',
    description: 'Get in touch with the LedgerSpear team.',
    type: 'website',
  },
};

export default function ContactPage() {
  return (
    <>
      <Header />
      <main className="pt-16">
        <Contact />
      </main>
      <Footer />
    </>
  );
}
