import { Colors } from "@/constants/Colors";
import { RegistrationProvider } from "@/hooks/useRegistration";
import { Stack } from "expo-router";

export default function RegisterLayout() {
  return (
    <RegistrationProvider initialData={{}}>
      <Stack
        screenOptions={{
          contentStyle: { backgroundColor: Colors.background },
          headerShown: false,
          animation: "fade",
        }}
      />
    </RegistrationProvider>
  );
}
