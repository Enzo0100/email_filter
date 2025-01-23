"use client"

import { useState } from "react"
import { Save } from "lucide-react"

interface Settings {
  emailNotifications: boolean
  taskReminders: boolean
  darkMode: boolean
  language: string
  autoArchive: boolean
}

export default function SettingsPage() {
  const [settings, setSettings] = useState<Settings>({
    emailNotifications: true,
    taskReminders: true,
    darkMode: false,
    language: "pt-BR",
    autoArchive: false,
  })

  const handleSettingChange = (key: keyof Settings, value: any) => {
    setSettings((prev) => ({ ...prev, [key]: value }))
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    // TODO: Implementar integração com a API
    console.log("Configurações salvas:", settings)
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold mb-2">Configurações</h1>
        <p className="text-gray-600 dark:text-gray-400">
          Gerencie suas preferências do sistema
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold mb-4">Notificações</h2>
          <div className="space-y-4">
            <label className="flex items-center space-x-3">
              <input
                type="checkbox"
                checked={settings.emailNotifications}
                onChange={(e) =>
                  handleSettingChange("emailNotifications", e.target.checked)
                }
                className="rounded border-gray-300"
              />
              <span>Receber notificações por email</span>
            </label>

            <label className="flex items-center space-x-3">
              <input
                type="checkbox"
                checked={settings.taskReminders}
                onChange={(e) =>
                  handleSettingChange("taskReminders", e.target.checked)
                }
                className="rounded border-gray-300"
              />
              <span>Lembretes de tarefas</span>
            </label>
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold mb-4">Preferências</h2>
          <div className="space-y-4">
            <label className="flex items-center space-x-3">
              <input
                type="checkbox"
                checked={settings.darkMode}
                onChange={(e) => handleSettingChange("darkMode", e.target.checked)}
                className="rounded border-gray-300"
              />
              <span>Modo escuro</span>
            </label>

            <div>
              <label className="block mb-2">Idioma</label>
              <select
                value={settings.language}
                onChange={(e) => handleSettingChange("language", e.target.value)}
                className="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800"
              >
                <option value="pt-BR">Português (Brasil)</option>
                <option value="en">English</option>
                <option value="es">Español</option>
              </select>
            </div>

            <label className="flex items-center space-x-3">
              <input
                type="checkbox"
                checked={settings.autoArchive}
                onChange={(e) =>
                  handleSettingChange("autoArchive", e.target.checked)
                }
                className="rounded border-gray-300"
              />
              <span>Arquivar emails automaticamente após processamento</span>
            </label>
          </div>
        </div>

        <div className="flex justify-end">
          <button
            type="submit"
            className="flex items-center space-x-2 px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
          >
            <Save className="w-4 h-4" />
            <span>Salvar Configurações</span>
          </button>
        </div>
      </form>
    </div>
  )
}