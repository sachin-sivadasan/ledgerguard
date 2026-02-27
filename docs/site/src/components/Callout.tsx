import { AlertCircle, Info as InfoIcon, AlertTriangle, CheckCircle } from 'lucide-react'
import clsx from 'clsx'

interface CalloutProps {
  type?: 'note' | 'warning' | 'info' | 'success'
  title?: string
  children: React.ReactNode
}

const icons = {
  note: InfoIcon,
  warning: AlertTriangle,
  info: AlertCircle,
  success: CheckCircle,
}

const styles = {
  note: 'callout-note',
  warning: 'callout-warning',
  info: 'callout-info',
  success: 'bg-emerald-50 dark:bg-emerald-950/30 border-emerald-500 text-emerald-900 dark:text-emerald-100',
}

export function Callout({ type = 'note', title, children }: CalloutProps) {
  const Icon = icons[type]

  return (
    <div className={clsx('callout', styles[type])}>
      <div className="flex gap-3">
        <Icon className="w-5 h-5 flex-shrink-0 mt-0.5" />
        <div>
          {title && <p className="font-semibold mb-1">{title}</p>}
          <div className="text-sm opacity-90">{children}</div>
        </div>
      </div>
    </div>
  )
}

export function Note({ children, title }: { children: React.ReactNode; title?: string }) {
  return <Callout type="note" title={title}>{children}</Callout>
}

export function Warning({ children, title }: { children: React.ReactNode; title?: string }) {
  return <Callout type="warning" title={title}>{children}</Callout>
}

export function Info({ children, title }: { children: React.ReactNode; title?: string }) {
  return <Callout type="info" title={title}>{children}</Callout>
}
