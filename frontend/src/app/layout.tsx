"use client"

import { Inter } from "next/font/google"
import { Providers } from "@/lib/providers"
import "./globals.css"

const inter = Inter({ subsets: ["latin"] })

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="pt-BR">
      <body className={inter.className}>
        <div className="min-h-screen bg-gray-100 dark:bg-gray-900">
          <Providers>
            {children}
          </Providers>
        </div>
      </body>
    </html>
  )
}
