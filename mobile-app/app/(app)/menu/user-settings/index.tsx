import {
  Keyboard,
  KeyboardAvoidingView,
  Modal,
  Platform,
  RefreshControl,
  ScrollView,
  StyleSheet,
  TextInput,
  TouchableWithoutFeedback,
  View,
} from "react-native";
import { AnimatedHapticButton } from "@/components/AnimatedHapticButton";
import { Ionicons } from "@expo/vector-icons";
import { Colors } from "@/constants/Colors";
import { router } from "expo-router";
import { ThemedText } from "@/components/ThemedText";
import React, { useCallback, useEffect, useState } from "react";
import { Spinner } from "@/components/Spinner";
import { userDataFetch, userDataUpdate, userDeleteAccount } from "@/api/Api";
import { SafeAreaProvider, SafeAreaView } from "react-native-safe-area-context";
import { useAuth } from "@/hooks/useAuth";

export default function MenuUserSettings() {
  const { logoutUser } = useAuth();
  const [originalUserData, setOriginalUserData] = useState({
    email: "",
    firstName: "",
    lastName: "",
  });
  const [isAppleUser, setIsAppleUser] = useState(false);
  const [email, setEmail] = useState("");
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [emailError, setEmailError] = useState("");
  const [firstNameError, setFirstNameError] = useState("");
  const [lastNameError, setLastNameError] = useState("");
  const [globalError, setGlobalError] = useState("");
  const [isFetching, setIsFetching] = useState(false);
  const [refreshing, setRefreshing] = useState(false);

  const [showDeleteAccountConfirmation, setShowDeleteAccountConfirmation] = useState(false);

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await fetchUserData();
    setRefreshing(false);
  }, []);

  useEffect(() => {
    fetchUserData();
  }, []);

  async function fetchUserData() {
    const { success, data, error } = await userDataFetch();
    if (success) {
      setOriginalUserData(data);
      setIsAppleUser(!!data.appleId);
      setEmail(data.email || "");
      setFirstName(data.firstName || "");
      setLastName(data.lastName || "");
    } else {
      setGlobalError(error.message || "Fehler beim Laden der Benutzerdaten");
    }
  }

  async function handleSave() {
    let hasError = false;

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

    if (!isAppleUser && !email.trim()) {
      setEmailError("Bitte gib deine E-Mail-Adresse ein");
      hasError = true;
    } else {
      setEmailError("");
    }

    if (!hasError) {
      setIsFetching(true);

      const body = {
        email: email !== originalUserData.email ? email : undefined,
        firstName: firstName !== originalUserData.firstName ? firstName : undefined,
        lastName: lastName !== originalUserData.lastName ? lastName : undefined,
      };

      const performLogout = !isAppleUser && email !== originalUserData.email;
      const { success, error } = await userDataUpdate(body, performLogout);
      setIsFetching(false);

      if (success) {
        setGlobalError("");
        if (performLogout) {
          router.push({
            pathname: "/auth/register/verify",
            params: { email },
          });
        } else {
          await fetchUserData();
        }
      } else {
        setGlobalError(error.message || "Fehler beim Speichern der Benutzerdaten");
      }
    }
  }

  const hasDataChanged = () => {
    return email !== originalUserData.email || firstName !== originalUserData.firstName || lastName !== originalUserData.lastName;
  };

  const handleBack = () => {
    router.back();
  };

  return (
    <React.Fragment>
      <KeyboardAvoidingView
        style={styles.container}
        behavior={Platform.OS === "ios" ? "padding" : "height"}
        keyboardVerticalOffset={Platform.OS === "ios" ? 100 : 0}
      >
        <TouchableWithoutFeedback onPress={Keyboard.dismiss}>
          <ScrollView
            contentContainerStyle={styles.inner}
            alwaysBounceVertical={true}
            refreshControl={<RefreshControl refreshing={refreshing} onRefresh={onRefresh} tintColor={Colors.text} colors={[Colors.text]} />}
          >
            <View style={styles.header}>
              <View style={styles.headerButtons}>
                <AnimatedHapticButton style={styles.buttonOutline} onPress={handleBack}>
                  <Ionicons name="arrow-back" size={20} color={Colors.text} />
                </AnimatedHapticButton>
                <ThemedText type="title" style={styles.headerTitle}>
                  Benutzerdaten
                </ThemedText>
              </View>
            </View>
            <View style={styles.form}>
              {globalError && (
                <View style={styles.globalError}>
                  <ThemedText style={styles.globalErrorText}>{globalError}</ThemedText>
                </View>
              )}
              <View>
                <TextInput
                  placeholder="Vorname"
                  value={firstName}
                  onChangeText={setFirstName}
                  style={[styles.input, firstNameError ? styles.inputError : null]}
                  placeholderTextColor={Colors.text}
                  autoCapitalize="words"
                  autoComplete="given-name"
                />
                {firstNameError && <ThemedText style={styles.errorText}>{firstNameError}</ThemedText>}
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
                {lastNameError && <ThemedText style={styles.errorText}>{lastNameError}</ThemedText>}
              </View>
              {!isAppleUser && (
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
                  {emailError && <ThemedText style={styles.errorText}>{emailError}</ThemedText>}
                </View>
              )}
              <View style={styles.buttonContainer}>
                <AnimatedHapticButton disabled={!hasDataChanged() || isFetching} onPress={handleSave} style={styles.button}>
                  {isFetching ? <Spinner fill={Colors.background} /> : <ThemedText style={styles.buttonText}>Speichern</ThemedText>}
                </AnimatedHapticButton>
              </View>
              <AnimatedHapticButton style={styles.buttonDeleteAccount} onPress={() => setShowDeleteAccountConfirmation(true)}>
                <View style={styles.buttonDeleteAccountTextContainer}>
                  <ThemedText style={styles.buttonDeleteAccountText}>Konto löschen</ThemedText>
                </View>
              </AnimatedHapticButton>
            </View>
            <View style={{ flex: 1 }} />
          </ScrollView>
        </TouchableWithoutFeedback>
      </KeyboardAvoidingView>
      <Modal
        animationType="slide"
        transparent={false}
        visible={showDeleteAccountConfirmation}
        onRequestClose={() => setShowDeleteAccountConfirmation(false)}
      >
        <SafeAreaProvider>
          <SafeAreaView style={styles.modalContainer}>
            <View style={styles.modalButtonContainer}>
              <AnimatedHapticButton onPress={() => setShowDeleteAccountConfirmation(false)} style={styles.modalButton} useHaptics={false}>
                <Ionicons name="close" size={24} color={Colors.text} />
              </AnimatedHapticButton>
            </View>
            <View style={styles.modalContent}>
              <ThemedText type="title" style={{ fontSize: 34 }}>
                Konto löschen
              </ThemedText>
              <ThemedText style={{ marginTop: 16 }}>
                Bist du sicher, dass du dein Konto löschen möchtest? Dies kann nicht rückgängig gemacht werden.
              </ThemedText>
              <View style={styles.modalActionButtonContainer}>
                <AnimatedHapticButton
                  onPress={async () => {
                    setIsFetching(true);
                    setGlobalError("");
                    const { success, error } = await userDeleteAccount();
                    if (success) {
                      await logoutUser();
                      router.push("/");
                    } else {
                      setGlobalError(error.message || "Fehler beim Löschen des Kontos");
                      setShowDeleteAccountConfirmation(false);
                    }
                  }}
                  style={[styles.modalActionButton, { backgroundColor: Colors.error }]}
                  disabled={isFetching}
                >
                  {isFetching ? <Spinner fill={Colors.background} /> : <ThemedText style={styles.buttonText}>Ja, Konto löschen</ThemedText>}
                </AnimatedHapticButton>
                <AnimatedHapticButton
                  onPress={() => {
                    setShowDeleteAccountConfirmation(false);
                  }}
                  style={[styles.modalActionButton]}
                >
                  <ThemedText style={styles.buttonText}>Nein, abbrechen</ThemedText>
                </AnimatedHapticButton>
              </View>
            </View>
          </SafeAreaView>
        </SafeAreaProvider>
      </Modal>
    </React.Fragment>
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
  },
  header: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    paddingVertical: 16,
  },
  headerButtons: {
    flex: 1,
    flexDirection: "column",
    alignItems: "flex-start",
  },
  headerTitle: {
    fontSize: 32,
    marginTop: 24,
  },
  form: {
    gap: 12,
    width: "100%",
    marginTop: 24,
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
    height: 45,
    borderColor: Colors.text,
    borderWidth: 2,
    borderRadius: 15,
    justifyContent: "center",
    alignItems: "center",
    paddingHorizontal: 10,
  },
  buttonText: {
    color: Colors.background,
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
  buttonDeleteAccount: {
    height: 60,
  },
  buttonDeleteAccountTextContainer: {
    flexDirection: "row",
    alignItems: "center",
    gap: 10,
  },
  buttonDeleteAccountText: {
    color: Colors.error,
    fontSize: 16,
  },
  modalContainer: {
    flex: 1,
    backgroundColor: Colors.background,
  },
  modalButtonContainer: {
    flexDirection: "column",
    alignItems: "flex-end",
    paddingHorizontal: 16,
  },
  modalButton: {
    height: 48,
    width: 48,
  },
  modalContent: {
    flex: 1,
    alignItems: "center",
    marginTop: 24,
    paddingHorizontal: 16,
  },
  modalActionButtonContainer: {
    flexDirection: "column",
    width: "100%",
    gap: 12,
    marginTop: 24,
  },
  modalActionButton: {
    height: 48,
    backgroundColor: Colors.text,
    borderRadius: 15,
    justifyContent: "center",
    alignItems: "center",
  },
});
