import { ThemedText } from "@/components/ThemedText";
import { Colors } from "@/constants/Colors";
import { useState, useEffect } from "react";
import {
  View,
  StyleSheet,
  KeyboardAvoidingView,
  Platform,
  TouchableWithoutFeedback,
  Keyboard,
  Pressable,
  Text,
  Modal,
  ScrollView,
  Linking,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { AnimatedHapticButton } from "@/components/AnimatedHapticButton";
import { router } from "expo-router";
import analytics from "@react-native-firebase/analytics";
import { saveTermsAcceptance, CURRENT_TERMS_VERSION, TERMS_ACCEPTANCE_KEY } from "@/utils/TermsManager";
import { useRegistration } from "@/hooks/useRegistration";
import AsyncStorage from "@react-native-async-storage/async-storage";
import * as SecureStore from "expo-secure-store";
import { needsTermsUpdate, getTermsChangesSince } from "@/utils/TermsManager";
import { termsUpdate } from "@/api/Api";

export default function RegisterUserInfo() {
  const { data: registrationData, updateData } = useRegistration();
  const [termsAccepted, setTermsAccepted] = useState(false);
  const [trackingAccepted, setTrackingAccepted] = useState(false);
  const [isUpdate, setIsUpdate] = useState(false);
  const [changedSince, setChangedSince] = useState<{ version: string; releaseDate: string }[]>([]);

  // Check if this is an update to existing terms
  useEffect(() => {
    const checkIfUpdate = async () => {
      try {
        const needsUpdate = await needsTermsUpdate();
        if (needsUpdate) {
          // It's an update if we need an update but already have acceptance data
          const termsDataStr = await AsyncStorage.getItem(TERMS_ACCEPTANCE_KEY);
          setIsUpdate(!!termsDataStr && termsDataStr !== "true");

          // Get tracking preference if it's an update
          if (termsDataStr && termsDataStr !== "true") {
            const termsData = JSON.parse(termsDataStr);
            setTrackingAccepted(termsData.trackingEnabled);

            // Get history of changes since last accepted version
            const changes = await getTermsChangesSince();
            setChangedSince(changes);
          }
        }
      } catch (error) {
        console.error("Error checking terms status:", error);
      }
    };

    checkIfUpdate();
  }, []);

  const handleContinue = async () => {
    if (trackingAccepted) {
      await analytics().setAnalyticsCollectionEnabled(true);
      await analytics().setConsent({
        ad_storage: false,
        ad_user_data: false,
        ad_personalization: false,
        analytics_storage: true,
        functionality_storage: true,
        personalization_storage: true,
        security_storage: true
      });      
    } else {
      await analytics().setAnalyticsCollectionEnabled(false);
    }

    if (!termsAccepted) return;

    // Save terms acceptance with version info
    await saveTermsAcceptance(trackingAccepted);

    // Store tracking preference in registration context for later use
    updateData({ trackingEnabled: trackingAccepted });

    // Also update on server if user is logged in (for version updates)
    try {
      const accessToken = await SecureStore.getItemAsync("accessToken");
      if (accessToken) {
        await termsUpdate(CURRENT_TERMS_VERSION, trackingAccepted);
      }
    } catch (error) {
      console.error("Failed to update terms on server:", error);
    }

    // If the user is not in registration process, navigate to home
    if (registrationData.isRegistration) {
      router.push("/auth/register/club-code");
    } else {
      if (router.canDismiss()) {
        router.dismissAll();
      }
      router.push("/");
    }
  };

  const handleBack = () => {
    router.push("/auth/register/start");
  };

  const Section: React.FC<{ title: string; children: React.ReactNode }> = ({ title, children }) => (
    <View style={styles.sectionContainer}>
      <ThemedText style={styles.sectionTitle}>{title}</ThemedText>
      {children}
    </View>
  );

  const handleLink = (url: string) => {
    Linking.openURL(url).catch(() => {
      console.error("Failed to open URL:", url);
    });
  };

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === "ios" ? "padding" : "height"}
      keyboardVerticalOffset={Platform.OS === "ios" ? 100 : 0}
    >
      <TouchableWithoutFeedback onPress={Keyboard.dismiss}>
        <View style={styles.inner}>
          <View>
            <ThemedText type="title" style={{ textAlign: "center", fontSize: 32 }}>
              {isUpdate ? "Aktualisierte Nutzungsbedingungen & Datenschutz" : "Nutzungsbedingungen & Datenschutz"}
            </ThemedText>
            <ThemedText
              style={{
                textAlign: "center",
                fontSize: 14,
                marginTop: 8,
                color: Colors.muted,
              }}
            >
              Version {CURRENT_TERMS_VERSION}
            </ThemedText>
          </View>

          {isUpdate && changedSince.length > 0 && (
            <ScrollView style={styles.changesContainer}>
              <ThemedText style={styles.changesTitle}>Änderungen seit deiner letzten Zustimmung:</ThemedText>
              {changedSince.map((change) => (
                <View style={styles.changeItem} key={change.version}>
                  <ThemedText style={styles.changeVersion}>
                    Version {change.version} ({change.releaseDate})
                  </ThemedText>
                </View>
              ))}
            </ScrollView>
          )}

          <View style={styles.form}>
            <View style={styles.termsContainer}>
              <Pressable style={styles.checkbox} onPress={() => setTermsAccepted(!termsAccepted)}>
                <View style={[styles.checkboxInner, termsAccepted && styles.checkboxChecked]}>
                  {termsAccepted && <Text style={styles.checkmark}>✓</Text>}
                </View>
              </Pressable>
              <ThemedText style={styles.termsText}>
                Ich akzeptiere die{" "}
                <Text style={styles.termsLink} onPress={() => handleLink("https://tt-plus.ch/nutzungsbedinungen")}>
                  Nutzungsbedingungen
                </Text>{" "}
                und{" "}
                <Text style={styles.termsLink} onPress={() => handleLink("https://tt-plus.ch/datenschutz")}>
                  Datenschutzerklärung
                </Text>
              </ThemedText>
            </View>

            <View style={styles.termsContainer}>
              <Pressable style={styles.checkbox} onPress={() => setTrackingAccepted(!trackingAccepted)}>
                <View style={[styles.checkboxInner, trackingAccepted && styles.checkboxChecked]}>
                  {trackingAccepted && <Text style={styles.checkmark}>✓</Text>}
                </View>
              </Pressable>
              <ThemedText style={styles.termsText}>Ich erlaube TT+, meine Nutzungsdaten zu analysieren, um die App zu verbessern.</ThemedText>
            </View>

            <View style={styles.buttonContainer}>
              {registrationData.isRegistration && (
                <AnimatedHapticButton style={styles.buttonOutline} onPress={handleBack}>
                  <Ionicons name="arrow-back" size={20} color={Colors.text} />
                </AnimatedHapticButton>
              )}
              <AnimatedHapticButton
                disabled={!termsAccepted}
                onPress={handleContinue}
                style={[styles.button, !termsAccepted && styles.buttonDisabled]}
              >
                <ThemedText style={styles.buttonText}>Weiter</ThemedText>
              </AnimatedHapticButton>
            </View>
          </View>
          <View style={{ flex: 1 }} />
        </View>
      </TouchableWithoutFeedback>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    paddingHorizontal: 24,
  },
  inner: {
    flex: 1,
    flexGrow: 1,
    justifyContent: "flex-end",
    marginTop: 128,
  },
  form: {
    gap: 6,
    width: "100%",
    marginTop: 200,
  },
  input: {
    borderColor: Colors.text,
    borderWidth: 2,
    borderRadius: 15,
    paddingVertical: 12,
    paddingHorizontal: 16,
    fontFamily: "Bilo",
    width: "100%",
  },
  buttonContainer: {
    flexDirection: "row",
    width: "100%",
    gap: 12,
    justifyContent: "space-between",
    flexWrap: "wrap",
  },
  button: {
    flex: 1,
    height: 48,
    backgroundColor: Colors.text,
    borderRadius: 15,
    justifyContent: "center",
    alignItems: "center",
  },
  buttonOutline: {
    height: 48,
    borderColor: Colors.text,
    borderWidth: 2,
    borderRadius: 15,
    justifyContent: "center",
    alignItems: "center",
    paddingHorizontal: 12,
  },
  buttonDisabled: {
    opacity: 0.5,
  },
  buttonText: {
    color: Colors.background,
  },
  resendButton: {
    height: 48,
    justifyContent: "center",
    alignItems: "center",
  },
  inputError: {
    borderColor: Colors.error,
  },
  errorText: {
    color: Colors.error,
    fontSize: 14,
    marginTop: 2,
  },
  globalError: {
    backgroundColor: Colors.error,
    padding: 10,
    borderRadius: 15,
  },
  globalErrorText: {
    color: Colors.background,
    textAlign: "center",
  },
  termsContainer: {
    flexDirection: "row",
    alignItems: "center",
    marginBottom: 16,
  },
  checkbox: {
    marginRight: 8,
  },
  checkboxInner: {
    width: 20,
    height: 20,
    borderWidth: 2,
    borderColor: Colors.text,
    borderRadius: 4,
    justifyContent: "center",
    alignItems: "center",
  },
  checkboxChecked: {
    backgroundColor: Colors.text,
  },
  checkmark: {
    color: Colors.background,
    fontSize: 14,
    fontWeight: "bold",
  },
  termsText: {
    fontSize: 16,
    flex: 1,
  },
  termsLink: {
    color: Colors.text,
    textDecorationLine: "underline",
  },
  modalContainer: {
    flex: 1,
    backgroundColor: Colors.background,
    marginTop: 70,
    paddingHorizontal: 24,
  },
  modalHeader: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    paddingTop: 16,
    paddingBottom: 16,
  },
  modalTitle: {
    fontSize: 24,
  },
  closeButton: {
    width: 48,
    height: 48,
    justifyContent: "center",
    alignItems: "center",
  },
  modalContent: {
    flex: 1,
    padding: 24,
  },
  modalText: {
    fontSize: 16,
    lineHeight: 24,
    marginBottom: 16,
  },
  changesContainer: {
    maxHeight: 150,
    marginTop: 20,
    marginBottom: 10,
    borderWidth: 1,
    borderColor: Colors.muted,
    borderRadius: 8,
    padding: 10,
  },
  changesTitle: {
    fontWeight: "bold",
    marginBottom: 8,
  },
  changeItem: {
    marginBottom: 10,
  },
  changeVersion: {
    fontWeight: "bold",
  },
  changeDescription: {
    fontSize: 14,
  },
  sectionContainer: {
    marginBottom: 24,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: "600",
    marginBottom: 8,
  },
  paragraph: {
    fontSize: 14,
    lineHeight: 22,
    marginBottom: 12,
    color: "#111827",
  },
  list: {
    marginBottom: 12,
    paddingLeft: 12,
  },
  listItem: {
    fontSize: 14,
    lineHeight: 22,
    marginBottom: 4,
    color: "#111827",
  },
  externalLink: {
    alignSelf: "center",
    marginVertical: 4,
  },
  link: {
    fontSize: 14,
    color: Colors.muted,
    textDecorationLine: "underline",
  },
  linkInline: {
    color: Colors.muted,
    textDecorationLine: "underline",
  },
  italic: {
    fontStyle: "italic",
  },
});
