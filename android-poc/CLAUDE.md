# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
This is an Android Point-of-Sale (POS) terminal proof-of-concept built with Kotlin and Jetpack Compose. The app integrates with SumUp payment SDK for card payments and includes thermal receipt printing functionality via Bluetooth.

## Development Commands

### Build and Run
```bash
# Build the project
./gradlew build

# Build debug APK
./gradlew assembleDebug

# Install and run on connected device/emulator
./gradlew installDebug

# Run tests
./gradlew test

# Run instrumented tests (requires device/emulator)
./gradlew connectedAndroidTest
```

### Gradle Tasks
```bash
# Clean build artifacts
./gradlew clean

# Check for dependency updates
./gradlew dependencyUpdates

# Generate lint report
./gradlew lint
```

## Architecture

### Core Components
- **MainActivity**: Main UI controller handling SumUp payment flow and thermal printing
- **ShopApplication**: Application class that initializes SumUp SDK
- **PaymentScreen**: Composable UI for payment interface
- **EnsureBtPermission**: Composable for Bluetooth permission handling

### Key Integrations
- **SumUp Merchant SDK**: Card payment processing (affiliate key in BuildConfig)
- **ESCPOS Thermal Printer**: Bluetooth thermal receipt printing with custom ESC/POS commands
- **Jetpack Compose**: Modern Android UI toolkit

### Payment Flow
1. User taps "Pay with card" 
2. If not logged in, SumUp login flow is triggered
3. Payment amount (CHF 1.00) is processed via SumUp
4. Receipt data is stored in activity state
5. User can print receipt via Bluetooth thermal printer

### Printing Implementation
- Direct Bluetooth connection to thermal printer (MAC: B0:B0:09:43:96:16)
- Custom ESC/POS raster printing implementation
- 384-pixel wide bitmap generation from text
- Phomemo-specific printer commands and timing

## Configuration
- **Package**: ch.leys.bless2n
- **Min SDK**: 34 (Android 14)
- **Target SDK**: 35
- **Compile SDK**: 35
- **Java Version**: 11

## Permissions
The app requires extensive permissions for payment and printing:
- Bluetooth (BLUETOOTH, BLUETOOTH_CONNECT, BLUETOOTH_SCAN)
- Location (ACCESS_COARSE_LOCATION, ACCESS_FINE_LOCATION) - required for Bluetooth discovery
- Network (INTERNET, ACCESS_WIFI_STATE, CHANGE_WIFI_STATE)
- NFC - for potential card reader integration

## Testing
- Unit tests: `./gradlew test`
- Instrumented tests: `./gradlew connectedAndroidTest`
- Test files located in `app/src/test/` and `app/src/androidTest/`