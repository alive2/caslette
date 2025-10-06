import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/poker_provider.dart';

class CreateTableDialog extends ConsumerStatefulWidget {
  const CreateTableDialog({super.key});

  @override
  ConsumerState<CreateTableDialog> createState() => _CreateTableDialogState();
}

class _CreateTableDialogState extends ConsumerState<CreateTableDialog> {
  final _formKey = GlobalKey<FormState>();
  final _nameController = TextEditingController();
  final _passwordController = TextEditingController();

  int _maxPlayers = 6;
  int _smallBlind = 10;
  int _bigBlind = 20;
  bool _autoStart = true;
  bool _isPrivate = false;

  @override
  void dispose() {
    _nameController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      child: Container(
        constraints: const BoxConstraints(maxWidth: 500),
        padding: const EdgeInsets.all(24),
        child: Form(
          key: _formKey,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Icon(Icons.add_circle, color: Colors.green[600], size: 28),
                  const SizedBox(width: 12),
                  const Text(
                    'Create New Table',
                    style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
                  ),
                  const Spacer(),
                  IconButton(
                    onPressed: () => Navigator.of(context).pop(),
                    icon: const Icon(Icons.close),
                  ),
                ],
              ),

              const SizedBox(height: 24),

              // Table Name
              TextFormField(
                controller: _nameController,
                decoration: const InputDecoration(
                  labelText: 'Table Name',
                  hintText: 'Enter table name',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.table_restaurant),
                ),
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return 'Please enter a table name';
                  }
                  if (value.trim().length < 3) {
                    return 'Table name must be at least 3 characters';
                  }
                  return null;
                },
              ),

              const SizedBox(height: 16),

              // Max Players
              Row(
                children: [
                  const Expanded(
                    child: Text(
                      'Max Players',
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ),
                  DropdownButton<int>(
                    value: _maxPlayers,
                    items: [2, 3, 4, 5, 6, 7, 8].map((int value) {
                      return DropdownMenuItem<int>(
                        value: value,
                        child: Text('$value players'),
                      );
                    }).toList(),
                    onChanged: (int? newValue) {
                      if (newValue != null) {
                        setState(() {
                          _maxPlayers = newValue;
                        });
                      }
                    },
                  ),
                ],
              ),

              const SizedBox(height: 16),

              // Blinds
              Row(
                children: [
                  Expanded(
                    child: TextFormField(
                      initialValue: _smallBlind.toString(),
                      decoration: const InputDecoration(
                        labelText: 'Small Blind',
                        prefixText: '\$',
                        border: OutlineInputBorder(),
                      ),
                      keyboardType: TextInputType.number,
                      validator: (value) {
                        if (value == null || value.isEmpty) {
                          return 'Required';
                        }
                        final amount = int.tryParse(value);
                        if (amount == null || amount <= 0) {
                          return 'Must be positive';
                        }
                        return null;
                      },
                      onChanged: (value) {
                        final amount = int.tryParse(value);
                        if (amount != null && amount > 0) {
                          setState(() {
                            _smallBlind = amount;
                            _bigBlind = amount * 2;
                          });
                        }
                      },
                    ),
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: TextFormField(
                      initialValue: _bigBlind.toString(),
                      decoration: const InputDecoration(
                        labelText: 'Big Blind',
                        prefixText: '\$',
                        border: OutlineInputBorder(),
                      ),
                      keyboardType: TextInputType.number,
                      validator: (value) {
                        if (value == null || value.isEmpty) {
                          return 'Required';
                        }
                        final amount = int.tryParse(value);
                        if (amount == null || amount <= 0) {
                          return 'Must be positive';
                        }
                        if (amount <= _smallBlind) {
                          return 'Must be > small blind';
                        }
                        return null;
                      },
                      onChanged: (value) {
                        final amount = int.tryParse(value);
                        if (amount != null && amount > _smallBlind) {
                          setState(() {
                            _bigBlind = amount;
                          });
                        }
                      },
                    ),
                  ),
                ],
              ),

              const SizedBox(height: 16),

              // Settings
              CheckboxListTile(
                title: const Text('Auto-start game'),
                subtitle: const Text(
                  'Start game automatically when enough players join',
                ),
                value: _autoStart,
                onChanged: (bool? value) {
                  setState(() {
                    _autoStart = value ?? true;
                  });
                },
                controlAffinity: ListTileControlAffinity.leading,
                contentPadding: EdgeInsets.zero,
              ),

              CheckboxListTile(
                title: const Text('Private table'),
                subtitle: const Text('Require password to join'),
                value: _isPrivate,
                onChanged: (bool? value) {
                  setState(() {
                    _isPrivate = value ?? false;
                  });
                },
                controlAffinity: ListTileControlAffinity.leading,
                contentPadding: EdgeInsets.zero,
              ),

              // Password field (only if private)
              if (_isPrivate) ...[
                const SizedBox(height: 8),
                TextFormField(
                  controller: _passwordController,
                  decoration: const InputDecoration(
                    labelText: 'Password',
                    hintText: 'Enter table password',
                    border: OutlineInputBorder(),
                    prefixIcon: Icon(Icons.lock),
                  ),
                  obscureText: true,
                  validator: _isPrivate
                      ? (value) {
                          if (value == null || value.trim().isEmpty) {
                            return 'Password required for private tables';
                          }
                          if (value.trim().length < 4) {
                            return 'Password must be at least 4 characters';
                          }
                          return null;
                        }
                      : null,
                ),
              ],

              const SizedBox(height: 24),

              // Action buttons
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: [
                  TextButton(
                    onPressed: () => Navigator.of(context).pop(),
                    child: const Text('Cancel'),
                  ),
                  const SizedBox(width: 12),
                  ElevatedButton.icon(
                    onPressed: _createTable,
                    icon: const Icon(Icons.add),
                    label: const Text('Create Table'),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: Colors.green[600],
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(
                        horizontal: 24,
                        vertical: 12,
                      ),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _createTable() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    try {
      final table = await ref
          .read(pokerTablesProvider.notifier)
          .createTable(
            name: _nameController.text.trim(),
            maxPlayers: _maxPlayers,
            smallBlind: _smallBlind,
            bigBlind: _bigBlind,
            autoStart: _autoStart,
            isPrivate: _isPrivate,
            password: _isPrivate ? _passwordController.text.trim() : null,
          );

      if (mounted && table != null) {
        Navigator.of(context).pop();

        // Since the user is automatically joined when creating a table,
        // we can navigate directly to the table screen
        Navigator.of(context).pushNamed('/poker/table');

        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Table created successfully! Welcome to your table.'),
            backgroundColor: Colors.green,
          ),
        );
      }
    } catch (error) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to create table: $error'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }
}
