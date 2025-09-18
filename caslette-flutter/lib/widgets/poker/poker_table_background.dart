import 'package:flutter/material.dart';

/// A responsive poker table background that adapts to different screen sizes and orientations
class PokerTableBackground extends StatelessWidget {
  final Widget child;
  final bool useGradientFallback;

  const PokerTableBackground({
    super.key,
    required this.child,
    this.useGradientFallback = false,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      height: double.infinity,
      decoration: _buildBackgroundDecoration(context),
      child: child,
    );
  }

  BoxDecoration _buildBackgroundDecoration(BuildContext context) {
    final screenSize = MediaQuery.of(context).size;

    if (useGradientFallback) {
      return _buildGradientBackground();
    }

    // For now, use improved image handling
    return BoxDecoration(
      image: DecorationImage(
        image: const AssetImage('assets/images/poker_table_bg.jpg'),
        fit: _getOptimalFit(screenSize),
        alignment: Alignment.center,
        repeat: ImageRepeat.noRepeat,
        // Use a color filter to ensure good contrast with UI elements
        colorFilter: ColorFilter.mode(
          Colors.black.withOpacity(0.1),
          BlendMode.darken,
        ),
      ),
    );
  }

  /// Determine the best BoxFit for landscape poker tables
  BoxFit _getOptimalFit(Size screenSize) {
    // Always use cover to fill the entire screen and avoid black edges
    // This ensures the background completely fills any screen size
    return BoxFit.cover;
  }

  /// Mobile-optimized gradient background that adapts to any screen size
  BoxDecoration _buildGradientBackground() {
    return BoxDecoration(
      gradient: RadialGradient(
        center: Alignment.center,
        radius: 1.0, // Tighter radius for mobile
        colors: [
          const Color(
            0xFF1F7A44,
          ), // Lighter green center (more visible on mobile)
          const Color(0xFF0F5132), // Standard poker green
          const Color(0xFF0A3D26), // Darker green
          const Color(0xFF052E1B), // Very dark green (edges)
        ],
        stops: const [0.0, 0.3, 0.7, 1.0],
      ),
    );
  }
}

/// A mobile-optimized poker table background with subtle table indication
class MobilePokerBackground extends StatelessWidget {
  final Widget child;
  final bool showTableOutline;

  const MobilePokerBackground({
    super.key,
    required this.child,
    this.showTableOutline = true,
  });

  @override
  Widget build(BuildContext context) {
    final screenSize = MediaQuery.of(context).size;
    final isLandscape = screenSize.width > screenSize.height;

    return Container(
      width: double.infinity,
      height: double.infinity,
      decoration: BoxDecoration(
        gradient: RadialGradient(
          center: Alignment.center,
          radius: isLandscape ? 0.8 : 1.2, // Adapt to orientation
          colors: [
            const Color(0xFF1F7A44), // Brighter center for mobile visibility
            const Color(0xFF0F5132), // Standard poker green
            const Color(0xFF0A3D26), // Darker edges
            const Color(0xFF000000), // Black background
          ],
          stops: const [0.0, 0.4, 0.8, 1.0],
        ),
      ),
      child: showTableOutline ? _buildWithTableOutline(context) : child,
    );
  }

  Widget _buildWithTableOutline(BuildContext context) {
    final screenSize = MediaQuery.of(context).size;
    final isLandscape = screenSize.width > screenSize.height;

    // Calculate responsive table area
    final tableWidth = screenSize.width * (isLandscape ? 0.7 : 0.85);
    final tableHeight = screenSize.height * (isLandscape ? 0.8 : 0.6);

    return Stack(
      children: [
        // Subtle table outline (optional visual guide)
        Center(
          child: Container(
            width: tableWidth,
            height: tableHeight,
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(tableHeight / 2.5),
              border: Border.all(
                color: const Color(0xFF2D5A3D).withOpacity(0.3),
                width: 2,
              ),
              // Very subtle inner glow
              boxShadow: [
                BoxShadow(
                  color: const Color(0xFF1F7A44).withOpacity(0.1),
                  blurRadius: 20,
                  spreadRadius: -5,
                ),
              ],
            ),
          ),
        ),
        // Game content
        child,
      ],
    );
  }
}

/// A programmatic poker table background with felt texture
class ProgrammaticPokerBackground extends StatelessWidget {
  final Widget child;
  final Color primaryColor;
  final Color accentColor;

  const ProgrammaticPokerBackground({
    super.key,
    required this.child,
    this.primaryColor = const Color(0xFF0F5132),
    this.accentColor = const Color(0xFF1F7A44),
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      height: double.infinity,
      decoration: BoxDecoration(
        gradient: RadialGradient(
          center: Alignment.center,
          radius: 1.5,
          colors: [
            primaryColor.withOpacity(0.9),
            primaryColor,
            primaryColor.withOpacity(0.8),
            Colors.black.withOpacity(0.9),
          ],
          stops: const [0.0, 0.3, 0.7, 1.0],
        ),
      ),
      child: CustomPaint(
        painter: FeltTexturePainter(
          primaryColor: primaryColor,
          accentColor: accentColor,
        ),
        child: child,
      ),
    );
  }
}

/// Custom painter for creating a felt-like texture programmatically
class FeltTexturePainter extends CustomPainter {
  final Color primaryColor;
  final Color accentColor;

  FeltTexturePainter({required this.primaryColor, required this.accentColor});

  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..style = PaintingStyle.fill
      ..strokeWidth = 1.0;

    // Create a subtle texture pattern
    const spacing = 8.0;
    final rows = (size.height / spacing).ceil();
    final cols = (size.width / spacing).ceil();

    for (int row = 0; row < rows; row++) {
      for (int col = 0; col < cols; col++) {
        final x = col * spacing + (row % 2) * (spacing / 2);
        final y = row * spacing;

        if (x < size.width && y < size.height) {
          // Vary the opacity based on position for organic look
          final opacity = 0.05 + (((row + col) % 3) * 0.02);
          paint.color = accentColor.withOpacity(opacity);

          canvas.drawCircle(Offset(x, y), 0.5 + ((row + col) % 3) * 0.3, paint);
        }
      }
    }

    // Add some subtle diagonal lines for felt effect
    paint.color = accentColor.withOpacity(0.03);
    paint.strokeWidth = 0.5;

    for (double i = -size.height; i < size.width + size.height; i += 15) {
      canvas.drawLine(
        Offset(i, 0),
        Offset(i + size.height, size.height),
        paint,
      );
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}
