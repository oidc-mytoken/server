package api

// AllTokeninfoActions holds all defined TokenInfo strings
var AllTokeninfoActions = [...]string{TokeninfoActionIntrospect, TokeninfoActionEventHistory, TokeninfoActionSubtokenTree, TokeninfoActionListMytokens}

// TokeninfoActions
const (
	TokeninfoActionIntrospect   = "introspect"
	TokeninfoActionEventHistory = "event_history"
	TokeninfoActionSubtokenTree = "subtoken_tree"
	TokeninfoActionListMytokens = "list_mytokens"
)
