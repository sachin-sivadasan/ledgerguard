import clsx from 'clsx'

interface EndpointProps {
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
  path: string
}

const methodStyles = {
  GET: 'endpoint-get',
  POST: 'endpoint-post',
  PUT: 'bg-amber-100 dark:bg-amber-950 text-amber-700 dark:text-amber-400',
  DELETE: 'endpoint-delete',
  PATCH: 'bg-purple-100 dark:bg-purple-950 text-purple-700 dark:text-purple-400',
}

export function Endpoint({ method, path }: EndpointProps) {
  return (
    <div className="flex items-center gap-3 my-4">
      <span className={clsx('endpoint-badge', methodStyles[method])}>
        <span className="font-semibold">{method}</span>
      </span>
      <code className="text-sm font-mono text-gray-700 dark:text-gray-300">{path}</code>
    </div>
  )
}
