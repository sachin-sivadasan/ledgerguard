import '../models/review_model.dart';

ReviewSentiment _sentiment(int rating) {
  if (rating <= 2) return ReviewSentiment.negative;
  if (rating == 3) return ReviewSentiment.neutral;
  return ReviewSentiment.positive;
}

/// 50 real reviews scraped from the Shopify App Store (Zoko app, 5 pages).
/// Distributed across app-1 (Zoko), app-2 (ReviewBoost), app-3 (ShipTracker).
final mockReviews = <AppReview>[
  // ── app-1 (Zoko) — 30 reviews ──
  AppReview(
    id: 'rev-1', appId: 'app-1', author: 'K-JO', rating: 5,
    text: 'We\'ve had a very positive experience with Zoko so far. The onboarding process was smooth, and their team was extremely supportive. They took the time to answer all my questions and made sure everything was properly set up.\n\nSince implementing Zoko, our WhatsApp conversations have become much more organized and efficient. The integration with Shopify works very well, and the overall usage has been simple and reliable for our daily operations.\n\nIt has definitely improved the way we communicate with our customers.\n\nI\'m very happy with the results and would confidently recommend Zoko to other Shopify store owners.',
    date: DateTime(2026, 2, 26), sentiment: _sentiment(5),
    location: 'United Arab Emirates', timeUsing: '4 months using the app',
  ),
  AppReview(
    id: 'rev-2', appId: 'app-1', author: 'NIIMBOT UK', rating: 5,
    text: 'The most incredible app that integrates so well with Shopify. Our customers love being kept up to date with their shipping notifications via WhatsApp. The team at Zoko are incredibly responsive and helpful.',
    date: DateTime(2026, 2, 26), sentiment: _sentiment(5),
    location: 'United Kingdom', timeUsing: '3 months using the app',
  ),
  AppReview(
    id: 'rev-3', appId: 'app-1', author: 'FeetPicks', rating: 5,
    text: 'We have been using zoko for a few months now and I must say its amazing. Especially the team - Aaquib has been extremely helpful and responsive to all our queries.',
    date: DateTime(2026, 2, 27), sentiment: _sentiment(5),
    location: 'India', timeUsing: '3 months using the app',
  ),
  AppReview(
    id: 'rev-4', appId: 'app-1', author: 'OriginalLympBrush', rating: 5,
    text: 'Zoko has been a game-changer for our Shopify store. The AI features are incredible, they handle customer queries instantly and accurately. The WhatsApp integration is seamless.',
    date: DateTime(2026, 3, 3), sentiment: _sentiment(5),
    location: 'India', timeUsing: '3 months using the app',
  ),
  AppReview(
    id: 'rev-5', appId: 'app-1', author: 'Ecosand Romania', rating: 5,
    text: 'Very good product and great support. They helped me with initial setup and even quickly added an extra language that was missing. Highly recommended!',
    date: DateTime(2026, 2, 23), sentiment: _sentiment(5),
    location: 'Romania', timeUsing: '2 months using the app',
  ),
  AppReview(
    id: 'rev-6', appId: 'app-1', author: 'Aishwaryam Oils', rating: 5,
    text: 'Zoko has delivered exactly what we expected, Special thanks to Muzaina Tasneem for her proactive and outstanding support throughout our journey.',
    date: DateTime(2026, 2, 17), sentiment: _sentiment(5),
    location: 'India', timeUsing: '2 months using the app',
  ),
  AppReview(
    id: 'rev-7', appId: 'app-1', author: 'Peacoski', rating: 4,
    text: 'I recently started using Zoko and its a very useful software. My customer interactions and management has got so much easier with Zoko.',
    date: DateTime(2026, 2, 24), sentiment: _sentiment(4),
    location: 'India', timeUsing: '2 months using the app',
  ),
  AppReview(
    id: 'rev-8', appId: 'app-1', author: 'Oje\' Living', rating: 5,
    text: 'Really smooth experience with Zoko so far. It\'s made WhatsApp marketing and customer communication much easier for us, and the Shopify integration is solid.',
    date: DateTime(2026, 3, 19), sentiment: _sentiment(5),
    location: 'India', timeUsing: 'About 1 month using the app',
  ),
  AppReview(
    id: 'rev-9', appId: 'app-1', author: 'The resha / الرِّيشَة', rating: 5,
    text: 'We had a great experience working with Zoko. Their team is very professional, supportive, and always ready to help whenever needed.',
    date: DateTime(2026, 2, 13), sentiment: _sentiment(5),
    location: 'United Arab Emirates', timeUsing: 'About 1 month using the app',
  ),
  AppReview(
    id: 'rev-10', appId: 'app-1', author: 'Flurr', rating: 3,
    text: 'We\'ve been using Zoko for Flurr, and it has streamlined our WhatsApp communication somewhat. From order updates to customer support, it works but there\'s room for improvement.',
    date: DateTime(2026, 2, 23), sentiment: _sentiment(3),
    location: 'India', timeUsing: 'About 1 month using the app',
  ),
  AppReview(
    id: 'rev-11', appId: 'app-1', author: 'NIIMBOT.CO.ZA', rating: 5,
    text: 'Just the best app ever!! Super efficient with great ROI results. Zoko\'s customer support is brilliant and their team are always there to help.',
    date: DateTime(2026, 2, 18), sentiment: _sentiment(5),
    location: 'South Africa', timeUsing: 'About 3 years using the app',
  ),
  AppReview(
    id: 'rev-12', appId: 'app-1', author: 'THE DUNGEON GEAR', rating: 5,
    text: 'I\'m extremely pleased with Zoko, especially with the outstanding support from Sevin and Arpit. They consistently take time to address our needs.',
    date: DateTime(2025, 11, 29), sentiment: _sentiment(5),
    location: 'United Arab Emirates', timeUsing: 'About 1 year using the app',
  ),
  AppReview(
    id: 'rev-13', appId: 'app-1', author: 'Yo Baby India', rating: 5,
    text: 'Zoko has been an excellent choice for my business specially since they are other options as well when it comes to WhatsApp API providers. Great support team.',
    date: DateTime(2025, 12, 18), sentiment: _sentiment(5),
    location: 'India', timeUsing: '2 months using the app',
  ),
  AppReview(
    id: 'rev-14', appId: 'app-1', author: 'aavyaa', rating: 2,
    text: 'The integration was harder to set up than expected. Documentation could be better. Support eventually helped but it took a while to get things working.',
    date: DateTime(2025, 12, 5), sentiment: _sentiment(2),
    location: 'India', timeUsing: '3 months using the app',
  ),
  AppReview(
    id: 'rev-15', appId: 'app-1', author: 'Clothivia', rating: 5,
    text: 'We\'ve been using Zoko for some time now, and it\'s been working amazingly for us! The platform makes it super easy to manage WhatsApp conversations.',
    date: DateTime(2025, 11, 30), sentiment: _sentiment(5),
    location: 'Qatar', timeUsing: '3 months using the app',
  ),
  AppReview(
    id: 'rev-16', appId: 'app-1', author: 'Nutrova', rating: 5,
    text: 'Zoko has made it incredibly easy for us to integrate WhatsApp and Instagram with our business and communicate seamlessly with our customers.',
    date: DateTime(2025, 11, 20), sentiment: _sentiment(5),
    location: 'India', timeUsing: '8 months using the app',
  ),
  AppReview(
    id: 'rev-17', appId: 'app-1', author: 'Retail Maharaj', rating: 4,
    text: 'The customer service of the team is very good. They understand your issue and provide solution that will cater to your requirements perfectly.',
    date: DateTime(2026, 3, 16), sentiment: _sentiment(4),
    location: 'India', timeUsing: '3 months using the app',
  ),
  AppReview(
    id: 'rev-18', appId: 'app-1', author: 'Shahd Beauty', rating: 5,
    text: 'Zoko is one of the most powerful WhatsApp automation platforms I\'ve worked with. The integration with Shopify is seamless and very reliable.',
    date: DateTime(2025, 12, 18), sentiment: _sentiment(5),
    location: 'United Arab Emirates', timeUsing: '8 days using the app',
  ),
  AppReview(
    id: 'rev-19', appId: 'app-1', author: 'TodsNTeens', rating: 5,
    text: 'Switched to Zoko recently and I couldn\'t be happier. One of the nicest decisions I\'ve made so far. Totally organizes chats and makes marketing easy.',
    date: DateTime(2025, 11, 26), sentiment: _sentiment(5),
    location: 'Pakistan', timeUsing: 'About 2 months using the app',
  ),
  AppReview(
    id: 'rev-20', appId: 'app-1', author: 'Froggmag', rating: 4,
    text: 'Our experience with the Zoko team has been very positive. We\'re still exploring the platform\'s full potential, but the support has been excellent so far.',
    date: DateTime(2026, 2, 4), sentiment: _sentiment(4),
    location: 'India', timeUsing: '3 months using the app',
  ),
  AppReview(
    id: 'rev-21', appId: 'app-1', author: 'DesiVidesi', rating: 5,
    text: 'Zoko has been an excellent addition to our business, but what has truly made the experience outstanding is the support we receive from their team.',
    date: DateTime(2025, 11, 24), sentiment: _sentiment(5),
    location: 'United Arab Emirates', timeUsing: 'About 2 months using the app',
  ),
  AppReview(
    id: 'rev-22', appId: 'app-1', author: 'Artiqulate Lifestyle', rating: 5,
    text: 'I really appreciate Zoko and the team for all the help they have given my company for our WhatsApp Marketing. The automation flows are powerful and easy to set up.',
    date: DateTime(2026, 2, 11), sentiment: _sentiment(5),
    location: 'India', timeUsing: 'About 2 months using the app',
  ),
  AppReview(
    id: 'rev-23', appId: 'app-1', author: 'House of Nirvana', rating: 1,
    text: 'Very disappointing experience. The app crashed multiple times during setup and the integration with our store was broken for days. Support was slow to respond.',
    date: DateTime(2025, 11, 20), sentiment: _sentiment(1),
    location: 'India', timeUsing: '3 months using the app',
  ),
  AppReview(
    id: 'rev-24', appId: 'app-1', author: 'NativeForever', rating: 4,
    text: 'Zoko is fantastic and we use it for order confirmation and recover cart also. It\'s incredibly easy to use and the Shopify integration works great.',
    date: DateTime(2025, 11, 19), sentiment: _sentiment(4),
    location: 'India', timeUsing: '3 months using the app',
  ),
  AppReview(
    id: 'rev-25', appId: 'app-1', author: 'Alaaya Decor', rating: 5,
    text: 'Amazing - 5 star service as of now! Kudos to Nishant and Team!',
    date: DateTime(2026, 2, 17), sentiment: _sentiment(5),
    location: 'India', timeUsing: 'About 1 month using the app',
  ),
  AppReview(
    id: 'rev-26', appId: 'app-1', author: 'Raw Mango', rating: 4,
    text: 'Well Zoko team is quite knowledgeable and friendly. The automation flows were quite easy to build and the setup UI is quite intuitive.',
    date: DateTime(2025, 11, 14), sentiment: _sentiment(4),
    location: 'India', timeUsing: '4 months using the app',
  ),
  AppReview(
    id: 'rev-27', appId: 'app-1', author: '24 Skin Bank - India', rating: 5,
    text: 'The service is amazing thanks to Muzaina Tasneem for helping out at every step. The AI does most of the work, the team looks after the rest.',
    date: DateTime(2026, 2, 25), sentiment: _sentiment(5),
    location: 'India', timeUsing: '28 days using the app',
  ),
  AppReview(
    id: 'rev-28', appId: 'app-1', author: 'Palamano', rating: 3,
    text: 'The app works okay for basic messaging but the advanced features are confusing. Documentation is lacking. Customer support is responsive though.',
    date: DateTime(2026, 2, 24), sentiment: _sentiment(3),
    location: 'Colombia', timeUsing: '2 months using the app',
  ),
  AppReview(
    id: 'rev-29', appId: 'app-1', author: 'The Golden Cascade', rating: 5,
    text: 'Zoko is a powerful and easy-to-use WhatsApp marketing tool. It helps businesses send broadcasts, automate messages, manage conversations efficiently.',
    date: DateTime(2025, 11, 14), sentiment: _sentiment(5),
    location: 'India', timeUsing: '3 months using the app',
  ),
  AppReview(
    id: 'rev-30', appId: 'app-1', author: 'Daily Taaza Food Products', rating: 5,
    text: 'I\'ve been using Zoko for my business, and the experience has been outstanding. What truly sets Zoko apart is their customer support team.',
    date: DateTime(2025, 11, 18), sentiment: _sentiment(5),
    location: 'India', timeUsing: 'About 2 months using the app',
  ),

  // ── app-2 (ReviewBoost) — 12 reviews ──
  AppReview(
    id: 'rev-31', appId: 'app-2', author: 'Statement Watches', rating: 5,
    text: 'We looked into various review solutions and finally settled on ReviewBoost since it integrates fully with Shopify and the pricing is fair.',
    date: DateTime(2025, 11, 24), sentiment: _sentiment(5),
    location: 'South Africa', timeUsing: '24 days using the app',
  ),
  AppReview(
    id: 'rev-32', appId: 'app-2', author: 'Nunuū', rating: 4,
    text: 'ReviewBoost has been a solid addition to our store. The review collection emails are well-designed and conversion rate for getting reviews improved significantly.',
    date: DateTime(2025, 11, 21), sentiment: _sentiment(4),
    location: 'Mexico', timeUsing: '24 days using the app',
  ),
  AppReview(
    id: 'rev-33', appId: 'app-2', author: 'Chappal', rating: 5,
    text: 'We are very pleased to have ReviewBoost integrated to our website. The photo review feature really helps build trust with customers.',
    date: DateTime(2026, 1, 16), sentiment: _sentiment(5),
    location: 'Türkiye', timeUsing: '3 months using the app',
  ),
  AppReview(
    id: 'rev-34', appId: 'app-2', author: 'Poshewix', rating: 3,
    text: 'The app works for basic review collection but the customization options are limited. Would love to see more template choices for the review request emails.',
    date: DateTime(2025, 11, 18), sentiment: _sentiment(3),
    location: 'Lebanon', timeUsing: 'About 1 month using the app',
  ),
  AppReview(
    id: 'rev-35', appId: 'app-2', author: 'Maltesitos', rating: 5,
    text: 'A Game-Changer for collecting social proof! The automated review requests have tripled our review count in just 2 months.',
    date: DateTime(2026, 1, 16), sentiment: _sentiment(5),
    location: 'Spain', timeUsing: '3 months using the app',
  ),
  AppReview(
    id: 'rev-36', appId: 'app-2', author: 'BioFrost', rating: 5,
    text: 'I\'m extremely happy with ReviewBoost. The interface is clean, intuitive, and the review widget on our product pages looks beautiful.',
    date: DateTime(2025, 11, 14), sentiment: _sentiment(5),
    location: 'India', timeUsing: 'About 1 month using the app',
  ),
  AppReview(
    id: 'rev-37', appId: 'app-2', author: 'Thou', rating: 2,
    text: 'The review import feature didn\'t work properly for us. Took several back-and-forth emails with support to get it resolved. Frustrating first experience.',
    date: DateTime(2025, 10, 30), sentiment: _sentiment(2),
    location: 'India', timeUsing: '7 months using the app',
  ),
  AppReview(
    id: 'rev-38', appId: 'app-2', author: 'flossycosmetics', rating: 5,
    text: 'ReviewBoost has been amazing! Our product pages now showcase real customer reviews with photos and it has noticeably improved our conversion rate.',
    date: DateTime(2025, 10, 27), sentiment: _sentiment(5),
    location: 'India', timeUsing: '8 months using the app',
  ),
  AppReview(
    id: 'rev-39', appId: 'app-2', author: 'Techmanistan', rating: 4,
    text: 'Solid review collection app. The analytics dashboard helps us understand which products need more reviews. Good value for the price.',
    date: DateTime(2025, 10, 29), sentiment: _sentiment(4),
    location: 'Pakistan', timeUsing: '6 months using the app',
  ),
  AppReview(
    id: 'rev-40', appId: 'app-2', author: 'Mojaa Maroc', rating: 5,
    text: 'Very nice and powerful review tool, with very good onboarding & assistance. The SEO benefits from structured review data are a great bonus.',
    date: DateTime(2026, 2, 26), sentiment: _sentiment(5),
    location: 'Morocco', timeUsing: '7 months using the app',
  ),
  AppReview(
    id: 'rev-41', appId: 'app-2', author: 'خنه', rating: 1,
    text: 'The app kept sending duplicate review request emails to our customers. Very embarrassing. Had to disable it until the bug was fixed.',
    date: DateTime(2025, 11, 24), sentiment: _sentiment(1),
    location: 'Kuwait', timeUsing: '3 months using the app',
  ),
  AppReview(
    id: 'rev-42', appId: 'app-2', author: 'True Mens', rating: 4,
    text: 'We\'ve been using ReviewBoost for some time now, and the interface is super easy to use. The review moderation tools are especially helpful.',
    date: DateTime(2025, 11, 21), sentiment: _sentiment(4),
    location: 'India', timeUsing: '4 months using the app',
  ),

  // ── app-3 (ShipTracker) — 8 reviews ──
  AppReview(
    id: 'rev-43', appId: 'app-3', author: 'Miniatures Shop', rating: 5,
    text: 'ShipTracker has been excellent for managing our shipping notifications. Customers love the real-time tracking page. The platform is intuitive and reliable.',
    date: DateTime(2025, 10, 29), sentiment: _sentiment(5),
    location: 'India', timeUsing: '18 days using the app',
  ),
  AppReview(
    id: 'rev-44', appId: 'app-3', author: 'Prickly pear me', rating: 5,
    text: 'We are super happy with ShipTracker. The branded tracking page has reduced our "where is my order?" support tickets by 60%.',
    date: DateTime(2025, 10, 13), sentiment: _sentiment(5),
    location: 'United Arab Emirates', timeUsing: '6 months using the app',
  ),
  AppReview(
    id: 'rev-45', appId: 'app-3', author: 'WayUp Sports', rating: 3,
    text: 'Tracking works well for major carriers but some regional carriers in our area are not supported. Would love to see more carrier integrations.',
    date: DateTime(2025, 11, 24), sentiment: _sentiment(3),
    location: 'Egypt', timeUsing: '2 months using the app',
  ),
  AppReview(
    id: 'rev-46', appId: 'app-3', author: 'Unfrosen.com', rating: 5,
    text: 'Switched from another tracking app to ShipTracker. The delivery notifications and tracking page are much better. Customers notice the difference.',
    date: DateTime(2025, 10, 14), sentiment: _sentiment(5),
    location: 'Romania', timeUsing: '5 months using the app',
  ),
  AppReview(
    id: 'rev-47', appId: 'app-3', author: 'Sip Club Company', rating: 4,
    text: 'Have been using ShipTracker for order tracking and delivery notifications. The estimated delivery dates feature is very helpful for our customers.',
    date: DateTime(2025, 10, 22), sentiment: _sentiment(4),
    location: 'India', timeUsing: '28 days using the app',
  ),
  AppReview(
    id: 'rev-48', appId: 'app-3', author: 'EveBeauty', rating: 5,
    text: 'The tracking page customization is excellent. We matched it perfectly with our brand colors. Customers love the estimated delivery countdown.',
    date: DateTime(2026, 2, 17), sentiment: _sentiment(5),
    location: 'Italy', timeUsing: '3 months using the app',
  ),
  AppReview(
    id: 'rev-49', appId: 'app-3', author: 'Scents Creation', rating: 2,
    text: 'Tracking updates are sometimes delayed by several hours compared to the actual carrier tracking. This confuses our customers. Needs improvement.',
    date: DateTime(2026, 1, 16), sentiment: _sentiment(2),
    location: 'United Arab Emirates', timeUsing: '3 months using the app',
  ),
  AppReview(
    id: 'rev-50', appId: 'app-3', author: 'Thinking Threads', rating: 5,
    text: 'ShipTracker has been great for us! The automated shipping notifications save us so much time and our customers always know where their orders are.',
    date: DateTime(2025, 10, 27), sentiment: _sentiment(5),
    location: 'India', timeUsing: '11 days using the app',
  ),
];
