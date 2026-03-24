import Image from "next/image";

const screenshots = [
  {
    src: "/screenshots/dashboard.png",
    alt: "LedgerSpear Revenue Dashboard",
    caption: "Revenue Dashboard",
    description: "MRR, renewal rate, and risk overview at a glance",
  },
  {
    src: "/screenshots/subscriptions.png",
    alt: "LedgerSpear Subscription Management",
    caption: "Subscription Management",
    description: "Active subscriptions with status and risk badges",
  },
  {
    src: "/screenshots/risk.png",
    alt: "LedgerSpear Risk Analysis",
    caption: "Store Health & Risk",
    description: "Store health scores and at-risk merchant identification",
  },
];

export default function ProductScreenshots() {
  return (
    <section className="py-20 lg:py-24 bg-white" id="screenshots">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <h2 className="text-3xl sm:text-4xl font-bold text-slate-900 mb-4">
            See LedgerSpear in Action
          </h2>
          <p className="text-lg text-slate-600 max-w-2xl mx-auto">
            A cross-platform app built with Flutter — available on web, iOS, and Android.
          </p>
        </div>

        <div className="grid md:grid-cols-3 gap-8">
          {screenshots.map((shot) => (
            <div key={shot.caption} className="group">
              <div className="relative aspect-[16/10] rounded-xl overflow-hidden bg-slate-100 border border-slate-200 shadow-sm">
                <Image
                  src={shot.src}
                  alt={shot.alt}
                  fill
                  className="object-cover object-top"
                  sizes="(max-width: 768px) 100vw, 50vw"
                />
              </div>
              <div className="mt-4">
                <h3 className="text-lg font-semibold text-slate-900">
                  {shot.caption}
                </h3>
                <p className="text-sm text-slate-600">{shot.description}</p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
