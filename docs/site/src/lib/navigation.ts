export interface NavItem {
  title: string
  href?: string
  items?: NavItem[]
}

export const navigation: NavItem[] = [
  {
    title: 'Getting Started',
    items: [
      { title: 'Introduction', href: '/docs' },
      { title: 'Quick Start', href: '/docs/quickstart' },
      { title: 'Authentication', href: '/docs/authentication' },
    ],
  },
  {
    title: 'Core Concepts',
    items: [
      { title: 'Subscription Status', href: '/docs/concepts/subscription-status' },
      { title: 'Risk States', href: '/docs/concepts/risk-states' },
      { title: 'Usage Billing', href: '/docs/concepts/usage-billing' },
    ],
  },
  {
    title: 'REST API',
    items: [
      { title: 'Overview', href: '/docs/api/overview' },
      { title: 'Get Subscription', href: '/docs/api/subscriptions/get' },
      { title: 'Get by Domain', href: '/docs/api/subscriptions/get-by-domain' },
      { title: 'Batch Subscriptions', href: '/docs/api/subscriptions/batch' },
      { title: 'Get Usage', href: '/docs/api/usage/get' },
      { title: 'Batch Usage', href: '/docs/api/usage/batch' },
    ],
  },
  {
    title: 'GraphQL',
    items: [
      { title: 'Overview', href: '/docs/graphql/overview' },
      { title: 'Schema', href: '/docs/graphql/schema' },
      { title: 'Queries', href: '/docs/graphql/queries' },
      { title: 'Examples', href: '/docs/graphql/examples' },
    ],
  },
  {
    title: 'Resources',
    items: [
      { title: 'Error Codes', href: '/docs/errors' },
      { title: 'Rate Limits', href: '/docs/rate-limits' },
      { title: 'Best Practices', href: '/docs/best-practices' },
      { title: 'Changelog', href: '/docs/changelog' },
    ],
  },
]
