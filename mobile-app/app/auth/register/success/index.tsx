import { ThemedText } from "@/components/ThemedText";
import { Colors } from "@/constants/Colors";
import { router } from "expo-router";
import { View, StyleSheet } from "react-native";
import ConfettiCannon from "react-native-confetti-cannon";
import { AnimatedHapticButton } from "@/components/AnimatedHapticButton";

export default function RegisterSuccess() {
  const handleComplete = async () => {
    router.dismissAll();
    router.replace("/");
  };

  return (
    <View style={styles.container}>
      <ConfettiCannon
        count={200}
        origin={{ x: 0, y: 0 }}
        fadeOut={true}
        fallSpeed={2500}
      />

      <View style={styles.header}>
        <ThemedText type="title" style={{ textAlign: "center" }}>
          Registrierung erfolgreich!
        </ThemedText>
      </View>

      <View style={styles.content}>
        <AnimatedHapticButton style={styles.button} onPress={handleComplete}>
          <ThemedText style={{ color: Colors.background }}>
            Zur Übersicht
          </ThemedText>
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
    marginBottom: 28,
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
