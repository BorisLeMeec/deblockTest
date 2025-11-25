package pkg

type TxMessage struct {
	UserID      string `json:"userId"`
	From        string `json:"from"`
	To          string `json:"to"`
	Amount      string `json:"amount"`
	Hash        string `json:"hash"`
	BlockNumber uint64 `json:"blockNumber"`
}
