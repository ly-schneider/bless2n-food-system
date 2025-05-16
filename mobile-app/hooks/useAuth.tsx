import React, { createContext, useContext, useEffect, useState } from "react";
import AsyncStorage from "@react-native-async-storage/async-storage";
import analytics from "@react-native-firebase/analytics";
import crashlytics from "@react-native-firebase/crashlytics";

type AuthContextType = {
  user: string | null;
  setUser: (user: string | null) => void;
  saveUser: (user: string) => Promise<void>;
  logoutUser: () => Promise<void>;
  isLoading: boolean;
};

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: React.ReactNode }) => {
  const [user, setUser] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const loadUser = async () => {
      const storedUser = await AsyncStorage.getItem("user");
      if (storedUser) {
        setUser(storedUser);
      }
      setIsLoading(false);
    };

    loadUser();
  }, []);

  const saveUser = async (user: string | null) => {
    if (user === null) {
      await AsyncStorage.removeItem("user");
      setUser(null);
      return;
    }
    await AsyncStorage.setItem("user", user);

    await analytics().setUserId(user);
    await crashlytics().setUserId(user);

    setUser(user);
  };

  const logoutUser = async () => {
    await saveUser(null);
  };

  return (
    <AuthContext.Provider
      value={{ user, setUser, saveUser, logoutUser, isLoading }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) throw new Error("useAuth must be used within an AuthProvider");
  return context;
};
