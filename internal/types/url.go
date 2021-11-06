package types

import (
	"encoding/json"
	"time"
)

type Url struct {
	ID           uint64    `db:"id" json:"id"`
	OriginalUrl  string    `db:"original_url" json:"original_url"`
	Slug         string    `db:"slug" json:"slug"`
	TotalVisits  uint64    `db:"total_visits" json:"total_visits"`
	UniqueVisits uint64    `db:"unique_visits" json:"unique_visits"`
	AccountID    uint64    `db:"account_id" json:"account_id"`
	Disabled     bool      `db:"disabled" json:"disabled"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

func (u *Url) String() string {
	bytes, err := json.Marshal(u)
	if err != nil {
		panic(err) // This cant happen.
	}

	return string(bytes)
}

func (u *Url) FromString(str string) error {
	return json.Unmarshal([]byte(str), u)
}
