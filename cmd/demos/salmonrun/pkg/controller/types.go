package controller

type Message struct {
	Type  string `json:"type,omitempty"`
	From  Player `json:"to,omitempty"`
	To    Player `json:"from",omitempty"`
	Nonce string `json:"id,omitempty"`
}

type Player struct {
	Name string `json:"name,omitempty"`
	UUID string `json:"uuid,omitempty"`
}

func (p *Player) Key() string {
	return p.Name + p.UUID
}
