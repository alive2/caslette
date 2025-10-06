// Card representation
class Card {
  final String suit; // 'hearts', 'diamonds', 'clubs', 'spades'
  final String rank; // '2', '3', ..., '10', 'J', 'Q', 'K', 'A'

  const Card({required this.suit, required this.rank});

  factory Card.fromJson(Map<String, dynamic> json) {
    return Card(suit: json['suit'] as String, rank: json['rank'] as String);
  }

  Map<String, dynamic> toJson() {
    return {'suit': suit, 'rank': rank};
  }

  String get displayName => '$rank of $suit';

  String get shortName =>
      '${rank.substring(0, 1)}${suit.substring(0, 1).toUpperCase()}';

  @override
  String toString() => shortName;

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is Card &&
          runtimeType == other.runtimeType &&
          suit == other.suit &&
          rank == other.rank;

  @override
  int get hashCode => suit.hashCode ^ rank.hashCode;
}

// Player in a poker game
class PokerPlayer {
  final String id;
  final String username;
  final int chips;
  final int currentBet;
  final bool isActive;
  final bool isFolded;
  final bool isAllIn;
  final String lastAction;
  final int position;
  final List<Card> holeCards;
  final bool isReady;

  const PokerPlayer({
    required this.id,
    required this.username,
    required this.chips,
    this.currentBet = 0,
    this.isActive = false,
    this.isFolded = false,
    this.isAllIn = false,
    this.lastAction = '',
    required this.position,
    this.holeCards = const [],
    this.isReady = false,
  });

  factory PokerPlayer.fromJson(Map<String, dynamic> json) {
    return PokerPlayer(
      id: json['id'] as String,
      username: json['username'] as String,
      chips: json['chips'] as int,
      currentBet: json['current_bet'] as int? ?? 0,
      isActive: json['is_active'] as bool? ?? false,
      isFolded: json['is_folded'] as bool? ?? false,
      isAllIn: json['is_all_in'] as bool? ?? false,
      lastAction: json['last_action'] as String? ?? '',
      position: json['position'] as int,
      holeCards:
          (json['hole_cards'] as List<dynamic>?)
              ?.map((card) => Card.fromJson(card as Map<String, dynamic>))
              .toList() ??
          [],
      isReady: json['is_ready'] as bool? ?? false,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'username': username,
      'chips': chips,
      'current_bet': currentBet,
      'is_active': isActive,
      'is_folded': isFolded,
      'is_all_in': isAllIn,
      'last_action': lastAction,
      'position': position,
      'hole_cards': holeCards.map((card) => card.toJson()).toList(),
      'is_ready': isReady,
    };
  }

  PokerPlayer copyWith({
    String? id,
    String? username,
    int? chips,
    int? currentBet,
    bool? isActive,
    bool? isFolded,
    bool? isAllIn,
    String? lastAction,
    int? position,
    List<Card>? holeCards,
    bool? isReady,
  }) {
    return PokerPlayer(
      id: id ?? this.id,
      username: username ?? this.username,
      chips: chips ?? this.chips,
      currentBet: currentBet ?? this.currentBet,
      isActive: isActive ?? this.isActive,
      isFolded: isFolded ?? this.isFolded,
      isAllIn: isAllIn ?? this.isAllIn,
      lastAction: lastAction ?? this.lastAction,
      position: position ?? this.position,
      holeCards: holeCards ?? this.holeCards,
      isReady: isReady ?? this.isReady,
    );
  }
}

// Poker game state
class PokerGameState {
  final String
  phase; // 'waiting', 'preflop', 'flop', 'turn', 'river', 'showdown', 'finished'
  final List<Card> communityCards;
  final int pot;
  final int currentBet;
  final String? activePlayerID;
  final List<PokerPlayer> players;
  final int dealerPosition;
  final int smallBlindPosition;
  final int bigBlindPosition;
  final int smallBlind;
  final int bigBlind;
  final String? winnerID;
  final Map<String, dynamic>? handResult;

  const PokerGameState({
    required this.phase,
    this.communityCards = const [],
    this.pot = 0,
    this.currentBet = 0,
    this.activePlayerID,
    this.players = const [],
    this.dealerPosition = 0,
    this.smallBlindPosition = 1,
    this.bigBlindPosition = 2,
    this.smallBlind = 10,
    this.bigBlind = 20,
    this.winnerID,
    this.handResult,
  });

  factory PokerGameState.fromJson(Map<String, dynamic> json) {
    return PokerGameState(
      phase: json['phase'] as String,
      communityCards:
          (json['community_cards'] as List<dynamic>?)
              ?.map((card) => Card.fromJson(card as Map<String, dynamic>))
              .toList() ??
          [],
      pot: json['pot'] as int? ?? 0,
      currentBet: json['current_bet'] as int? ?? 0,
      activePlayerID: json['active_player_id'] as String?,
      players:
          (json['players'] as List<dynamic>?)
              ?.map(
                (player) =>
                    PokerPlayer.fromJson(player as Map<String, dynamic>),
              )
              .toList() ??
          [],
      dealerPosition: json['dealer_position'] as int? ?? 0,
      smallBlindPosition: json['small_blind_position'] as int? ?? 1,
      bigBlindPosition: json['big_blind_position'] as int? ?? 2,
      smallBlind: json['small_blind'] as int? ?? 10,
      bigBlind: json['big_blind'] as int? ?? 20,
      winnerID: json['winner_id'] as String?,
      handResult: json['hand_result'] as Map<String, dynamic>?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'phase': phase,
      'community_cards': communityCards.map((card) => card.toJson()).toList(),
      'pot': pot,
      'current_bet': currentBet,
      'active_player_id': activePlayerID,
      'players': players.map((player) => player.toJson()).toList(),
      'dealer_position': dealerPosition,
      'small_blind_position': smallBlindPosition,
      'big_blind_position': bigBlindPosition,
      'small_blind': smallBlind,
      'big_blind': bigBlind,
      'winner_id': winnerID,
      'hand_result': handResult,
    };
  }

