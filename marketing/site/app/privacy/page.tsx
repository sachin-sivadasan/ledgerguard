import { Metadata } from 'next';
import Header from '@/components/Header';
import Footer from '@/components/Footer';

export const metadata: Metadata = {
  title: 'Privacy Policy | LedgerSpear',
  description: 'LedgerSpear Privacy Policy — how we collect, use, and protect your data.',
  openGraph: {
    title: 'Privacy Policy | LedgerSpear',
    description: 'LedgerSpear Privacy Policy.',
    type: 'website',
  },
};

export default function PrivacyPage() {
  return (
    <>
      <Header />
      <main className="pt-16">
        <div className="py-20 lg:py-24 bg-white">
          <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8">
            <h1 className="text-4xl font-bold text-slate-900 mb-2">
              Privacy Policy
            </h1>
            <p className="text-slate-500 mb-12">
              Last updated: March 16, 2026
            </p>

            <div className="prose prose-slate max-w-none">
              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                1. Introduction
              </h2>
              <p className="text-slate-600 mb-4">
                LedgerSpear (&quot;we&quot;, &quot;us&quot;, or &quot;our&quot;) operates LedgerSpear, a
                revenue intelligence platform for Shopify app developers. This
                Privacy Policy explains how we collect, use, store, and protect
                your information when you use our Service.
              </p>

              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                2. Information We Collect
              </h2>

              <h3 className="text-xl font-medium text-slate-800 mt-6 mb-3">
                2.1 Account Information
              </h3>
              <p className="text-slate-600 mb-4">
                When you create an account via Firebase Authentication, we
                collect:
              </p>
              <ul className="list-disc pl-6 text-slate-600 mb-4 space-y-2">
                <li>Email address</li>
                <li>Display name (if provided)</li>
                <li>Authentication provider details (Google, email/password)</li>
              </ul>

              <h3 className="text-xl font-medium text-slate-800 mt-6 mb-3">
                2.2 Shopify Partner Data
              </h3>
              <p className="text-slate-600 mb-4">
                When you connect your Shopify Partner account, we access and
                store:
              </p>
              <ul className="list-disc pl-6 text-slate-600 mb-4 space-y-2">
                <li>
                  App subscription data (plan names, pricing, status, creation
                  and cancellation dates)
                </li>
                <li>
                  Transaction data (charge amounts, types, dates — for revenue
                  calculation)
                </li>
                <li>
                  App listing metadata (app name, app ID)
                </li>
                <li>
                  Merchant identifiers (anonymized shop references for
                  subscription tracking)
                </li>
              </ul>
              <p className="text-slate-600 mb-4">
                We do <strong>not</strong> access merchant store data, customer
                information, product listings, or order details. Our access is
                limited to the Shopify Partner API scope required for revenue
                analytics.
              </p>

              <h3 className="text-xl font-medium text-slate-800 mt-6 mb-3">
                2.3 Usage Data
              </h3>
              <p className="text-slate-600 mb-4">
                We may collect basic usage analytics such as pages visited,
                features used, and session duration to improve the Service.
              </p>

              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                3. How We Use Your Information
              </h2>
              <p className="text-slate-600 mb-4">We use your information to:</p>
              <ul className="list-disc pl-6 text-slate-600 mb-4 space-y-2">
                <li>Provide and maintain the Service</li>
                <li>
                  Calculate revenue metrics (MRR, ARR, churn rates, risk scores)
                </li>
                <li>Generate daily revenue snapshots and AI-powered insights</li>
                <li>Send service notifications and updates</li>
                <li>Improve and optimize the Service</li>
                <li>Respond to support requests</li>
              </ul>

              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                4. Data Storage and Security
              </h2>
              <p className="text-slate-600 mb-4">
                Your data is stored on Google Cloud Platform (GCP)
                infrastructure. We implement industry-standard security measures
                including:
              </p>
              <ul className="list-disc pl-6 text-slate-600 mb-4 space-y-2">
                <li>Encryption in transit (TLS/HTTPS)</li>
                <li>Encryption at rest for stored data</li>
                <li>
                  Firebase Authentication for secure access control
                </li>
                <li>
                  Private database networking (no public IP access)
                </li>
              </ul>

              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                5. Shopify Data Handling
              </h2>
              <p className="text-slate-600 mb-4">
                We take special care with Shopify Partner data:
              </p>
              <ul className="list-disc pl-6 text-slate-600 mb-4 space-y-2">
                <li>
                  Data is fetched via the official Shopify Partner API using
                  your authorized credentials
                </li>
                <li>
                  Raw transaction data is stored immutably for audit purposes
                </li>
                <li>
                  Revenue ledgers are rebuilt deterministically from raw data —
                  the same input always produces the same output
                </li>
                <li>
                  Daily snapshots are retained as a permanent audit trail
                </li>
                <li>
                  We do not sell, share, or provide your Shopify data to third
                  parties
                </li>
              </ul>

              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                6. Data Retention
              </h2>
              <p className="text-slate-600 mb-4">
                We retain your data for as long as your account is active. Daily
                revenue snapshots are retained indefinitely as part of the
                audit trail. If you delete your account, we will remove your
                personal data and Shopify Partner data within 30 days, except
                where retention is required by law.
              </p>

              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                7. Third-Party Services
              </h2>
              <p className="text-slate-600 mb-4">
                We use the following third-party services:
              </p>
              <ul className="list-disc pl-6 text-slate-600 mb-4 space-y-2">
                <li>
                  <strong>Google Cloud Platform</strong> — infrastructure and
                  data storage
                </li>
                <li>
                  <strong>Firebase</strong> — authentication and hosting
                </li>
                <li>
                  <strong>Shopify Partner API</strong> — revenue data source
                </li>
              </ul>
              <p className="text-slate-600 mb-4">
                Each third-party service has its own privacy policy governing
                their handling of data.
              </p>

              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                8. Your Rights
              </h2>
              <p className="text-slate-600 mb-4">You have the right to:</p>
              <ul className="list-disc pl-6 text-slate-600 mb-4 space-y-2">
                <li>
                  <strong>Access</strong> — Request a copy of the personal data
                  we hold about you
                </li>
                <li>
                  <strong>Correction</strong> — Request correction of inaccurate
                  data
                </li>
                <li>
                  <strong>Deletion</strong> — Request deletion of your account
                  and associated data
                </li>
                <li>
                  <strong>Export</strong> — Request an export of your data in a
                  portable format
                </li>
                <li>
                  <strong>Disconnect</strong> — Revoke Shopify Partner API
                  access at any time
                </li>
              </ul>
              <p className="text-slate-600 mb-4">
                To exercise any of these rights, contact us at{' '}
                <a
                  href="mailto:accounts@ledgerspear.com"
                  className="text-blue-600 hover:text-blue-700 underline"
                >
                  accounts@ledgerspear.com
                </a>
                .
              </p>

              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                9. Cookies
              </h2>
              <p className="text-slate-600 mb-4">
                We use essential cookies for authentication and session
                management. We do not use third-party advertising or tracking
                cookies.
              </p>

              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                10. Children&apos;s Privacy
              </h2>
              <p className="text-slate-600 mb-4">
                The Service is not intended for use by individuals under the age
                of 18. We do not knowingly collect personal information from
                children.
              </p>

              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                11. Changes to This Policy
              </h2>
              <p className="text-slate-600 mb-4">
                We may update this Privacy Policy from time to time. We will
                notify you of material changes via email or through the Service.
                Continued use of the Service after changes constitutes
                acceptance of the updated policy.
              </p>

              <h2 className="text-2xl font-semibold text-slate-900 mt-10 mb-4">
                12. Contact
              </h2>
              <p className="text-slate-600 mb-4">
                If you have questions about this Privacy Policy or how we handle
                your data, please contact us at{' '}
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
