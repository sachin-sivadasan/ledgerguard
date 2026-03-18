# Stripe India — Invite Request Document

**Prepared for:** Stripe India Beta Access Request
**Date:** 2026-03-09 (updated 2026-03-17)
**Status:** In Progress — Stripe contact form submitted, GST registration pending

---

## 1. Business Information

| Field | Value |
|-------|-------|
| Business Name | [YOUR BUSINESS NAME] |
| Business Type | [Sole Proprietorship / LLP / Pvt Ltd] |
| Registered Address | [YOUR REGISTERED ADDRESS] |
| State | [STATE] |
| PIN Code | [PIN CODE] |
| PAN Number | [YOUR PAN] |
| GSTIN | [YOUR GSTIN — or "Applying"] |
| Website | https://ledgerspear.com |
| Contact Email | accounts@ledgerspear.com |
| Contact Phone | [YOUR PHONE] |

---

## 2. Business Description

**What does your business do?**

LedgerGuard is a Revenue Intelligence Platform for Shopify App Developers (B2B SaaS). It connects to the Shopify Partner API to provide app developers with real-time revenue analytics, subscription health monitoring, churn risk detection, and AI-powered insights for their Shopify app portfolio.

**Who are your customers?**

Independent software developers and companies who build and sell apps on the Shopify App Store. These are technology businesses (B2B) — not merchants or end consumers.

**What problem does it solve?**

Shopify app developers lack visibility into subscription health, churn risk, and revenue trends across their app portfolio. LedgerGuard aggregates Partner API data, classifies revenue by type (recurring, usage, one-time), detects at-risk subscriptions, and provides actionable intelligence to reduce churn and grow revenue.

---

## 3. Stripe Use Case

**What will you use Stripe for?**

Subscription billing for LedgerGuard's own SaaS plans. We will charge customers a monthly/annual subscription fee to access the platform.

**Billing model:**

| Plan | Price | Billing |
|------|-------|---------|
| Starter | $249/month | Monthly or Annual |
| Pro | $499/month | Monthly or Annual |

- New users get a **14-day free trial** on the Starter plan (no credit card required upfront)
- After trial, users subscribe via **Stripe Checkout** (hosted payment page)
- Self-service plan management via **Stripe Customer Portal**
- Upgrades are immediate with **proration**
- Downgrades are scheduled for the **end of billing period**

**Stripe products we plan to use:**

- Stripe Checkout (hosted payment page — no PCI burden)
- Stripe Billing (recurring subscriptions)
- Stripe Customer Portal (self-service plan management)
- Stripe Webhooks (subscription lifecycle events)

---

## 4. Payment & Volume Details

| Field | Value |
|-------|-------|
| Currency charged to customers | USD |
| Expected payout currency | INR |
| Average transaction size | $249–$499/month per customer |
| Expected monthly volume (Year 1) | $500–$5,000/month |
| Expected monthly volume (Year 2) | $5,000–$25,000/month |
| Payment methods needed | Credit/Debit Cards |
| Customer geography | Global (US, EU, India, SEA) |
| Refund policy | Pro-rated refunds within billing period |
| Business model | B2B SaaS subscription (recurring) |

---

## 5. Technical Integration

| Field | Value |
|-------|-------|
| Backend | Go (Golang) |
| Frontend | Flutter (Web + Mobile) |
| Hosting | GCP Cloud Run (staging), Hetzner (production) |
| Auth | Firebase Authentication |
| Integration approach | Stripe Checkout (hosted) + Webhooks |
| PCI compliance | SAQ-A (using Stripe Checkout, no card data touches our servers) |

**Webhook events we will handle:**
- `checkout.session.completed` — New subscription created
- `invoice.paid` — Recurring payment succeeded
- `invoice.payment_failed` — Payment failed, trigger dunning
- `customer.subscription.updated` — Plan changed
- `customer.subscription.deleted` — Subscription canceled

---

## 6. Why Stripe (Not Alternatives)

We evaluated multiple billing options:

| Option | Why Not |
|--------|---------|
| Shopify Billing API | Only works for merchant-installed apps. LedgerGuard is a partner-facing tool — no shop installation, no OAuth app context. Shopify Partner API has no billing endpoints. |
| Razorpay | Limited international subscription support; our customers are global (USD billing) |
| PayU | Primarily India-domestic; weak subscription/recurring billing features |
| Paddle | Higher fees; limited control over subscription lifecycle |

**Stripe is the right fit because:**
- Industry-standard for B2B SaaS subscription billing
- Stripe Checkout eliminates PCI compliance burden
- Customer Portal for self-service (reduces support load)
- Robust webhook system for real-time subscription state sync
- Supports USD billing with INR payouts
- Proration, trials, and dunning built-in

---

## 7. Compliance & Risk

