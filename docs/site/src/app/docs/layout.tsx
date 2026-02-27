import { Header } from '@/components/Header'
import { Sidebar } from '@/components/Sidebar'

export default function DocsLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <div className="min-h-screen">
      <Header />
      <Sidebar />
      <main className="pl-64 pt-16">
        <article className="max-w-3xl mx-auto px-8 py-12">
          {children}
        </article>
      </main>
    </div>
  )
}
