import { NativeModules } from "react-native"

import { ArgType } from "reactotron-core-client"
import { mst } from "reactotron-mst"
import mmkvPlugin from "reactotron-react-native-mmkv"

import { storage, clear } from "@/utils/storage"
import { goBack, resetRoot, navigate } from "@/navigators/navigationUtilities"

import { Reactotron } from "./ReactotronClient"
import { ReactotronReactNative } from "reactotron-react-native"

const reactotron = Reactotron.configure({
  name: require("../../package.json").name,
  onConnect: () => {
    Reactotron.clear()
  },
})

reactotron.use(
  mst({
    filter: (event) => /postProcessSnapshot|@APPLY_SNAPSHOT/.test(event.name) === false,
  }),
)

reactotron.use(mmkvPlugin<ReactotronReactNative>({ storage }))

reactotron.useReactNative({
  networking: {
    ignoreUrls: /symbolicate/,
  },
})

reactotron.onCustomCommand({
  title: "Show Dev Menu",
  description: "Opens the React Native dev menu",
  command: "showDevMenu",
  handler: () => {
    Reactotron.log("Showing React Native dev menu")
    NativeModules.DevMenu.show()
  },
})

reactotron.onCustomCommand({
  title: "Reset Root Store",
  description: "Resets the MST store",
  command: "resetStore",
  handler: () => {
    Reactotron.log("resetting store")
    clear()
  },
})

reactotron.onCustomCommand({
  title: "Reset Navigation State",
  description: "Resets the navigation state",
  command: "resetNavigation",
  handler: () => {
    Reactotron.log("resetting navigation state")
    resetRoot({ index: 0, routes: [] })
  },
})

reactotron.onCustomCommand<[{ name: "route"; type: ArgType.String }]>({
  command: "navigateTo",
  handler: (args) => {
    const { route } = args ?? {}
    if (route) {
      Reactotron.log(`Navigating to: ${route}`)
      navigate(route as any) // this should be tied to the navigator, but since this is for debugging, we can navigate to illegal routes
    } else {
      Reactotron.log("Could not navigate. No route provided.")
    }
  },
  title: "Navigate To Screen",
  description: "Navigates to a screen by name.",
  args: [{ name: "route", type: ArgType.String }],
})

reactotron.onCustomCommand({
  title: "Go Back",
  description: "Goes back",
  command: "goBack",
  handler: () => {
    Reactotron.log("Going back")
    goBack()
  },
})

/**
 * Add `console.tron` to the Reactotron object.
 *
 * @example
 * if (__DEV__) {
 *  console.tron.display({
 *    name: 'Error',
 *    preview: 'An error occurred',
 *    value: 'Something went wrong!',
 *    important: true
 *  })
 * }
 */
console.tron = reactotron

declare global {
  interface Console {
    /**
     * Reactotron client for logging, displaying, measuring performance, and more.
     * @see https://github.com/infinitered/reactotron
     * @example
     * if (__DEV__) {
     *  console.tron.display({
     *    name: 'Error',
     *    preview: 'An error occurred',
     *    value: 'Something went wrong!',
     *    important: true
     *  })
     * }
     */
    tron: typeof reactotron
  }
}

reactotron.connect()