| Field | Value |
|-------|-------|
| Business registration | [Registered under — Companies Act / LLP Act / Shops & Establishments Act] |
| Expected chargeback rate | <0.1% (B2B SaaS, no physical goods) |
| Fraud risk | Low (B2B customers, known identities, Firebase Auth) |
| Prohibited content | None — analytics SaaS platform |
| Terms of Service | https://ledgerspear.com/terms |
| Privacy Policy | https://ledgerspear.com/privacy |

---

## 8. Payment Provider Onboarding Q&A

**Reusable answers for Razorpay, Stripe, or any payment provider verification.**

| # | Question | Answer |
|---|----------|--------|
| 1 | What products or services do you offer? | We offer a B2B SaaS subscription service. LedgerSpear is a revenue intelligence platform for Shopify app developers. We provide monthly subscription plans ($249/month and $499/month) for analytics, churn risk detection, and AI-powered revenue insights. No physical goods — purely software delivered via web app. |
| 2 | Are your customers exclusively other businesses, or do you also serve individual consumers? | Our customers are exclusively other businesses — specifically Shopify app developers and software companies who build and sell apps on the Shopify App Store. We do not serve individual consumers. This is a B2B SaaS product. |
| 3 | Do you process payments only for your own SaaS platform, or do you also facilitate payments on behalf of your clients? | We process payments only for our own SaaS platform. We charge customers a monthly subscription fee for access to LedgerSpear. We do not facilitate payments on behalf of our clients — we are not a payment gateway or marketplace. |
| 4 | Is your SaaS platform focused solely on analytics and revenue intelligence, or does it also include other features such as CRM, marketing automation, or project management? | Our SaaS platform is focused solely on revenue analytics and intelligence for Shopify app developers. Core features include MRR tracking, subscription renewal monitoring, churn risk scoring, and AI-powered revenue briefs. We do not include CRM, marketing automation, or project management features. |
| 5 | Business category? | IT and software > Software as a Service (SaaS) |
| 6 | RBI Purpose Code for international payments? | **P0802** - Software consultancy/implementation. Standard code for SaaS businesses receiving international payments in India. Covers all software-related international receipts regardless of B2B/B2C. |

---

## 8. Documents Ready for Upload

Prepare these documents before applying:

- [ ] PAN Card (Business or Individual)
- [ ] Certificate of Incorporation / GST Registration / Shop & Establishment Certificate
- [ ] Bank account proof (cancelled cheque or bank statement)
- [ ] Address proof (utility bill or rental agreement)
- [ ] Identity proof (Aadhaar / Passport of authorized signatory)
- [ ] GSTIN certificate (if registered)
- [ ] Website with Terms of Service and Privacy Policy live

---

## 9. Request Email Template

> **Subject:** Stripe India Beta Access Request — LedgerGuard (B2B SaaS)
>
> Hi Stripe Team,
>
> I'm building **LedgerGuard**, a Revenue Intelligence Platform for Shopify App Developers. It's a B2B SaaS product that helps app developers monitor subscription health, detect churn risk, and optimize revenue.
>
> I'd like to request access to **Stripe India** for subscription billing. Here's a summary:
>
> - **Business type:** [Sole Proprietorship / LLP / Pvt Ltd]
> - **Use case:** Monthly/annual SaaS subscription billing via Stripe Checkout
> - **Customers:** Global (Shopify app developers), charged in USD
> - **Expected volume:** $500–$5,000/month initially, growing to $25,000+/month
> - **Products needed:** Stripe Checkout, Billing, Customer Portal, Webhooks
> - **Website:** https://ledgerguard.com
>
> I have all required documents ready (PAN, registration certificate, bank details). Happy to provide any additional information.
>
> Looking forward to hearing from you.
>
> Best regards,
> [YOUR NAME]
> [YOUR EMAIL]
> [YOUR PHONE]

---

## 10. Alternative Path: US LLC + Stripe US

If Stripe India invite takes too long or doesn't work out:

| Step | Action |
|------|--------|
| 1 | Form a US LLC (Delaware or Wyoming) via Stripe Atlas ($500) or Firstbase ($399) |
| 2 | Get EIN (free, IRS Form SS-4) |
| 3 | Open US business bank account (Mercury, Relay, or Wise Business) |
| 4 | Sign up for Stripe US directly (no invite needed) |
| 5 | Charge USD, receive USD in US account |
| 6 | Transfer to India via Wise / bank wire as needed |

**Pros:** No invite wait, full Stripe feature access, US entity looks professional for global B2B
**Cons:** $400–$500 setup cost, annual LLC maintenance ($50–$300/year), foreign remittance compliance

---

---

## 11. GST Registration Steps (Pre-requisite for Stripe India)

**Status:** Not started

### What you need before applying:
- PAN card (personal — for sole proprietorship)
- Aadhaar card
- Bank account statement or cancelled cheque (savings account is fine initially)
- Address proof for business (rent agreement + NOC from owner, OR electricity bill in your name)
- Passport-size photo
- Mobile number linked to Aadhaar (for OTP)

### Steps:

