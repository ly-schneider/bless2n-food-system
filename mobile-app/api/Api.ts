import axios, { AxiosInstance } from "axios";
import * as SecureStore from "expo-secure-store";
import { router } from "expo-router";
import AsyncStorage from "@react-native-async-storage/async-storage";
import { Match } from "@/constants/Match";
import { clearTermsAcceptance, saveTermsAcceptance } from "@/utils/TermsManager";

/* -------------------------------------------------------------
   Helper Functions & Configuration
------------------------------------------------------------- */
let httpClient: AxiosInstance | null = null;

const initializeHttpClient = async (): Promise<AxiosInstance> => {
  const baseUrl = process.env.EXPO_PUBLIC_API_URL;
  console.info("API Base URL: " + baseUrl);
  const client = axios.create({
    baseURL: baseUrl,
    headers: { "Content-Type": "application/json" },
  });

  client.interceptors.response.use(
    (response) => response,
    async (error) => {
      console.info("API Error:", error);

      if (error.response) {
        const status = error.response.status;
        if (status === 401) {
          try {
            const refreshToken = await SecureStore.getItemAsync("refreshToken");
            if (!refreshToken) throw new Error("No refresh token available");

            const res = await client.post(
              `/auth/refresh`,
              {},
              {
                headers: { refreshToken },
              }
            );

            if (res.data.success) {
              await SecureStore.setItemAsync("accessToken", res.headers["accesstoken"], {
                keychainAccessible: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
              });
              await SecureStore.setItemAsync("refreshToken", res.headers["refreshtoken"], {
                keychainAccessible: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
              });
              error.config.headers["accessToken"] = res.headers["accesstoken"];
              return client.request(error.config);
            }
          } catch (err) {
            console.info("Token refresh failed:", err);
            // Don't immediately logout on network errors, only on authentication failures
            if (axios.isAxiosError(err) && err.response && (err.response.status === 401 || err.response.status === 403)) {
              await AsyncStorage.removeItem("user");
              const accessToken = await SecureStore.getItemAsync("accessToken");
              if (accessToken) await SecureStore.deleteItemAsync("accessToken");
              const refreshToken = await SecureStore.getItemAsync("refreshToken");
              if (refreshToken) await SecureStore.deleteItemAsync("refreshToken");
              router.push("/auth/register/start");
            }
          }
        }
      }
      return Promise.reject(error);
    }
  );

  return client;
};

const getHttpClient = async (): Promise<AxiosInstance> => {
  if (!httpClient) {
    httpClient = await initializeHttpClient();
  }
  return httpClient;
};

/* -------------------------------------------------------------
   API Methods

   Authentication Methods
------------------------------------------------------------- */

// Email login: initiate login by sending the user's email.
export const authLoginWithEmail = async (email: string) => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.post("/auth/email", { email });
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

// Email registration: register with additional user details.
export const authRegisterWithEmail = async (email: string, firstName: string, lastName: string, termsVersion?: string, trackingEnabled?: boolean) => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.post("/auth/email", {
      email,
      firstName,
      lastName,
      termsVersion,
      trackingEnabled,
    });
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    console.error(error.response);
    return { success: false, error: error.response };
  }
};

// Verify OTP sent to email.
export const authVerifyOTP = async (email: string, otp: string) => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.post("/auth/verify-otp", { email, otp });
    if (res.data.success) {
      await SecureStore.setItemAsync("accessToken", res.headers["accesstoken"], {
        keychainAccessible: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
      });
      await SecureStore.setItemAsync("refreshToken", res.headers["refreshtoken"], {
        keychainAccessible: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
      });
    }
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

// Login using Apple credentials.
export const authLoginWithApple = async (
  appleId: string,
  email: string | null,
  emailManually: boolean,
  firstName?: string | null,
  lastName?: string | null,
  termsVersion?: string,
  trackingEnabled?: boolean
) => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.post("/auth/apple", {
      appleId,
      email,
      emailManually,
      firstName,
      lastName,
      termsVersion,
      trackingEnabled,
    });
    if (res.data.success && res.headers["accesstoken"] && res.headers["refreshtoken"]) {
      await SecureStore.setItemAsync("accessToken", res.headers["accesstoken"], {
        keychainAccessible: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
      });
      await SecureStore.setItemAsync("refreshToken", res.headers["refreshtoken"], {
        keychainAccessible: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
      });
    }
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

