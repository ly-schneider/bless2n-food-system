import React, { useEffect, useState, useRef } from "react";
import {
  View,
  StyleSheet,
  TextInput,
  Animated,
  Text,
  KeyboardAvoidingView,
  Platform,
  TouchableWithoutFeedback,
  Keyboard,
} from "react-native";
import * as AppleAuthentication from "expo-apple-authentication";
import { router } from "expo-router";
import { ThemedText } from "@/components/ThemedText";
import { AnimatedHapticButton } from "@/components/AnimatedHapticButton";
import { Colors } from "@/constants/Colors";
import { Spinner } from "@/components/Spinner";
import { useAuth } from "@/hooks/useAuth";
import { authLoginWithEmail, authLoginWithApple } from "@/api/Api";
import { useRegistration } from "@/hooks/useRegistration";
import { needsTermsUpdate, saveTermsAcceptance } from "@/utils/TermsManager";
import * as SecureStore from "expo-secure-store";
import AsyncStorage from "@react-native-async-storage/async-storage";
import { TERMS_ACCEPTANCE_KEY } from "@/utils/TermsManager";

const AnimatedThemedText = Animated.createAnimatedComponent(Text);

export default function RegisterStart() {
  const { saveUser } = useAuth();
  const { updateData } = useRegistration();
  const [email, setEmail] = useState("");
  const [emailError, setEmailError] = useState("");
  const [appleError, setAppleError] = useState("");
  const [isFetching, setIsFetching] = useState(false);

  useEffect(() => {
    updateData({ email: "", isRegistration: false });
  }, []);

  const handleEmailContinue = async () => {
    if (!email) return;

    setIsFetching(true);
    const { success, error, data } = await authLoginWithEmail(email);
    setIsFetching(false);

    if (success) {
      const termsData = {
        version: data.termsVersion,
        acceptedAt: data.termsAcceptedAt,
        trackingEnabled: data.trackingEnabled,
      };
      await AsyncStorage.setItem(
        TERMS_ACCEPTANCE_KEY,
        JSON.stringify(termsData)
      );
      await SecureStore.setItemAsync(
        TERMS_ACCEPTANCE_KEY,
        JSON.stringify(termsData),
        {
          keychainAccessible: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
        }
      );

      if (data.requiresTestingAccountPassword) {
        updateData({ userId: data._id });
        router.push("/auth/register/testing-account" as any);
        return;
      }

      updateData({ email });
      router.push("/auth/register/verify");
    } else if (error.status === 400) {
      updateData({ email, isRegistration: true });
      router.push("/auth/register/user-info");
    } else {
      setEmailError("Fehler bei der Anmeldung oder Registrierung");
    }
  };

  const handleAppleContinue = async () => {
    const appleAuthRequestResponse = await AppleAuthentication.signInAsync({
      requestedScopes: [
        AppleAuthentication.AppleAuthenticationScope.FULL_NAME,
        AppleAuthentication.AppleAuthenticationScope.EMAIL,
      ],
    });

    if (appleAuthRequestResponse) {
      const { success, data, error } = await authLoginWithApple(
        appleAuthRequestResponse.user,
        appleAuthRequestResponse.email,
        false,
        appleAuthRequestResponse.fullName?.givenName,
        appleAuthRequestResponse.fullName?.familyName
      );

      if (success) {
        if (data.requiresTestingAccountPassword) {
          updateData({ userId: data._id });
          router.push("/auth/register/testing-account");
          return;
        }

        await saveUser(data._id);
        if (data.newRegistration === true) {
          updateData({ isRegistration: true });
          router.push("/auth/register/terms");
        } else if (await needsTermsUpdate(data.termsVersion)) {
          if (router.canDismiss()) {
            router.dismissAll();
          }
          router.push("/auth/register/terms");
        } else {
          await saveTermsAcceptance(data.trackingEnabled);
          router.push("/");
        }
      } else if (error.status === 400) {
        updateData({
          appleUserId: appleAuthRequestResponse.user,
          isRegistration: true,
        });
        router.push("/auth/register/user-info");
      } else {
        setAppleError("Fehler bei der Apple Authentifizierung");
      }
    }
  };

  const features = [
    {
      title: "Matches Planen",
      subtitle: "Erstelle und verwalte Spiele mühelos - alles an einem Ort.",
    },
    {
      title: "Spielerinfo",
      subtitle:
        "Übersichtliche Verwaltung deiner Verfügbarkeit und Match-Details.",
    },
    {
      title: "Teams",
      subtitle: "Behalte immer den Überblick über dein Team.",
    },
  ];

  const [featureIndex, setFeatureIndex] = useState(0);
  const fadeAnim = useRef(new Animated.Value(1)).current;

  useEffect(() => {
    let isMounted = true;
    const cycleFeatures = () => {
      if (!isMounted) return;
      Animated.timing(fadeAnim, {
        toValue: 0,
        duration: 500,
        useNativeDriver: true,
      }).start(() => {
        if (!isMounted) return;
        setFeatureIndex((prev) => (prev + 1) % features.length);
        Animated.timing(fadeAnim, {
          toValue: 1,
          duration: 500,
          useNativeDriver: true,
        }).start(() => {
          if (!isMounted) return;
          setTimeout(cycleFeatures, 3000);
        });
      });
    };
    const timeout = setTimeout(cycleFeatures, 3000);
    return () => {
      isMounted = false;
      clearTimeout(timeout);
    };
  }, [fadeAnim, features.length]);

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === "ios" ? "padding" : "height"}
      keyboardVerticalOffset={Platform.OS === "ios" ? 100 : 0}
    >
      <TouchableWithoutFeedback onPress={Keyboard.dismiss}>
        <View style={styles.inner}>
          <View>
            <ThemedText style={{ textAlign: "center" }} type="subtitle">
              Willkommen!
            </ThemedText>
          </View>
          <View style={styles.features}>
            <AnimatedThemedText style={[styles.title, { opacity: fadeAnim }]}>
              {features[featureIndex].title}
            </AnimatedThemedText>
            <AnimatedThemedText style={[styles.text, { opacity: fadeAnim }]}>
              {features[featureIndex].subtitle}
            </AnimatedThemedText>
          </View>
          <View style={styles.login}>
            {Platform.OS === "ios" && (
              <>
                <View style={styles.oauth}>
                  {appleError && (
                    <View style={styles.appleError}>
                      <ThemedText style={styles.appleErrorText}>
                        {appleError}
                      </ThemedText>
                    </View>
                  )}
                  <AppleAuthentication.AppleAuthenticationButton
                    buttonType={
                      AppleAuthentication.AppleAuthenticationButtonType.CONTINUE
                    }
                    buttonStyle={
                      AppleAuthentication.AppleAuthenticationButtonStyle.BLACK
                    }
                    cornerRadius={15}
                    style={styles.apple}
                    onPress={handleAppleContinue}
                  />
                </View>
                <View
                  style={{ borderBottomColor: "#B0B0B0", borderBottomWidth: 1 }}
                />
              </>
            )}
            <View style={styles.start}>
              <ThemedText>Anmelden oder Registrieren</ThemedText>
              <View>
                <TextInput
                  placeholder="E-Mail-Adresse"
                  value={email}
                  onChangeText={setEmail}
                  style={[
                    styles.textInput,
                    emailError ? styles.textInputError : null,
                  ]}
                  placeholderTextColor="#000"
                  keyboardType="email-address"
                  autoCapitalize="none"
                  autoCorrect={false}
                />
                {emailError && (
                  <ThemedText style={styles.errorText}>{emailError}</ThemedText>
                )}
              </View>
              <AnimatedHapticButton
                disabled={!email || isFetching}
                onPress={handleEmailContinue}
                style={styles.button}
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
    marginTop: 72,
  },
  features: {
    marginTop: 32,
  },
  title: {
    fontSize: 42,
    fontFamily: "MozaicHUMVariable",
    textAlign: "center",
  },
  text: {
    fontSize: 16,
    lineHeight: 24,
    fontFamily: "Bilo",
    textAlign: "center",
    marginTop: 8,
    paddingHorizontal: 24,
  },
  login: {
    flexDirection: "column",
    gap: 24,
    marginTop: Platform.OS === "ios" ? 64 : 152,
  },
  oauth: {
    flexDirection: "column",
    gap: 12,
  },
  apple: {
    width: "100%",
    height: 48,
  },
  start: {
    flexDirection: "column",
    gap: 12,
  },
  textInput: {
    borderColor: "#000",
    borderWidth: 2,
    borderRadius: 15,
    paddingVertical: 12,
    paddingHorizontal: 16,
    fontFamily: "Bilo",
  },
  button: {
    width: "100%",
    height: 48,
    backgroundColor: "#000",
    borderRadius: 15,
    justifyContent: "center",
    alignItems: "center",
  },
  buttonText: {
    color: Colors.background,
  },
  textInputError: {
    borderColor: Colors.error,
  },
  errorText: {
    color: Colors.error,
    fontSize: 14,
    marginTop: 2,
  },
  appleError: {
    backgroundColor: Colors.error,
    padding: 10,
    borderRadius: 15,
  },
  appleErrorText: {
    color: Colors.background,
    textAlign: "center",
  },
});
