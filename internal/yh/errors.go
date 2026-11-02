package yh

import "errors"

// The kinds of failure, without the words: the window is what says them,
// in the language the desktop is in, so the client says what went wrong
// and leaves the sentence to the interface.
//
// A failure the server explained itself is an *APIError, whose message is
// the server's own and is shown as it came.
var (
	ErrNetwork   = errors.New("the request did not go through")
	ErrResponse  = errors.New("the answer could not be read")
	ErrStatus    = errors.New("the answer was not a success")
	ErrNoToken   = errors.New("the sign-in returned no token")
	ErrNoAccount = errors.New("the account could not be read")
	ErrNoCaptcha = errors.New("no picture to read out came back")
)
