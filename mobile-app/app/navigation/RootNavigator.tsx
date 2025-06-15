import { createNativeStackNavigator } from "@react-navigation/native-stack"

export type RootStackParamList = {
  home: undefined
}

export const RootNavigator = createNativeStackNavigator<RootStackParamList>()
