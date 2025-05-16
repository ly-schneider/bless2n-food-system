import { FC } from "react";
import { StyleSheet, View } from "react-native";
import { ThemedText } from "./ThemedText";
import { AnimatedHapticButton } from "./AnimatedHapticButton";
import { Colors } from "@/constants/Colors";
import { Ionicons } from "@expo/vector-icons";
import { router } from "expo-router";
import { Match } from "@/constants/Match";
import { User } from "@/constants/User";

interface MatchesContainerProps {
  isAdminInAnyClub: boolean;
  matches: Match[];
  selectedDate: Date;
}

export const MatchesContainer: FC<MatchesContainerProps> = ({
  isAdminInAnyClub,
  matches,
  selectedDate,
}) => {
  const filteredMatches = matches.filter((match) => {
    const matchDate = new Date(match.date || Date.now());
    return matchDate.toDateString() === selectedDate.toDateString();
  });

  filteredMatches.sort(
    (a, b) =>
      new Date(a.date || Date.now()).getTime() -
      new Date(b.date || Date.now()).getTime()
  );

  return (
    <View style={styles.container}>
      <View
        style={[styles.header, { paddingVertical: isAdminInAnyClub ? 0 : 8 }]}
      >
        <ThemedText type="title" style={styles.title}>
          Matches
        </ThemedText>
        {isAdminInAnyClub && (
          <AnimatedHapticButton
            style={styles.button}
            onPress={() => router.push("/match/create")}
          >
            <ThemedText style={{ color: Colors.background, fontSize: 14 }}>
              Erstellen
            </ThemedText>
          </AnimatedHapticButton>
        )}
      </View>
      {filteredMatches.length > 0 ? (
        <View style={styles.matchesList}>
          {filteredMatches.map((match) => (
            <AnimatedHapticButton
              useDefaultStyles={false}
              pressableStyle={styles.matchItem}
              onPress={() =>
                router.push({
                  pathname: "/match/detail/[id]",
                  params: { id: match._id },
                })
              }
              key={match._id}
            >
              <View style={styles.matchHeader}>
                <View style={styles.matchHeaderTextContainer}>
                  <ThemedText type="title" style={styles.matchHeaderText}>
                    {match.club.name} -{" "}
                  </ThemedText>
                  <ThemedText type="title" style={styles.matchHeaderText}>
                    {match.enemyClub.name}
                  </ThemedText>
                </View>
                <View style={styles.matchIsHomeGameContainer}>
                  <Ionicons
                    name="pin-outline"
                    size={16}
                    color={Colors.background}
                  />
                  <ThemedText style={styles.matchIsHomeGameText}>
                    {match.isHomeGame ? "Heim" : "Auswärts"}
                  </ThemedText>
                </View>
              </View>
              <ThemedText style={styles.matchDate}>
                {new Intl.DateTimeFormat("de-DE", {
                  hour: "2-digit",
                  minute: "2-digit",
                }).format(new Date(match.date || Date.now()))}{" "}
                Uhr
              </ThemedText>
              <View style={styles.footerContainer}>
                <View style={styles.participantsContainer}>
                  {match.participants === undefined ||
                  match.participants.length === 0 ? (
                    <ThemedText>Keine Spieler</ThemedText>
                  ) : (
                    (match.participants as User[]).map((participant) => (
                      <View
                        key={participant?._id}
                        style={styles.participantsItem}
                      >
                        <ThemedText style={styles.participantsItemText}>
                          {participant?.firstName[0]}
                          {participant?.lastName[0]}
                        </ThemedText>
                      </View>
                    ))
                  )}
                </View>
                <AnimatedHapticButton
                  style={styles.matchItemButton}
                  onPress={() =>
                    router.push({
                      pathname: "/match/detail/[id]",
                      params: { id: match._id },
                    })
                  }
                >
                  <Ionicons
                    name="arrow-forward"
                    size={18}
                    color={Colors.background}
                  />
                </AnimatedHapticButton>
              </View>
            </AnimatedHapticButton>
          ))}
        </View>
      ) : (
        <ThemedText style={styles.noMatchesText}>Keine Matches</ThemedText>
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    marginTop: 32,
    paddingHorizontal: 24,
    flexDirection: "column",
    gap: 16,
  },
  header: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    gap: 16,
  },
  title: {
    fontSize: 28,
  },
  button: {
    height: 38,
    backgroundColor: Colors.text,
    borderRadius: 12,
    justifyContent: "center",
    alignItems: "center",
    paddingHorizontal: 16,
  },
  matchesList: {
    flex: 1,
    flexDirection: "column",
    gap: 16,
  },
  matchItem: {
    padding: 16,
    backgroundColor: Colors.text,
    borderRadius: 15,
  },
  matchHeader: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "flex-start",
    gap: 16,
  },
  matchHeaderTextContainer: {
    flexDirection: "row",
    gap: 0,
    flexWrap: "wrap",
    width: "60%",
    alignSelf: "center",
  },
  matchHeaderText: {
    color: Colors.background,
    fontSize: 18,
  },
  matchIsHomeGameContainer: {
    flexDirection: "row",
    alignItems: "center",
    gap: 2,
  },
  matchIsHomeGameText: {
    color: Colors.background,
  },
  matchDate: {
    color: Colors.background,
    marginTop: -2,
    fontSize: 14,
  },
  footerContainer: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    marginTop: 12,
  },
  participantsContainer: {
    flexDirection: "row",
    gap: 8,
  },
  participantsItem: {
    height: 28,
    width: 28,
    borderRadius: 9,
    justifyContent: "center",
    alignItems: "center",
    backgroundColor: Colors.background,
  },
  participantsItemText: {
    color: Colors.text,
    fontSize: 14,
  },
  matchItemButton: {
    height: 32,
    width: 32,
    borderRadius: 12,
    justifyContent: "center",
    alignItems: "center",
    borderWidth: 2,
    borderColor: Colors.background,
  },
  noMatchesText: {
    color: Colors.muted,
    textAlign: "center",
    marginTop: 8,
  },
});
