class WebSocketService {
  constructor() {
    this.ws = null
    this.reconnectInterval = 3000
    this.maxReconnectAttempts = 10
    this.reconnectAttempts = 0
    this.listeners = new Map()
  }

  connect() {
    try {
      // Use the same base URL as the API, but convert to WebSocket protocol
      const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
      const wsUrl = apiBaseUrl.replace(/^https?:/, apiBaseUrl.startsWith('https:') ? 'wss:' : 'ws:') + '/ws'
      
      console.log('Connecting to WebSocket:', wsUrl)
      this.ws = new WebSocket(wsUrl)
      
      this.ws.onopen = () => {
        console.log('WebSocket connected')
        this.reconnectAttempts = 0
      }
      
      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          this.handleMessage(data)
        } catch (error) {
          console.error('Error parsing WebSocket message:', error)
        }
      }
      
      this.ws.onclose = () => {
        console.log('WebSocket disconnected')
        this.reconnect()
      }
      
      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error)
      }
    } catch (error) {
      console.error('Error connecting to WebSocket:', error)
      this.reconnect()
    }
  }

  reconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      console.log(`Attempting to reconnect... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`)
      setTimeout(() => this.connect(), this.reconnectInterval)
    }
  }

  handleMessage(data) {
    console.log('WebSocket message received:', data)
    const { type } = data
    console.log('Message type:', type)
    const callbacks = this.listeners.get(type)
    console.log('Callbacks for type:', callbacks ? callbacks.length : 0)
    if (callbacks) {
      callbacks.forEach(callback => callback(data))
    } else {
      console.warn('No callbacks registered for WebSocket message type:', type)
    }
  }

  on(eventType, callback) {
    if (!this.listeners.has(eventType)) {
      this.listeners.set(eventType, [])
    }
    this.listeners.get(eventType).push(callback)
  }

  off(eventType, callback) {
    const callbacks = this.listeners.get(eventType)
    if (callbacks) {
      const index = callbacks.indexOf(callback)
      if (index > -1) {
        callbacks.splice(index, 1)
      }
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }
}

export default new WebSocketService()
