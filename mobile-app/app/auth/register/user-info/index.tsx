import { ThemedText } from "@/components/ThemedText";
import { Colors } from "@/constants/Colors";
import { useState } from "react";
import {
  View,
  StyleSheet,
  TextInput,
  KeyboardAvoidingView,
  Platform,
  TouchableWithoutFeedback,
  Keyboard,
  Pressable,
  Text,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { AnimatedHapticButton } from "@/components/AnimatedHapticButton";
import { router } from "expo-router";
import { Spinner } from "@/components/Spinner";
import { authLoginWithApple, authRegisterWithEmail } from "@/api/Api";
import { CURRENT_TERMS_VERSION } from "@/utils/TermsManager";
import { useRegistration } from "@/hooks/useRegistration";

export default function RegisterUserInfo() {
  const { data: registrationData, updateData } = useRegistration();
  const [email, setEmail] = useState(registrationData.email || "");
  const [firstName, setFirstName] = useState(registrationData.firstName || "");
  const [lastName, setLastName] = useState(registrationData.lastName || "");
  const [emailError, setEmailError] = useState("");
  const [firstNameError, setFirstNameError] = useState("");
  const [lastNameError, setLastNameError] = useState("");
  const [globalError, setGlobalError] = useState("");
  const [isFetching, setIsFetching] = useState(false);

  const handleContinue = async () => {
    let hasError = false;

    if (registrationData.appleUserId && !email.trim()) {
      setEmailError("Bitte gib deine E-Mail-Adresse ein");
      hasError = true;
    } else {
      setEmailError("");
    }

    if (!firstName.trim()) {
      setFirstNameError("Bitte gib deinen Vornamen ein");
      hasError = true;
    } else {
      setFirstNameError("");
    }

    if (!lastName.trim()) {
      setLastNameError("Bitte gib deinen Nachnamen ein");
      hasError = true;
    } else {
      setLastNameError("");
    }

    if (!hasError) {
      setIsFetching(true);
      if (registrationData.email) {
        const { success, error } = await authRegisterWithEmail(
          registrationData.email,
          firstName,
          lastName,
          CURRENT_TERMS_VERSION, // Pass terms version
          registrationData.trackingEnabled // Pass tracking preference
        );
        setIsFetching(false);

        if (success) {
          updateData({ firstName, lastName });
          router.push("/auth/register/verify" as any);
        } else {
          setGlobalError(error.message || "Fehler bei der Registrierung");
        }
      } else if (registrationData.appleUserId) {
        const { success, error } = await authLoginWithApple(
          registrationData.appleUserId,
          email,
          true,
          firstName,
          lastName,
          CURRENT_TERMS_VERSION, // Pass terms version
          registrationData.trackingEnabled // Pass tracking preference
        );
        setIsFetching(false);

        if (success) {
          updateData({ email, firstName, lastName });
          router.push("/auth/register/verify" as any);
        } else {
          const message = JSON.parse(error.request._response);
          setGlobalError(message.message || "Fehler bei der Registrierung");
        }
      }
    }
  };

  const handleBack = () => {
    router.push("/auth/register/start");
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
              Sag uns, wer du bist!
            </ThemedText>
          </View>

          <View
            style={[
              styles.form,
              { marginTop: registrationData.appleUserId ? 38 : 90 },
            ]}
          >
            <ThemedText>
              Mit deinem Namen können dich deine Teamkollegen leichter finden.
            </ThemedText>
            {globalError && (
              <View style={styles.globalError}>
                <ThemedText style={styles.globalErrorText}>
                  {globalError}
                </ThemedText>
              </View>
            )}
            {registrationData.appleUserId && (
              <View>
                <TextInput
                  placeholder="E-Mail-Adresse"
                  value={email}
                  onChangeText={setEmail}
                  style={[styles.input, emailError ? styles.inputError : null]}
                  placeholderTextColor={Colors.text}
                  autoCapitalize="none"
                  keyboardType="email-address"
                  autoComplete="email"
                />
                {emailError && (
                  <ThemedText style={styles.errorText}>{emailError}</ThemedText>
                )}
              </View>
            )}

            <View>
              <TextInput
                placeholder="Vorname"
                value={firstName}
                onChangeText={setFirstName}
                style={[
                  styles.input,
                  firstNameError ? styles.inputError : null,
                ]}
                placeholderTextColor={Colors.text}
                autoCapitalize="words"
                autoComplete="given-name"
              />
              {firstNameError && (
                <ThemedText style={styles.errorText}>
                  {firstNameError}
                </ThemedText>
              )}
            </View>

            <View>
              <TextInput
                placeholder="Nachname"
                value={lastName}
                onChangeText={setLastName}
                style={[styles.input, lastNameError ? styles.inputError : null]}
                placeholderTextColor={Colors.text}
                autoCapitalize="words"
                autoComplete="family-name"
              />
              {lastNameError && (
                <ThemedText style={styles.errorText}>
                  {lastNameError}
                </ThemedText>
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
                disabled={
                  !!(!firstName.trim() || !lastName.trim() || isFetching)
                }
                onPress={handleContinue}
                style={[
                  styles.button,
                  !!(!firstName.trim() || !lastName.trim())
                    ? styles.buttonDisabled
                    : null,
                ]}
              >
                {isFetching ? (
                  <Spinner fill={Colors.background} />
                ) : (
                  <ThemedText style={styles.buttonText}>Weiter</ThemedText>
                )}
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
    gap: 12,
    width: "100%",
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
  buttonText: {
    color: Colors.background,
  },
  buttonDisabled: {
    opacity: 0.5,
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
});
