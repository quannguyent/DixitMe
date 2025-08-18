export interface Player {
  id: string;
  name: string;
  score: number;
  position: number;
  hand: number[];
  is_connected: boolean;
  is_active: boolean;
}

export interface Card {
  id: number;
  url: string;
}

export interface RevealedCard {
  card_id: number;
  player_id: string;
  vote_count: number;
}

export interface Round {
  id: string;
  round_number: number;
  storyteller_id: string;
  clue: string;
  status: 'storytelling' | 'submitting' | 'voting' | 'scoring' | 'completed';
  storyteller_card?: number;
  submissions: Record<string, CardSubmission>;
  votes: Record<string, Vote>;
  revealed_cards?: RevealedCard[];
  created_at: string;
}

export interface CardSubmission {
  player_id: string;
  card_id: number;
}

export interface Vote {
  player_id: string;
  card_id: number;
}

export interface GameState {
  id: string;
  room_code: string;
  players: Record<string, Player>;
  current_round?: Round;
  status: 'waiting' | 'in_progress' | 'completed' | 'abandoned';
  round_number: number;
  max_rounds: number;
  created_at: string;
}

export interface GameMessage {
  type: string;
  payload: any;
}

// WebSocket message types
export const MessageTypes = {
  // From server
  CONNECTION_ESTABLISHED: 'connection_established',
  PLAYER_JOINED: 'player_joined',
  PLAYER_LEFT: 'player_left',
  GAME_STARTED: 'game_started',
  ROUND_STARTED: 'round_started',
  CLUE_SUBMITTED: 'clue_submitted',
  CARD_SUBMITTED: 'card_submitted',
  VOTING_STARTED: 'voting_started',
  VOTE_SUBMITTED: 'vote_submitted',
  ROUND_COMPLETED: 'round_completed',
  GAME_COMPLETED: 'game_completed',
  ERROR: 'error',
  GAME_STATE: 'game_state',
  
  // To server
  CREATE_GAME: 'create_game',
  JOIN_GAME: 'join_game',
  START_GAME: 'start_game',
  SUBMIT_CLUE: 'submit_clue',
  SUBMIT_CARD: 'submit_card',
  SUBMIT_VOTE: 'submit_vote',
  LEAVE_GAME: 'leave_game',
} as const;

export type MessageType = typeof MessageTypes[keyof typeof MessageTypes];
