import { Redirect, router, Stack } from "expo-router";
import { Colors } from "@/constants/Colors";
import { useAuth } from "@/hooks/useAuth";
import { useEffect, useState } from "react";
import {
  hasAcceptedCurrentTerms
} from "@/utils/TermsManager";

export default function AppLayout() {
  const { user, isLoading } = useAuth();
  const [checkingTerms, setCheckingTerms] = useState(true);

  useEffect(() => {
    const checkAcceptedTerms = async () => {
      setCheckingTerms(true);

      if (!user) {
        setCheckingTerms(false);
        return;
      }
      
      try {
        const termsUpToDate = await hasAcceptedCurrentTerms();

        if (!termsUpToDate) {
          // Use a small timeout to ensure navigation happens after render
          setTimeout(() => {
            if (router.canDismiss()) router.dismissAll();
            router.push("/auth/register/terms");
          }, 100);
        }
      } catch (error) {
        console.error("Error checking terms acceptance:", error);
      } finally {
        setCheckingTerms(false);
      }
    };

    checkAcceptedTerms();
  }, [user]);

  if (isLoading || checkingTerms) {
    return null;
  }

  if (!user) {
    return <Redirect href="/auth/register/start" />;
  }

  return (
    <Stack
      screenOptions={{
        contentStyle: { backgroundColor: Colors.background },
        headerShown: false,
        animation: "slide_from_right",
      }}
    />
  );
}
