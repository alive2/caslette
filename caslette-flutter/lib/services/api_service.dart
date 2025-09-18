import "package:dio/dio.dart";
import "package:flutter_riverpod/flutter_riverpod.dart";

class ApiService {
  final Dio _dio = Dio();

  ApiService() {
    _dio.options.baseUrl = "http://localhost:8081/api/v1";
    _dio.options.connectTimeout = const Duration(seconds: 5);
    _dio.options.receiveTimeout = const Duration(seconds: 3);
  }

  Future<Map<String, dynamic>?> login(String username, String password) async {
    try {
      print("Attempting login to: ${_dio.options.baseUrl}/auth/login");
      print("Username: $username");

      final response = await _dio.post(
        "/auth/login",
        data: {"username": username, "password": password},
      );

      print("Login response status: ${response.statusCode}");
      print("Login response data: ${response.data}");

      return response.data;
    } catch (e) {
      print("Login failed with error: ${e.toString()}");
      if (e is DioError) {
        print("DioError type: ${e.type}");
        print("DioError message: ${e.message}");
        print("DioError response: ${e.response?.data}");
        print("DioError status code: ${e.response?.statusCode}");
      }
      return null;
    }
  }

  Future<Map<String, dynamic>?> getBalance(String userId, String token) async {
    try {
      print("Fetching balance for user: $userId");
      final response = await _dio.get(
        "/diamonds/user/$userId",
        options: Options(headers: {"Authorization": "Bearer $token"}),
      );
      print("Balance response: ${response.data}");
      return response.data;
    } catch (e) {
      print("Get balance failed: $e");
      if (e is DioError) {
        print("DioError type: ${e.type}");
        print("DioError message: ${e.message}");
        print("DioError response: ${e.response?.data}");
        print("DioError status code: ${e.response?.statusCode}");
      }
      return null;
    }
  }
}

final apiServiceProvider = Provider<ApiService>((ref) {
  return ApiService();
});
