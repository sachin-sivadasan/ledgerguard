'use client'

import { useState } from 'react'
import { Check, Copy } from 'lucide-react'
import clsx from 'clsx'

interface CodeBlockProps {
  children: string
  language?: string
  filename?: string
  showLineNumbers?: boolean
}

export function CodeBlock({ children, language = 'bash', filename, showLineNumbers = false }: CodeBlockProps) {
  const [copied, setCopied] = useState(false)

  const copyToClipboard = async () => {
    await navigator.clipboard.writeText(children)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="code-block my-4">
      {filename && (
        <div className="px-4 py-2 text-xs text-gray-400 border-b border-gray-800 bg-gray-900">
          {filename}
        </div>
      )}
      <div className="relative">
        <pre className={clsx(showLineNumbers && 'line-numbers')}>
          <code className={`language-${language}`}>{children}</code>
        </pre>
        <button
          onClick={copyToClipboard}
          className="copy-button"
          aria-label="Copy code"
        >
          {copied ? (
            <Check className="w-4 h-4 text-green-400" />
          ) : (
            <Copy className="w-4 h-4" />
          )}
        </button>
      </div>
    </div>
  )
}

interface CodeTabsProps {
  tabs: {
    label: string
    language: string
    code: string
  }[]
}

export function CodeTabs({ tabs }: CodeTabsProps) {
  const [activeTab, setActiveTab] = useState(0)
  const [copied, setCopied] = useState(false)

  const copyToClipboard = async () => {
    await navigator.clipboard.writeText(tabs[activeTab].code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="code-block my-4">
      <div className="language-tabs">
        {tabs.map((tab, index) => (
          <button
            key={tab.label}
            onClick={() => setActiveTab(index)}
            className={clsx('language-tab', index === activeTab && 'active')}
          >
            {tab.label}
          </button>
        ))}
      </div>
      <div className="relative">
        <pre>
          <code className={`language-${tabs[activeTab].language}`}>
            {tabs[activeTab].code}
          </code>
        </pre>
        <button
          onClick={copyToClipboard}
          className="copy-button"
          aria-label="Copy code"
        >
          {copied ? (
            <Check className="w-4 h-4 text-green-400" />
          ) : (
            <Copy className="w-4 h-4" />
          )}
        </button>
      </div>
    </div>
  )
}
