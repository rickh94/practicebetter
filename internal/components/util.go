package components

import "fmt"

func HxCsrfHeader(csrf string) string {
	return fmt.Sprintf("{\"X-Csrf-Token\": \"%s\"}", csrf)
}

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

func getStageColor(stage string) string {
	switch stage {
	case "repeat":
		return "bg-violet-100/80 hover:bg-violet-200"
	case "random":
		return "bg-indigo-100/80 hover:bg-indigo-200"
	case "interleave":
		return "bg-sky-100/80 hover:bg-sky-200"
	case "interleave_days":
		return "bg-teal-100/80 hover:bg-teal-200"
	case "completed":
		return "bg-green-100/80 hover:bg-green-200"
	default:
		return "bg-white/80 hover:bg-white"
	}
}
