interface ResponseFieldProps {
  name: string
  type: string
  required?: boolean
  children: React.ReactNode
}

export function ResponseField({ name, type, required = false, children }: ResponseFieldProps) {
  return (
    <div className="py-3 border-b border-gray-100 dark:border-gray-800 last:border-b-0">
      <div className="flex items-center gap-2 mb-1">
        <code className="text-sm font-semibold text-primary-600 dark:text-primary-400">{name}</code>
        <span className="text-xs text-gray-500 dark:text-gray-400 font-mono">{type}</span>
        {required && (
          <span className="text-xs text-red-500 font-medium">required</span>
        )}
      </div>
      <div className="text-sm text-gray-600 dark:text-gray-400">{children}</div>
    </div>
  )
}

export function ParamField({ name, type, required, children, location }: ResponseFieldProps & { location?: string }) {
  return (
    <div className="py-3 border-b border-gray-100 dark:border-gray-800 last:border-b-0">
      <div className="flex items-center gap-2 mb-1">
        <code className="text-sm font-semibold text-primary-600 dark:text-primary-400">{name}</code>
        <span className="text-xs text-gray-500 dark:text-gray-400 font-mono">{type}</span>
        {location && (
          <span className="text-xs text-gray-400 dark:text-gray-500">in {location}</span>
        )}
        {required && (
          <span className="text-xs text-red-500 font-medium">required</span>
        )}
      </div>
      <div className="text-sm text-gray-600 dark:text-gray-400">{children}</div>
    </div>
  )
}
