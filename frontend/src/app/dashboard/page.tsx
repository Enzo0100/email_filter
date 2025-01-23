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

  const { data: emails = [], isLoading, error } = useQuery({
    queryKey: ["emails", filters],
    queryFn: () => emailsApi.getEmails(filters),
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

  if (error) {
    return (
      <div className="text-center py-12">
        <p className="text-red-500">
          Erro ao carregar emails. Por favor, tente novamente mais tarde.
        </p>
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