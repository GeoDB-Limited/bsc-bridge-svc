package data

type Transfer struct {
	ID      int64  `db:"id" json:"id"`
	Address string `db:"address" json:"address"`
	Amount  string `db:"amount,omitempty" json:"amount,omitempty"`
	Denom   string `db:"denom,omitempty" json:"denom"`
	Status  Status `db:"status" json:"status"`
	UserID  int64  `db:"user_id" json:"user_id"`
}

func (u Transfer) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"address": u.Address,
		"amount":  u.Amount,
		"denom":   u.Denom,
		"status":  u.Status,
		"user_id": u.UserID,
	}

	return result
}

func (u Transfer) ToReturn() map[string]interface{} {
	result := map[string]interface{}{
		"id":      u.ID,
		"address": u.Address,
		"amount":  u.Amount,
		"denom":   u.Denom,
		"status":  u.Status,
	}

	return result
}
