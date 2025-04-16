import { createClient } from "@/utils/supabase/server";
import { AdminDashboard } from "@/components/admin/dashboard";
import { redirect } from "next/navigation";
import { Suspense } from "react";

// Skeleton loader component for better UX during loading
function AdminSkeleton() {
  return (
    <div className="space-y-8">
      <div className="h-10 w-1/3 bg-gray-200 rounded animate-pulse"></div>
      <div className="border rounded-md p-4">
        <div className="h-8 w-full bg-gray-200 rounded mb-6 animate-pulse"></div>
        <div className="grid grid-cols-1 gap-6">
          <div className="h-[400px] bg-gray-200 rounded animate-pulse"></div>
        </div>
      </div>
    </div>
  );
}

export default async function AdminPage() {
  const supabase = await createClient();

  const {
    data: { user },
  } = await supabase.auth.getUser();

  if (!user) {
    redirect("/admin/sign-in");
  }

  return (
    <Suspense fallback={<AdminSkeleton />}>
      <AdminDashboard />
    </Suspense>
  );
}
