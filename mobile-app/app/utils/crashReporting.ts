/**
 * TODO: Implement crashlytics: https://rnfirebase.io/crashlytics/usage
 * import crashlytics from "@react-native-firebase/crashlytics"
 */

/**
 *  Initialization code to call in `./app/app.tsx`
 */
export const initCrashReporting = () => {}

/**
 * Error classifications used to sort errors on error reporting services.
 */
export enum ErrorType {
  /**
   * An error that would normally cause a red screen in dev
   * and force the user to sign out and restart.
   */
  FATAL = "Fatal",
  /**
   * An error caught by try/catch where defined using Reactotron.tron.error.
   */
  HANDLED = "Handled",
}

/**
 * Manually report a handled error.
 */
export const reportCrash = (error: Error, type: ErrorType = ErrorType.FATAL) => {
  if (__DEV__) {
    // Log to console and Reactotron in development
    const message = error.message || "Unknown"
    console.error(error)
    console.log(message, type)
  } else {
    // In production, utilize crash reporting service of choice below:
    // crashlytics().recordError(error)
  }
}
