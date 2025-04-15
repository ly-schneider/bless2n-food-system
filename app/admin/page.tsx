import { createClient } from "@/utils/supabase/server";
import { AdminDashboard } from "@/components/admin/dashboard";
import { redirect } from "next/navigation";

export default async function AdminPage() {
  const supabase = await createClient();

  const {
    data: { user },
  } = await supabase.auth.getUser();

  if (!user) {
    redirect("/login");
  }

  return <AdminDashboard />;
}
