import { Modal, SafeAreaView, StyleSheet, View } from "react-native";
import { ThemedText } from "./ThemedText";
import { Match } from "@/constants/Match";
import { User } from "@/constants/User";
import { Colors } from "@/constants/Colors";
import { Ionicons } from "@expo/vector-icons";
import { SafeAreaProvider } from "react-native-safe-area-context";
import { AnimatedHapticButton } from "./AnimatedHapticButton";
import { useEffect, useState } from "react";
import { Club } from "@/constants/Club";
import { format } from "date-fns";
import {
  clubFetchParticipants,
  matchAddParticipant,
  matchRemoveParticipant,
} from "@/api/Api";

export default function ParticipantsContainer({
  club,
  matchData,
  isAdmin,
}: {
  club: Club;
  matchData: Match;
  isAdmin: boolean;
}) {
  const [match, setMatch] = useState<Match>(matchData);
  const [players, setPlayers] = useState<User[]>([]);
  const [showPickerModal, setShowPickerModal] = useState<boolean>(false);

  useEffect(() => {
    setMatch(matchData);
  }, [matchData]);

  useEffect(() => {
    fetchPlayers();
  }, []);

  async function fetchPlayers() {
    const { success, data } = await clubFetchParticipants(club.code);
    if (success) {
      setPlayers(data);
    }
  }

  async function addPlayer(player: User) {
    const { success, error } = await matchAddParticipant(match._id, player._id);
    if (success) {
      setShowPickerModal(false);
      setMatch({
        ...match,
        participants: [...(match.participants as User[]), player],
      });
    }
  }

  async function removePlayer(player: User) {
    const { success, error } = await matchRemoveParticipant(
      match._id,
      player._id
    );
    if (success) {
      setMatch({
        ...match,
        participants: (match.participants as User[]).filter(
          (p) => p._id !== player._id
        ),
      });
    }
  }

  const isPlayerAvailable = (player: User) => {
    if (match.date === undefined) return true;

    const weekday = format(
      match.date as Date,
      "EEEE"
    ).toLowerCase() as keyof typeof player.absenceDays;
    if (player.absenceDays?.[weekday]) return false;

    if (player.absencePeriods) {
      return !player.absencePeriods.some((period) => {
        const startDay = new Date(
          new Date(period.startDate).setHours(0, 0, 0, 0)
        );
        const endDay = new Date(
          new Date(period.endDate).setHours(23, 59, 59, 999)
        );
        return (
          (match.date as Date) >= startDay && (match.date as Date) <= endDay
        );
      });
    }

    return true;
  };

  const availablePlayers = players.filter(
    (player) =>
      !(match.participants as User[])?.map((p) => p._id).includes(player._id)
  );

  return (
    <View>
      <View style={styles.headerContainer}>
        <ThemedText type="title" style={styles.mainTitle}>
          Spieler
        </ThemedText>
      </View>
      <View style={styles.participantsList}>
        {match.participants === undefined || match.participants.length === 0 ? (
          <ThemedText>Keine Spieler</ThemedText>
        ) : (
          (match.participants as User[]).map((participant) => (
            <View style={styles.participantItem} key={participant._id}>
              <View style={styles.participantItemInner}>
                <View style={styles.participantItemBadge}>
                  <ThemedText style={styles.participantItemBadgeText}>
                    {participant.firstName[0]}
                    {participant.lastName[0]}
                  </ThemedText>
                </View>
                <ThemedText>
                  {participant.firstName} {participant.lastName}
                </ThemedText>
              </View>
              {isAdmin && (
                <AnimatedHapticButton
                  onPress={() => removePlayer(participant)}
                  style={styles.participantItemBadge}
                >
                  <Ionicons
                    name="close"
                    size={16}
                    style={styles.participantItemBadgeIcon}
                  />
                </AnimatedHapticButton>
              )}
            </View>
          ))
        )}
      </View>
      {isAdmin && (
        <AnimatedHapticButton
          style={[styles.buttonOutline, { marginTop: 16 }]}
          onPress={() => setShowPickerModal(true)}
        >
          <ThemedText style={{ color: Colors.text }}>
            Spieler hinzufügen
          </ThemedText>
        </AnimatedHapticButton>
      )}
      <Modal
        animationType="slide"
        transparent={false}
        visible={showPickerModal}
        onRequestClose={() => setShowPickerModal(false)}
      >
        <SafeAreaProvider>
          <SafeAreaView style={styles.modalContainer}>
            <View style={{ paddingHorizontal: 24 }}>
              <View style={styles.modalButtonContainer}>
                <AnimatedHapticButton
                  onPress={() => setShowPickerModal(false)}
                  style={styles.modalButton}
                  useHaptics={false}
                >
                  <Ionicons name="close" size={24} color={Colors.text} />
                </AnimatedHapticButton>
              </View>
              <View>
                <ThemedText type="title" style={styles.modalHeaderTitle}>
                  Spieler hinzufügen
                </ThemedText>
                <View style={styles.modalFormContainer}>
                  {availablePlayers.length === 0 && (
                    <ThemedText style={{ textAlign: "center", marginTop: 24 }}>
                      Keine Spieler verfügbar
                    </ThemedText>
                  )}
                  <View style={styles.participantsList}>
                    {availablePlayers.map((player: User) => {
                      const isAvailable = isPlayerAvailable(player);
                      return (
                        <View style={styles.participantItem} key={player._id}>
                          <View style={styles.participantItemInner}>
                            <View style={styles.participantItemBadge}>
                              <ThemedText style={styles.participantItemBadgeText}>
                                {player.firstName[0]}
                                {player.lastName[0]}
                              </ThemedText>
                            </View>
                            <View>
                              <ThemedText style={{ lineHeight: isAvailable ? undefined : 16 }}>
                                {player.firstName} {player.lastName}
                              </ThemedText>
                              {!isAvailable && (
                                <ThemedText style={{ color: Colors.error, fontSize: 12, lineHeight: 16 }}>
                                  Nicht verfügbar
                                </ThemedText>
                              )}
                            </View>
                          </View>
                          {isAdmin && (
                            <AnimatedHapticButton
                              onPress={() => isPlayerAvailable(player) && addPlayer(player)}
                              style={[styles.participantItemBadge, !isPlayerAvailable(player) && { opacity: 0.5 }]}
                              disabled={!isPlayerAvailable(player)}
                            >
                              <Ionicons
                                name="add"
                                size={18}
                                style={styles.participantItemBadgeIcon}
                              />
                            </AnimatedHapticButton>
                          )}
                        </View>
                      );
                    })}
                  </View>
                </View>
                <View style={styles.modalActionButtonContainer}>
                  <AnimatedHapticButton
                    onPress={() => setShowPickerModal(false)}
                    style={styles.buttonOutline}
                  >
                    <ThemedText style={{ color: Colors.text }}>
                      Abbrechen
                    </ThemedText>
                  </AnimatedHapticButton>
                </View>
              </View>
            </View>
          </SafeAreaView>
        </SafeAreaProvider>
      </Modal>
    </View>
  );
}

