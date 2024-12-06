package webentities

// Bootstrap text color classes
const (
	textColorGreen  = "text-success"
	textColorYellow = "text-warning"
	textColorRed    = "text-danger"
)

const (
	dangerLow = iota
	dangerMid
	dangerHigh
)

func textColorByDanger(danger int) string {
	switch danger {
	case dangerLow:
		return textColorGreen
	case dangerMid:
		return textColorYellow
	case dangerHigh:
		return textColorRed
	default:
		return ""
	}
}
