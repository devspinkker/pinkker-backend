package utils

import (
	"github.com/gofiber/websocket/v2"
)

// Define una estructura para Client
type Client struct {
	// Campos relevantes para tu aplicación
	Connection *websocket.Conn
}

// ChatService maneja las operaciones del chat
type ChatService struct {
	// Mapa de salas donde la clave es el ID de la sala y el valor es una lista de clientes
	rooms map[string][]*Client
	// Otros campos y métodos relevantes...
}

// Instancia única de ChatService
var chatService *ChatService

func NewChatService() *ChatService {
	if chatService == nil {
		chatService = &ChatService{
			rooms: make(map[string][]*Client),
		}
	}
	return chatService
}

// AddClientToRoom agrega un cliente a una sala
func (s *ChatService) AddClientToRoom(roomID string, client *Client) {
	s.rooms[roomID] = append(s.rooms[roomID], client)
}

// RemoveClientFromRoom elimina un cliente de una sala
func (s *ChatService) RemoveClientFromRoom(roomID string, client *Client) {
	clients := s.rooms[roomID]
	for i, c := range clients {
		if c == client {
			// Elimina el cliente de la lista
			s.rooms[roomID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
}

// GetWebSocketClientsInRoom devuelve una lista de conexiones WebSocket de clientes en una sala
func (s *ChatService) GetWebSocketClientsInRoom(roomID string) ([]*websocket.Conn, error) {
	clients := s.rooms[roomID]
	var connections []*websocket.Conn
	for _, client := range clients {
		connections = append(connections, client.Connection)
	}
	return connections, nil
}
