import Link from "next/link";
import { Button } from "./ui/button";
import { createClient } from "@/utils/supabase/server";

export default async function AuthButton() {
  const supabase = await createClient();

  const {
    data: { user },
  } = await supabase.auth.getUser();

  return user && (
    <div className="flex gap-2">
      <Button asChild size="sm" variant={"default"}>
        <Link href="/admin">Admin</Link>
      </Button>
    </div>
  );
}
