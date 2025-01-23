"use client"

import { useState } from "react"
import Link from "next/link"
import { Mail, Lock, Building, User } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

export default function RegisterPage() {
  const [formData, setFormData] = useState({
    name: "",
    email: "",
    password: "",
    confirmPassword: "",
    companyName: "",
  })
  const [isLoading, setIsLoading] = useState(false)
  const [errors, setErrors] = useState<Record<string, string>>({})

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setErrors({})

    if (formData.password !== formData.confirmPassword) {
      setErrors({ confirmPassword: "As senhas não coincidem" })
      return
    }

    setIsLoading(true)
    try {
      // TODO: Implementar integração com a API de registro
      console.log("Registro:", formData)
      await new Promise((resolve) => setTimeout(resolve, 1000)) // Simulação de delay
    } catch (error) {
      console.error("Erro ao registrar:", error)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-100 dark:bg-gray-900 px-4">
      <div className="max-w-md w-full">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold">Email Filter</h1>
          <p className="text-gray-600 dark:text-gray-400 mt-2">
            Crie sua conta para começar
          </p>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
          <h2 className="text-2xl font-semibold mb-6">Registro</h2>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-2">Nome</label>
              <Input
                type="text"
                value={formData.name}
                onChange={(e) =>
                  setFormData((prev) => ({ ...prev, name: e.target.value }))
                }
                icon={<User className="w-5 h-5" />}
                placeholder="Seu nome completo"
                required
                disabled={isLoading}
                error={errors.name}
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">Email</label>
              <Input
                type="email"
                value={formData.email}
                onChange={(e) =>
                  setFormData((prev) => ({ ...prev, email: e.target.value }))
                }
                icon={<Mail className="w-5 h-5" />}
                placeholder="seu@email.com"
                required
                disabled={isLoading}
                error={errors.email}
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">Nome da Empresa</label>
              <Input
                type="text"
                value={formData.companyName}
                onChange={(e) =>
                  setFormData((prev) => ({ ...prev, companyName: e.target.value }))
                }
                icon={<Building className="w-5 h-5" />}
                placeholder="Nome da sua empresa"
                required
                disabled={isLoading}
                error={errors.companyName}
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">Senha</label>
              <Input
                type="password"
                value={formData.password}
                onChange={(e) =>
                  setFormData((prev) => ({ ...prev, password: e.target.value }))
                }
                icon={<Lock className="w-5 h-5" />}
                placeholder="••••••••"
                required
                disabled={isLoading}
                error={errors.password}
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">Confirmar Senha</label>
              <Input
                type="password"
                value={formData.confirmPassword}
                onChange={(e) =>
                  setFormData((prev) => ({ ...prev, confirmPassword: e.target.value }))
                }
                icon={<Lock className="w-5 h-5" />}
                placeholder="••••••••"
                required
                disabled={isLoading}
                error={errors.confirmPassword}
              />
            </div>

            <Button
              type="submit"
              className="w-full"
              isLoading={isLoading}
            >
              Criar Conta
            </Button>
          </form>

          <div className="mt-6 text-center">
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Já tem uma conta?{" "}
              <Link
                href="/login"
                className="text-primary-500 hover:text-primary-600"
              >
                Faça login
              </Link>
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}