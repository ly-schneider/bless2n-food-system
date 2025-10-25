# Bless2n Android App

# Keystore

```bash
keytool -genkeypair -v -keystore bfs-android-upload.jks -storepass "<STORE-PASSWORD>" -alias bfsUpload -keypass "<ALIAS-PASSWORD>" -keyalg RSA -keysize 4096 -validity 365 -storetype JKS -dname "CN=BFS Android,OU=Mobile Development,O=Bless2n Food System"
```

Change to PKCS12

```bash
keytool -importkeystore -srckeystore bfs-android-upload.jks -destkeystore bfs-android-upload.jks -deststoretype pkcs12
```

# Base64

```bash
base64 -i bfs-android-upload.jks | pbcopy
```

# ADB Forwarding Port 3000

```bash
adb forward tcp:3000 tcp:3000
```