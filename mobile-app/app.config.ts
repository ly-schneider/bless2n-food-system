import { ExpoConfig, ConfigContext } from "@expo/config"

require("ts-node/register")

module.exports = ({ config }: ConfigContext): Partial<ExpoConfig> => {
  const existingPlugins = config.plugins ?? []

  return {
    ...config,
    plugins: [...existingPlugins, require("./plugins/withSplashScreen").withSplashScreen],
  }
}
