import { redirect } from "next/navigation"

export default function HomePage() {
  // TODO: Verificar autenticação
  // Por enquanto, sempre redireciona para o login
  redirect("/login")
}
