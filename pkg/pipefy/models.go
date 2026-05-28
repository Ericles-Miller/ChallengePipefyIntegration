package pipefy

type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type graphQLResponse[T any] struct {
	Data   T              `json:"data"`
	Errors []graphQLError `json:"errors"`
}

type graphQLError struct {
	Message string `json:"message"`
}

type CardResult struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type createCardData struct {
	CreateCard struct {
		Card CardResult `json:"card"`
	} `json:"createCard"`
}

type moveCardToPhaseData struct {
	MoveCardToPhase struct {
		Card struct {
			ID           string `json:"id"`
			CurrentPhase struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"current_phase"`
		} `json:"card"`
	} `json:"moveCardToPhase"`
}

type updateCardFieldData struct {
	UpdateCardField struct {
		Card    struct{ ID string `json:"id"` } `json:"card"`
		Success bool                            `json:"success"`
	} `json:"updateCardField"`
}
