"use client";

import { useState, FormEvent } from "react";

const contactMethods = [
  {
    title: "Email",
    detail: "accounts@ledgerspear.com",
    href: "mailto:accounts@ledgerspear.com",
    description: "For general inquiries, partnerships, and support.",
    icon: (
      <svg
        className="w-6 h-6 text-blue-500"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
        />
      </svg>
    ),
  },
  {
    title: "Location",
    detail: "Kochi, India",
    description: "Operating remotely with a global focus on Shopify developers.",
    icon: (
      <svg
        className="w-6 h-6 text-emerald-500"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"
        />
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"
        />
      </svg>
    ),
  },
  {
    title: "Response Time",
    detail: "Within 24 hours",
    description:
      "We aim to respond to all inquiries within one business day.",
    icon: (
      <svg
        className="w-6 h-6 text-violet-500"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
        />
      </svg>
    ),
  },
];

export default function Contact() {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [message, setMessage] = useState("");

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const subject = encodeURIComponent(`Contact from ${name}`);
    const body = encodeURIComponent(
      `Name: ${name}\nEmail: ${email}\n\n${message}`
    );
    window.location.href = `mailto:accounts@ledgerspear.com?subject=${subject}&body=${body}`;
  }

  return (
    <div>
      <section className="py-20 lg:py-24 bg-white">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h1 className="text-4xl sm:text-5xl font-bold text-slate-900 mb-6">
              Contact Us
            </h1>
            <p className="text-xl text-slate-600 max-w-2xl mx-auto">
              Have a question about LedgerSpear? Want to discuss a partnership?
              We&apos;d love to hear from you.
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-8">
            {contactMethods.map((method) => (
              <div
                key={method.title}
                className="bg-slate-50 rounded-xl p-8 text-center"
              >
                <div className="inline-flex items-center justify-center w-12 h-12 rounded-full bg-white shadow-sm mb-4">
                  {method.icon}
                </div>
                <h3 className="text-lg font-semibold text-slate-900 mb-1">
                  {method.title}
                </h3>
                {method.href ? (
                  <a
                    href={method.href}
                    className="text-blue-600 font-medium hover:text-blue-700 transition-colors"
                  >
                    {method.detail}
                  </a>
                ) : (
                  <p className="text-slate-900 font-medium">{method.detail}</p>
                )}
                <p className="text-slate-500 text-sm mt-2">
                  {method.description}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Contact Form */}
      <section className="py-20 lg:py-24 bg-slate-50">
        <div className="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8">
          <h2 className="text-3xl font-bold text-slate-900 text-center mb-4">
            Send Us a Message
          </h2>
          <p className="text-slate-600 text-center mb-12">
            Fill out the form below and we&apos;ll get back to you promptly.
          </p>

          <form onSubmit={handleSubmit} className="bg-white rounded-xl p-8 space-y-6">
            <div>
              <label
                htmlFor="name"
                className="block text-sm font-medium text-slate-700 mb-2"
              >
                Name
              </label>
              <input
                type="text"
                id="name"
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full px-4 py-3 rounded-lg border border-slate-300 text-slate-900 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="Your name"
              />
            </div>

            <div>
              <label
                htmlFor="email"
                className="block text-sm font-medium text-slate-700 mb-2"
              >
                Email
              </label>
              <input
                type="email"
                id="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full px-4 py-3 rounded-lg border border-slate-300 text-slate-900 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="you@example.com"
              />
            </div>

            <div>
              <label
                htmlFor="message"
                className="block text-sm font-medium text-slate-700 mb-2"
              >
                Message
              </label>
              <textarea
                id="message"
                required
                rows={5}
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                className="w-full px-4 py-3 rounded-lg border border-slate-300 text-slate-900 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
                placeholder="How can we help?"
              />
            </div>

            <button
              type="submit"
              className="w-full py-3 px-6 text-white font-semibold bg-blue-600 rounded-lg hover:bg-blue-700 transition-colors"
            >
              Send Message
            </button>
          </form>
        </div>
      </section>
    </div>
  );
}
