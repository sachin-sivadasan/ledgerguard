import '../models/playbook_model.dart';

final mockPlaybooks = <RecoveryPlaybook>[
  const RecoveryPlaybook(
    id: 'pb-1',
    name: 'Re-engagement Email Sequence',
    description: 'Send a 3-part email sequence highlighting new features and offering a discount.',
    successRate: 0.42,
    steps: [
      RecoveryAction(label: 'Send "We miss you" email', description: 'Personalized email with usage stats and new features since they left.'),
      RecoveryAction(label: 'Wait 3 days', description: 'Allow time for the merchant to engage.'),
      RecoveryAction(label: 'Send feature highlight', description: 'Showcase the top 3 features released this quarter.'),
      RecoveryAction(label: 'Offer 20% discount', description: 'Send a limited-time 20% off coupon for the next billing cycle.'),
    ],
  ),
  const RecoveryPlaybook(
    id: 'pb-2',
    name: 'Direct Outreach Call',
    description: 'Personal phone/video call to understand pain points and offer solutions.',
    successRate: 0.58,
    steps: [
      RecoveryAction(label: 'Research the store', description: 'Review their usage history, last interactions, and any support tickets.'),
      RecoveryAction(label: 'Schedule a call', description: 'Send a calendar invite for a 15-minute check-in.'),
      RecoveryAction(label: 'Conduct the call', description: 'Ask about pain points, demo new features, offer migration help.'),
      RecoveryAction(label: 'Follow up', description: 'Send a summary email with action items and any offered discounts.'),
    ],
  ),
  const RecoveryPlaybook(
    id: 'pb-3',
    name: 'Plan Downgrade Offer',
    description: 'Offer a lower-tier plan to retain the customer rather than lose them entirely.',
    successRate: 0.35,
    steps: [
      RecoveryAction(label: 'Analyze usage', description: 'Check if the merchant is using features above the Basic tier.'),
      RecoveryAction(label: 'Send downgrade offer', description: 'Email offering a switch to the Basic plan at current pricing.'),
      RecoveryAction(label: 'Process if accepted', description: 'Update the subscription to the lower tier immediately.'),
    ],
  ),
];