const styles = StyleSheet.create({
  headerContainer: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
  },
  addButton: {
    height: 38,
    width: 38,
    borderRadius: 12,
    backgroundColor: Colors.text,
    justifyContent: "center",
    alignItems: "center",
  },
  mainTitle: {
    fontSize: 24,
  },
  participantsList: {
    flexDirection: "column",
    gap: 8,
    marginTop: 12,
  },
  participantItem: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    gap: 8,
  },
  participantItemInner: {
    flexDirection: "row",
    alignItems: "center",
    gap: 8,
  },
  participantItemBadge: {
    width: 28,
    height: 28,
    borderRadius: 9,
    backgroundColor: Colors.text,
    justifyContent: "center",
    alignItems: "center",
  },
  participantItemBadgeText: {
    color: Colors.background,
    fontSize: 12,
  },
  participantItemBadgeIcon: {
    color: Colors.background,
  },
  modalContainer: {
    flex: 1,
    backgroundColor: Colors.background,
    paddingHorizontal: 24,
  },
  modalButtonContainer: {
    flexDirection: "row",
    justifyContent: "flex-end",
  },
  modalButton: {
    height: 48,
    width: 48,
  },
  modalHeaderTitle: {
    fontSize: 32,
    marginTop: 8,
    color: Colors.text,
  },
  modalHeaderText: {
    fontSize: 18,
    marginVertical: 20,
  },
  modalSwitchesList: {
    gap: 12,
  },
  modalSwitchItem: {
    flexDirection: "row",
    alignItems: "center",
    gap: 12,
  },
  modalActionButtonContainer: {
    marginTop: 12,
  },
  modalFormContainer: {
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
  inputButtonOutline: {
    height: 52,
    borderColor: Colors.text,
    borderWidth: 2,
    borderRadius: 15,
    paddingHorizontal: 16,
    paddingVertical: 12,
  },
  clubPickerDefault: {
    height: 52,
  },
  clubPickerPressableStyle: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "space-between",
  },
  clubPickerSelected: {
    borderColor: Colors.text,
    borderWidth: 2,
    boxShadow: "none",
  },
  pickerItem: {
    fontSize: 16,
  },
  errorText: {
    color: Colors.error,
    fontSize: 14,
    marginTop: 2,
  },
  modalOverlay: {
    flex: 1,
    backgroundColor: "rgba(0, 0, 0, 0.5)",
    justifyContent: "center",
    alignItems: "center",
  },
  pickerModal: {
    width: "90%",
    maxHeight: "80%",
    backgroundColor: Colors.background,
    borderRadius: 20,
    padding: 20,
  },
  pickerHeader: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    marginBottom: 20,
  },
  pickerTitle: {
    fontSize: 24,
  },
  clubList: {
    maxHeight: "100%",
  },
  clubItem: {
    padding: 15,
    borderRadius: 10,
    marginBottom: 8,
    borderWidth: 1,
    borderColor: Colors.text,
  },
  selectedClubItem: {
    backgroundColor: Colors.text,
  },
  clubItemText: {
    fontSize: 16,
  },
  selectedClubItemText: {
    color: Colors.background,
  },
  inputError: {
    borderColor: Colors.error,
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
});
