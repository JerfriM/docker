package domain

import "time"

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Matricula     string    `json:"matricula"`
	Password  string    `json:"-"` // El '-' evita que la contraseña se muestre en los JSON de respuesta
	CreatedAt time.Time `json:"created_at"`
}
