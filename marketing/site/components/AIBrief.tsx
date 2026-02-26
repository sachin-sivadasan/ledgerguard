export default function AIBrief() {
  return (
    <section className="py-20 lg:py-24 bg-slate-900">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-12">
          <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-blue-500/10 text-blue-400 border border-blue-500/20">
            Pro Feature
          </span>
          <h2 className="mt-6 text-3xl sm:text-4xl font-bold text-white">
            Your Daily Revenue Brief, Powered by AI
          </h2>
          <p className="mt-4 text-lg text-slate-400 max-w-2xl mx-auto">
            Every morning, get an 80-word executive summary of your revenue
            health. What changed, what needs attention, and what&apos;s
            trending—delivered to your inbox or Slack.
          </p>
        </div>
        <div className="max-w-2xl mx-auto">
          <div className="bg-slate-800 rounded-2xl p-6 sm:p-8 border border-slate-700">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-10 h-10 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
                <svg
                  className="w-5 h-5 text-white"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M13 10V3L4 14h7v7l9-11h-7z"
                  />
                </svg>
              </div>
              <div>
                <p className="text-sm font-medium text-white">Daily Revenue Brief</p>
                <p className="text-xs text-slate-500">Today at 8:00 AM</p>
              </div>
            </div>
            <div className="text-slate-300 leading-relaxed">
              <p>
                Your MRR is <span className="text-green-400 font-medium">$12,450</span>,
                up 3.2% from last week. <span className="text-amber-400 font-medium">2 subscriptions</span> moved
                to at-risk status yesterday—both from annual plans approaching renewal.
                Consider reaching out to <span className="text-white font-medium">store-abc.myshopify.com</span> and{" "}
                <span className="text-white font-medium">mega-store.myshopify.com</span>.
                Renewal success rate remains strong at 94.2%.
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
