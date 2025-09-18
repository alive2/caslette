import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/auth_provider.dart';

class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});

  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen> {
  final _usernameController = TextEditingController();
  final _passwordController = TextEditingController();
  String _errorMessage = '';
  bool _isLoading = false;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFF0F172A), // Slate-900
      body: Container(
        decoration: const BoxDecoration(
          gradient: RadialGradient(
            center: Alignment.topRight,
            radius: 1.5,
            colors: [
              Color(0xFF4C1D95), // Purple-800
              Color(0xFF0F172A), // Slate-900
            ],
          ),
        ),
        child: OrientationBuilder(
          builder: (context, orientation) {
            if (orientation == Orientation.landscape) {
              return _buildLandscapeLayout();
            } else {
              return _buildPortraitLayout();
            }
          },
        ),
      ),
    );
  }

  Widget _buildPortraitLayout() {
    return Center(
      child: SingleChildScrollView(
        padding: const EdgeInsets.all(32.0),
        child: _buildLoginCard(isLandscape: false),
      ),
    );
  }

  Widget _buildLandscapeLayout() {
    return Row(
      children: [
        // Left side - Logo and branding
        Expanded(
          flex: 1,
          child: Container(
            padding: const EdgeInsets.all(40),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                _buildLogo(isLandscape: true),
                const SizedBox(height: 24),
                const Text(
                  'Welcome to the Cosmic Casino',
                  style: TextStyle(
                    fontSize: 24,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                  ),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 12),
                Text(
                  'Experience the ultimate poker adventure in the stars',
                  style: TextStyle(
                    fontSize: 16,
                    color: Colors.white.withOpacity(0.8),
                  ),
                  textAlign: TextAlign.center,
                ),
              ],
            ),
          ),
        ),
        // Right side - Login form
        Expanded(
          flex: 1,
          child: Center(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(32.0),
              child: _buildLoginCard(isLandscape: true),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildLogo({bool isLandscape = false}) {
    return Container(
      padding: EdgeInsets.all(isLandscape ? 12 : 16),
      decoration: const BoxDecoration(
        shape: BoxShape.circle,
        gradient: LinearGradient(
          colors: [
            Color(0xFF9333EA), // Purple-600
            Color(0xFFF59E0B), // Amber-500
          ],
        ),
      ),
      child: Icon(
        Icons.casino,
        size: isLandscape ? 48 : 64,
        color: Colors.white,
      ),
    );
  }

  Widget _buildLoginCard({bool isLandscape = false}) {
    return ConstrainedBox(
      constraints: BoxConstraints(maxWidth: isLandscape ? 400 : 500),
      child: Card(
        color: const Color(0xFF1F2937), // Gray-800
        elevation: 8,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        child: Padding(
          padding: EdgeInsets.all(isLandscape ? 24.0 : 32.0),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              if (!isLandscape) ...[
                // Logo/Title (only show in portrait, as landscape has it on the left)
                _buildLogo(),
                const SizedBox(height: 16),
                const Text(
                  'CASLETTE',
                  style: TextStyle(
                    fontSize: 32,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                    letterSpacing: 2,
                  ),
                ),
                const SizedBox(height: 8),
                Text(
                  'Cosmic Casino Experience',
                  style: TextStyle(
                    fontSize: 16,
                    color: Color(0xFF9CA3AF), // Gray-400
                  ),
                ),
                const SizedBox(height: 32),
              ] else ...[
                // Compact header for landscape
                const Text(
                  'Sign In',
                  style: TextStyle(
                    fontSize: 24,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                  ),
                ),
                const SizedBox(height: 20),
              ],

              _buildLoginForm(isLandscape: isLandscape),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildLoginForm({bool isLandscape = false}) {
    return Column(
      children: [
        // Username field
        TextField(
          controller: _usernameController,
          style: const TextStyle(color: Colors.white),
          decoration: InputDecoration(
            labelText: 'Username',
            labelStyle: const TextStyle(color: Color(0xFF9CA3AF)),
            prefixIcon: const Icon(Icons.person, color: Color(0xFF9333EA)),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: Color(0xFF374151)),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: Color(0xFF374151)),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: Color(0xFF9333EA)),
            ),
            filled: true,
            fillColor: const Color(0xFF374151),
          ),
        ),
        const SizedBox(height: 16),

        // Password field
        TextField(
          controller: _passwordController,
          obscureText: true,
          style: const TextStyle(color: Colors.white),
          decoration: InputDecoration(
            labelText: 'Password',
            labelStyle: const TextStyle(color: Color(0xFF9CA3AF)),
            prefixIcon: const Icon(Icons.lock, color: Color(0xFF9333EA)),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: Color(0xFF374151)),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: Color(0xFF374151)),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: Color(0xFF9333EA)),
            ),
            filled: true,
            fillColor: const Color(0xFF374151),
          ),
        ),

        if (_errorMessage.isNotEmpty) ...[
          const SizedBox(height: 16),
          Text(
            _errorMessage,
            style: const TextStyle(
              color: Color(0xFFEF4444), // Red-500
              fontSize: 14,
            ),
            textAlign: TextAlign.center,
          ),
        ],

        const SizedBox(height: 24),

        // Login button
        SizedBox(
          width: double.infinity,
          height: 48,
          child: ElevatedButton(
            onPressed: _isLoading
                ? null
                : () async {
                    final username = _usernameController.text.trim();
                    final password = _passwordController.text.trim();

                    if (username.isEmpty || password.isEmpty) {
                      if (mounted) {
                        setState(() {
                          _errorMessage =
                              'Please enter both username and password';
                        });
                      }
                      return;
                    }

                    if (mounted) {
                      setState(() {
                        _isLoading = true;
                        _errorMessage = '';
                      });
                    }

                    try {
                      final success = await ref
                          .read(authProvider.notifier)
                          .login(username, password);
                      if (!success && mounted) {
                        setState(() {
                          _errorMessage = 'Invalid username or password.';
                        });
                      }
                    } catch (e) {
                      if (mounted) {
                        setState(() {
                          _errorMessage = 'Login failed: ${e.toString()}';
                        });
                      }
                    } finally {
                      if (mounted) {
                        setState(() {
                          _isLoading = false;
                        });
                      }
                    }
                  },
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFF9333EA),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
            child: _isLoading
                ? const Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          valueColor: AlwaysStoppedAnimation<Color>(
                            Colors.white,
                          ),
                        ),
                      ),
                      SizedBox(width: 12),
                      Text(
                        'LOGGING IN...',
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                          color: Colors.white,
                          letterSpacing: 1,
                        ),
                      ),
                    ],
                  )
                : const Text(
                    'ENTER CASINO',
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                      letterSpacing: 1,
                    ),
                  ),
          ),
        ),

        const SizedBox(height: 12),

        // Test Login button (for development)
        SizedBox(
          width: double.infinity,
          height: 42,
          child: OutlinedButton(
            onPressed: _isLoading
                ? null
                : () async {
                    if (mounted) {
                      setState(() {
                        _isLoading = true;
                        _errorMessage = '';
                      });
                    }

                    try {
                      // Use the test credentials we created
                      final success = await ref
                          .read(authProvider.notifier)
                          .login('david', 'password123');
                      if (!success && mounted) {
                        setState(() {
                          _errorMessage = 'Test login failed.';
                        });
                      }
                    } catch (e) {
                      if (mounted) {
                        setState(() {
                          _errorMessage = 'Test login failed: ${e.toString()}';
                        });
                      }
                    } finally {
                      if (mounted) {
                        setState(() {
                          _isLoading = false;
                        });
                      }
                    }
                  },
            style: OutlinedButton.styleFrom(
              side: const BorderSide(color: Color(0xFF9333EA)),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
            child: const Text(
              'TEST LOGIN (david/password123)',
              style: TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.bold,
                color: Color(0xFF9333EA),
                letterSpacing: 1,
              ),
            ),
          ),
        ),
      ],
    );
  }

  @override
  void dispose() {
    _usernameController.dispose();
    _passwordController.dispose();
    super.dispose();
  }
}
