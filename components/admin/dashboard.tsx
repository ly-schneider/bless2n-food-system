"use client";

import { useState } from "react";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { StatisticsPanel } from "./statistics";
import { ProductManagement } from "./product-management";

export function AdminDashboard() {
  const [activeTab, setActiveTab] = useState("products");

  return (
    <div className="">
      <h1 className="text-3xl font-bold mb-8">Admin Dashboard</h1>
      
      <Tabs defaultValue="products" value={activeTab} onValueChange={setActiveTab} className="space-y-8">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="products">Produktverwaltung</TabsTrigger>
          <TabsTrigger value="statistics">Statistiken</TabsTrigger>
        </TabsList>

        <TabsContent value="products">
          <ProductManagement />
        </TabsContent>
        
        <TabsContent value="statistics" className="space-y-8">
          <StatisticsPanel />
        </TabsContent>
      </Tabs>
    </div>
  );
}