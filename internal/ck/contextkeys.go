package ck

type ContextKey string

var (
	ActivePlanKey  = ContextKey("activePlanID")
	UserKey        = ContextKey("user")
	CurrentPathKey = ContextKey("currentPath")
)
