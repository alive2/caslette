#!/bin/bash

# Caslette Casino - Web Deployment Script

echo "🎰 Building Caslette Casino for Web..."

# Clean previous builds
echo "🧹 Cleaning previous builds..."
flutter clean
flutter pub get

# Build for web production
echo "🌐 Building for web..."
flutter build web --release

echo "✅ Build complete!"
echo "📁 Output location: build/web/"
echo ""
echo "🚀 To serve locally for testing:"
echo "   cd build/web && python -m http.server 8000"
echo ""
echo "📤 To deploy:"
echo "   - Upload build/web/ contents to your web server"
echo "   - Or use: firebase deploy (if using Firebase Hosting)"
echo "   - Or use: netlify deploy --prod --dir=build/web"