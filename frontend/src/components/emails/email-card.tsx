import { formatDate, getPriorityColor } from "@/lib/utils"
import { Mail, Tag, Clock, CheckCircle } from "lucide-react"

interface EmailCardProps {
  email: {
    id: string
    subject: string
    from: string
    content: string
    priority: string
    category: string
    labels: string[]
    processedAt: string
    createdAt: string
  }
}

export function EmailCard({ email }: EmailCardProps) {
  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-4">
      <div className="flex items-start justify-between mb-2">
        <div>
          <h3 className="text-lg font-semibold mb-1">{email.subject}</h3>
          <p className="text-sm text-gray-600 dark:text-gray-400">
            De: {email.from}
          </p>
        </div>
        <div className={`px-2 py-1 rounded text-sm ${getPriorityColor(email.priority)}`}>
          {email.priority.charAt(0).toUpperCase() + email.priority.slice(1)}
        </div>
      </div>

      <p className="text-gray-700 dark:text-gray-300 mb-4 line-clamp-2">
        {email.content}
      </p>

      <div className="flex flex-wrap gap-2 mb-3">
        {email.labels.map((label) => (
          <span
            key={label}
            className="inline-flex items-center px-2 py-1 rounded-full text-xs bg-gray-100 dark:bg-gray-700"
          >
            <Tag className="w-3 h-3 mr-1" />
            {label}
          </span>
        ))}
      </div>

      <div className="flex items-center justify-between text-sm text-gray-500 dark:text-gray-400">
        <div className="flex items-center space-x-4">
          <span className="flex items-center">
            <Mail className="w-4 h-4 mr-1" />
            {email.category}
          </span>
          <span className="flex items-center">
            <Clock className="w-4 h-4 mr-1" />
            {formatDate(email.createdAt)}
          </span>
        </div>
        <div className="flex items-center">
          <CheckCircle className="w-4 h-4 mr-1" />
          Processado em {formatDate(email.processedAt)}
        </div>
      </div>
    </div>
  )
}