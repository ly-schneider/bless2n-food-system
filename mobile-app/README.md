# Rentro POS Mobile App

## Install App on Android Device

### 1 Prepare the tablet
#### 2.1 Allow “unknown” installs

Android blocks manual installs by default. On Android 12+:

1. Settings ▶ Apps ▶ Special app access ▶ Install unknown apps.
2. Grant your chosen installer (Files, Chrome, Drive, etc.) permission.
3. Keep Google Play Protect scanning switched on so sideloaded apps are scanned for malware.

#### 2.2 (ADB) enable debugging

1. Settings ▶ About tablet ▶ Build number (tap 7 times).
2. Developer options ▶ USB debugging.

### 2 Install the APK

Instant ADB sideload

```bash
# Plug the tablet in (or adb connect <IP>:5555 for wireless)
adb devices        # should list one “device”
adb install -r path/to/my-app-debug.apk
```
- -r lets you overwrite an older build in place
- Success looks like Success in the terminal; the app icon appears on the tablet.