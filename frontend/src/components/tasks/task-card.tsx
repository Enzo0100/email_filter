import { formatDate, getPriorityColor } from "@/lib/utils"
import { Calendar, CheckSquare, Square } from "lucide-react"

interface TaskCardProps {
  task: {
    id: string
    description: string
    dueDate: string
    priority: string
    status: string
    createdAt: string
  }
  onStatusChange?: (id: string, status: string) => void
}

export function TaskCard({ task, onStatusChange }: TaskCardProps) {
  const isCompleted = task.status === "completed"

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-4">
      <div className="flex items-start gap-3">
        <button
          onClick={() => onStatusChange?.(task.id, isCompleted ? "pending" : "completed")}
          className="mt-1"
        >
          {isCompleted ? (
            <CheckSquare className="w-5 h-5 text-green-500" />
          ) : (
            <Square className="w-5 h-5 text-gray-400" />
          )}
        </button>

        <div className="flex-1">
          <div className="flex items-start justify-between mb-2">
            <p className={`text-lg ${isCompleted ? "line-through text-gray-500" : ""}`}>
              {task.description}
            </p>
            <span className={`px-2 py-1 rounded text-sm ${getPriorityColor(task.priority)}`}>
              {task.priority.charAt(0).toUpperCase() + task.priority.slice(1)}
            </span>
          </div>

          <div className="flex items-center gap-4 text-sm text-gray-500 dark:text-gray-400">
            <span className="flex items-center">
              <Calendar className="w-4 h-4 mr-1" />
              Vence em: {formatDate(task.dueDate)}
            </span>
          </div>
        </div>
      </div>
    </div>
  )
}