// Login using testing account password.
export const authLoginWithTestingAccount = async (userId: string, password: string) => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.post("/auth/testing-account", {
      userId,
      password,
    });
    if (res.data.success && res.headers["accesstoken"] && res.headers["refreshtoken"]) {
      await SecureStore.setItemAsync("accessToken", res.headers["accesstoken"], {
        keychainAccessible: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
      });
      await SecureStore.setItemAsync("refreshToken", res.headers["refreshtoken"], {
        keychainAccessible: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
      });
    }
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

// Logout the user and clear tokens.
export const authLogout = async () => {
  try {
    const httpClient = await getHttpClient();
    const accessToken = await SecureStore.getItemAsync("accessToken");
    if (accessToken) {
      const res = await httpClient.post(
        "/auth/logout",
        {}, // Empty body
        {
          headers: {
            accessToken: accessToken,
          },
        }
      );

      // Clear all authentication tokens
      await SecureStore.deleteItemAsync("accessToken");
      await SecureStore.deleteItemAsync("refreshToken");

      // Also clear terms acceptance when logging out
      await clearTermsAcceptance();

      return { success: res.data.success, data: res.data.data };
    }
    return { success: true, data: null };
  } catch (error: any) {
    // Still clear tokens even if the server request fails
    await SecureStore.deleteItemAsync("accessToken");
    await SecureStore.deleteItemAsync("refreshToken");
    await clearTermsAcceptance();
    return { success: false, error: error.response };
  }
};

/* -------------------------------------------------------------
   Club Methods
------------------------------------------------------------- */

// Fetch the user's clubs.
export const clubsFetch = async () => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.get("/clubs", {
      headers: {
        accessToken: await SecureStore.getItemAsync("accessToken"),
      },
    });
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

