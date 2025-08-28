import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { GameState, Player, Card, GameMessage, MessageTypes, ChatMessage } from '../types/game';

interface GameStore {
  // Connection state
  socket: WebSocket | null;
  isConnected: boolean;
  playerId: string | null;
  playerName: string | null;
  
  // Game state
  gameState: GameState | null;
  currentPlayer: Player | null;
  cards: Card[];
  
  // Chat state
  chatMessages: ChatMessage[];
  
  // UI state
  isLoading: boolean;
  error: string | null;
  
  // Actions
  setPlayerInfo: (id: string, name: string) => void;
  connect: () => void;
  disconnect: () => void;
  createGame: (roomCode: string, playerName: string) => void;
  joinGame: (roomCode: string, playerName: string) => void;
  addBot: (roomCode: string, botName?: string, difficulty?: string) => void;
  removePlayer: (roomCode: string, playerId: string) => Promise<void>;
  leaveGame: (roomCode: string) => Promise<void>;
  leaveGameViaWebSocket: (roomCode: string) => void;
  deleteGame: (roomCode: string) => Promise<void>;
  startGame: (roomCode: string) => void;
  submitClue: (roomCode: string, clue: string, cardId: number) => void;
  submitCard: (roomCode: string, cardId: number) => void;
  submitVote: (roomCode: string, cardId: number) => void;
  sendChatMessage: (roomCode: string, message: string, messageType?: string) => void;
  getChatHistory: (roomCode: string, phase?: string, limit?: number) => void;
  setGameState: (gameState: GameState) => void;
  setChatMessages: (messages: ChatMessage[]) => void;
  addChatMessage: (message: ChatMessage) => void;
  setCards: (cards: Card[]) => void;
  setError: (error: string | null) => void;
  setLoading: (loading: boolean) => void;
}

const WEBSOCKET_URL = process.env.NODE_ENV === 'production' 
  ? `wss://${window.location.host}/ws`
  : 'ws://localhost:8080/ws';

const API_BASE_URL = process.env.NODE_ENV === 'production'
  ? ''
  : 'http://localhost:8080';

