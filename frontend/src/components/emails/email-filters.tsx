import { priorities } from "@/lib/utils"

interface EmailFiltersProps {
  onFilterChange: (filters: {
    priority?: string
    category?: string
    status?: string
  }) => void
}

export function EmailFilters({ onFilterChange }: EmailFiltersProps) {
  return (
    <div className="flex flex-wrap gap-4 mb-6">
      <select
        className="px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800"
        onChange={(e) => onFilterChange({ priority: e.target.value })}
      >
        <option value="">Todas Prioridades</option>
        {Object.entries(priorities).map(([key, value]) => (
          <option key={key} value={key}>
            {value.label}
          </option>
        ))}
      </select>

      <select
        className="px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800"
        onChange={(e) => onFilterChange({ category: e.target.value })}
      >
        <option value="">Todas Categorias</option>
        <option value="trabalho">Trabalho</option>
        <option value="pessoal">Pessoal</option>
        <option value="financeiro">Financeiro</option>
        <option value="outros">Outros</option>
      </select>

      <select
        className="px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800"
        onChange={(e) => onFilterChange({ status: e.target.value })}
      >
        <option value="">Todos Status</option>
        <option value="pendente">Pendente</option>
        <option value="processado">Processado</option>
        <option value="arquivado">Arquivado</option>
      </select>
    </div>
  )
}