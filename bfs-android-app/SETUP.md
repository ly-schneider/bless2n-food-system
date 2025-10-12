Setup (WebView POS Wrapper)

Prereqs
- Android Studio (latest), SDK 35 installed
- SumUp Affiliate Key available as env/Gradle property
- Optional: paired Bluetooth ESC/POS printer

Config
- CI-friendly inputs (recommended):
  - Env vars: `SUMUP_AFFILIATE_KEY`, `POS_URL_RELEASE`
  - Or pass Gradle props: `-PsumupAffiliateKey=... -PposUrlRelease=...`
- Local development: you can also export env vars before building.

Run (dev)
- Start the web app locally: `cd bfs-web-app && pnpm dev` (port 3000)
- Run the Android `dev` variant. Choose one of:
  1) Emulator: use default `POS_URL=http://10.0.2.2:3000/pos` (works out of the box).
  2) Physical device via USB + adb reverse:
     - `adb reverse tcp:3000 tcp:3000`
     - Build with `-PposUrlDev=http://127.0.0.1:3000/pos` (or set `POS_URL_DEV` env)
  3) Physical device on same Wi‑Fi (no adb reverse):
     - Find your host IP (e.g., `192.168.1.23`)
     - Build with `-PposUrlDev=http://192.168.1.23:3000/pos`
- Cleartext HTTP is allowed for the dev build (network security config is wide‑open in dev only).

Permissions
- The app requests Bluetooth permissions on-demand for ESC/POS printing on Android 12+.
- Camera permissions are used by the existing demo activity and are not required for the WebView POS.

Printing
- System mode: prints HTML via Android PrintManager (available if called by the page via BFS shim).
- ESC/POS mode: uses the DantSu library and selects the first paired Bluetooth printer by default.
  - For reliable printing, pair the target printer in Android settings first.

SumUp
- The bridge triggers SumUp login if necessary, then opens checkout with the amount and currency provided by the page.
- Results are posted back to the page via a `bfs:sumup:result` CustomEvent.

CI/CD
- A GitHub Actions workflow is included at `.github/workflows/android.yml`:
  - Reads `SUMUP_AFFILIATE_KEY` and `POS_URL_RELEASE` from repository secrets.
  - Runs `clean test assembleDev assembleRelease` and uploads APK artifacts.
- For other CI systems, export the same env variables or pass Gradle properties to the build step.
- Release signing: the current workflow exports unsigned release APKs. If you want signing, we can add a signingConfig that reads keystore secrets from CI (keystore base64 + passwords) and produce signed artifacts.
