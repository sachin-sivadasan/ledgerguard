export default function Hero() {
  return (
    <section className="bg-gradient-to-b from-slate-50 to-white py-20 lg:py-32">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
        <h1 className="text-4xl sm:text-5xl lg:text-6xl font-bold text-slate-900 tracking-tight">
          Stop Guessing.{" "}
          <span className="text-blue-600">Start Knowing.</span>
        </h1>
        <p className="mt-6 text-lg sm:text-xl text-slate-600 max-w-3xl mx-auto">
          Revenue intelligence for Shopify app developers. Track renewals,
          predict churn, and protect your MRR.
        </p>
        <div className="mt-10 flex flex-col sm:flex-row gap-4 justify-center">
          <a
            href="#pricing"
            className="inline-flex items-center justify-center px-8 py-4 text-lg font-semibold text-white bg-blue-600 rounded-lg hover:bg-blue-700 transition-colors"
          >
            Connect Shopify Partner
          </a>
          <a
            href="#features"
            className="inline-flex items-center justify-center px-8 py-4 text-lg font-semibold text-slate-700 bg-white border border-slate-300 rounded-lg hover:bg-slate-50 transition-colors"
          >
            See How It Works
          </a>
        </div>
      </div>
    </section>
  );
}
