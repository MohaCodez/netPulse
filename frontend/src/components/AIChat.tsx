import { useState, useRef, useEffect } from 'react';
import './AIChat.css';

interface Message {
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
}

export function AIChat() {
  const [messages, setMessages] = useState<Message[]>([
    {
      role: 'assistant',
      content: 'Hi! I\'m your network assistant powered by Qwen (running locally via Ollama). Ask me anything about your network, the charts, or networking concepts. I have access to your real-time metrics.',
      timestamp: new Date(),
    },
  ]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSend = async () => {
    const question = input.trim();
    if (!question || loading) return;

    setInput('');
    setMessages((prev) => [...prev, { role: 'user', content: question, timestamp: new Date() }]);
    setLoading(true);

    try {
      if (window.go?.main?.App?.AskAI) {
        const response = await window.go.main.App.AskAI(question);
        setMessages((prev) => [...prev, { role: 'assistant', content: response, timestamp: new Date() }]);
      } else {
        setMessages((prev) => [...prev, { role: 'assistant', content: 'AI not available — make sure Ollama is running (`ollama serve`).', timestamp: new Date() }]);
      }
    } catch (err: any) {
      setMessages((prev) => [...prev, { role: 'assistant', content: `Error: ${err.message || err}`, timestamp: new Date() }]);
    } finally {
      setLoading(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const suggestions = [
    "Why is my latency spiking?",
    "Explain my Wi-Fi signal quality",
    "Should I switch to 5 GHz?",
    "What's causing packet loss?",
    "How can I reduce jitter for gaming?",
    "Explain bufferbloat",
  ];

  return (
    <div className="ai-chat">
      <div className="chat-header">
        <h3>🤖 AI Network Assistant</h3>
        <span className="chat-model">qwen2.5-coder:3b (local)</span>
      </div>

      <div className="chat-messages">
        {messages.map((msg, i) => (
          <div key={i} className={`chat-message ${msg.role}`}>
            <div className="message-bubble">
              <pre className="message-content">{msg.content}</pre>
            </div>
            <span className="message-time">
              {msg.timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
            </span>
          </div>
        ))}
        {loading && (
          <div className="chat-message assistant">
            <div className="message-bubble">
              <span className="typing-indicator">Thinking<span className="dots">...</span></span>
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {messages.length <= 1 && (
        <div className="chat-suggestions">
          {suggestions.map((s) => (
            <button key={s} className="suggestion-btn" onClick={() => { setInput(s); }}>
              {s}
            </button>
          ))}
        </div>
      )}

      <div className="chat-input-row">
        <textarea
          className="chat-input"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Ask about your network..."
          rows={1}
          disabled={loading}
        />
        <button className="chat-send" onClick={handleSend} disabled={loading || !input.trim()}>
          ➤
        </button>
      </div>
    </div>
  );
}
