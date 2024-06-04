package components

import (
	"context"
	"fmt"
)

func HxCsrfHeader(csrf string) string {
	return fmt.Sprintf("{\"X-Csrf-Token\": \"%s\"}", csrf)
}

func HxHeader(name string, value string) string {
	return fmt.Sprintf("{\"%s\": \"%s\"}", name, value)
}

func getStageColor(stage string) string {
	switch stage {
	case "repeat":
		return "from-amber-200 to-amber-100 hover:from-amber-300 hover:to-amber-200 hover:shadow-amber-300"
	case "extra_repeat":
		return "from-orange-200 to-orange-100 hover:from-orange-300 hover:to-orange-200 hover:shadow-orange-300"
	case "random":
		return "from-pink-200 to-pink-100 hover:from-pink-300 hover:to-pink-200 hover:shadow-pink-300"
	case "interleave":
		return "from-indigo-200 to-indigo-100 hover:from-indigo-300 hover:to-indigo-200 hover:shadow-indigo-300"
	case "interleave_days":
		return "from-sky-200 to-sky-100 hover:from-sky-300 hover:to-sky-200 hover:shadow-sky-300"
	case "completed":
		return "from-green-200 to-green-100 hover:from-green-300 hover:to-green-200 hover:shadow-green-300"
	default:
		return "bg-white hover:bg-white hover:shadow-black"
	}
}

const basePlanCardClass = "flex flex-col gap-2 p-4 text-black bg-white rounded-lg shadow-sm focusable hover:shadow border-2 hover:shadow-violet-400  shadow-black/20"

func GetPlanCardClass(ctx context.Context, planID string) string {
	if planID == GetActivePracticePlan(ctx) {
		return basePlanCardClass + " border-violet-700"
	}
	return basePlanCardClass + " border-transparent"

}

const pieceCardClass = "flex py-4 px-6 text-black rounded-xl border shadow-sm transition-all duration-200 shadow-black/20 focusable"

func getPieceCardClass(completed bool) string {
	if completed {
		return pieceCardClass + " bg-green-200 border-green-300"
	} else {
		return pieceCardClass + " border-neutral-300 bg-gradient-to-br from-neutral-50 to-white hover:shadow hover:shadow-indigo-400"
	}
}

const readingCardClass = "flex py-4 w-full h-full px-6 text-black rounded-xl border shadow-sm transition-all duration-200 shadow-black/20 focusable bg-gradient-to-br"

func getReadingCardClass(completed bool) string {
	if completed {
		return readingCardClass + " to-green-100 from-green-200 border-green-300"
	} else {
		return readingCardClass + " border-teal-300 from-teal-100 to-teal-50 hover:shadow hover:shadow-teal-400"
	}
}

func getSessionsFromIntensity(intensity string) string {
	switch intensity {
	case "light":
		return "1"
	case "medium":
		return "3"
	case "heavy":
		return "5"
	default:
		return "1"
	}
}

func GetPiecePracticeUrl(completed, isActive bool, pieceID, practiceType, intensity string) string {
	if completed || !isActive {
		return "/library/pieces/" + pieceID
	}
	switch practiceType {
	case "random_spots":
		return "/library/pieces/" + pieceID + "/practice/random?resume=true&skipSetup=true&numSessions=" + getSessionsFromIntensity(intensity)
	case "starting_point":
		return "/library/pieces/" + pieceID + "/practice/starting-point?resume=true&skipSetup=true&numSessions=" + getSessionsFromIntensity(intensity)

	default:
		return "/library/pieces/" + pieceID
	}
}

func GetKeySignatureIconName(keyName, mode string) string {
	if mode == "Major (Ionian)" {
		switch keyName {
		case "G♯/A♭":
			return "icon-[key--af-major]"
		case "A":
			return "icon-[key--a-major]"
		case "A♯/B♭":
			return "icon-[key--bf-major]"
		case "B":
			return "icon-[key--b-major]"
		case "C":
			return "icon-[key--c-major]"
		case "C♯/D♭":
			return "icon-[key--df-major]"
		case "D":
			return "icon-[key--d-major]"
		case "D♯/E♭":
			return "icon-[key--ef-major]"
		case "E":
			return "icon-[key--e-major]"
		case "F":
			return "icon-[key--f-major]"
		case "F♯/G♭":
			return "icon-[key--gf-major]"
		case "G":
			return "icon-[key--g-major]"
		}
	} else if mode == "Minor (Aeolian)" {
		switch keyName {
		case "G♯/A♭":
			return "icon-[key--gs-minor]"
		case "A":
			return "icon-[key--a-minor]"
		case "A♯/B♭":
			return "icon-[key--bf-minor]"
		case "B":
			return "icon-[key--b-minor]"
		case "C":
			return "icon-[key--c-minor]"
		case "C♯/D♭":
			return "icon-[key--cs-minor]"
		case "D":
			return "icon-[key--d-minor]"
		case "D♯/E♭":
			return "icon-[key--ef-minor]"
		case "E":
			return "icon-[key--e-minor]"
		case "F":
			return "icon-[key--f-minor]"
		case "F♯/G♭":
			return "icon-[key--fs-minor]"
		case "G":
			return "icon-[key--g-minor]"
		}
	}
	return ""
}

const scaleCardClass = "flex py-4 w-full h-full px-6 text-black rounded-xl border shadow-sm transition-all duration-200 shadow-black/20 focusable bg-gradient-to-br"

func getScaleCardClass(completed bool) string {
	if completed {
		return scaleCardClass + " from-green-200 to-green-100 border-green-300"
	} else {
		return scaleCardClass + " border-rose-300 from-rose-100 to-rose-50 hover:shadow hover:shadow-rose-400"
	}
}
