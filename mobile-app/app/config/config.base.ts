export interface ConfigBaseProps {
  persistNavigation: "always" | "dev" | "prod" | "never"
  catchErrors: "always" | "dev" | "prod" | "never"
  exitRoutes: string[]
}

export type PersistNavigationConfig = ConfigBaseProps["persistNavigation"]

const BaseConfig: ConfigBaseProps = {
  persistNavigation: "never",
  catchErrors: "always",

  /**
   * List of all the route names that will exit the app if the back button is pressed. Only affects Android.
   */
  exitRoutes: ["Home"],
}

export default BaseConfig
