package components

import "fmt"

func HxCsrfHeader(csrf string) string {
	return fmt.Sprintf("{\"X-Csrf-Token\": \"%s\"}", csrf)
}

func getStageColor(stage string) string {
	switch stage {
	case "repeat":
		return "bg-amber-100 hover:bg-amber-200"
	case "extra_repeat":
		return "bg-amber-100 hover:bg-amber-200"
	case "random":
		return "bg-pink-100 hover:bg-pink-200"
	case "interleave":
		return "bg-indigo-100 hover:bg-indigo-200"
	case "interleave_days":
		return "bg-sky-100 hover:bg-sky-200"
	case "completed":
		return "bg-green-100 hover:bg-green-200"
	default:
		return "bg-white/80 hover:bg-white"
	}
}
