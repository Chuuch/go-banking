package util

// Constants for all supported currencies
const (
	USD = "USD"
	EUR = "EUR"
	CAD = "CAD"
)

// IsSupported() returns true if the provided currency is supported
func IsSupported(currency string) bool {
	switch currency {
	case USD, EUR, CAD:
		return true
	}
	return false
}
