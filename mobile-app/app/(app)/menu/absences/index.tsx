import {
  Modal,
  RefreshControl,
  ScrollView,
  StyleSheet,
  Switch,
  TouchableOpacity,
  View,
} from "react-native";
import { AnimatedHapticButton } from "@/components/AnimatedHapticButton";
import { Ionicons } from "@expo/vector-icons";
import { Colors } from "@/constants/Colors";
import { router } from "expo-router";
import { ThemedText } from "@/components/ThemedText";
import { useCallback, useEffect, useState } from "react";
import {
  absenceCreatePeriod,
  absenceDeletePeriod,
  absenceUpdateDays,
  absenceFetchRecords,
} from "@/api/Api";
import { format } from "date-fns";
import { SafeAreaProvider, SafeAreaView } from "react-native-safe-area-context";
import { Spinner } from "@/components/Spinner";
import DateTimePickerModal from "react-native-modal-datetime-picker";

interface AbsencePeriod {
  _id: string;
  startDate: Date;
  endDate: Date;
}

export default function MenuAbsences() {
  const [absenceDays, setAbsenceDays] = useState<Record<string, boolean>>({});
  const [absencePeriods, setAbsencePeriods] = useState<AbsencePeriod[]>([]);
  const [showModalEditAbsenceDays, setShowModalEditAbsenceDays] =
    useState(false);
  const [showModalAddAbsencePeriod, setShowModalAddAbsencePeriod] =
    useState(false);
  const [isFetchingEditAbsenceDays, setIsFetchingEditAbsenceDays] =
    useState(false);
  const [isFetchingAddAbsencePeriod, setIsFetchingAddAbsencePeriod] =
    useState(false);
  const [startDate, setStartDate] = useState<Date | null>(null);
  const [endDate, setEndDate] = useState<Date | null>(null);
  const [showStartPicker, setShowStartPicker] = useState(false);
  const [showEndPicker, setShowEndPicker] = useState(false);
  const [dateError, setDateError] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await fetchAbsencePeriods();
    setRefreshing(false);
  }, []);

  const weekdays = [
    { key: "monday", label: "Montag" },
    { key: "tuesday", label: "Dienstag" },
    { key: "wednesday", label: "Mittwoch" },
    { key: "thursday", label: "Donnerstag" },
    { key: "friday", label: "Freitag" },
    { key: "saturday", label: "Samstag" },
    { key: "sunday", label: "Sonntag" },
  ];

  useEffect(() => {
    fetchAbsencePeriods();
  }, []);

  async function fetchAbsencePeriods() {
    const { success, data, error } = await absenceFetchRecords();
    if (success) {
      const sortedPeriods = data.absencePeriods.sort(
        (a: AbsencePeriod, b: AbsencePeriod) =>
          new Date(a.startDate).getTime() - new Date(b.startDate).getTime()
      );
      setAbsenceDays(data.absenceDays);
      setAbsencePeriods(sortedPeriods);
    } else {
      console.error(error);
    }
  }

  const displayDateRange = (startDate: Date, endDate: Date) => {
    const formattedStart = format(new Date(startDate), "dd.MM.yyyy");
    const formattedEnd = format(new Date(endDate), "dd.MM.yyyy");
    return `${formattedStart} - ${formattedEnd}`;
  };

  const formatAbsenceDays = (absenceDays: Record<string, boolean>): string => {
    const dayAbbreviations: Record<string, string> = {
      monday: "Mo.",
      tuesday: "Di.",
      wednesday: "Mi.",
      thursday: "Do.",
      friday: "Fr.",
      saturday: "Sa.",
      sunday: "So.",
    };
    const activeDays = Object.entries(absenceDays)
      .filter(([_, isActive]) => isActive)
      .map(([day]) => dayAbbreviations[day] || day);
    return activeDays.length === 0
      ? "Keine Tage ausgewählt"
      : `Jeden ${activeDays.join(", ")}`;
  };

  const handleBack = () => {
    router.back();
  };

  const handleEditAbsenceDays = () => {
    setShowModalEditAbsenceDays(true);
  };

  const handleSaveAbsenceDays = async () => {
    setIsFetchingEditAbsenceDays(true);
    const { success, error } = await absenceUpdateDays(absenceDays);
    setIsFetchingEditAbsenceDays(false);
    if (success) {
      setShowModalEditAbsenceDays(false);
    } else {
      console.error(error);
    }
  };

  const handleDeleteAbsencePeriod = async (id: string) => {
    const { success, error } = await absenceDeletePeriod(id);
    if (success) {
      setAbsencePeriods((prev) => prev.filter((period) => period._id !== id));
    } else {
      console.error(error);
    }
  };

  const handleAddAbsencePeriod = () => {
    setShowModalAddAbsencePeriod(true);
  };

  const handleSaveAbsencePeriod = async () => {
    setDateError(null);
    if (!startDate || !endDate) {
      setDateError("Start- und Enddatum sind erforderlich.");
      return;
    }
    if (endDate < startDate) {
      setDateError("Startdatum muss vor dem Enddatum liegen.");
      return;
    }
    startDate.setHours(0, 0, 0, 0);
    endDate.setHours(23, 59, 59, 999);
    setIsFetchingAddAbsencePeriod(true);
    const { success, error } = await absenceCreatePeriod(startDate, endDate);
    setIsFetchingAddAbsencePeriod(false);
    if (success) {
      await fetchAbsencePeriods();
      setShowModalAddAbsencePeriod(false);
    } else {
      console.error(error);
    }
  };

  const handleToggleAbsenceDay = (day: string) => {
    setAbsenceDays((prev) => ({ ...prev, [day]: !prev[day] }));
  };

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
            Abwesenheitszeiten
          </ThemedText>
        </View>
      </View>
      <View style={styles.menuContainer}>
        <View>
          <View style={styles.menuItem}>
            <ThemedText style={styles.menuItemText}>
              {formatAbsenceDays(absenceDays)}
            </ThemedText>
            <AnimatedHapticButton
              style={styles.menuItemButton}
              onPress={handleEditAbsenceDays}
            >
              <Ionicons
                name="pencil-outline"
                size={20}
                color={Colors.background}
              />
            </AnimatedHapticButton>
          </View>
        </View>
        <View style={styles.divider}></View>
        <View style={styles.menuContainerInner}>
          {absencePeriods.map((period) => (
            <View style={styles.menuItem} key={period._id}>
              <ThemedText style={styles.menuItemText}>
                {displayDateRange(period.startDate, period.endDate)}
              </ThemedText>
              <AnimatedHapticButton
                style={styles.menuItemButton}
                onPress={() => handleDeleteAbsencePeriod(period._id)}
              >
                <Ionicons name="trash-outline" size={20} color={Colors.error} />
              </AnimatedHapticButton>
            </View>
          ))}
          {absencePeriods.length === 0 && (
            <ThemedText style={styles.absencePeriodsEmpty}>
              Keine Einträge vorhanden
            </ThemedText>
          )}
          <AnimatedHapticButton
            onPress={handleAddAbsencePeriod}
            style={styles.button}
          >
            <ThemedText style={styles.buttonText}>
              Eintrag hinzufügen
            </ThemedText>
          </AnimatedHapticButton>
        </View>
      </View>
      <Modal
        animationType="slide"
        transparent={false}
        visible={showModalEditAbsenceDays}
        onRequestClose={() => setShowModalEditAbsenceDays(false)}
      >
        <SafeAreaProvider>
          <SafeAreaView style={styles.modalContainer}>
            <View style={styles.modalButtonContainer}>
              <AnimatedHapticButton
                onPress={() => setShowModalEditAbsenceDays(false)}
                style={styles.modalButton}
                useHaptics={false}
              >
                <Ionicons name="close" size={24} color={Colors.text} />
              </AnimatedHapticButton>
            </View>
            <View style={styles.modalContent}>
              <ThemedText type="title" style={styles.modalHeaderTitle}>
                Abwesenheitszeiten
              </ThemedText>
              <ThemedText style={styles.modalHeaderText}>
                Du hast hier die Möglichkeit Wochentage zu definieren an denen
                du immer abwesend bist.
              </ThemedText>
              <View style={styles.modalSwitchesList}>
                {weekdays.map(({ key, label }) => (
                  <View style={styles.modalSwitchItem} key={key}>
                    <Switch
                      trackColor={{ false: undefined, true: Colors.text }}
                      onValueChange={() => handleToggleAbsenceDay(key)}
                      value={!!absenceDays[key]}
                    />
                    <ThemedText style={{ fontSize: 18 }}>{label}</ThemedText>
                  </View>
                ))}
              </View>
              <View style={styles.modalActionButtonContainer}>
                <AnimatedHapticButton
                  disabled={isFetchingEditAbsenceDays}
                  onPress={handleSaveAbsenceDays}
                  style={styles.button}
                >
                  {isFetchingEditAbsenceDays ? (
                    <Spinner fill={Colors.background} />
                  ) : (
                    <ThemedText style={styles.buttonText}>Speichern</ThemedText>
                  )}
                </AnimatedHapticButton>
                <AnimatedHapticButton
                  onPress={() => setShowModalEditAbsenceDays(false)}
                  style={styles.buttonOutline}
                >
                  <ThemedText style={{ color: Colors.text }}>
                    Abbrechen
                  </ThemedText>
                </AnimatedHapticButton>
              </View>
            </View>
          </SafeAreaView>
        </SafeAreaProvider>
      </Modal>
      <Modal
        animationType="slide"
        transparent={false}
        visible={showModalAddAbsencePeriod}
        onRequestClose={() => setShowModalAddAbsencePeriod(false)}
      >
        <SafeAreaProvider>
          <SafeAreaView style={styles.modalContainer}>
            <View style={styles.modalButtonContainer}>
              <AnimatedHapticButton
                onPress={() => setShowModalAddAbsencePeriod(false)}
                style={styles.modalButton}
                useHaptics={false}
              >
                <Ionicons name="close" size={24} color={Colors.text} />
              </AnimatedHapticButton>
            </View>
            <View style={styles.modalContent}>
              <ThemedText type="title" style={styles.modalHeaderTitle}>
                Abwesenheitszeiten
              </ThemedText>
              <ThemedText style={styles.modalHeaderText}>
                Wähle ein Start- und Enddatum:
              </ThemedText>
              <View style={styles.modalFormContainer}>
                <TouchableOpacity
                  style={styles.input}
                  onPress={() => setShowStartPicker(true)}
                >
                  <ThemedText style={{ color: Colors.text }}>
                    {startDate
                      ? format(startDate, "dd.MM.yyyy")
                      : "Startdatum auswählen"}
                  </ThemedText>
                </TouchableOpacity>
                <DateTimePickerModal
                  isVisible={showStartPicker}
                  mode="date"
                  display="spinner"
                  onConfirm={(date: Date) => {
                    setStartDate(date);
                    setShowStartPicker(false);
                  }}
                  onCancel={() => setShowStartPicker(false)}
                />
                <TouchableOpacity
                  style={[styles.input, { marginTop: 12 }]}
                  onPress={() => setShowEndPicker(true)}
                >
                  <ThemedText style={{ color: Colors.text }}>
                    {endDate
                      ? format(endDate, "dd.MM.yyyy")
                      : "Enddatum auswählen"}
                  </ThemedText>
                </TouchableOpacity>
                <DateTimePickerModal
                  isVisible={showEndPicker}
                  mode="date"
                  display="spinner"
                  onConfirm={(date: Date) => {
                    setEndDate(date);
                    setShowEndPicker(false);
                  }}
                  onCancel={() => setShowEndPicker(false)}
                />
                {dateError && (
                  <ThemedText style={styles.errorText}>{dateError}</ThemedText>
                )}
              </View>
              <View style={styles.modalActionButtonContainer}>
                <AnimatedHapticButton
                  disabled={isFetchingAddAbsencePeriod}
                  onPress={handleSaveAbsencePeriod}
                  style={styles.button}
                >
                  {isFetchingAddAbsencePeriod ? (
                    <Spinner fill={Colors.background} />
                  ) : (
                    <ThemedText style={styles.buttonText}>Speichern</ThemedText>
                  )}
                </AnimatedHapticButton>
                <AnimatedHapticButton
                  onPress={() => setShowModalAddAbsencePeriod(false)}
                  style={styles.buttonOutline}
                >
                  <ThemedText style={{ color: Colors.text }}>
                    Abbrechen
                  </ThemedText>
                </AnimatedHapticButton>
              </View>
            </View>
          </SafeAreaView>
        </SafeAreaProvider>
      </Modal>
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
  menuItemButton: {
    height: 38,
    width: 38,
    borderRadius: 12,
    justifyContent: "center",
    alignItems: "center",
    backgroundColor: Colors.text,
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
  divider: {
    borderBottomWidth: 1,
    borderBottomColor: "#B0B0B0",
    marginVertical: 16,
  },
  absencePeriodsEmpty: {
    color: Colors.muted,
    fontSize: 16,
    textAlign: "center",
    paddingBottom: 10,
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
  modalContent: {
    flex: 1,
  },
  modalHeaderTitle: {
    fontSize: 32,
    marginTop: 8,
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
    gap: 0,
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
  errorText: {
    color: Colors.error,
    fontSize: 14,
    marginTop: 2,
  },
});
