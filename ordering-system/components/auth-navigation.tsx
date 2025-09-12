"use client";

import Link from "next/link";
import { Button } from "./ui/button";
import { createClient } from "@/utils/supabase/client";
import { useEffect, useState } from "react";
import { User } from "@supabase/supabase-js";
import { usePathname } from "next/navigation";
import { ArrowLeft } from "lucide-react";

export default function AuthNavigation() {
  const supabase = createClient();
  const pathname = usePathname();
  const isAdminPage = pathname.startsWith("/admin");
  const [user, setUser] = useState<User>();

  const fetchUser = async () => {
    const {
      data: { user },
    } = await supabase.auth.getUser();

    if (!user) {
      return;
    }

    setUser(user);
  };

  useEffect(() => {
    fetchUser();
  }, []);

  if (!user) {
    return;
  }

  return isAdminPage ? (
    <div className="flex gap-2">
      <Button asChild size="sm" variant={"default"}>
        <Link href="/">
          <ArrowLeft className="h-4 w-4" />
          Zurück
        </Link>
      </Button>
    </div>
  ) : (
    <div className="flex gap-2">
      <Button asChild size="sm" variant={"outline"}>
        <Link href="/qr-codes">QR-Codes</Link>
      </Button>
      <Button asChild size="sm" variant={"default"}>
        <Link href="/admin">Admin</Link>
      </Button>
    </div>
  );
}
