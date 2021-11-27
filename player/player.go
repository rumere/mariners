package player

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Player struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	Phone      string   `json:"phone"`
	Email      string   `json:"email"`
	GhinNumber string   `json:"ghin_number"`
	Nicknames  []string `json:"nicknames"`
}

func AddPlayer(w http.ResponseWriter, r *http.Request) {
	p := Player{}

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Do something with the Person struct...
	fmt.Fprintf(w, "Person: %+v", p)
	w.WriteHeader(http.StatusOK)
}
