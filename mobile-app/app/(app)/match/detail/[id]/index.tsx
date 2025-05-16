import {
  Modal,
  RefreshControl,
  ScrollView,
  StyleSheet,
  View,
} from "react-native";
import { AnimatedHapticButton } from "@/components/AnimatedHapticButton";
import { Ionicons } from "@expo/vector-icons";
import { Colors } from "@/constants/Colors";
import { router, useFocusEffect, useLocalSearchParams } from "expo-router";
import { ThemedText } from "@/components/ThemedText";
import { useCallback, useState } from "react";
import { clubFetch, matchFetch } from "@/api/Api";
import { SafeAreaProvider, SafeAreaView } from "react-native-safe-area-context";
import { MatchesForm } from "@/components/MatchesForm";
import { Match } from "@/constants/Match";
import ParticipantsContainer from "@/components/ParticipantsContainer";

export default function MatchDetail() {
  const { id } = useLocalSearchParams<{ id: string }>();

  const [match, setMatch] = useState<Match>();
  const [isAdmin, setIsAdmin] = useState(false);
  const [showModalEditMatch, setShowModalEditMatch] = useState(false);
  const [refreshing, setRefreshing] = useState(false);

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await fetchMatch();
    setRefreshing(false);
  }, []);

  useFocusEffect(
    useCallback(() => {
      fetchMatch();

      return () => {
        setMatch(undefined);
      };
    }, [])
  );

  async function fetchMatch() {
    const { success, data } = await matchFetch(id);
    if (success) {
      setMatch(data);
      fetchClub(data.club.code);
    }
  }

  async function fetchClub(clubCode: string) {
    const { success, data } = await clubFetch(clubCode);
    if (success) {
      if (data.role === "admin") {
        setIsAdmin(true);
      }
    }
  }

  const handleBack = () => {
    router.back();
  };

  const handleEdit = () => {
    setShowModalEditMatch(true);
  };

  const formatReadableDateTime = (isoString: Date) => {
    const date = new Date(isoString);

    const formattedDate = new Intl.DateTimeFormat("de-DE", {
      weekday: "long",
      day: "numeric",
      month: "short",
    }).format(date);

    const formattedTime =
      new Intl.DateTimeFormat("de-DE", {
        hour: "2-digit",
        minute: "2-digit",
      }).format(date) + " Uhr";

    return { formattedDate, formattedTime };
  };

  if (!match) {
    return <View />;
  }

  return (
    <ScrollView
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
      <View style={styles.container}>
        <View style={styles.header}>
          <View style={styles.headerButtons}>
            <AnimatedHapticButton
              style={styles.buttonOutline}
              onPress={handleBack}
            >
              <Ionicons name="arrow-back" size={20} color={Colors.text} />
            </AnimatedHapticButton>
            {isAdmin && (
              <AnimatedHapticButton
                style={styles.buttonOutline}
                onPress={handleEdit}
              >
                <Ionicons name="pencil-outline" size={20} color={Colors.text} />
              </AnimatedHapticButton>
            )}
          </View>
          <View style={styles.headerTitleContainer}>
            <ThemedText type="title" style={styles.headerTitle}>
              {match.club.name} -{" "}
            </ThemedText>
            <ThemedText type="title" style={styles.headerTitle}>
              {match.enemyClub.name}
            </ThemedText>
          </View>
          <View style={styles.headerTextContainer}>
            <ThemedText style={styles.headerText}>
              {formatReadableDateTime(match.date as Date).formattedDate}
            </ThemedText>
            <ThemedText style={styles.headerText}>&bull;</ThemedText>
            <ThemedText style={styles.headerText}>
              {formatReadableDateTime(match.date as Date).formattedTime}
            </ThemedText>
            <ThemedText style={styles.headerText}>&bull;</ThemedText>
            <View
              style={{ flexDirection: "row", alignItems: "center", gap: 2 }}
            >
              <Ionicons name="pin-outline" size={16} color={Colors.text} />
              <ThemedText style={styles.headerText}>
                {match.isHomeGame ? "Heim" : "Auswärts"}
              </ThemedText>
            </View>
          </View>
        </View>
        <View style={styles.mainContainer}>
          {match.enemyClub.address && (
            <View>
              <ThemedText type="title" style={styles.mainTitle}>
                Adresse
              </ThemedText>
              <ThemedText>{match.enemyClub.address}</ThemedText>
            </View>
          )}
          <ParticipantsContainer matchData={match} club={match.club} isAdmin={isAdmin} />
        </View>
      </View>
      <Modal
        animationType="slide"
        transparent={false}
        visible={showModalEditMatch}
        onRequestClose={() => setShowModalEditMatch(false)}
      >
        <SafeAreaProvider>
          <SafeAreaView style={styles.modalContainer}>
            <MatchesForm
              matchData={match}
              fetchMatch={fetchMatch}
              setShowModalEditMatch={setShowModalEditMatch}
            />
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
    paddingVertical: 16,
    flex: 1,
    flexDirection: "column",
    alignItems: "flex-start",
  },
  headerButtons: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    width: "100%",
  },
  headerTitleContainer: {
    flexDirection: "row",
    alignItems: "center",
    gap: 0,
    flexWrap: "wrap",
    marginTop: 24,
  },
  headerTitle: {
    fontSize: 32,
  },
  headerTextContainer: {
    flexDirection: "row",
    flexWrap: "wrap",
    gap: 10,
    marginTop: 2,
  },
  headerText: {
    color: Colors.text,
    fontSize: 16,
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
  mainContainer: {
    marginTop: 24,
    flexDirection: "column",
    gap: 28,
  },
  mainTitle: {
    fontSize: 24,
  },
  modalContainer: {
    flex: 1,
    backgroundColor: Colors.background,
  },
});
