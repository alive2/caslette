# Caslette Casino - Web Deployment Script (PowerShell)

Write-Host "🎰 Building Caslette Casino for Web..." -ForegroundColor Cyan

# Clean previous builds
Write-Host "🧹 Cleaning previous builds..." -ForegroundColor Yellow
flutter clean
flutter pub get

# Build for web production
Write-Host "🌐 Building for web..." -ForegroundColor Green
flutter build web --release

Write-Host "✅ Build complete!" -ForegroundColor Green
Write-Host "📁 Output location: build/web/" -ForegroundColor White
Write-Host ""
Write-Host "🚀 To serve locally for testing:" -ForegroundColor Cyan
Write-Host "   cd build/web && python -m http.server 8000" -ForegroundColor White
Write-Host ""
Write-Host "📤 To deploy:" -ForegroundColor Cyan
Write-Host "   - Upload build/web/ contents to your web server" -ForegroundColor White
Write-Host "   - Or use: firebase deploy (if using Firebase Hosting)" -ForegroundColor White
Write-Host "   - Or use: netlify deploy --prod --dir=build/web" -ForegroundColor White