// Validate a club code.
export const clubFetch = async (clubCode: string) => {
  try {
    const httpClient = await getHttpClient();
    const accessToken = await SecureStore.getItemAsync("accessToken");
    const res = await httpClient.get(`/clubs/${clubCode}`, {
      headers: {
        accessToken: accessToken,
      },
    });
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

// Join a club using its code.
export const clubJoin = async (clubCode: string) => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.put(
      `/clubs/${clubCode}/join`,
      {},
      {
        headers: {
          accessToken: await SecureStore.getItemAsync("accessToken"),
        },
      }
    );
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

// Return the role of the user in the club.
export const clubFetchParticipants = async (clubCode: string) => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.get(`/clubs/${clubCode}/participants`, {
      headers: {
        accessToken: await SecureStore.getItemAsync("accessToken"),
      },
    });
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

/* -------------------------------------------------------------
   Match Methods
------------------------------------------------------------- */

// Retrieve all matches for the user.
export const matchesFetchAll = async () => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.get("/matches", {
      headers: {
        accessToken: await SecureStore.getItemAsync("accessToken"),
      },
    });
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

// Retrieve a specific match by its ID.
export const matchFetch = async (id: string) => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.get(`/matches/${id}`, {
      headers: {
        accessToken: await SecureStore.getItemAsync("accessToken"),
      },
    });
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

// Create a new match.
export const matchCreate = async (match: Match) => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.post("/matches", match, {
      headers: {
        accessToken: await SecureStore.getItemAsync("accessToken"),
      },
    });
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

// Update an existing match.
export const matchUpdate = async (match: Match) => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.patch(`/matches/${match._id}`, match, {
      headers: {
        accessToken: await SecureStore.getItemAsync("accessToken"),
      },
    });
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

// Delete an existing match.
export const matchDelete = async (id: string) => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.delete(`/matches/${id}`, {
      headers: {
        accessToken: await SecureStore.getItemAsync("accessToken"),
      },
    });
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

// Add a participant to a match.
export const matchAddParticipant = async (id: string, participant: string) => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.put(`/matches/${id}/participants/${participant}`, {
      headers: {
        accessToken: await SecureStore.getItemAsync("accessToken"),
      },
    });
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

// Remove a participant from a match.
export const matchRemoveParticipant = async (id: string, participant: string) => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.delete(`/matches/${id}/participants/${participant}`, {
      headers: {
        accessToken: await SecureStore.getItemAsync("accessToken"),
      },
    });
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

/* -------------------------------------------------------------
   User Absence Methods
------------------------------------------------------------- */

// Retrieve the user's absence records.
export const absenceFetchRecords = async () => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.get("/users/absences", {
      headers: {
        accessToken: await SecureStore.getItemAsync("accessToken"),
      },
    });
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

// Create a new absence period.
export const absenceCreatePeriod = async (startDate: Date, endDate: Date) => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.post(
      "/users/absences/periods",
      { startDate, endDate },
      {
        headers: {
          accessToken: await SecureStore.getItemAsync("accessToken"),
        },
      }
    );
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

// Delete an existing absence period.
export const absenceDeletePeriod = async (id: string) => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.delete(`/users/absences/periods/id/${id}`, {
      headers: {
        accessToken: await SecureStore.getItemAsync("accessToken"),
      },
    });
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

// Update the absence days (e.g., marking specific days as absent).
export const absenceUpdateDays = async (absenceDays: Record<string, boolean>) => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.put(
      "/users/absences/days",
      { absenceDays },
      {
        headers: {
          accessToken: await SecureStore.getItemAsync("accessToken"),
        },
      }
    );
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

/* -------------------------------------------------------------
   User Data Methods
------------------------------------------------------------- */

// Retrieve the user's data.
export const userDataFetch = async () => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.get("/users", {
      headers: {
        accessToken: await SecureStore.getItemAsync("accessToken"),
      },
    });
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

// Update the user's data.
export const userDataUpdate = async (
  data: {
    email?: string;
    firstName?: string;
    lastName?: string;
    pushNotificationToken?: string;
    pushNotificationBadgeCount?: number;
  },
  performLogout: boolean
) => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.patch("/users", data, {
      headers: {
        accessToken: await SecureStore.getItemAsync("accessToken"),
      },
    });
    if (res.data.success && performLogout) {
      await authLogout();
      await authLoginWithEmail(data.email || "");
    }
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

// Delete the user's account.
export const userDeleteAccount = async () => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.delete("/users", {
      headers: {
        accessToken: await SecureStore.getItemAsync("accessToken"),
      },
    });

    if (res.data.success) {
      await SecureStore.deleteItemAsync("accessToken");
      await SecureStore.deleteItemAsync("refreshToken");

      await clearTermsAcceptance();
    }
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

/* -------------------------------------------------------------
   Notification Settings Methods
------------------------------------------------------------- */

// Fetch notification settings
export const notificationFetchSettings = async () => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.get("/users/notifications", {
      headers: {
        accessToken: await SecureStore.getItemAsync("accessToken"),
      },
    });
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

// Update notification settings
export const notificationUpdateSettings = async (pushEnabled: boolean, emailEnabled: boolean) => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.put(
      "/users/notifications",
      { pushEnabled, emailEnabled },
      {
        headers: {
          accessToken: await SecureStore.getItemAsync("accessToken"),
        },
      }
    );
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

/* -------------------------------------------------------------
   Terms & Conditions Methods
------------------------------------------------------------- */

// Get user's current terms acceptance status
export const termsFetchStatus = async () => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.get("/users/terms", {
      headers: {
        accessToken: await SecureStore.getItemAsync("accessToken"),
      },
    });
    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};

// Update terms acceptance status
export const termsUpdate = async (version: string, trackingEnabled: boolean) => {
  try {
    const httpClient = await getHttpClient();
    const res = await httpClient.put(
      "/users/terms",
      { version, trackingEnabled },
      {
        headers: {
          accessToken: await SecureStore.getItemAsync("accessToken"),
        },
      }
    );

    // If successful, also update local storage
    if (res.data.success) {
      await saveTermsAcceptance(trackingEnabled);
    }

    return { success: res.data.success, data: res.data.data };
  } catch (error: any) {
    return { success: false, error: error.response };
  }
};
