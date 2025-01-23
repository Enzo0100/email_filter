import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatDate(date: Date | string) {
  return new Date(date).toLocaleDateString('pt-BR', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

export const priorities = {
  high: { label: 'Alta', color: 'text-red-500' },
  medium: { label: 'Média', color: 'text-yellow-500' },
  low: { label: 'Baixa', color: 'text-green-500' }
}

export function getPriorityColor(priority: string) {
  return priorities[priority as keyof typeof priorities]?.color || 'text-gray-500'
}