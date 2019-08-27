package controller

type Message struct {
	Type  string `json:"type,omitempty"`
	From  Player `json:"from,omitempty"`
	To    Player `json:"to",omitempty"`
	Nonce string `json:"nonce,omitempty"`
}

type Player struct {
	Name string `json:"name,omitempty"`
	UUID string `json:"uuid,omitempty"`
}

func (p *Player) Key() string {
	return p.Name + p.UUID
}
