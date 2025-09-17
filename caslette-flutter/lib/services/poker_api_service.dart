import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/poker/poker_models.dart';

class PokerApiService {
  static const String baseUrl = 'http://localhost:8080/api/v1';

  /// Fetch list of available poker tables
  static Future<List<PokerTable>> fetchTables(String token) async {
    print('DEBUG: Making HTTP API call to fetch tables...');
    print('DEBUG: URL: $baseUrl/poker/tables');
    print('DEBUG: Token: ${token.substring(0, 20)}...');

    try {
      final response = await http.get(
        Uri.parse('$baseUrl/poker/tables'),
        headers: {
          'Authorization': 'Bearer $token',
          'Content-Type': 'application/json',
        },
      );

      print('DEBUG: HTTP response status: ${response.statusCode}');
      print('DEBUG: HTTP response body: ${response.body}');

      if (response.statusCode == 200) {
        final Map<String, dynamic> data = jsonDecode(response.body);
        final List<dynamic> tablesJson = data['tables'] ?? [];

        return tablesJson
            .map((tableData) => PokerTable.fromJson(tableData))
            .toList();
      } else {
        throw Exception('Failed to fetch tables: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Error fetching tables: $e');
    }
  }
}
