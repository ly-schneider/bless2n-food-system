"use client";

import { usePathname } from "next/navigation";
import { LockButton } from "./lock-provider";
import AuthNavigation from "./auth-navigation";
import { RefreshButton } from "./refresh-button";

export default function Navigation() {
  const pathname = usePathname();
  const isSignInPage = pathname.startsWith("/admin/sign-in");

  return (
    <>
      {!isSignInPage && (
        <>
          <RefreshButton />
          <LockButton />
        </>
      )}
      <AuthNavigation />
    </>
  );
}
