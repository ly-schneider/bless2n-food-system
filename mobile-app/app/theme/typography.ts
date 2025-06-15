export const customFontsToLoad = {
  mozaicHum: require("../../assets/fonts/Mozaic-HUM-Variable.ttf"),
}

const fonts = {
  mozaicHum: {
    regular: "mozaicHum",
  },
}

export const typography = {
  fonts,
  primary: fonts.mozaicHum,
  heading: { fontFamily: "mozaicHum", fontWeight: "700" },
  bold: { fontFamily: "mozaicHum", fontWeight: "700" },
}
