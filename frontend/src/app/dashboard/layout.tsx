import Link from "next/link"
import { Inbox, Settings, CheckSquare } from "lucide-react"

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <div className="flex h-screen">
      {/* Sidebar */}
      <aside className="w-64 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700">
        <div className="h-16 flex items-center px-6 border-b border-gray-200 dark:border-gray-700">
          <h1 className="text-xl font-bold">Email Filter</h1>
        </div>
        <nav className="p-4 space-y-2">
          <Link 
            href="/dashboard"
            className="flex items-center space-x-3 px-3 py-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700"
          >
            <Inbox className="w-5 h-5" />
            <span>Emails</span>
          </Link>
          <Link 
            href="/dashboard/tasks"
            className="flex items-center space-x-3 px-3 py-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700"
          >
            <CheckSquare className="w-5 h-5" />
            <span>Tarefas</span>
          </Link>
          <Link 
            href="/settings"
            className="flex items-center space-x-3 px-3 py-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700"
          >
            <Settings className="w-5 h-5" />
            <span>Configurações</span>
          </Link>
        </nav>
      </aside>

      {/* Main content */}
      <main className="flex-1 overflow-auto">
        <div className="h-16 border-b border-gray-200 dark:border-gray-700 flex items-center px-6">
          <div className="ml-auto flex items-center space-x-4">
            <span className="text-sm">Usuário</span>
            <button className="text-sm text-red-500 hover:text-red-600">
              Sair
            </button>
          </div>
        </div>
        <div className="p-6">
          {children}
        </div>
      </main>
    </div>
  )
}