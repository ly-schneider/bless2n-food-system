import AsyncStorage from "@react-native-async-storage/async-storage";
import * as SecureStore from "expo-secure-store";

// Current version of the terms
export const CURRENT_TERMS_VERSION = "1.1.0";

// Storage keys
export const TERMS_ACCEPTANCE_KEY = "termsAcceptance";

interface TermsAcceptance {
  version: string;
  acceptedAt: string; // ISO date string
  trackingEnabled: boolean;
}

// Terms version history log - tracks when terms were updated
const TERMS_HISTORY = [
  {
    version: "1.0.0",
    releaseDate: "12. April 2025"
  },
  {
    version: "1.0.4",
    releaseDate: "5. Mai 2025"
  },
  {
    version: "1.1.0",
    releaseDate: "6. Mai 2025"
  },
];

/**
 * Save terms acceptance status
 */
export const saveTermsAcceptance = async (
  trackingEnabled: boolean
): Promise<void> => {
  try {
    const termsData: TermsAcceptance = {
      version: CURRENT_TERMS_VERSION,
      acceptedAt: new Date().toISOString(),
      trackingEnabled,
    };

    // Store in both places for redundancy
    await AsyncStorage.setItem(TERMS_ACCEPTANCE_KEY, JSON.stringify(termsData));
    await SecureStore.setItemAsync(
      TERMS_ACCEPTANCE_KEY,
      JSON.stringify(termsData),
      {
        keychainAccessible: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
      }
    );
  } catch (error) {
    console.error("Failed to save terms acceptance:", error);
  }
};

/**
 * Check if the user has accepted the current terms
 */
export const hasAcceptedCurrentTerms = async (): Promise<boolean> => {
  try {
    // Try AsyncStorage first
    let termsDataStr = await AsyncStorage.getItem(TERMS_ACCEPTANCE_KEY);
    console.log("termsDataStr", termsDataStr);

    // If not found in AsyncStorage, try SecureStore
    if (!termsDataStr) {
      termsDataStr = await SecureStore.getItemAsync(TERMS_ACCEPTANCE_KEY);

      // If found in SecureStore but not in AsyncStorage, restore to AsyncStorage
      if (termsDataStr) {
        await AsyncStorage.setItem(TERMS_ACCEPTANCE_KEY, termsDataStr);
      }
    }

    // Check for legacy format (just storing "true")
    if (termsDataStr === "true") {
      return false; // Force reacceptance to update to new format
    }

    if (!termsDataStr) {
      return false;
    }

    const termsData: TermsAcceptance = JSON.parse(termsDataStr);

    // Compare versions
    return termsData.version === CURRENT_TERMS_VERSION;
  } catch (error) {
    console.error("Error checking terms acceptance:", error);
    return false;
  }
};

/**
 * Get tracking preference from terms acceptance
 */
export const getTrackingPreference = async (): Promise<boolean> => {
  try {
    // Try AsyncStorage first
    let termsDataStr = await AsyncStorage.getItem(TERMS_ACCEPTANCE_KEY);

    // If not found in AsyncStorage, try SecureStore
    if (!termsDataStr) {
      termsDataStr = await SecureStore.getItemAsync(TERMS_ACCEPTANCE_KEY);
    }

    if (!termsDataStr || termsDataStr === "true") {
      return false; // Default to false if not found or legacy format
    }

    const termsData: TermsAcceptance = JSON.parse(termsDataStr);
    return termsData.trackingEnabled;
  } catch (error) {
    console.error("Error getting tracking preference:", error);
    return false;
  }
};

/**
 * Clear terms acceptance (useful for testing or logout)
 */
export const clearTermsAcceptance = async (): Promise<void> => {
  try {
    await AsyncStorage.removeItem(TERMS_ACCEPTANCE_KEY);
    await SecureStore.deleteItemAsync(TERMS_ACCEPTANCE_KEY);
  } catch (error) {
    console.error("Failed to clear terms acceptance:", error);
  }
};

/**
 * Check if a new terms version is available that requires user acceptance
 * @param providedVersion Optional explicit version to check against current version
 */
export const needsTermsUpdate = async (
  providedVersion?: string
): Promise<boolean> => {
  try {
    // If a specific version was provided, compare it directly
    if (providedVersion) {
      return compareVersions(providedVersion, CURRENT_TERMS_VERSION) < 0;
    }

    // Otherwise check stored terms data
    const termsDataStr = await AsyncStorage.getItem(TERMS_ACCEPTANCE_KEY);
    if (!termsDataStr) return true;

    // If in legacy format, needs update
    if (termsDataStr === "true") return true;

    // Parse the terms data and compare versions
    const termsData: TermsAcceptance = JSON.parse(termsDataStr);
    return compareVersions(termsData.version, CURRENT_TERMS_VERSION) < 0;
  } catch (error) {
    console.error("Error checking terms version:", error);
    return true; // If in doubt, request acceptance
  }
};

/**
 * Compare two semantic version strings
 * @returns negative if v1 < v2, positive if v1 > v2, 0 if equal
 */
export function compareVersions(v1: string, v2: string): number {
  const parts1 = v1.split(".").map(Number);
  const parts2 = v2.split(".").map(Number);

  for (let i = 0; i < Math.max(parts1.length, parts2.length); i++) {
    const p1 = i < parts1.length ? parts1[i] : 0;
    const p2 = i < parts2.length ? parts2[i] : 0;

    if (p1 !== p2) {
      return p1 - p2;
    }
  }

  return 0;
}

/**
 * Get terms history changes since the user's accepted version
 */
export const getTermsChangesSince = async (): Promise<
  { version: string; releaseDate: string }[]
> => {
  try {
    const termsDataStr = await AsyncStorage.getItem(TERMS_ACCEPTANCE_KEY);
    if (!termsDataStr) return TERMS_HISTORY;

    // If legacy format or no data, return all history
    if (termsDataStr === "true") return TERMS_HISTORY;

    const termsData: TermsAcceptance = JSON.parse(termsDataStr);
    return TERMS_HISTORY.filter(
      (entry) => compareVersions(entry.version, termsData.version) > 0
    );
  } catch (error) {
    console.error("Error getting terms changes:", error);
    return TERMS_HISTORY; // Return all if there's an error
  }
};