export const useGameStore = create<GameStore>()(
  devtools(
    (set, get) => ({
      // Initial state
      socket: null,
      isConnected: false,
      playerId: null,
      playerName: null,
      gameState: null,
      currentPlayer: null,
      cards: [],
      chatMessages: [],
      isLoading: false,
      error: null,

      // Actions
      setPlayerInfo: (id, name) => {
        set({ playerId: id, playerName: name }, false, 'setPlayerInfo');
      },

      connect: () => {
        const state = get();
        if (state.socket?.readyState === WebSocket.OPEN) {
          return; // Already connected
        }

        // Get auth token from localStorage (since auth store uses persist)
        let token = null;
        try {
          const authStorage = localStorage.getItem('auth-store');
          if (authStorage) {
            const parsed = JSON.parse(authStorage);
            token = parsed.state?.token;
          }
        } catch (error) {
          console.warn('Failed to get auth token:', error);
        }

        let wsUrl = WEBSOCKET_URL;
        const params = new URLSearchParams();
        
        if (state.playerId) {
          params.append('player_id', state.playerId);
        }
        
        if (token) {
          params.append('token', token);
        }
        
        if (params.toString()) {
          wsUrl += '?' + params.toString();
        }

        const socket = new WebSocket(wsUrl);

        socket.onopen = () => {
          console.log('WebSocket connected');
          set({ isConnected: true, error: null }, false, 'websocket-connected');
        };

        socket.onmessage = (event) => {
          try {
            const message: GameMessage = JSON.parse(event.data);
            handleWebSocketMessage(message, set, get);
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error);
          }
        };

        socket.onclose = () => {
          console.log('WebSocket disconnected');
          set({ isConnected: false }, false, 'websocket-disconnected');
          
          // Attempt to reconnect after 3 seconds
          setTimeout(() => {
            const currentState = get();
            if (!currentState.isConnected) {
              currentState.connect();
            }
          }, 3000);
        };

        socket.onerror = (error) => {
          console.error('WebSocket error:', error);
          set({ error: 'Connection error' }, false, 'websocket-error');
        };

        set({ socket }, false, 'set-socket');
      },

      disconnect: () => {
        const state = get();
        if (state.socket) {
          state.socket.close();
          set({ socket: null, isConnected: false }, false, 'disconnect');
        }
      },

      createGame: (roomCode, playerName) => {
        const state = get();
        if (!state.socket || !state.isConnected) {
          set({ error: 'Not connected to server' }, false, 'create-game-error');
          return;
        }

        const message = {
          type: MessageTypes.CREATE_GAME,
          payload: { room_code: roomCode, player_name: playerName }
        };

        state.socket.send(JSON.stringify(message));
        set({ isLoading: true, error: null }, false, 'creating-game');
      },

      joinGame: (roomCode, playerName) => {
        const state = get();
        if (!state.socket || !state.isConnected) {
          set({ error: 'Not connected to server' }, false, 'join-game-error');
          return;
        }

        const message = {
          type: MessageTypes.JOIN_GAME,
          payload: { room_code: roomCode, player_name: playerName }
        };

        state.socket.send(JSON.stringify(message));
        set({ isLoading: true, error: null }, false, 'joining-game');
      },

      addBot: (roomCode: string, botName = 'Bot', difficulty = 'medium') => {
        const state = get();
        if (!state.socket || !state.isConnected) {
          set({ error: 'Not connected to server' }, false, 'add-bot-error');
          return;
        }

        const message = {
          type: MessageTypes.ADD_BOT,
          payload: { 
            room_code: roomCode, 
            bot_level: difficulty 
          }
        };

        state.socket.send(JSON.stringify(message));
        set({ isLoading: true, error: null }, false, 'adding-bot');
      },

      // Remove player/bot from game
      removePlayer: async (roomCode: string, playerId: string) => {
        try {
          set({ isLoading: true, error: null }, false, 'removing-player');

          const response = await fetch(`${API_BASE_URL}/api/v1/games/remove-player`, {
            method: 'DELETE',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({
              room_code: roomCode,
              player_id: playerId
            }),
          });

          if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Failed to remove player');
          }

          // Player removed successfully - the server will send a game state update via WebSocket
          set({ isLoading: false }, false, 'player-removed');
        } catch (error) {
          console.error('Error removing player:', error);
          set({ 
            error: error instanceof Error ? error.message : 'Failed to remove player',
            isLoading: false 
          }, false, 'player-remove-error');
        }
      },

      // Leave game (for current player)
      leaveGame: async (roomCode: string) => {
        try {
          set({ isLoading: true, error: null }, false, 'leaving-game');

          const response = await fetch(`${API_BASE_URL}/api/v1/games/leave`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({
              room_code: roomCode
            }),
          });

          if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Failed to leave game');
          }

          // Successfully left game - redirect to lobby or home
          set({ 
            gameState: null,
            isLoading: false 
          }, false, 'left-game');
        } catch (error) {
          console.error('Error leaving game:', error);
          set({ 
            error: error instanceof Error ? error.message : 'Failed to leave game',
            isLoading: false 
          }, false, 'leave-game-error');
        }
      },

      // Delete game (for lobby manager/creator)
      deleteGame: async (roomCode: string) => {
        try {
          set({ isLoading: true, error: null }, false, 'deleting-game');

          const response = await fetch(`${API_BASE_URL}/api/v1/games/${roomCode}`, {
            method: 'DELETE',
            headers: {
              'Content-Type': 'application/json',
            },
          });

          if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Failed to delete game');
          }

          // Successfully deleted game - redirect to home
          set({ 
            gameState: null,
            isLoading: false 
          }, false, 'game-deleted');
        } catch (error) {
          console.error('Error deleting game:', error);
          set({ 
            error: error instanceof Error ? error.message : 'Failed to delete game',
            isLoading: false 
          }, false, 'delete-game-error');
        }
      },

      startGame: (roomCode) => {
        const state = get();
        if (!state.socket || !state.isConnected) return;

        const message = {
          type: MessageTypes.START_GAME,
          payload: { room_code: roomCode }
        };

        state.socket.send(JSON.stringify(message));
      },

      submitClue: (roomCode, clue, cardId) => {
        const state = get();
        if (!state.socket || !state.isConnected) return;

        const message = {
          type: MessageTypes.SUBMIT_CLUE,
          payload: { room_code: roomCode, clue, card_id: cardId }
        };

        state.socket.send(JSON.stringify(message));
      },

      submitCard: (roomCode, cardId) => {
        const state = get();
        if (!state.socket || !state.isConnected) return;

        const message = {
          type: MessageTypes.SUBMIT_CARD,
          payload: { room_code: roomCode, card_id: cardId }
        };

        state.socket.send(JSON.stringify(message));
      },

      submitVote: (roomCode, cardId) => {
        const state = get();
        if (!state.socket || !state.isConnected) return;

        const message = {
          type: MessageTypes.SUBMIT_VOTE,
          payload: { room_code: roomCode, card_id: cardId }
        };

        state.socket.send(JSON.stringify(message));
      },

      leaveGameViaWebSocket: (roomCode) => {
        const state = get();
        if (!state.socket || !state.isConnected) return;

        const message = {
          type: MessageTypes.LEAVE_GAME,
          payload: { room_code: roomCode }
        };

        state.socket.send(JSON.stringify(message));
      },

      setGameState: (gameState) => {
        const state = get();
        const currentPlayer = state.playerId && gameState.players[state.playerId] 
          ? gameState.players[state.playerId] 
          : null;

        set({ 
          gameState, 
          currentPlayer,
          isLoading: false 
        }, false, 'set-game-state');
      },

      setCards: (cards) => {
        set({ cards }, false, 'set-cards');
      },

      setError: (error) => {
        set({ error, isLoading: false }, false, 'set-error');
      },

      setLoading: (isLoading) => {
        set({ isLoading }, false, 'set-loading');
      },

      sendChatMessage: (roomCode, message, messageType = 'chat') => {
        const state = get();
        if (!state.socket || !state.isConnected) return;

        const chatMessage = {
          type: MessageTypes.SEND_CHAT,
          payload: {
            room_code: roomCode,
            message: message.trim(),
            message_type: messageType
          }
        };

        state.socket.send(JSON.stringify(chatMessage));
      },

      getChatHistory: (roomCode, phase = 'all', limit = 50) => {
        const state = get();
        if (!state.socket || !state.isConnected) return;

        const message = {
          type: MessageTypes.GET_CHAT_HISTORY,
          payload: {
            room_code: roomCode,
            phase,
            limit
          }
        };

        state.socket.send(JSON.stringify(message));
      },

      setChatMessages: (messages) => {
        set({ chatMessages: messages }, false, 'set-chat-messages');
      },

      addChatMessage: (message) => {
        const state = get();
        const newMessages = [...state.chatMessages, message];
        set({ chatMessages: newMessages }, false, 'add-chat-message');
      },
    }),
    {
      name: 'game-store',
    }
  )
);

