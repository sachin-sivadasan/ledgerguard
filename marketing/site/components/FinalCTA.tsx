export default function FinalCTA() {
  return (
    <section className="py-20 lg:py-24 bg-blue-600">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
        <h2 className="text-3xl sm:text-4xl font-bold text-white">
          Ready to Protect Your Revenue?
        </h2>
        <p className="mt-4 text-lg text-blue-100">
          Connect your Shopify Partner account in 60 seconds. No credit card
          required.
        </p>
        <a
          href="#"
          className="mt-8 inline-flex items-center justify-center px-8 py-4 text-lg font-semibold text-blue-600 bg-white rounded-lg hover:bg-blue-50 transition-colors"
        >
          Connect Shopify Partner
          <svg
            className="ml-2 w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M13 7l5 5m0 0l-5 5m5-5H6"
            />
          </svg>
        </a>
      </div>
    </section>
  );
}
