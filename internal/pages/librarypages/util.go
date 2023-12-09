package librarypages

import "database/sql"

func getStageDisplayName(stage string) string {
	switch stage {
	case "repeat":
		return "Repeat Practice"
	case "more_repeat":
		return "Extra Repeat Practice"
	case "random":
		return "Random Practice"
	case "interleave":
		return "Interleaved Practice"
	case "interleave_days":
		return "Interleave Between Days"
	case "completed":
		return "Completed"
	default:
		return "Unknown"
	}
}

func SpotMeasuresOrEmpty(measures sql.NullString) string {
	if measures.Valid {
		return measures.String
	}
	return ""
}
