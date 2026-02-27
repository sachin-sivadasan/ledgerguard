'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { navigation, NavItem } from '@/lib/navigation'
import clsx from 'clsx'

function NavLink({ item, depth = 0 }: { item: NavItem; depth?: number }) {
  const pathname = usePathname()
  const isActive = pathname === item.href

  if (item.items) {
    return (
      <div className={clsx(depth > 0 && 'ml-4')}>
        <h4 className="px-3 py-2 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
          {item.title}
        </h4>
        <div className="space-y-1">
          {item.items.map((subItem) => (
            <NavLink key={subItem.href || subItem.title} item={subItem} depth={depth + 1} />
          ))}
        </div>
      </div>
    )
  }

  return (
    <Link
      href={item.href!}
      className={clsx(
        'sidebar-link',
        isActive && 'active'
      )}
    >
      {item.title}
    </Link>
  )
}

export function Sidebar() {
  return (
    <aside className="fixed left-0 top-16 bottom-0 w-64 border-r border-gray-200 dark:border-gray-800 overflow-y-auto bg-white dark:bg-gray-950 z-40">
      <nav className="p-4 space-y-6">
        {navigation.map((section) => (
          <NavLink key={section.title} item={section} />
        ))}
      </nav>
    </aside>
  )
}
