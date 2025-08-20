import React, { useState, useEffect, useRef } from 'react';
import { useGameStore } from '../store/gameStore';
import { useAuthStore } from '../store/authStore';
import { ChatMessage } from '../types/game';
import styles from './Chat.module.css';

interface ChatProps {
  isOpen: boolean;
  onToggle: () => void;
}

const Chat: React.FC<ChatProps> = ({ isOpen, onToggle }) => {
  const [message, setMessage] = useState('');
  const [showEmojiPicker, setShowEmojiPicker] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const { gameState, isConnected, chatMessages, sendChatMessage, getChatHistory } = useGameStore();
  const { user } = useAuthStore();

  // Common emojis for quick selection
  const emojis = ['😀', '😂', '😍', '🤔', '😢', '😡', '👍', '👎', '❤️', '🎉', '🤝', '👀'];

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    scrollToBottom();
  }, [chatMessages]);

  // Load chat history when component mounts or game state changes
  useEffect(() => {
    if (gameState && isConnected) {
      getChatHistory(gameState.room_code);
    }
  }, [gameState, isConnected, getChatHistory]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const handleSendMessage = () => {
    if (!message.trim() || !gameState || !user) return;

    sendChatMessage(gameState.room_code, message.trim(), 'chat');
    setMessage('');
    setShowEmojiPicker(false);
    inputRef.current?.focus();
  };

  const handleSendEmoji = (emoji: string) => {
    if (!gameState || !user) return;

    sendChatMessage(gameState.room_code, emoji, 'emote');
    setShowEmojiPicker(false);
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  const isSystemMessage = (msg: ChatMessage) => {
    return msg.message_type === 'system' || msg.player_name === 'System';
  };

  const isEmoteMessage = (msg: ChatMessage) => {
    return msg.message_type === 'emote';
  };

  const processMessage = (text: string) => {
    // Simple mention detection (you can enhance this)
    return text.replace(/@(\w+)/g, '<span class="mention">@$1</span>');
  };

  if (!isOpen) {
    return (
      <div className={styles.chatToggle} onClick={onToggle}>
        <span className={styles.chatIcon}>💬</span>
        <span className={styles.chatLabel}>Chat</span>
        {chatMessages.length > 0 && (
          <span className={styles.messageCount}>{chatMessages.length}</span>
        )}
      </div>
    );
  }

  return (
    <div className={styles.chatContainer}>
      <div className={styles.chatHeader}>
        <div className={styles.chatTitle}>
          <span className={styles.chatIcon}>💬</span>
          Game Chat
        </div>
        <button onClick={onToggle} className={styles.closeButton}>
          ×
        </button>
      </div>

      <div className={styles.messagesContainer}>
        {chatMessages.length === 0 ? (
          <div className={styles.emptyState}>
            <span className={styles.emptyIcon}>💭</span>
            <p>No messages yet. Start the conversation!</p>
          </div>
        ) : (
          chatMessages.map((msg, index) => (
            <div
              key={msg.id || index}
              className={`${styles.message} ${
                isSystemMessage(msg) ? styles.systemMessage : ''
              } ${
                isEmoteMessage(msg) ? styles.emoteMessage : ''
              } ${
                msg.player_id === user?.id ? styles.ownMessage : ''
              }`}
            >
              {!isSystemMessage(msg) && (
                <div className={styles.messageHeader}>
                  <span className={styles.playerName}>{msg.player_name}</span>
                  <span className={styles.timestamp}>
                    {formatTimestamp(msg.timestamp)}
                  </span>
                </div>
              )}
              
              <div 
                className={styles.messageContent}
                dangerouslySetInnerHTML={{ 
                  __html: isSystemMessage(msg) ? msg.message : processMessage(msg.message)
                }}
              />
            </div>
          ))
        )}
        <div ref={messagesEndRef} />
      </div>

      <div className={styles.inputContainer}>
        {showEmojiPicker && (
          <div className={styles.emojiPicker}>
            <div className={styles.emojiGrid}>
              {emojis.map((emoji) => (
                <button
                  key={emoji}
                  onClick={() => handleSendEmoji(emoji)}
                  className={styles.emojiButton}
                >
                  {emoji}
                </button>
              ))}
            </div>
          </div>
        )}

        <div className={styles.inputRow}>
          <button
            onClick={() => setShowEmojiPicker(!showEmojiPicker)}
            className={styles.emojiToggle}
            title="Add emoji"
          >
            😀
          </button>
          
          <input
            ref={inputRef}
            type="text"
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            onKeyPress={handleKeyPress}
            placeholder="Type a message..."
            className={styles.messageInput}
            maxLength={500}
            disabled={!isConnected}
          />
          
          <button
            onClick={handleSendMessage}
            disabled={!message.trim() || !isConnected}
            className={styles.sendButton}
            title="Send message"
          >
            📤
          </button>
        </div>

        <div className={styles.inputHint}>
          {!isConnected ? (
            <span className={styles.disconnected}>Disconnected</span>
          ) : (
            <span>Press Enter to send • Use @username to mention</span>
          )}
        </div>
      </div>
    </div>
  );
};

export default Chat;
