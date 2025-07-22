//package models

package models

type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password,omitempty"`
	//Password string `json:"-"` //hidden in all responses
	Age int `json:"age"`
}
