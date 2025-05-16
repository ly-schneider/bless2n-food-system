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
} from "react-native";
import { AnimatedHapticButton } from "@/components/AnimatedHapticButton";
import { router } from "expo-router";
import { Spinner } from "@/components/Spinner";
import { authLoginWithTestingAccount } from "@/api/Api";
import { useAuth } from "@/hooks/useAuth";
import { Ionicons } from "@expo/vector-icons";
import { useRegistration } from "@/hooks/useRegistration";

export default function RegisterTestingAccount() {
  const { data: registrationData } = useRegistration();
  const { saveUser } = useAuth();
  const [password, setPassword] = useState("");
  const [passwordError, setPasswordError] = useState("");
  const [globalError, setGlobalError] = useState("");
  const [isFetching, setIsFetching] = useState(false);

  const handleContinue = async () => {
    if (!password.trim()) {
      setPasswordError("Bitte gib dein Passwort ein");
      return;
    }

    setPasswordError("");
    setIsFetching(true);

    const { success, data, error } = await authLoginWithTestingAccount(
      registrationData.userId!,
      password
    );

    setIsFetching(false);

    if (success) {
      await saveUser(data._id);
      if (router.canDismiss()) {
        router.dismissAll();
      }
      router.replace("/");
    } else {
      const message = JSON.parse(error.request._response);
      setGlobalError(message.message || "Fehler bei der Anmeldung");
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
              Test-Account Passwort
            </ThemedText>
          </View>

          <View style={styles.form}>
            <ThemedText>
              Bitte gib dein Test-Account Passwort ein, um fortzufahren.
            </ThemedText>
            <View>
              <TextInput
                placeholder="Passwort"
                value={password}
                onChangeText={(text) => {
                  setPassword(text);
                  setPasswordError("");
                }}
                style={[styles.input, passwordError ? styles.inputError : null]}
                placeholderTextColor={Colors.text}
                secureTextEntry
                autoCapitalize="none"
                autoComplete="password"
              />
              {passwordError && (
                <ThemedText style={styles.errorText}>
                  {passwordError}
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
                disabled={!password.trim() || isFetching}
                onPress={handleContinue}
                style={styles.button}
              >
                {isFetching ? (
                  <Spinner fill={Colors.background} />
                ) : (
                  <ThemedText style={{ color: Colors.background }}>
                    Fortfahren
                  </ThemedText>
                )}
              </AnimatedHapticButton>
            </View>
            {globalError && (
              <View style={{ marginTop: 12 }}>
                <ThemedText style={styles.globalErrorText}>
                  {globalError}
                </ThemedText>
              </View>
            )}
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
  inputError: {
    borderColor: Colors.error,
  },
  errorText: {
    color: Colors.error,
    fontSize: 14,
    marginTop: 2,
  },
  globalErrorText: {
    color: Colors.error,
    textAlign: "center",
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
  email: {
    fontSize: 16,
    fontFamily: "Bilo",
    textAlign: "center",
    opacity: 0.7,
  },
});
