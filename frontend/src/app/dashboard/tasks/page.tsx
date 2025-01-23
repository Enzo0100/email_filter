"use client"

import { useState } from "react"
import { TaskCard } from "@/components/tasks/task-card"

// Dados mockados para exemplo
const mockTasks = [
  {
    id: "1",
    description: "Preparar apresentação para reunião de vendas",
    dueDate: new Date(Date.now() + 86400000).toISOString(), // Amanhã
    priority: "high",
    status: "pending",
    createdAt: new Date().toISOString(),
  },
  {
    id: "2",
    description: "Revisar relatório mensal",
    dueDate: new Date(Date.now() + 172800000).toISOString(), // 2 dias
    priority: "medium",
    status: "pending",
    createdAt: new Date().toISOString(),
  },
  {
    id: "3",
    description: "Responder email do departamento financeiro",
    dueDate: new Date(Date.now() + 86400000).toISOString(), // Amanhã
    priority: "low",
    status: "completed",
    createdAt: new Date().toISOString(),
  },
]

export default function TasksPage() {
  const [tasks, setTasks] = useState(mockTasks)
  const [filter, setFilter] = useState<"all" | "pending" | "completed">("all")

  const handleStatusChange = (taskId: string, newStatus: string) => {
    setTasks((currentTasks) =>
      currentTasks.map((task) =>
        task.id === taskId ? { ...task, status: newStatus } : task
      )
    )
  }

  const filteredTasks = tasks.filter((task) => {
    if (filter === "all") return true
    if (filter === "pending") return task.status === "pending"
    if (filter === "completed") return task.status === "completed"
    return true
  })

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold mb-2">Tarefas</h1>
        <p className="text-gray-600 dark:text-gray-400">
          Gerencie suas tarefas geradas a partir dos emails
        </p>
      </div>

      <div className="mb-6">
        <div className="flex gap-2">
          <button
            onClick={() => setFilter("all")}
            className={`px-4 py-2 rounded-lg ${
              filter === "all"
                ? "bg-blue-500 text-white"
                : "bg-gray-100 dark:bg-gray-700"
            }`}
          >
            Todas
          </button>
          <button
            onClick={() => setFilter("pending")}
            className={`px-4 py-2 rounded-lg ${
              filter === "pending"
                ? "bg-blue-500 text-white"
                : "bg-gray-100 dark:bg-gray-700"
            }`}
          >
            Pendentes
          </button>
          <button
            onClick={() => setFilter("completed")}
            className={`px-4 py-2 rounded-lg ${
              filter === "completed"
                ? "bg-blue-500 text-white"
                : "bg-gray-100 dark:bg-gray-700"
            }`}
          >
            Concluídas
          </button>
        </div>
      </div>

      <div className="space-y-4">
        {filteredTasks.map((task) => (
          <TaskCard
            key={task.id}
            task={task}
            onStatusChange={handleStatusChange}
          />
        ))}
      </div>

      {filteredTasks.length === 0 && (
        <div className="text-center py-12">
          <p className="text-gray-500 dark:text-gray-400">
            Nenhuma tarefa encontrada com o filtro selecionado
          </p>
        </div>
      )}
    </div>
  )
}