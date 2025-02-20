"use client"

import { useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { EmailFilters } from "@/components/emails/email-filters"
import { EmailCard } from "@/components/emails/email-card"
import { emailsApi } from "@/lib/api"

export default function DashboardPage() {
  const [filters, setFilters] = useState({
    priority: "",
    category: "",
    status: "",
  })

  const { data: emails = [], isLoading, error, isError } = useQuery({
    queryKey: ["emails", filters],
    queryFn: () => emailsApi.getEmails(filters),
    retry: 2,
    retryDelay: 1000,
    staleTime: 30000
  })

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 dark:border-white mx-auto mb-4"></div>
          <p className="text-gray-600 dark:text-gray-400">Carregando emails...</p>
        </div>
      </div>
    )
  }

  if (isError) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <div className="bg-red-50 dark:bg-red-900/20 rounded-lg p-4 mb-4">
          <p className="text-red-600 dark:text-red-400 text-center">
            {error instanceof Error ? error.message : 'Erro ao carregar emails. Por favor, tente novamente mais tarde.'}
          </p>
        </div>
        <button
          onClick={() => window.location.reload()}
          className="mt-4 px-4 py-2 bg-gray-100 dark:bg-gray-800 rounded-md hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
        >
          Tentar Novamente
        </button>
      </div>
    )
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold mb-2">Emails Classificados</h1>
        <p className="text-gray-600 dark:text-gray-400">
          Gerencie seus emails classificados e tarefas associadas
        </p>
      </div>

      <EmailFilters
        onFilterChange={(newFilters) => {
          setFilters((prev) => ({ ...prev, ...newFilters }))
        }}
      />

      <div className="space-y-4">
        {emails.map((email) => (
          <EmailCard key={email.id} email={email} />
        ))}
      </div>

      {emails.length === 0 && (
        <div className="text-center py-12">
          <p className="text-gray-500 dark:text-gray-400">
            Nenhum email encontrado com os filtros selecionados
          </p>
        </div>
      )}
    </div>
  )
}