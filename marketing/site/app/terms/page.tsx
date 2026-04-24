import { Metadata } from 'next';
import Header from '@/components/Header';
import Footer from '@/components/Footer';

export const metadata: Metadata = {
  title: 'Terms of Service | LedgerSpear',
  description: 'LedgerSpear Terms of Service — the rules governing use of our revenue intelligence platform.',
  openGraph: {
    title: 'Terms of Service | LedgerSpear',
    description: 'LedgerSpear Terms of Service.',
    type: 'website',
  },
};

export default function TermsPage() {
  return (
    <>
      <Header />
      <main className="pt-16">
        <div className="py-20 lg:py-24 bg-white">
          <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8">
            <h1 className="text-4xl font-bold text-slate-900 mb-2">
              Terms of Service
            </h1>
            <p className="text-slate-500 mb-12">
              Last updated: March 16, 2026
            </p>

            <div className="prose prose-slate max-w-none">
              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                1. Agreement to Terms
              </h2>
              <p className="text-slate-600 mb-4">
                By accessing or using LedgerSpear (&quot;the Service&quot;), operated by
                LedgerSpear (&quot;we&quot;, &quot;us&quot;, or &quot;our&quot;), you agree to be bound by
                these Terms of Service. If you do not agree, do not use the
                Service.
              </p>

              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                2. Description of Service
              </h2>
              <p className="text-slate-600 mb-4">
                LedgerSpear is a revenue intelligence platform for Shopify app
                developers. The Service connects to the Shopify Partner API to
                retrieve subscription and transaction data, processes that data
                to calculate revenue metrics (MRR, churn, risk scores), and
                presents insights through a web dashboard and AI-powered
                briefings.
              </p>

              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                3. Account Registration
              </h2>
              <p className="text-slate-600 mb-4">
                To use the Service, you must create an account using Firebase
                Authentication. You are responsible for maintaining the
                confidentiality of your account credentials and for all
                activities that occur under your account. You must provide
                accurate and complete information during registration.
              </p>

              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                4. Shopify Partner API Access
              </h2>
              <p className="text-slate-600 mb-4">
                The Service requires you to authorize access to your Shopify
                Partner account data. By connecting your account, you confirm
                that you have the authority to grant this access and that your
                use complies with Shopify&apos;s Partner Program Agreement and API
                Terms of Service.
              </p>

              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                5. Acceptable Use
              </h2>
              <p className="text-slate-600 mb-4">You agree not to:</p>
              <ul className="list-disc pl-6 text-slate-600 mb-4 space-y-2">
                <li>
                  Use the Service for any unlawful purpose or in violation of
                  any applicable laws
                </li>
                <li>
                  Attempt to gain unauthorized access to the Service or its
                  related systems
                </li>
                <li>
                  Interfere with or disrupt the integrity or performance of the
                  Service
                </li>
                <li>
                  Reverse engineer, decompile, or disassemble any part of the
                  Service
                </li>
                <li>
                  Share your account credentials with third parties
                </li>
                <li>
                  Use the Service to collect data about other users without
                  their consent
                </li>
              </ul>

              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                6. Intellectual Property
              </h2>
              <p className="text-slate-600 mb-4">
                The Service, including its design, code, algorithms, and
                content, is owned by LedgerSpear and protected by intellectual
                property laws. You retain ownership of your data. We do not
                claim any rights over the Shopify Partner data you connect to
                the Service.
              </p>

              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                7. Data and Privacy
              </h2>
              <p className="text-slate-600 mb-4">
                Your use of the Service is also governed by our{' '}
                <a
                  href="/privacy"
                  className="text-blue-600 hover:text-blue-700 underline"
                >
                  Privacy Policy
                </a>
                , which describes how we collect, use, and protect your data.
              </p>

              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                8. Service Availability
              </h2>
              <p className="text-slate-600 mb-4">
                We strive to maintain high availability but do not guarantee
                uninterrupted access to the Service. We may perform maintenance,
                updates, or modifications that temporarily affect availability.
                The accuracy of revenue metrics depends on the data provided by
                the Shopify Partner API.
              </p>

              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                9. Limitation of Liability
              </h2>
              <p className="text-slate-600 mb-4">
                To the maximum extent permitted by law, LedgerSpear shall not be
                liable for any indirect, incidental, special, consequential, or
                punitive damages arising from your use of the Service. Our total
                liability shall not exceed the amount you paid for the Service
                in the twelve months preceding the claim. The Service provides
                revenue analytics for informational purposes and should not be
                the sole basis for financial decisions.
              </p>

              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                10. Termination
              </h2>
              <p className="text-slate-600 mb-4">
                You may terminate your account at any time by contacting us. We
                may suspend or terminate your access if you violate these Terms.
                Upon termination, your right to use the Service ceases
                immediately. We will retain your data for a reasonable period to
                allow data export, after which it will be deleted.
              </p>

              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                11. Changes to Terms
              </h2>
              <p className="text-slate-600 mb-4">
                We may update these Terms from time to time. We will notify you
                of material changes via email or through the Service. Continued
                use of the Service after changes constitutes acceptance of the
                updated Terms.
              </p>

              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                12. Governing Law
              </h2>
              <p className="text-slate-600 mb-4">
                These Terms are governed by and construed in accordance with the
                laws of India. Any disputes arising from these Terms shall be
                subject to the exclusive jurisdiction of the courts in
                Kochi, Kerala, India.
              </p>

              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                13. Contact
              </h2>
              <p className="text-slate-600 mb-4">
                If you have questions about these Terms, please contact us at{' '}
                <a
                  href="mailto:accounts@ledgerspear.com"
                  className="text-blue-600 hover:text-blue-700 underline"
                >
                  accounts@ledgerspear.com
                </a>
                .
              </p>
            </div>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}
