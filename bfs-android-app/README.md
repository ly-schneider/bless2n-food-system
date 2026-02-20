# BFS Android App

Android POS terminal application for the Bless2n Food System — provides on-site payment processing via SumUp card terminals, QR code scanning, and thermal receipt printing.

## Architecture

The app is a **Kotlin + Jetpack Compose** application that wraps the web-based POS interface in a WebView, extending it with native Android capabilities for hardware integration.

### Key Design Decisions

- **WebView-based POS** — the core POS UI comes from the Next.js web app (`/pos`), keeping business logic centralized
- **SumUp SDK** for card terminal payments — connects to physical SumUp readers
- **ZXing** for barcode/QR code scanning — reads order codes and product barcodes
- **ESCPOS thermal printing** — prints receipts on compatible thermal printers
- **Multi-build-type** configuration — debug, dev, staging, and release variants with environment-specific URLs

## Tech Stack

| Category | Technology |
|----------|-----------|
| Language | Kotlin |
| UI | Jetpack Compose (Material 3) |
| Payments | SumUp Merchant SDK 6.0 |
| QR / Barcode | ZXing Core 3.5 + Android Embedded 4.3 |
| Printing | ESCPOS-ThermalPrinter-Android 3.4 |
| WebView | AndroidX WebKit 1.15 |
| Min SDK | 26 (Android 8.0) |
| Target SDK | 35 |
| Compile SDK | 36 |

## Project Structure

```
bfs-android-app/
├── app/
│   ├── src/
│   │   ├── main/               Main source code & resources
│   │   │   ├── java/           Kotlin code
│   │   │   ├── res/            Layouts, drawables, strings
│   │   │   └── AndroidManifest.xml
│   │   ├── dev/                Dev-specific resources
│   │   ├── test/               Unit tests
│   │   └── androidTest/        Instrumentation tests
│   ├── build.gradle.kts        App configuration
│   └── proguard-rules.pro      R8 obfuscation rules
├── build.gradle.kts            Project configuration
├── gradle/                     Gradle wrapper
├── settings.gradle.kts
└── dist/                       Distribution APK/AAB files
```

## Build Variants

| Variant | App Name | POS URL | Debuggable |
|---------|----------|---------|------------|
| **debug** | BlessThun Food (Debug) | http://localhost:3000/pos | Yes |
| **dev** | BlessThun Food (Dev) | http://localhost:3000/pos | Yes |
| **staging** | BlessThun Food (Staging) | Configured via build | No |
| **release** | BlessThun Food | Configured via `POS_URL` | No |

## Prerequisites

- **Android Studio** (latest stable)
- **JDK 11+**
- Physical or emulated Android device (SDK 26+)
- **SumUp developer account** (for payment integration)

## Development Setup

```bash
# Open in Android Studio
# Sync Gradle
# Select "debug" build variant
# Run on device/emulator
```

### ADB Port Forwarding

When developing against a local backend, forward the required ports to the Android device:

```bash
adb reverse tcp:3000 tcp:3000     # Next.js web app
adb reverse tcp:8080 tcp:8080     # Go backend API
adb reverse tcp:10000 tcp:10000   # Azurite blob storage
```

## Versioning

Version codes are automatically derived from git tags:

```
Tag: v2.1.2
Version Name: 2.1.2
Version Code: 20102  (major×10000 + minor×100 + patch)
```

## Release Signing

Release builds require a PKCS12 keystore configured via environment variables:

| Variable | Purpose |
|----------|---------|
| `BFS_UPLOAD_STORE_FILE` | Path to keystore file |
| `BFS_UPLOAD_STORE_PASSWORD` | Keystore password |
| `BFS_UPLOAD_KEY_ALIAS` | Key alias |
| `BFS_UPLOAD_KEY_PASSWORD` | Key password |
| `SUMUP_AFFILIATE_KEY` | SumUp affiliate key (required for release) |
| `POS_URL` | Production POS URL (required for release) |

### Generating a Keystore

```bash
keytool -genkeypair -v \
  -keystore bfs-android-upload.jks \
  -storepass "<password>" \
  -alias bfsUpload \
  -keypass "<password>" \
  -keyalg RSA -keysize 4096 \
  -validity 365 -storetype JKS \
  -dname "CN=BFS Android,OU=Mobile Development,O=Bless2n Food System"

# Convert to PKCS12
keytool -importkeystore \
  -srckeystore bfs-android-upload.jks \
  -destkeystore bfs-android-upload.jks \
  -deststoretype pkcs12

# Base64 encode for CI secrets
base64 -i bfs-android-upload.jks | pbcopy
```

## Hardware Integration

- **SumUp Card Reader** — Bluetooth-connected card terminal for in-person payments
- **Thermal Printer** — ESCPOS-compatible printers for receipt printing
- **Camera** — QR code and barcode scanning via ZXing
