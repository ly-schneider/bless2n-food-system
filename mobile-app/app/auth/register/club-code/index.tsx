import { ThemedText } from "@/components/ThemedText";
import { Colors } from "@/constants/Colors";
import { router, useGlobalSearchParams } from "expo-router";
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
import { Ionicons } from "@expo/vector-icons";
import { AnimatedHapticButton } from "@/components/AnimatedHapticButton";
import { Spinner } from "@/components/Spinner";
import { clubJoin, clubFetch, authLogout } from "@/api/Api";
import { useAuth } from "@/hooks/useAuth";

export default function RegisterClubCode() {
  const { from } = useGlobalSearchParams<{ from?: string }>();

  const { logoutUser } = useAuth();

  const [clubCode, setClubCode] = useState("");
  const [clubData, setClubData] = useState<{
    _id: string;
    name: string;
  } | null>(null);
  const [clubCodeError, setClubCodeError] = useState("");
  const [isFetching, setIsFetching] = useState(false);
  const [isFetchingJoin, setIsFetchingJoin] = useState(false);

  const handleFetchClub = async () => {
    if (clubCode.length !== 6) return;

    setIsFetching(true);
    const { data, success } = await clubFetch(clubCode);
    setIsFetching(false);

    if (success) {
      setClubData(data);
    } else {
      setClubCodeError("Club nicht gefunden");
    }
  };

  const handleBack = () => {
    router.back();
  };

  const closeClubData = () => {
    setClubData(null);
    setClubCode("");
    setClubCodeError("");
  };

  const handleClubJoin = async () => {
    if (!clubData) return;

    setIsFetchingJoin(true);
    const { success } = await clubJoin(clubCode);
    setIsFetchingJoin(false);

    if (success) {
      router.push("/auth/register/notifications");
    } else {
      setClubCodeError("Fehler beim Beitreten des Clubs");
    }
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
              Tritt deinem Club bei!
            </ThemedText>
          </View>

          {clubData ? (
            <View style={[styles.form, { marginTop: 202 }]}>
              {clubCodeError && (
                <View style={styles.clubCodeError}>
                  <ThemedText style={styles.clubCodeErrorText}>
                    {clubCodeError}
                  </ThemedText>
                </View>
              )}
              <View style={styles.clubDataContainer}>
                <ThemedText style={styles.clubDataText}>
                  {clubData.name}
                </ThemedText>
                <AnimatedHapticButton
                  style={styles.clubDataClose}
                  onPress={closeClubData}
                >
                  <Ionicons
                    name="close-outline"
                    size={22}
                    color={Colors.text}
                  />
                </AnimatedHapticButton>
              </View>
              <View style={styles.buttonContainer}>
                <AnimatedHapticButton
                  style={styles.buttonOutline}
                  onPress={handleBack}
                >
                  <Ionicons name="arrow-back" size={20} color={Colors.text} />
                </AnimatedHapticButton>
                <AnimatedHapticButton
                  disabled={clubData === undefined || isFetchingJoin}
                  onPress={handleClubJoin}
                  style={styles.button}
                >
                  {isFetchingJoin ? (
                    <Spinner fill={Colors.background} />
                  ) : (
                    <ThemedText style={{ color: Colors.background }}>
                      Beitreten
                    </ThemedText>
                  )}
                </AnimatedHapticButton>
              </View>
            </View>
          ) : (
            <View style={styles.form}>
              <ThemedText>
                Gib den Club-Code ein, den du von deinem Verein erhalten hast.
              </ThemedText>
              <View>
                <TextInput
                  placeholder="Club-Code eingeben"
                  value={clubCode}
                  onChangeText={(text) => {
                    setClubCode(text);
                    setClubCodeError("");
                  }}
                  style={[
                    styles.input,
                    clubCodeError ? styles.inputError : null,
                  ]}
                  placeholderTextColor={Colors.text}
                  maxLength={6}
                  autoComplete="off"
                  autoCorrect={false}
                  autoCapitalize="none"
                />
                {clubCodeError && (
                  <ThemedText style={styles.errorText}>
                    {clubCodeError}
                  </ThemedText>
                )}
              </View>
              <View style={styles.buttonContainer}>
                {from !== "app" && (
                  <AnimatedHapticButton
                    style={styles.buttonOutline}
                    onPress={handleBack}
                  >
                    <Ionicons name="arrow-back" size={20} color={Colors.text} />
                  </AnimatedHapticButton>
                )}
                <AnimatedHapticButton
                  disabled={clubCode.length !== 6 || isFetching}
                  onPress={handleFetchClub}
                  style={styles.button}
                >
                  {isFetching ? (
                    <Spinner fill={Colors.background} />
                  ) : (
                    <ThemedText style={{ color: Colors.background }}>
                      Weiter
                    </ThemedText>
                  )}
                </AnimatedHapticButton>
              </View>
              {from === "app" && (
                <AnimatedHapticButton
                  style={styles.menuLogoutItem}
                  onPress={async () => {
                    await logoutUser();
                    const { success, error } = await authLogout();
                    if (success) {
                      router.push("/auth/register/start");
                    } else if (error) {
                      router.push("/auth/register/start");
                    }
                  }}
                >
                  <View style={styles.menuTextContainer}>
                    <ThemedText style={styles.menuLogoutItemText}>
                      Logout
                    </ThemedText>
                    <Ionicons
                      name="log-out-outline"
                      size={20}
                      color={Colors.error}
                    />
                  </View>
                </AnimatedHapticButton>
              )}
            </View>
          )}
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
    marginTop: 2,
  },
  clubDataContainer: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    paddingVertical: 12,
    paddingHorizontal: 12,
    backgroundColor: Colors.text,
    borderRadius: 15,
    height: 52,
  },
  clubDataText: {
    color: Colors.background,
  },
  clubDataClose: {
    padding: 3,
    backgroundColor: Colors.background,
    borderRadius: 8,
  },
  clubCodeError: {
    padding: 10,
    backgroundColor: Colors.error,
    borderRadius: 15,
  },
  clubCodeErrorText: {
    color: Colors.background,
    textAlign: "center",
  },
  menuLogoutItem: {
    height: 60,
  },
  menuLogoutItemText: {
    color: Colors.error,
    fontSize: 16,
  },
  menuTextContainer: {
    flexDirection: "row",
    alignItems: "center",
    gap: 10,
  },
});
