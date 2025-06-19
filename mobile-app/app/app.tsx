import "./utils/gestureHandler"
import { initI18n } from "./i18n"
import { useFonts } from "expo-font"
import { useEffect, useState } from "react"
import { initialWindowMetrics, SafeAreaProvider } from "react-native-safe-area-context"
import * as Linking from "expo-linking"
import { AppNavigator, useNavigationPersistence } from "./navigators"
import * as storage from "./utils/storage"
import { customFontsToLoad } from "./theme"
import { KeyboardProvider } from "react-native-keyboard-controller"
import * as ScreenOrientation from "expo-screen-orientation"
if (__DEV__) {
  // Load Reactotron in development only.
  require("./devtools/ReactotronConfig.ts")
}

export const NAVIGATION_PERSISTENCE_KEY = "NAVIGATION_STATE"

const prefix = Linking.createURL("/")
const config = {
  screens: {},
}

export function App() {
  const {
    initialNavigationState,
    onNavigationStateChange,
    isRestored: isNavigationStateRestored,
  } = useNavigationPersistence(storage, NAVIGATION_PERSISTENCE_KEY)

  const [areFontsLoaded, fontLoadError] = useFonts(customFontsToLoad)
  const [isI18nInitialized, setIsI18nInitialized] = useState(false)

  useEffect(() => {
    initI18n().then(() => setIsI18nInitialized(true))
  }, [])

  useEffect(() => {
    ScreenOrientation.lockAsync(ScreenOrientation.OrientationLock.LANDSCAPE)
  }, [])

  if (!isNavigationStateRestored || !isI18nInitialized || (!areFontsLoaded && !fontLoadError)) {
    return null
  }

  const linking = {
    prefixes: [prefix],
    config,
  }

  return (
    <SafeAreaProvider initialMetrics={initialWindowMetrics}>
      <KeyboardProvider>
        <AppNavigator
          linking={linking}
          initialState={initialNavigationState}
          onStateChange={onNavigationStateChange}
        />
      </KeyboardProvider>
    </SafeAreaProvider>
  )
}
