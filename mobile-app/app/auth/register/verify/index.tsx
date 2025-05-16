import { ThemedText } from "@/components/ThemedText";
import { Colors } from "@/constants/Colors";
import { router, useGlobalSearchParams } from "expo-router";
import { useEffect, useState } from "react";
import {
  View,
  StyleSheet,
  TextInput,
  KeyboardAvoidingView,
  Platform,
  TouchableWithoutFeedback,
  Keyboard,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { AnimatedHapticButton } from "@/components/AnimatedHapticButton";
import { Spinner } from "@/components/Spinner";
import { useAuth } from "@/hooks/useAuth";
import { authLoginWithEmail, authVerifyOTP } from "@/api/Api";
import { useRegistration } from "@/hooks/useRegistration";
import { hasAcceptedCurrentTerms } from "@/utils/TermsManager";

export default function RegisterVerify() {
  const { email } = useGlobalSearchParams<{ email: string }>();

  const { saveUser } = useAuth();
  const { data: registrationData, updateData } = useRegistration();
  const [otp, setOtp] = useState("");
  const [resendCooldown, setResendCooldown] = useState(60);
  const [isResendActive, setIsResendActive] = useState(false);
  const [otpError, setOtpError] = useState("");
  const [isFetching, setIsFetching] = useState(false);

  useEffect(() => {
    if (resendCooldown > 0 && !isResendActive) {
      const timer = setInterval(() => {
        setResendCooldown((prev) => prev - 1);
      }, 1000);
      return () => clearInterval(timer);
    } else if (resendCooldown === 0) {
      setIsResendActive(true);
    }
  }, [resendCooldown, isResendActive]);

  const handleVerifyOTP = async () => {
    if (otp.length !== 6) return;

    const emailToVerify = email || registrationData.email;

    if (!emailToVerify) {
      router.push("/auth/register/start");
      return;
    }

    setIsFetching(true);
    const { success, data } = await authVerifyOTP(emailToVerify, otp);
    setIsFetching(false);

    if (success) {
      await saveUser(data._id);
      if (registrationData.isRegistration === true) {
        router.push("/auth/register/club-code");
      } else if (email) {
        router.replace("/menu/user-data");
      } else {
        if (router.canDismiss()) {
          router.dismissAll();
        }
        if (await hasAcceptedCurrentTerms()) {
          router.replace("/");
        } else {
          router.replace("/auth/register/terms");
        }
      }
    } else {
      setOtpError("Fehler bei der Bestätigung");
    }
  };

  const handleResendOTP = async () => {
    const emailToVerify = email || registrationData.email;

    if (!emailToVerify) {
      router.push("/auth/register/start");
      return;
    }

    if (!isResendActive) return;

    const { success, error } = await authLoginWithEmail(emailToVerify);

    if (!success) {
      setOtpError(error);
    }

    setResendCooldown(60);
    setIsResendActive(false);
  };

  const handleBack = () => {
    router.back();
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
            <ThemedText type="title" style={{ textAlign: "center" }}>
              {registrationData.isRegistration === true
                ? "E-Mail-Adresse bestätigen"
                : "Zwei-Faktor-Authentifizierung"}
            </ThemedText>
          </View>

          <View style={styles.form}>
            <ThemedText>
              Schau in dein Postfach und gib den Bestätigungscode ein.
            </ThemedText>
            <View>
              <TextInput
                placeholder="Bestätigungscode eingeben"
                value={otp}
                onChangeText={(text) => setOtp(text.replace(/[^0-9]/g, ""))}
                style={[styles.input, otpError ? styles.inputError : null]}
                placeholderTextColor={Colors.text}
                keyboardType="number-pad"
                maxLength={6}
                autoComplete="one-time-code"
              />
              {otpError && (
                <ThemedText style={styles.errorText}>{otpError}</ThemedText>
              )}
            </View>
            <View style={styles.buttonContainer}>
              <AnimatedHapticButton
                style={styles.buttonOutline}
                onPress={handleBack}
              >
                <Ionicons name="arrow-back" size={20} color={Colors.text} />
              </AnimatedHapticButton>
              <AnimatedHapticButton
                disabled={otp.length !== 6 || isFetching}
                onPress={handleVerifyOTP}
                style={styles.button}
              >
                {isFetching ? (
                  <Spinner fill={Colors.background} />
                ) : (
                  <ThemedText style={styles.buttonText}>Weiter</ThemedText>
                )}
              </AnimatedHapticButton>
            </View>
            <AnimatedHapticButton
              style={[
                styles.resendButton,
                !isResendActive && styles.buttonDisabled,
              ]}
              onPress={handleResendOTP}
              disabled={!isResendActive}
            >
              <ThemedText style={styles.resendButtonText}>
                {isResendActive
                  ? "Erneut senden"
                  : `Erneut senden in ${resendCooldown}s`}
              </ThemedText>
            </AnimatedHapticButton>
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
    gap: 12,
    width: "100%",
    marginTop: 148,
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
  resendButtonText: {
    color: Colors.text,
  },
  inputError: {
    borderColor: Colors.error,
  },
  errorText: {
    color: Colors.error,
    fontSize: 14,
    marginTop: 2,
  },
});
