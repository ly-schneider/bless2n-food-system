import { MatchesForm } from "@/components/MatchesForm";
import { Colors } from "@/constants/Colors";
import { useCallback, useState } from "react";
import { RefreshControl, ScrollView } from "react-native";

export default function MatchCreate() {
  const [refreshing, setRefreshing] = useState(false);

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await new Promise((resolve) => setTimeout(resolve, 2000));
    setRefreshing(false);
  }, []);

  return (
    <ScrollView
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
      {!refreshing && <MatchesForm />}
    </ScrollView>
  );
}