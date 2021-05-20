package data

type User struct {
	ID      int64  `db:"id" json:"id"`
	Address string `db:"address" json:"address"`
	Amount  string `db:"amount,omitempty" json:"amount,omitempty"`
	Denom   string `db:"denom,omitempty" json:"denom"`
}

func (u User) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"address": u.Address,
		"amount":  u.Amount,
		"denom":   u.Denom,
	}

	return result
}

func (u User) ToReturn() map[string]interface{} {
	result := map[string]interface{}{
		"id":      u.ID,
		"address": u.Address,
		"amount":  u.Amount,
		"denom":   u.Denom,
	}

	return result
}