  PokerPlayer? get activePlayer {
    if (activePlayerID == null) return null;
    try {
      return players.firstWhere((player) => player.id == activePlayerID);
    } catch (e) {
      return null;
    }
  }

  PokerPlayer? getPlayerById(String playerId) {
    try {
      return players.firstWhere((player) => player.id == playerId);
    } catch (e) {
      return null;
    }
  }

  bool get isGameActive => phase != 'waiting' && phase != 'finished';
  bool get isWaitingForPlayers => phase == 'waiting';
  bool get isGameFinished => phase == 'finished';
}

class PlayerSlot {
  final int position;
  final String? playerId;
  final String? username;
  final bool isReady;
  final DateTime? joinedAt;

  const PlayerSlot({
    required this.position,
    this.playerId,
    this.username,
    this.isReady = false,
    this.joinedAt,
  });

  factory PlayerSlot.fromJson(Map<String, dynamic> json) {
    return PlayerSlot(
      position: json['position'] as int,
      playerId: json['player_id'] as String?,
      username: json['username'] as String?,
      isReady: json['is_ready'] as bool? ?? false,
      joinedAt: json['joined_at'] != null
          ? DateTime.parse(json['joined_at'] as String)
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'position': position,
      'player_id': playerId,
      'username': username,
      'is_ready': isReady,
      'joined_at': joinedAt?.toIso8601String(),
    };
  }
}

// Poker table information
class PokerTable {
  final String id;
  final String name;
  final String gameType;
  final int maxPlayers;
  final int currentPlayers;
  final int observerCount;
  final List<PlayerSlot> playerSlots;
  final Map<String, dynamic> settings;
  final String status; // 'waiting', 'playing', 'finished'
  final String? creatorID;
  final bool isPrivate;
  final DateTime createdAt;

  const PokerTable({
    required this.id,
    required this.name,
    required this.gameType,
    this.maxPlayers = 8,
    this.currentPlayers = 0,
    this.observerCount = 0,
    this.playerSlots = const [],
    this.settings = const {},
    this.status = 'waiting',
    this.creatorID,
    this.isPrivate = false,
    required this.createdAt,
  });

  factory PokerTable.fromJson(Map<String, dynamic> json) {
    final playerSlots =
        (json['player_slots'] as List<dynamic>?)
            ?.map((slot) => PlayerSlot.fromJson(slot as Map<String, dynamic>))
            .toList() ??
        [];

    // Calculate current players from player slots
    final currentPlayers = playerSlots
        .where((slot) => slot.playerId != null)
        .length;

    return PokerTable(
      id: json['id'] as String,
      name: json['name'] as String,
      gameType: json['game_type'] as String,
      maxPlayers: json['max_players'] as int? ?? 8,
      currentPlayers: json['player_count'] as int? ?? currentPlayers,
      observerCount: json['observer_count'] as int? ?? 0,
      playerSlots: playerSlots,
      settings: json['settings'] as Map<String, dynamic>? ?? {},
      status: json['status'] as String? ?? 'waiting',
      creatorID: json['creator_id'] as String?,
      isPrivate: json['is_private'] as bool? ?? false,
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'game_type': gameType,
      'max_players': maxPlayers,
      'current_players': currentPlayers,
      'observer_count': observerCount,
      'player_slots': playerSlots.map((slot) => slot.toJson()).toList(),
      'settings': settings,
      'status': status,
      'creator_id': creatorID,
      'is_private': isPrivate,
      'created_at': createdAt.toIso8601String(),
    };
  }

  bool get isFull => currentPlayers >= maxPlayers;
  bool get canJoin => !isFull && status == 'waiting';
  bool get isPlaying => status == 'playing';
}

// Poker action types
enum PokerActionType { fold, check, call, raise, bet, allIn }

extension PokerActionTypeExtension on PokerActionType {
  String get name {
    switch (this) {
      case PokerActionType.fold:
        return 'fold';
      case PokerActionType.check:
        return 'check';
      case PokerActionType.call:
        return 'call';
      case PokerActionType.raise:
        return 'raise';
      case PokerActionType.bet:
        return 'bet';
      case PokerActionType.allIn:
        return 'all_in';
    }
  }

  static PokerActionType fromString(String value) {
    switch (value.toLowerCase()) {
      case 'fold':
        return PokerActionType.fold;
      case 'check':
        return PokerActionType.check;
      case 'call':
        return PokerActionType.call;
      case 'raise':
        return PokerActionType.raise;
      case 'bet':
        return PokerActionType.bet;
      case 'all_in':
        return PokerActionType.allIn;
      default:
        throw ArgumentError('Unknown poker action: $value');
    }
  }
}
