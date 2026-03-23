package connection

import (
	"errors"
	"virturalDevice/pkg/vds/domain/connection"
)

type Store struct {
	connections []connection.Connection
}

func NewStore() *Store {
	return &Store{
		connections: make([]connection.Connection, 0),
	}
}

func (s *Store) CreateConnection(conn connection.Connection) error {

	configurable, ok := conn.(Configurable)
	if !ok {
		return errors.New("connection is not configurable")
	}

	config := configurable.Config()

	switch conn := conn.(type) {
	case *UDPConnection:
		s.connections = append(s.connections, conn)
	} // todo : 还没完成
}