| Step | Action | Time |
|------|--------|------|
| 1 | Go to https://reg.gst.gov.in → "New Registration" | 5 min |
| 2 | Select "Taxpayer" → Fill Part A (PAN, mobile, email) → Get TRN | 10 min |
| 3 | Fill Part B with TRN → Business details, bank info, upload docs | 30 min |
| 4 | E-sign with Aadhaar OTP or DSC | 5 min |
| 5 | Wait for GSTIN approval | **7-10 working days** |
| 6 | Download GST certificate from portal | 5 min |

### Key fields for GST form:
| Field | Value |
|-------|-------|
| Type of registration | Regular |
| Constitution of business | Sole Proprietorship |
| Trade name | LedgerSpear |
| Principal place of business | Your address in Kochi |
| Nature of business | IT Software Services |
| HSN/SAC code | **998314** (IT consulting & software services) |

### After GST is approved:
1. Open a **current account** (business bank account) using GST cert + PAN
2. Sign up for **Razorpay** (primary) — no invite needed, fast onboarding
3. Complete **Stripe India** application (secondary) — if/when invite arrives

### Cost:
- GST registration: **Free** (govt portal) or ~2K-3K INR via a CA
- Current account: Free (most banks)

---

## 12. Razorpay as Primary Payment Provider

**Why Razorpay first:**
- No invite needed (Stripe India is invite-only)
- Same pricing (2% domestic, 3% international)
- Supports recurring subscriptions
- Faster onboarding (~2-3 days after docs)
- Can migrate to Stripe later if needed

### Razorpay Subscription Integration

**API endpoints we'll use:**

| Razorpay Feature | Purpose |
|------------------|---------|
| Plans API | Create Starter ($249) and Pro ($499) plans |
| Subscriptions API | Subscribe customers to plans |
| Hosted Checkout | Payment page (no PCI burden) |
| Webhooks | Subscription lifecycle events |

**Webhook events to handle:**
- `subscription.activated` — New subscription started
- `subscription.charged` — Recurring payment succeeded
- `subscription.pending` — Payment pending/retrying
- `subscription.halted` — Payment failed after retries
- `subscription.cancelled` — Subscription cancelled
- `subscription.completed` — Subscription ended

**What we need to build ourselves (not built-in like Stripe):**
- Customer portal UI (plan management, cancel, upgrade)
- Proration logic on plan upgrades
- Invoice display page

### Razorpay Onboarding Requirements

| Document | Status |
|----------|--------|
| PAN card | Have |
| GSTIN certificate | Pending (need GST registration) |
| Business bank account (current) | Pending (need GST first) |
| Website with pricing, terms, privacy | Done (ledgerspear.com) |
| Cancelled cheque / bank statement | After opening current account |

### Razorpay Sign-up Steps

| Step | Action | Time |
|------|--------|------|
| 1 | Sign up at https://dashboard.razorpay.com/signup | 5 min |
| 2 | Use **test mode** immediately (no docs needed) | Instant |
| 3 | Submit KYC docs for live mode (after GST) | 2-3 days approval |
| 4 | Create Plans (Starter, Pro) | 10 min |
| 5 | Integrate checkout + webhooks in backend | Development time |

---

## 13. Action Items Tracker

| # | Task | Status | Deadline |
|---|------|--------|----------|
| 1 | Submit Stripe India contact form | Done (2026-03-17) | - |
| 2 | Sign up Razorpay (test mode) | Done (2026-03-18) | - |
| 3 | Apply for GST registration | Not started | ASAP |
| 4 | Get GST certificate | Blocked on #3 | ~7-10 days after #3 |
| 5 | Open current (business) bank account | Blocked on #4 | ~3 days after #4 |
| 6 | Razorpay KYC (go live) | Blocked on #4, #5 | ~2-3 days after #5 |
| 7 | Complete Stripe India onboarding (if invited) | Blocked on #4, #5 | When invite arrives |
| 8 | Integrate Razorpay subscriptions in backend | **Done** (2026-03-18) | - |
| 9 | Integrate Razorpay billing UI in Flutter | **Done** (2026-03-18) | - |
| 10 | E2E test: create plans in Razorpay dashboard, test checkout flow | Next | - |

---

### GST Ongoing Compliance Reminder

| Obligation | Frequency | Penalty if missed |
|------------|-----------|-------------------|
| GSTR-1 (sales) | Monthly | Rs 50/day |
| GSTR-3B (summary + tax) | Monthly | Rs 50/day (Rs 20 for nil) |
| GSTR-9 (annual) | Yearly | Rs 200/day (max 0.25% of turnover) |
| Income Tax Return | Yearly | Rs 5,000 late fee |

**Tip:** Hire a CA (~Rs 1,000-2,000/month) or use ClearTax/Zoho Books to automate filings.

**Export of services (customers outside India) = 0% GST** — but you must still file returns monthly.

---

*This document is for internal preparation. Fill in all [PLACEHOLDER] fields before submitting to Stripe/Razorpay.*
