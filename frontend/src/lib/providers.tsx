import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { ReactNode } from "react"

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: (failureCount, error) => {
        if (error instanceof Error) {
          // Não tentar novamente para erros de autenticação ou permissão
          if (error.message.includes('Sessão expirada') || error.message.includes('permissão')) {
            return false;
          }
          // Não tentar novamente para erros 404
          if (error.message.includes('Nenhum email encontrado')) {
            return false;
          }
        }
        // Tentar no máximo 3 vezes para outros erros
        return failureCount < 3;
      },
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
      refetchOnWindowFocus: false,
      staleTime: 30000, // 30 segundos
      gcTime: 5 * 60 * 1000, // 5 minutos
    },
  },
})

export function Providers({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  )
}