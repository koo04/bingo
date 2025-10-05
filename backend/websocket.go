package main

import "github.com/labstack/echo/v4"

func webSocketHandler(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	connMutex.Lock()
	connections[ws] = true
	connMutex.Unlock()

	defer func() {
		connMutex.Lock()
		delete(connections, ws)
		connMutex.Unlock()
	}()

	// Keep connection alive
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}

	return nil
}

func broadcastUpdate(eventType string, item any) {
	connMutex.RLock()
	defer connMutex.RUnlock()

	message := map[string]any{
		"type": eventType,
		"data": item,
	}

	for conn := range connections {
		if err := conn.WriteJSON(message); err != nil {
			delete(connections, conn)
			conn.Close()
		}
	}
}
