package storage

func (s *Store) Register(keys []string, bc *BlockedClient) {
	for _, key := range keys {
		s.Waiting[key] = append(s.Waiting[key], bc)
	}
}

func (s *Store) Unregister(bc *BlockedClient) {
	for _, key := range bc.Keys {
		clients := s.Waiting[key]
		for i, c := range clients {
			if c == bc {
				s.Waiting[key] = append(clients[:i], clients[i+1:]...)
				break
			}
		}
	}
}