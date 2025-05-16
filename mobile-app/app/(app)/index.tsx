import { ThemedText } from "@/components/ThemedText";
import {
  StyleSheet,
  View,
  ScrollView,
  Modal,
  SafeAreaView,
  RefreshControl,
} from "react-native";
import { AnimatedHapticButton } from "@/components/AnimatedHapticButton";
import { Ionicons } from "@expo/vector-icons";
import { Colors } from "@/constants/Colors";
import { useState, useRef, useEffect, useCallback } from "react";
import { CalendarList } from "react-native-calendars";
import { SafeAreaProvider } from "react-native-safe-area-context";
import { router, useFocusEffect } from "expo-router";
import { clubsFetch, matchesFetchAll } from "@/api/Api";
import { MatchesContainer } from "@/components/MatchesContainer";
import { Match } from "@/constants/Match";

const ITEM_WIDTH = 50;
const ITEM_GAP = 10;
const HORIZONTAL_PADDING = 16;
const ITEM_SPACING = ITEM_WIDTH + ITEM_GAP;

type CalendarItem = { type: "date"; date: Date };

export default function Home() {
  const initialDate = new Date();
  const [selectedDate, setSelectedDate] = useState(initialDate);
  const [dates, setDates] = useState<Date[]>([]);
  const [scrollViewWidth, setScrollViewWidth] = useState(0);
  const [contentWidth, setContentWidth] = useState(0);
  const [showDatePicker, setShowDatePicker] = useState(false);
  const [currentMonth, setCurrentMonth] = useState(
    initialDate.toISOString().split("T")[0]
  );
  const [isAdminInAnyClub, setIsAdminInAnyClub] = useState(false);
  const [matches, setMatches] = useState<Match[]>([]);
  const [refreshing, setRefreshing] = useState(false);

  const scrollViewRef = useRef<ScrollView>(null);

  useEffect(() => {
    const datesArray: Date[] = [];
    const year = selectedDate.getFullYear();
    const month = selectedDate.getMonth();
    const firstDay = new Date(year, month, 1);
    const lastDay = new Date(year, month + 1, 0);
    let current = new Date(firstDay);
    while (current <= lastDay) {
      datesArray.push(new Date(current));
      current.setDate(current.getDate() + 1);
    }
    setDates(datesArray);
  }, [selectedDate]);

  useFocusEffect(
    useCallback(() => {
      fetchClubs();
      fetchAllMatches();

      return () => {
        setMatches([]);
      };
    }, [])
  );

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    const now = new Date();
    setSelectedDate(now);
    setCurrentMonth(now.toISOString().split("T")[0]);
    setShowDatePicker(false);
    await fetchClubs();
    await fetchAllMatches();
    setRefreshing(false);
  }, []);

  const fetchClubs = async () => {
    const { success, data } = await clubsFetch();
    if (success) {
      if (data.length === 0) {
        if (router.canDismiss()) router.dismissAll();
        router.push({
          pathname: "/auth/register/club-code",
          params: { from: "app" },
        });
      } else {
        setIsAdminInAnyClub(
          data.some((club: { role: string }) => club.role === "admin")
        );
      }
    }
  };

  const fetchAllMatches = async () => {
    const { success, data } = await matchesFetchAll();
    if (success) {
      setMatches(data);
    }
  };

  const nextMonthDate = new Date(
    selectedDate.getFullYear(),
    selectedDate.getMonth() + 1,
    1
  );
  const nextMonthName = new Intl.DateTimeFormat("de-CH", {
    month: "short",
  }).format(nextMonthDate);

  const prevMonthDate = new Date(
    selectedDate.getFullYear(),
    selectedDate.getMonth() - 1,
    1
  );
  const prevMonthName = new Intl.DateTimeFormat("de-CH", {
    month: "short",
  }).format(prevMonthDate);

  const items: CalendarItem[] = dates.map(
    (d): CalendarItem => ({
      type: "date",
      date: d,
    })
  );

  const isSelected = (date: Date) =>
    date.toDateString() === selectedDate.toDateString();

  const isToday = (date: Date) => {
    const today = new Date();
    return date.toDateString() === today.toDateString();
  };

  const calculateOffset = (index: number) => {
    const itemCenter =
      HORIZONTAL_PADDING + index * ITEM_SPACING + ITEM_WIDTH / 2;
    const idealOffset = itemCenter - scrollViewWidth / 2;
    const maxOffset = Math.max(0, contentWidth - scrollViewWidth);
    return Math.max(0, Math.min(idealOffset, maxOffset));
  };

  useEffect(() => {
    if (
      scrollViewRef.current &&
      scrollViewWidth &&
      contentWidth &&
      dates.length > 0
    ) {
      const index = dates.findIndex(
        (d) => d.toDateString() === selectedDate.toDateString()
      );
      if (index !== -1) {
        const offset = calculateOffset(index + 1);
        scrollViewRef.current.scrollTo({ x: offset, animated: false });
      } else {
        const todayIndex = dates.findIndex(
          (d) => d.toDateString() === new Date().toDateString()
        );
        if (todayIndex !== -1) {
          const offset = calculateOffset(todayIndex + 1);
          scrollViewRef.current.scrollTo({ x: offset, animated: false });
        }
      }
    }
  }, [selectedDate, scrollViewWidth, contentWidth, dates]);

  const handleDateSelect = (date: Date) => {
    setSelectedDate(date);
  };

  const openCalendar = () => {
    setCurrentMonth(selectedDate.toISOString().split("T")[0]);
    setShowDatePicker(true);
  };

  const switchToNextMonth = () => {
    const nextMonth = new Date(
      selectedDate.getFullYear(),
      selectedDate.getMonth() + 1,
      1
    );
    setSelectedDate(nextMonth);
  };

  const switchToPreviousMonth = () => {
    const prevMonth = new Date(
      selectedDate.getFullYear(),
      selectedDate.getMonth() - 1,
      1
    );
    setSelectedDate(prevMonth);
  };

  const openMenu = () => {
    router.push("/menu");
  };

  const formatDayName = (date: Date) =>
    new Intl.DateTimeFormat("de-DE", { weekday: "short" }).format(date);

  const formatDay = (date: Date) => date.getDate().toString();

  const getMarkedDates = () => {
    const marked: { [date: string]: any } = {};
    matches.forEach((match) => {
      const dateKey = new Date(match.date as Date).toISOString().split("T")[0];
      marked[dateKey] = {
        marked: true,
        dotColor: Colors.text,
      };
    });
    const selectedKey = selectedDate.toISOString().split("T")[0];
    marked[selectedKey] = {
      ...(marked[selectedKey] || {}),
      selected: true,
      selectedColor: Colors.text,
      dotColor: Colors.background,
    };
    return marked;
  };

  return (
    <ScrollView
      style={styles.container}
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
        <View style={styles.headerTexts}>
          <ThemedText type="title" style={styles.headerTitle}>
            {new Intl.DateTimeFormat("de-DE", { weekday: "long" }).format(
              selectedDate
            )}
          </ThemedText>
          <ThemedText type="subtitle" style={styles.headerText}>
            {new Intl.DateTimeFormat("de-DE", {
              day: "numeric",
              month: "short",
              year: "numeric",
            }).format(selectedDate)}
          </ThemedText>
        </View>
        <View style={styles.headerButtons}>
          <AnimatedHapticButton
            style={styles.buttonOutline}
            onPress={openCalendar}
          >
            <Ionicons
              name="calendar-clear-outline"
              size={24}
              color={Colors.text}
            />
          </AnimatedHapticButton>
          <AnimatedHapticButton style={styles.button} onPress={openMenu}>
            <Ionicons name="menu-outline" size={26} color={Colors.background} />
          </AnimatedHapticButton>
        </View>
      </View>

      <ScrollView
        ref={scrollViewRef}
        horizontal
        showsHorizontalScrollIndicator={false}
        contentContainerStyle={styles.calendarButtonsContainer}
        onLayout={(e) => setScrollViewWidth(e.nativeEvent.layout.width)}
        onContentSizeChange={(w, h) => setContentWidth(w)}
      >
        <AnimatedHapticButton
          style={[
            styles.calendarButtonOutline,
            { minWidth: ITEM_WIDTH, paddingTop: 6 },
          ]}
          onPress={switchToPreviousMonth}
        >
          <Ionicons name="arrow-back-outline" size={20} color={Colors.muted} />
          <ThemedText style={styles.calendarButtonMoreText}>
            {prevMonthName}
          </ThemedText>
        </AnimatedHapticButton>

        {items.map((item) => {
          const date = item.date;
          const selected = isSelected(date);
          const today = isToday(date);
          const hasMatch = matches.some(
            (match) =>
              new Date(match.date as Date).toDateString() ===
              date.toDateString()
          );
          let buttonStyle, textStyle;
          if (selected) {
            buttonStyle = styles.calendarButton;
            textStyle = styles.calendarButtonText;
          } else if (today) {
            buttonStyle = styles.todayButton;
            textStyle = styles.todayText;
          } else {
            buttonStyle = styles.calendarButtonOutline;
            textStyle = styles.calendarButtonTextLight;
          }
          return (
            <AnimatedHapticButton
              key={date.toISOString()}
              style={[buttonStyle, { minWidth: ITEM_WIDTH }]}
              onPress={() => handleDateSelect(date)}
            >
              <ThemedText style={[textStyle, { fontSize: 14 }]}>
                {formatDayName(date)}
              </ThemedText>
              <ThemedText style={[textStyle, { fontSize: 22, marginTop: 2 }]}>
                {formatDay(date)}
              </ThemedText>
              <View
                style={[
                  styles.dot,
                  {
                    backgroundColor: selected
                      ? Colors.background
                      : today
                      ? Colors.text
                      : Colors.muted,
                    opacity: hasMatch ? 100 : 0,
                  },
                ]}
              />
            </AnimatedHapticButton>
          );
        })}

        <AnimatedHapticButton
          style={[
            styles.calendarButtonOutline,
            { minWidth: ITEM_WIDTH, paddingTop: 6 },
          ]}
          onPress={switchToNextMonth}
        >
          <Ionicons
            name="arrow-forward-outline"
            size={20}
            color={Colors.muted}
          />
          <ThemedText style={styles.calendarButtonMoreText}>
            {nextMonthName}
          </ThemedText>
        </AnimatedHapticButton>
      </ScrollView>

      <MatchesContainer
        isAdminInAnyClub={isAdminInAnyClub}
        matches={matches}
        selectedDate={selectedDate}
      />

      <Modal
        animationType="slide"
        transparent={false}
        visible={showDatePicker}
        onRequestClose={() => setShowDatePicker(false)}
      >
        <SafeAreaProvider>
          <SafeAreaView style={styles.modalContainer}>
            <View style={styles.modalButtonContainer}>
              <AnimatedHapticButton
                onPress={() => setShowDatePicker(false)}
                style={styles.modalButton}
                useHaptics={false}
              >
                <Ionicons name="close" size={24} color={Colors.text} />
              </AnimatedHapticButton>
            </View>
            <View style={styles.modalContent}>
              <CalendarList
                current={currentMonth}
                onDayPress={(day) => {
                  const newSelectedDate = new Date(day.timestamp);
                  handleDateSelect(newSelectedDate);
                  setShowDatePicker(false);
                }}
                pastScrollRange={50}
                futureScrollRange={50}
                scrollEnabled={true}
                horizontal={false}
                showScrollIndicator={true}
                markingType="dot"
                markedDates={getMarkedDates()}
                style={styles.calendarList}
                firstDay={1}
                theme={{
                  textDayFontFamily: "Bilo",
                  textDayHeaderFontFamily: "Bilo",
                  todayButtonFontFamily: "Bilo",
                  textMonthFontFamily: "MozaicHUMVariable",
                  textMonthFontSize: 20,
                  todayBackgroundColor: "#D1D1D1",
                  todayButtonTextColor: Colors.text,
                  todayTextColor: Colors.text,
                  todayDotColor: "#D1D1D1",
                  calendarBackground: Colors.background,
                }}
              />
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
  },
  header: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    paddingVertical: 16,
    paddingHorizontal: 24,
  },
  headerTexts: {
    flexDirection: "column",
    gap: 6,
  },
  headerTitle: {
    fontSize: 28,
  },
  headerText: {
    fontSize: 14,
    color: Colors.muted,
  },
  headerButtons: {
    flexDirection: "row",
    gap: 8,
    justifyContent: "space-between",
    alignItems: "center",
  },
  button: {
    height: 48,
    backgroundColor: Colors.text,
    borderRadius: 15,
    justifyContent: "center",
    alignItems: "center",
    paddingHorizontal: 11,
  },
  buttonOutline: {
    height: 48,
    borderColor: Colors.text,
    borderWidth: 2,
    borderRadius: 15,
    justifyContent: "center",
    alignItems: "center",
    paddingHorizontal: 11,
  },
  calendarButtonsContainer: {
    paddingHorizontal: HORIZONTAL_PADDING,
    paddingVertical: 10,
    gap: 10,
    height: 85,
    alignItems: "center",
    flexDirection: "row",
  },
  calendarButton: {
    height: "auto",
    backgroundColor: Colors.text,
    borderRadius: 15,
    justifyContent: "center",
    alignItems: "center",
  },
  calendarButtonOutline: {
    height: "auto",
    backgroundColor: Colors.background,
    borderRadius: 15,
    justifyContent: "center",
    alignItems: "center",
    boxShadow: "0px 0px 10px rgba(0, 0, 0, 0.15)",
  },
  calendarButtonText: {
    color: Colors.background,
  },
  calendarButtonTextLight: {
    color: Colors.muted,
  },
  calendarButtonMoreText: {
    color: Colors.muted,
    fontSize: 12,
    marginTop: 0,
  },
  todayButton: {
    height: "auto",
    backgroundColor: Colors.background,
    borderRadius: 15,
    justifyContent: "center",
    alignItems: "center",
    borderWidth: 2,
    borderColor: Colors.text,
  },
  todayText: {
    color: Colors.text,
  },
  dot: {
    width: 6,
    height: 6,
    borderRadius: 3,
    marginTop: 4,
  },
  modalContainer: {
    flex: 1,
    backgroundColor: Colors.background,
  },
  modalButtonContainer: {
    flexDirection: "row",
    justifyContent: "flex-end",
    paddingHorizontal: 16,
  },
  modalButton: {
    height: 48,
    width: 48,
  },
  modalContent: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
  },
  calendarList: {
    width: "100%",
    height: 600,
  },
});
