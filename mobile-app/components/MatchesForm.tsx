import { FC, useEffect, useState } from "react";
import {
  Keyboard,
  KeyboardAvoidingView,
  Modal,
  Platform,
  StyleSheet,
  Switch,
  TextInput,
  TouchableOpacity,
  TouchableWithoutFeedback,
  View,
} from "react-native";
import { ThemedText } from "./ThemedText";
import { AnimatedHapticButton } from "./AnimatedHapticButton";
import { Colors } from "@/constants/Colors";
import { Ionicons } from "@expo/vector-icons";
import { Match } from "@/constants/Match";
import { Spinner } from "./Spinner";
import { Club } from "@/constants/Club";
import {
  clubFetchParticipants,
  clubsFetch,
  matchCreate,
  matchUpdate,
} from "@/api/Api";
import DateTimePickerModal from "react-native-modal-datetime-picker";
import { format } from "date-fns";
import { SafeAreaProvider, SafeAreaView } from "react-native-safe-area-context";
import { router } from "expo-router";
import { User } from "@/constants/User";

export const MatchesForm: FC<{
  matchData?: Match;
  fetchMatch?: () => Promise<void>;
  setShowModalEditMatch?: (value: boolean) => void;
}> = ({ matchData, fetchMatch, setShowModalEditMatch = () => {} }) => {
  const defaultMatch: Match = {
    _id: "",
    club: { _id: "", name: "", code: "" },
    enemyClub: { name: "", address: "" },
    isHomeGame: true,
    date: undefined,
    participants: [],
  };

  const [match, setMatch] = useState<Match>(matchData || defaultMatch);
  const [clubs, setClubs] = useState<Club[]>([]);
  const [players, setPlayers] = useState<User[]>([]);
  const [selectedClub, setSelectedClub] = useState<Club | null>(
    matchData?.club || null
  );
  const [showDateTimePicker, setShowDateTimePicker] = useState(false);
  const [selectedClubError, setSelectedClubError] = useState("");
  const [dateError, setDateError] = useState("");
  const [enemyClubNameError, setEnemyClubNameError] = useState("");
  const [playerSelectionError, setPlayerSelectionError] = useState("");
  const [isFetching, setIsFetching] = useState(false);
  const [showPickerModal, setShowPickerModal] = useState<
    "club" | "player" | null
  >(null);

  useEffect(() => {
    fetchClubs();
  }, []);

  useEffect(() => {
    if (selectedClub) {
      fetchPlayers(selectedClub);
    }
  }, [selectedClub]);

  async function fetchClubs() {
    const { success, data } = await clubsFetch();
    if (success) {
      const clubsWhereAdmin = data.filter(
        (club: Club) => club.role === "admin"
      );
      setClubs(clubsWhereAdmin);
      if (clubsWhereAdmin.length === 1) {
        setMatch({ ...match, club: clubsWhereAdmin[0] });
        setSelectedClub(clubsWhereAdmin[0]);
      }
    }
  }

  async function fetchPlayers(club: Club) {
    const { success, data } = await clubFetchParticipants(club.code);
    if (success) {
      setPlayers(data);
    }
  }

  const handleSaveMatch = async () => {
    setSelectedClubError("");
    setDateError("");
    setEnemyClubNameError("");
    setPlayerSelectionError("");

    if (!selectedClub) {
      setSelectedClubError("Club muss ausgewählt sein");
      return;
    }
    if (!match.date) {
      setDateError("Datum muss ausgewählt sein");
      return;
    }
    if (!match.enemyClub.name) {
      setEnemyClubNameError("Gegner Name darf nicht leer sein");
      return;
    }
    if (
      matchData === undefined &&
      (!match.participants || match.participants.length === 0)
    ) {
      setPlayerSelectionError("Mindestens ein Spieler muss ausgewählt sein");
      return;
    }
    setIsFetching(true);

    if (matchData === undefined) {
      const newMatch: Match = {
        ...match,
        club: selectedClub || match.club,
      };

      const { success, data, error } = await matchCreate(newMatch);
      if (success) {
        router.dismissAll();
        router.push("/");
      }
    } else {
      const updatedMatch: Match = {
        ...match,
        club: selectedClub || match.club,
        participants: undefined,
      };

      const { success, data, error } = await matchUpdate(updatedMatch);
      if (success) {
        if (fetchMatch) {
          await fetchMatch();
        }
        setShowModalEditMatch(false);
      }
    }

    setIsFetching(false);
  };

  const handleSaveClub = () => {
    if (!selectedClub) return;
    setMatch({ ...match, club: selectedClub });
    setShowPickerModal(null);
  };

  const handleSaveParticipants = () => {
    setShowPickerModal(null);
  };

  const togglePlayer = (player: User) => {
    const isSelected = match.participants?.includes(player._id);
    if (isSelected) {
      setMatch({
        ...match,
        participants: match.participants?.filter((id) => id !== player._id),
      });
    } else {
      setMatch({
        ...match,
        participants: [...(match.participants || []), player._id],
      });
    }
  };

  const isPlayerAvailable = (player: User) => {
    if (match.date === undefined) return true;

    const weekday = format(
      match.date,
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

  const handleBack = () => {
    router.back();
  };

  if (clubs.length === 0) {
    return <Spinner />;
  }

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === "ios" ? "padding" : "height"}
      keyboardVerticalOffset={Platform.OS === "ios" ? 100 : 0}
    >
      <TouchableWithoutFeedback onPress={Keyboard.dismiss}>
        <View>
          {matchData !== undefined && (
            <View style={styles.modalButtonContainer}>
              <AnimatedHapticButton
                onPress={() => setShowModalEditMatch(false)}
                style={styles.modalButton}
                useHaptics={false}
              >
                <Ionicons name="close" size={24} color={Colors.text} />
              </AnimatedHapticButton>
            </View>
          )}
          <View style={styles.header}>
            <View style={styles.headerButtons}>
              <AnimatedHapticButton
                style={styles.buttonOutline}
                onPress={handleBack}
              >
                <Ionicons name="arrow-back" size={20} color={Colors.text} />
              </AnimatedHapticButton>
              {matchData === undefined && (
                <ThemedText type="title" style={styles.headerTitle}>
                  Match erstellen
                </ThemedText>
              )}
            </View>
          </View>
          <View>
            {matchData !== undefined && (
              <ThemedText type="title" style={styles.modalHeaderTitle}>
                Match bearbeiten
              </ThemedText>
            )}
            <View style={styles.modalFormContainer}>
              {matchData === undefined && (
                <View>
                  <AnimatedHapticButton
                    onPress={() => setShowPickerModal("club")}
                    style={[
                      styles.inputButtonOutline,
                      selectedClubError ? styles.inputError : null,
                    ]}
                    pressableStyle={styles.clubPickerPressableStyle}
                  >
                    <ThemedText style={{ color: Colors.text }}>
                      {match.club.name || "Club auswählen"}
                    </ThemedText>
                    <Ionicons
                      name="chevron-forward-outline"
                      size={20}
                      color={Colors.text}
                    />
                  </AnimatedHapticButton>
                  {selectedClubError && (
                    <ThemedText style={styles.errorText}>
                      {selectedClubError}
                    </ThemedText>
                  )}
                </View>
              )}
              <View>
                <TouchableOpacity
                  style={[styles.input, dateError ? styles.inputError : null]}
                  onPress={() => setShowDateTimePicker(true)}
                >
                  <ThemedText style={{ color: Colors.text }}>
                    {match.date
                      ? format(match.date, "dd.MM.yyyy HH:mm")
                      : "Datum & Zeit auswählen"}
                  </ThemedText>
                </TouchableOpacity>
                {dateError && (
                  <ThemedText style={styles.errorText}>{dateError}</ThemedText>
                )}
              </View>
              <DateTimePickerModal
                date={new Date(match.date || Date.now())}
                isVisible={showDateTimePicker}
                mode="datetime"
                display="spinner"
                minuteInterval={15}
                onConfirm={(date: Date) => {
                  const roundedDate = new Date(
                    date.setMinutes(Math.ceil(date.getMinutes() / 15) * 15)
                  );
                  setMatch({ ...match, date: roundedDate });
                  setShowDateTimePicker(false);
                  setDateError("");
                }}
                onCancel={() => setShowDateTimePicker(false)}
              />
              <View>
                <TextInput
                  placeholder="Gegner Name"
                  value={match.enemyClub.name}
                  onChangeText={(text) => {
                    setMatch({
                      ...match,
                      enemyClub: { ...match.enemyClub, name: text },
                    });
                    setEnemyClubNameError("");
                  }}
                  style={[
                    styles.input,
                    enemyClubNameError ? styles.inputError : null,
                  ]}
                  placeholderTextColor={Colors.text}
                  autoCapitalize="words"
                  autoComplete="off"
                />
                {enemyClubNameError && (
                  <ThemedText style={styles.errorText}>
                    {enemyClubNameError}
                  </ThemedText>
                )}
              </View>
              <View style={styles.modalSwitchItem}>
                <Switch
                  trackColor={{ false: undefined, true: Colors.text }}
                  onValueChange={() => {
                    setMatch({ ...match, isHomeGame: !match.isHomeGame });
                  }}
                  value={match.isHomeGame}
                />
                <ThemedText style={{ fontSize: 16 }}>Heimspiel</ThemedText>
              </View>
              {!match.isHomeGame && (
                <TextInput
                  placeholder="Gegner Adresse (optional)"
                  value={match.enemyClub.address}
                  onChangeText={(text) => {
                    setMatch({
                      ...match,
                      enemyClub: { ...match.enemyClub, address: text },
                    });
                  }}
                  style={styles.input}
                  placeholderTextColor={Colors.text}
                  autoCapitalize="words"
                  autoComplete="off"
                />
              )}
              {matchData === undefined && selectedClub && match.date && (
                <View>
                  <AnimatedHapticButton
                    onPress={() => setShowPickerModal("player")}
                    style={[
                      styles.inputButtonOutline,
                      playerSelectionError ? styles.inputError : null,
                    ]}
                    pressableStyle={styles.clubPickerPressableStyle}
                  >
                    <ThemedText style={{ color: Colors.text }}>
                      Spieler auswählen
                    </ThemedText>
                    <Ionicons
                      name="chevron-forward-outline"
                      size={20}
                      color={Colors.text}
                    />
                  </AnimatedHapticButton>
                  {match.participants !== undefined &&
                    match.participants.length > 0 && (
                      <ThemedText>
                        {match.participants.length} Ausgewählt
                      </ThemedText>
                    )}
                  {playerSelectionError && (
                    <ThemedText style={styles.errorText}>
                      {playerSelectionError}
                    </ThemedText>
                  )}
                </View>
              )}
            </View>
            <View style={styles.modalActionButtonContainer}>
              <AnimatedHapticButton
                disabled={isFetching}
                onPress={handleSaveMatch}
                style={styles.button}
              >
                {isFetching ? (
                  <Spinner fill={Colors.background} />
                ) : (
                  <ThemedText style={styles.buttonText}>
                    {matchData === undefined ? "Erstellen" : "Speichern"}
                  </ThemedText>
                )}
              </AnimatedHapticButton>
              <AnimatedHapticButton
                onPress={() => setShowModalEditMatch(false)}
                style={styles.buttonOutline}
              >
                <ThemedText style={{ color: Colors.text }}>
                  Abbrechen
                </ThemedText>
              </AnimatedHapticButton>
            </View>
          </View>
          <Modal
            animationType="slide"
            transparent={false}
            visible={showPickerModal !== null}
            onRequestClose={() => setShowPickerModal(null)}
          >
            <SafeAreaProvider>
              <SafeAreaView style={styles.modalContainer}>
                <View>
                  <View style={styles.modalButtonContainer}>
                    <AnimatedHapticButton
                      onPress={() => setShowPickerModal(null)}
                      style={styles.modalButton}
                      useHaptics={false}
                    >
                      <Ionicons name="close" size={24} color={Colors.text} />
                    </AnimatedHapticButton>
                  </View>
                  <View>
                    <ThemedText type="title" style={styles.modalHeaderTitle}>
                      {showPickerModal === "club" ? "Club" : "Spieler"}{" "}
                      auswählen
                    </ThemedText>
                    <View
                      style={[styles.modalFormContainer, { marginTop: 24 }]}
                    >
                      {showPickerModal === "club"
                        ? clubs.map((club) => (
                            <AnimatedHapticButton
                              onPress={() => setSelectedClub(club)}
                              style={
                                selectedClub?._id === club._id
                                  ? [
                                      styles.clubPickerDefault,
                                      styles.clubPickerSelected,
                                    ]
                                  : styles.clubPickerDefault
                              }
                              pressableStyle={styles.clubPickerPressableStyle}
                              key={club._id}
                            >
                              {selectedClub?._id === club._id ? (
                                <Ionicons
                                  name="ellipse"
                                  size={20}
                                  color={Colors.text}
                                />
                              ) : (
                                <Ionicons
                                  name="ellipse-outline"
                                  size={20}
                                  color={Colors.muted}
                                />
                              )}
                              <ThemedText
                                style={{
                                  color:
                                    selectedClub?._id === club._id
                                      ? Colors.text
                                      : Colors.muted,
                                }}
                              >
                                {club.name}
                              </ThemedText>
                            </AnimatedHapticButton>
                          ))
                        : players.map((player: User) => {
                            const isAvailable = isPlayerAvailable(player);
                            return (
                              <AnimatedHapticButton
                                onPress={() =>
                                  isAvailable && togglePlayer(player)
                                }
                                style={[
                                  styles.clubPickerDefault,
                                  match.participants?.includes(player._id) &&
                                    styles.clubPickerSelected,
                                  !isAvailable && { opacity: 0.5 },
                                ]}
                                pressableStyle={styles.clubPickerPressableStyle}
                                key={player._id}
                              >
                                {match.participants?.includes(player._id) ? (
                                  <Ionicons
                                    name="ellipse"
                                    size={20}
                                    color={Colors.text}
                                  />
                                ) : (
                                  <Ionicons
                                    name="ellipse-outline"
                                    size={20}
                                    color={
                                      isAvailable ? Colors.muted : Colors.error
                                    }
                                  />
                                )}
                                <View style={{ flex: 1 }}>
                                  <ThemedText
                                    style={{
                                      color: match.participants?.includes(
                                        player._id
                                      )
                                        ? Colors.text
                                        : isAvailable
                                        ? Colors.muted
                                        : Colors.error,
                                      lineHeight: 16,
                                    }}
                                  >
                                    {player.firstName} {player.lastName}
                                  </ThemedText>
                                  {!isAvailable && (
                                    <ThemedText
                                      style={{
                                        fontSize: 12,
                                        color: Colors.error,
                                        lineHeight: 16,
                                      }}
                                    >
                                      Nicht verfügbar
                                    </ThemedText>
                                  )}
                                </View>
                              </AnimatedHapticButton>
                            );
                          })}
                    </View>
                    <View style={styles.modalActionButtonContainer}>
                      <AnimatedHapticButton
                        disabled={
                          (showPickerModal === "club" && !selectedClub) ||
                          (showPickerModal === "player" &&
                            (!match.participants ||
                              match.participants.length === 0))
                        }
                        onPress={
                          matchData === undefined
                            ? handleSaveClub
                            : handleSaveParticipants
                        }
                        style={styles.button}
                      >
                        <ThemedText style={styles.buttonText}>
                          Speichern
                        </ThemedText>
                      </AnimatedHapticButton>
                      <AnimatedHapticButton
                        onPress={() => setShowPickerModal(null)}
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
      </TouchableWithoutFeedback>
    </KeyboardAvoidingView>
  );
};

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
  buttonOutline: {
    height: 45,
    borderColor: Colors.text,
    borderWidth: 2,
    borderRadius: 15,
    justifyContent: "center",
    alignItems: "center",
    paddingHorizontal: 10,
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
    marginTop: 24,
    gap: 10,
  },
  modalFormContainer: {
    gap: 12,
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
    borderRadius: 15,
    paddingHorizontal: 16,
    paddingVertical: 12,
    boxShadow: "0px 0px 10px rgba(0, 0, 0, 0.15)",
    borderWidth: 2,
    borderColor: "transparent",
  },
  clubPickerPressableStyle: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "flex-start",
    gap: 8,
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
});
