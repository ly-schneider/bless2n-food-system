Setup (WebView POS Wrapper)

`env $(grep -v '^#' secrets.properties | xargs) ./gradlew :app:installDev --continuous`
`adb logcat --pid=$(adb shell pidof -s ch.leys.bless2n.dev) -v time`