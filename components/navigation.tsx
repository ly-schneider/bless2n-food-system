"use client";

import { usePathname } from "next/navigation";
import { LockButton } from "./lock-provider";
import AuthNavigation from "./auth-navigation";

export default function Navigation() {
  const pathname = usePathname();
  const isSignInPage = pathname.startsWith("/admin/sign-in");

  return (
    <>
      {!isSignInPage && <LockButton />}
      <AuthNavigation />
    </>
  );
}
