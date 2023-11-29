package librarypages

func getStageDisplayName(stage string) string {
	switch stage {
	case "repeat":
		return "Repeat Practice"
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
