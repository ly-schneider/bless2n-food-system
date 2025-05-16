import "expo-dev-client";
import { useFonts } from "expo-font";
import { Stack } from "expo-router";
import * as SplashScreen from "expo-splash-screen";
import { useEffect } from "react";
import "react-native-reanimated";
import { AppState, Platform, StyleSheet } from "react-native";
import { SafeAreaProvider, SafeAreaView } from "react-native-safe-area-context";
import { Colors } from "@/constants/Colors";
import { AuthProvider } from "@/hooks/useAuth";
import { userDataUpdate } from "@/api/Api";
import * as Notifications from "expo-notifications";
import { router } from "expo-router";
import AsyncStorage from "@react-native-async-storage/async-storage";
import crashlytics from "@react-native-firebase/crashlytics";
import messaging from "@react-native-firebase/messaging";

SplashScreen.preventAutoHideAsync();

interface RemoteMessage {
  data?: {
    matchLink?: string;
  };
}

export default function RootLayout() {
  const [loaded] = useFonts({
    MozaicHUMVariable: require("../assets/fonts/Mozaic-HUM-Variable-Regular.ttf"),
    Bilo: require("../assets/fonts/Bilo-Medium.ttf"),
  });

  const resetPushNotificationBadgeCount = async () => {
    const storedUser = await AsyncStorage.getItem("user");
    if (!storedUser) return;

    const { success } = await userDataUpdate(
      { pushNotificationBadgeCount: 0 },
      false
    );
    if (!success) {
      console.info("Failed to reset push notification badge count");
    }
  };

  useEffect(() => {
    if (loaded) {
      SplashScreen.hideAsync();
    }
  }, [loaded]);

  useEffect(() => {
    const initializeCrashlytics = async () => {
      await crashlytics().setCrashlyticsCollectionEnabled(true);
    };

    initializeCrashlytics();
  }, []);

  useEffect(() => {
    const subscription = AppState.addEventListener(
      "change",
      async (nextAppState) => {
        if (nextAppState === "active") {
          if (Platform.OS === "ios") {
            Notifications.setBadgeCountAsync(0);
          } else if (Platform.OS === "android") {
            Notifications.dismissAllNotificationsAsync();
          }

          await resetPushNotificationBadgeCount();
        }
      }
    );
    return () => {
      subscription.remove();
    };
  }, []);

  useEffect(() => {
    // Safe initialization of Firebase messaging with error handling
    const initializeMessaging = async () => {
      try {
        // Request permission first (required for iOS)
        if (Platform.OS === "ios") {
          const authStatus = await messaging().hasPermission();
          const enabled =
            authStatus === messaging.AuthorizationStatus.AUTHORIZED ||
            authStatus === messaging.AuthorizationStatus.PROVISIONAL;

          if (!enabled) {
            console.log("Notification permissions not granted");
            return;
          }
        }

        // Handle notification when app is in foreground
        const unsubscribeForeground = messaging().onMessage(
          async (remoteMessage: RemoteMessage) => {
            if (
              remoteMessage?.data?.matchLink &&
              typeof remoteMessage.data.matchLink === "string"
            ) {
              const matchId = remoteMessage.data.matchLink.split("/").pop();
              if (matchId) {
                router.push({
                  pathname: "/(app)/match/detail/[id]",
                  params: { id: matchId },
                });
              }
            }
          }
        );

        // Handle notification when app is opened from background
        const unsubscribeBackground = messaging().onNotificationOpenedApp(
          (remoteMessage: RemoteMessage) => {
            if (
              remoteMessage?.data?.matchLink &&
              typeof remoteMessage.data.matchLink === "string"
            ) {
              const matchId = remoteMessage.data.matchLink.split("/").pop();
              if (matchId) {
                router.push({
                  pathname: "/(app)/match/detail/[id]",
                  params: { id: matchId },
                });
              }
            }
          }
        );

        // Handle notification when app is opened from terminated state
        const initialNotification = await messaging().getInitialNotification();
        if (
          initialNotification?.data?.matchLink &&
          typeof initialNotification.data.matchLink === "string"
        ) {
          const matchId = initialNotification.data.matchLink.split("/").pop();
          if (matchId) {
            router.push({
              pathname: "/(app)/match/detail/[id]",
              params: { id: matchId },
            });
          }
        }

        return () => {
          unsubscribeForeground();
          unsubscribeBackground();
        };
      } catch (error) {
        // Log error but don't crash
        console.error("Firebase messaging initialization error:", error);
        if (crashlytics) {
          crashlytics().recordError(error as Error);
        }
        return () => {};
      }
    };

    const unsubscribe = initializeMessaging();
    return () => {
      if (unsubscribe && typeof unsubscribe.then === "function") {
        unsubscribe.then((cleanup) => cleanup && cleanup());
      }
    };
  }, []);

  return (
    <SafeAreaProvider>
      <SafeAreaView style={styles.safeArea}>
        {loaded ? (
          <AuthProvider>
            <Stack
              screenOptions={{
                headerShown: false,
                animation: "fade",
                contentStyle: { backgroundColor: Colors.background },
              }}
            />
          </AuthProvider>
        ) : null}
      </SafeAreaView>
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    paddingTop: 4,
    backgroundColor: Colors.background,
  },
  scrollView: {
    flex: 1,
  },
});
