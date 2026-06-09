package main

// allowedStatus define los estados permitidos para el campo status.
var allowedStatus = map[string]bool{
	"completado": true,
	"jugando":    true,
	"pendiente":  true,
	"abandonado": true,
}

// ValidateStatus retorna true si el status es uno de los valores permitidos.
func ValidateStatus(status string) bool {
	return allowedStatus[status]
}

// ValidatePersonalScore retorna true si el puntaje está en el rango 1..10.
func ValidatePersonalScore(score int) bool {
	return score >= 1 && score <= 10
}
