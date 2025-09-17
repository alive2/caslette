# Caslette Casino - Multi-Platform Build Guide

## Build Targets

### Web (✅ Working)

```bash
flutter build web
```

- Output: `build/web/`
- Ready for deployment to any web server
- Optimized with tree-shaking
- Production ready

### Windows Desktop (⚠️ Requires Developer Mode)

```bash
# Enable Developer Mode first:
start ms-settings:developers

# Then build:
flutter build windows
```

- Output: `build/windows/x64/runner/Release/`
- Native Windows application
- Full desktop experience

### Android (📱 Mobile Ready)

```bash
flutter build apk
# or for app bundle:
flutter build appbundle
```

- Requires Android SDK setup
- Cross-platform mobile deployment

### iOS (🍎 Mobile Ready)

```bash
flutter build ios
```

- Requires macOS and Xcode
- App Store deployment ready

## Development Commands

### Run on Different Platforms

```bash
# Web development
flutter run -d chrome

# Windows desktop
flutter run -d windows

# Android emulator
flutter run -d android

# iOS simulator (macOS only)
flutter run -d ios
```

## Deployment

### Web Deployment

1. Build: `flutter build web`
2. Deploy `build/web/` to any static hosting:
   - Netlify
   - Vercel
   - Firebase Hosting
   - GitHub Pages
   - Your own server

### Desktop Deployment

1. Build for target platform
2. Distribute the executable from build folder
3. Consider using packaging tools for installers

### Mobile Deployment

1. Build APK/iOS app
2. Upload to Google Play Store / Apple App Store
3. Follow platform-specific guidelines

## Configuration Notes

- All platforms share the same Dart/Flutter codebase
- WebSocket connections work across all platforms
- UI automatically adapts to platform conventions
- State management (Riverpod) works consistently
- API and backend integration identical on all platforms
