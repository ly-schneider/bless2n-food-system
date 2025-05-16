import { ThemedText } from "@/components/ThemedText";
import { Colors } from "@/constants/Colors";
import { View, StyleSheet } from "react-native";
import { AnimatedHapticButton } from "@/components/AnimatedHapticButton";
import messaging from "@react-native-firebase/messaging";
import { getApp } from "@react-native-firebase/app";
import { userDataUpdate } from "@/api/Api";
import { router } from "expo-router";

export default function RegisterNotifications() {
  const handleConfirm = async () => {
    try {
      const authStatus = await getApp().messaging().requestPermission();
      const enabled =
        authStatus === messaging.AuthorizationStatus.AUTHORIZED ||
        authStatus === messaging.AuthorizationStatus.PROVISIONAL;

      if (!enabled) {
        return;
      }

      const fcmToken = await getApp().messaging().getToken();

      const { success } = await userDataUpdate(
        { pushNotificationToken: fcmToken },
        false
      );

      if (!success) {
        console.error("Failed to update push notification token");
        return;
      }
    } catch (error) {
      console.error("Error requesting notification permissions:", error);
    }

    router.push("/auth/register/success");
  };

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <ThemedText type="title" style={{ textAlign: "center", fontSize: 36 }}>
          Erlaube Benachrichtigungen
        </ThemedText>
      </View>

      <View style={styles.content}>
        <AnimatedHapticButton style={styles.button} onPress={handleConfirm}>
          <ThemedText style={{ color: Colors.background }}>Erlauben</ThemedText>
        </AnimatedHapticButton>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    paddingHorizontal: 24,
  },
  header: {
    marginTop: 128,
    marginBottom: 32,
  },
  content: {
    flex: 1,
    justifyContent: "center",
    gap: 24,
  },
  button: {
    width: "100%",
    height: 48,
    backgroundColor: Colors.text,
    borderRadius: 15,
    justifyContent: "center",
    alignItems: "center",
  },
});
