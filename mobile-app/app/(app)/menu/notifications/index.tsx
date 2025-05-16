import {
  RefreshControl,
  ScrollView,
  StyleSheet,
  Switch,
  View,
  Linking,
  Platform,
  Alert,
} from "react-native";
import { AnimatedHapticButton } from "@/components/AnimatedHapticButton";
import { Ionicons } from "@expo/vector-icons";
import { Colors } from "@/constants/Colors";
import { router } from "expo-router";
import { ThemedText } from "@/components/ThemedText";
import { useCallback, useEffect, useState } from "react";
import { Spinner } from "@/components/Spinner";
import {
  notificationFetchSettings,
  notificationUpdateSettings,
  userDataUpdate,
} from "@/api/Api";
import messaging from "@react-native-firebase/messaging";
import { getApp } from "@react-native-firebase/app";

interface NotificationSettings {
  pushEnabled: boolean;
  emailEnabled: boolean;
}

export default function MenuNotifications() {
  const [notificationSettings, setNotificationSettings] =
    useState<NotificationSettings>({
      pushEnabled: true,
      emailEnabled: true,
    });
  const [isFetchingPush, setIsFetchingPush] = useState(false);
  const [isFetchingEmail, setIsFetchingEmail] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [devicePermissionEnabled, setDevicePermissionEnabled] = useState(true);

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await fetchNotificationSettings();
    await checkDevicePermissions();
    setRefreshing(false);
  }, []);

  useEffect(() => {
    fetchNotificationSettings();
    checkDevicePermissions();
  }, []);

  async function checkDevicePermissions() {
    try {
      const authStatus = await messaging().hasPermission();
      const enabled =
        authStatus === messaging.AuthorizationStatus.AUTHORIZED ||
        authStatus === messaging.AuthorizationStatus.PROVISIONAL;
      setDevicePermissionEnabled(enabled);
    } catch (error) {
      console.error("Error checking notification permissions:", error);
      setDevicePermissionEnabled(false);
    }
  }

  async function fetchNotificationSettings() {
    const { success, data, error } = await notificationFetchSettings();
    if (success) {
      setNotificationSettings(data);
    } else {
      console.error(error);
    }
  }

  const handleBack = () => {
    router.back();
  };

  const handleTogglePush = async () => {
    setIsFetchingPush(true);

    if (!notificationSettings.pushEnabled) {
      // If we're enabling, check device permissions first
      await handleEnableNotifications();
    } else {
      // If we're disabling, just update the setting
      const { success, data, error } = await notificationUpdateSettings(
        false,
        notificationSettings.emailEnabled
      );
      if (success) {
        setNotificationSettings(data);
      } else {
        console.error(error);
      }
    }

    setIsFetchingPush(false);
  };

  const handleToggleEmail = async () => {
    setIsFetchingEmail(true);
    const { success, data, error } = await notificationUpdateSettings(
      notificationSettings.pushEnabled,
      !notificationSettings.emailEnabled
    );
    if (success) {
      setNotificationSettings(data);
    } else {
      console.error(error);
    }
    setIsFetchingEmail(false);
  };

  const handleEnableNotifications = async () => {
    setIsFetchingPush(true);
    try {
      const authStatus = await getApp().messaging().requestPermission();
      const enabled =
        authStatus === messaging.AuthorizationStatus.AUTHORIZED ||
        authStatus === messaging.AuthorizationStatus.PROVISIONAL;

      setDevicePermissionEnabled(enabled);

      if (!enabled) {
        // When not enabled open system settings
        const message = Platform.select({
          ios: "Um Benachrichtigungen zu aktivieren, öffne die Einstellungen und erlaube Benachrichtigungen für TT-Plus.",
          android:
            "Um Benachrichtigungen zu aktivieren, öffne die App-Einstellungen und erlaube Benachrichtigungen für TT-Plus.",
          default:
            "Um Benachrichtigungen zu aktivieren, öffne die App-Einstellungen und erlaube Benachrichtigungen für TT-Plus.",
        });

        Alert.alert("Benachrichtigungen erlauben", message, [
          { text: "Abbrechen", style: "cancel" },
          {
            text: "Einstellungen öffnen",
            onPress: () => {
              if (Platform.OS === "ios") {
                Linking.openURL("app-settings:");
              } else {
                Linking.openSettings();
              }
            },
          },
        ]);

        setIsFetchingPush(false);
        return;
      }

      const fcmToken = await getApp().messaging().getToken();

      await userDataUpdate({ pushNotificationToken: fcmToken }, false);

      const { success, data, error } = await notificationUpdateSettings(
        true,
        notificationSettings.emailEnabled
      );
      if (success) {
        setNotificationSettings(data);
      } else {
        console.error(error);
      }
    } catch (error) {
      console.error("Error requesting notification permissions:", error);
    }
    setIsFetchingPush(false);
  };

  // Show Erlauben button if backend says enabled but device says disabled,
  // or if backend says disabled
  const shouldShowAllowButton =
    !devicePermissionEnabled || !notificationSettings.pushEnabled;

  return (
    <ScrollView
      style={styles.container}
      contentContainerStyle={{ flexGrow: 1, minHeight: "100%" }}
      alwaysBounceVertical={true}
      refreshControl={
        <RefreshControl
          refreshing={refreshing}
          onRefresh={onRefresh}
          tintColor={Colors.text}
          colors={[Colors.text]}
        />
      }
    >
      <View style={styles.header}>
        <View style={styles.headerButtons}>
          <AnimatedHapticButton
            style={styles.buttonOutline}
            onPress={handleBack}
          >
            <Ionicons name="arrow-back" size={20} color={Colors.text} />
          </AnimatedHapticButton>
          <ThemedText type="title" style={styles.headerTitle}>
            Benachrichtigungen
          </ThemedText>
        </View>
      </View>
      <View style={styles.menuContainer}>
        <View style={styles.menuContainerInner}>
          <View style={styles.menuItem}>
            <ThemedText style={styles.menuItemText}>
              Push-Benachrichtigungen
            </ThemedText>
            {isFetchingPush ? (
              <Spinner fill={Colors.text} />
            ) : shouldShowAllowButton ? (
              <AnimatedHapticButton
                style={styles.allowButton}
                onPress={handleEnableNotifications}
                disabled={isFetchingPush}
              >
                <ThemedText style={styles.allowButtonText}>Erlauben</ThemedText>
              </AnimatedHapticButton>
            ) : (
              <Switch
                trackColor={{ false: undefined, true: Colors.text }}
                onValueChange={handleTogglePush}
                value={
                  notificationSettings.pushEnabled && devicePermissionEnabled
                }
              />
            )}
          </View>
          <View style={styles.menuItem}>
            <ThemedText style={styles.menuItemText}>
              E-Mail-Benachrichtigungen
            </ThemedText>
            {isFetchingEmail ? (
              <Spinner fill={Colors.text} />
            ) : (
              <Switch
                trackColor={{ false: undefined, true: Colors.text }}
                onValueChange={handleToggleEmail}
                value={notificationSettings.emailEnabled}
              />
            )}
          </View>
        </View>
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    marginTop: 0,
    paddingHorizontal: 24,
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
  buttonOutline: {
    height: 45,
    borderColor: Colors.text,
    borderWidth: 2,
    borderRadius: 15,
    justifyContent: "center",
    alignItems: "center",
    paddingHorizontal: 10,
  },
  menuContainer: {
    flex: 1,
    marginTop: 24,
    gap: 6,
  },
  menuContainerInner: {
    flex: 1,
    gap: 12,
  },
  menuItem: {
    height: 60,
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    borderRadius: 15,
    backgroundColor: Colors.background,
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderWidth: 2,
    borderColor: Colors.text,
  },
  menuItemText: {
    color: Colors.text,
    fontSize: 18,
  },
  allowButton: {
    borderRadius: 15,
    backgroundColor: Colors.text,
    justifyContent: "center",
    alignItems: "center",
    paddingHorizontal: 12,
  },
  allowButtonText: {
    color: Colors.background,
    fontSize: 14,
  },
});
