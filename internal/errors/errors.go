package prerrors

import "fmt"

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func New(code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}


var (
	ErrTeamExists = New(
		"TEAM_EXISTS",
		"team already exists",
	)

	ErrPRExists = New(
		"PR_EXISTS",
		"pull request already exists",
	)

	ErrPRMerged = New(
		"PR_MERGED",
		"pull request is already merged",
	)

	ErrNotAssigned = New(
		"NOT_ASSIGNED",
		"no reviewers are assigned",
	)

	ErrNoCandidate = New(
		"NO_CANDIDATE",
		"no available candidate for reassignment",
	)

	ErrNotFound = New(
		"NOT_FOUND",
		"resource not found",
	)

	ErrServer = New(
		"SERVER_ERROR",
		"internal server error",
	)
)