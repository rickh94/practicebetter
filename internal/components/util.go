package components

import "fmt"

func HxCsrfHeader(csrf string) string {
	return fmt.Sprintf("{\"X-Csrf-Token\": \"%s\"}", csrf)
}

func getStageColor(stage string) string {
	switch stage {
	case "repeat":
		return "bg-amber-100 hover:bg-amber-200 hover:shadow-amber-300"
	case "extra_repeat":
		return "bg-orange-100 hover:bg-orange-200 hover:shadow-orange-300"
	case "random":
		return "bg-pink-100 hover:bg-pink-200 hover:shadow-pink-300"
	case "interleave":
		return "bg-indigo-100 hover:bg-indigo-200 hover:shadow-indigo-300"
	case "interleave_days":
		return "bg-sky-100 hover:bg-sky-200 hover:shadow-sky-300"
	case "completed":
		return "bg-green-100 hover:bg-green-200 hover:shadow-green-300"
	default:
		return "bg-white hover:bg-white hover:shadow-black"
	}
}
