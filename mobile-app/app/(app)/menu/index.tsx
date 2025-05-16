import { StyleSheet, View } from "react-native";
import { AnimatedHapticButton } from "@/components/AnimatedHapticButton";
import { Ionicons } from "@expo/vector-icons";
import { Colors } from "@/constants/Colors";
import { router } from "expo-router";
import { ThemedText } from "@/components/ThemedText";
import { useAuth } from "@/hooks/useAuth";
import { authLogout } from "@/api/Api";

export default function MenuHome() {
  const { logoutUser } = useAuth();

  const handleBack = () => {
    router.back();
  };

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <View style={styles.headerButtons}>
          <AnimatedHapticButton
            style={styles.buttonOutline}
            onPress={handleBack}
          >
            <Ionicons name="arrow-back" size={20} color={Colors.text} />
          </AnimatedHapticButton>
          <ThemedText type="title" style={styles.headerTitle}>
            Menu
          </ThemedText>
        </View>
      </View>
      <View style={styles.menuContainer}>
        <AnimatedHapticButton
          useDefaultStyles={false}
          pressableStyle={styles.menuItem}
          onPress={() => router.push("/menu/absences")}
        >
          <View style={styles.menuTextContainer}>
            <Ionicons
              name="close-circle-outline"
              size={22}
              color={Colors.background}
            />
            <ThemedText style={styles.menuItemText}>
              Abwesenheitszeiten
            </ThemedText>
          </View>
          <AnimatedHapticButton
            style={styles.menuItemButton}
            onPress={() => router.push("/menu/absences")}
          >
            <Ionicons name="arrow-forward" size={20} color={Colors.text} />
          </AnimatedHapticButton>
        </AnimatedHapticButton>
        <AnimatedHapticButton
          useDefaultStyles={false}
          pressableStyle={styles.menuItem}
          onPress={() => router.push("/menu/user-settings")}
        >
          <View style={styles.menuTextContainer}>
            <Ionicons
              name="person-outline"
              size={20}
              color={Colors.background}
            />
            <ThemedText style={styles.menuItemText}>Benutzereinstellungen</ThemedText>
          </View>
          <AnimatedHapticButton
            style={styles.menuItemButton}
            onPress={() => router.push("/menu/user-settings")}
          >
            <Ionicons name="arrow-forward" size={20} color={Colors.text} />
          </AnimatedHapticButton>
        </AnimatedHapticButton>
        <AnimatedHapticButton
          useDefaultStyles={false}
          pressableStyle={styles.menuItem}
          onPress={() => router.push("/menu/notifications")}
        >
          <View style={styles.menuTextContainer}>
            <Ionicons
              name="notifications-outline"
              size={20}
              color={Colors.background}
            />
            <ThemedText style={styles.menuItemText}>
              Benachrichtigungen
            </ThemedText>
          </View>
          <AnimatedHapticButton
            style={styles.menuItemButton}
            onPress={() => router.push("/menu/notifications")}
          >
            <Ionicons name="arrow-forward" size={20} color={Colors.text} />
          </AnimatedHapticButton>
        </AnimatedHapticButton>
        <AnimatedHapticButton
          style={styles.menuLogoutItem}
          onPress={async () => {
            await logoutUser();
            await authLogout();
            if (router.canDismiss()) router.dismissAll();
            router.push("/auth/register/start");
          }}
        >
          <View style={styles.menuTextContainer}>
            <ThemedText style={styles.menuLogoutItemText}>Logout</ThemedText>
            <Ionicons name="log-out-outline" size={22} color={Colors.error} />
          </View>
        </AnimatedHapticButton>
      </View>
    </View>
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
    gap: 10,
  },
  menuItem: {
    height: 60,
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    borderRadius: 15,
    backgroundColor: Colors.text,
    paddingHorizontal: 16,
    paddingVertical: 8,
  },
  menuTextContainer: {
    flexDirection: "row",
    alignItems: "center",
    gap: 10,
  },
  menuItemText: {
    color: Colors.background,
    fontSize: 20,
  },
  menuItemButton: {
    height: 38,
    width: 38,
    borderRadius: 12,
    justifyContent: "center",
    alignItems: "center",
    backgroundColor: Colors.background,
  },
  menuLogoutItem: {
    height: 60,
  },
  menuLogoutItemText: {
    color: Colors.error,
    fontSize: 20,
  },
});
