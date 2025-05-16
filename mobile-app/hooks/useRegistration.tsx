import { createContext, useContext, useState } from "react";

interface RegistrationData {
  email?: string;
  isRegistration?: boolean;
  userId?: string;
  firstName?: string | null;
  lastName?: string | null;
  clubData?: {
    _id: string;
    name: string;
  };
  appleUserId?: string | null;
  trackingEnabled?: boolean;
}

type RegistrationContextType = {
  data: RegistrationData;
  updateData: (newData: Partial<RegistrationData>) => void;
};

const RegistrationContext = createContext<RegistrationContextType | undefined>(
  undefined
);

export function useRegistration() {
  const context = useContext(RegistrationContext);
  if (!context) {
    throw new Error("useRegistration must be used within RegistrationProvider");
  }
  return context;
}

export function RegistrationProvider({
  children,
  initialData,
}: {
  children: React.ReactNode;
  initialData: RegistrationData;
}) {
  const [data, setData] = useState<RegistrationData>(initialData);

  const updateData = (newData: Partial<RegistrationData>) => {
    setData((prev) => ({ ...prev, ...newData }));
  };

  return (
    <RegistrationContext.Provider value={{ data, updateData }}>
      {children}
    </RegistrationContext.Provider>
  );
}