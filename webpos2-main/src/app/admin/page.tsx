'use client'

import { useState } from 'react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { ArticlesSection } from '@/components/admin/ArticlesSection'
import { StatistikSection } from '@/components/admin/StatistikSection'

export default function AdminPage() {
  const [activeTab, setActiveTab] = useState("articles")

  return (
    <div className="flex h-screen">
      {/* Left sidebar */}
      <div className="w-64 h-full border-r bg-gray-50/40">
        <div className="p-6">
          <h1 className="text-2xl font-semibold text-gray-900">Admin</h1>
        </div>
        <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full" orientation="vertical">
          <TabsList className="flex flex-col h-auto w-full bg-transparent rounded-none p-0 justify-start">
            <TabsTrigger 
              value="articles" 
              className="w-full justify-start rounded-none px-6 py-3 font-medium data-[state=active]:bg-white data-[state=active]:shadow-none"
            >
              Artikel
            </TabsTrigger>
            <TabsTrigger 
              value="statistik" 
              className="w-full justify-start rounded-none px-6 py-3 font-medium data-[state=active]:bg-white data-[state=active]:shadow-none"
            >
              Statistik
            </TabsTrigger>
            {/* Add more sections here in the future */}
          </TabsList>
        </Tabs>
      </div>

      {/* Main content area */}
      <div className="flex-1 h-full overflow-auto">
        <Tabs value={activeTab} onValueChange={setActiveTab} orientation="vertical">
          <TabsContent value="articles" className="m-0 outline-none">
            <ArticlesSection />
          </TabsContent>
          <TabsContent value="statistik" className="m-0 outline-none">
            <StatistikSection />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}