function handleWebSocketMessage(
  message: GameMessage, 
  set: any, 
  get: () => GameStore
) {
  console.log('Received message:', message);

  switch (message.type) {
    
    case MessageTypes.CONNECTION_ESTABLISHED:
      const { player_id } = message.payload;
      set({ playerId: player_id }, false, 'connection-established');
      break;

    case MessageTypes.GAME_STATE:
      const { game_state } = message.payload;
      get().setGameState(game_state);
      break;

    case MessageTypes.ERROR:
      const { message: errorMessage } = message.payload;
      
      // If game was closed due to inactivity, also clear the game state
      if (errorMessage.includes('Game closed')) {
        set({ 
          error: errorMessage, 
          isLoading: false,
          gameState: null,
          currentPlayer: null
        }, false, 'game-closed-error');
      } else {
        set({ error: errorMessage, isLoading: false }, false, 'websocket-error');
      }
      break;

    case MessageTypes.CHAT_MESSAGE:
      const chatMessage = message.payload as ChatMessage;
      get().addChatMessage(chatMessage);
      break;

    case MessageTypes.CHAT_HISTORY:
      const { messages } = message.payload;
      get().setChatMessages(messages || []);
      break;

    case MessageTypes.PLAYER_JOINED:
    case MessageTypes.PLAYER_LEFT:
    case MessageTypes.GAME_STARTED:
    case MessageTypes.ROUND_STARTED:
    case MessageTypes.CLUE_SUBMITTED:
    case MessageTypes.CARD_SUBMITTED:
    case MessageTypes.VOTING_STARTED:
    case MessageTypes.VOTE_SUBMITTED:
    case MessageTypes.ROUND_COMPLETED:
    case MessageTypes.GAME_COMPLETED:
      // These messages will trigger game state updates from the server
      break;

    default:
      console.warn('Unknown message type:', message.type);
  }
}
