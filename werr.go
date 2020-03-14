// Package werr provides some convenient error wrapping for use in HTTP handlers.
package werr

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
)

// ErrLoc wraps an error so it's Error() method will return the same text prefixed
// with the file and line number it was called from.  A nil error value will return nil.
func ErrLoc(err error) error {

	if err == nil {
		return nil
	}

	// TODO: we really should make this smart enough to return the Code()
	// and other things as part of the overall error concept, but we'd need some permutations of types for
	// that, probably only for the code and public message, data can always exist and just return
	// nil if the caller doesn't support it

	_, file, line, _ := runtime.Caller(1)
	return fmt.Errorf("%s:%v :: %w", file, line, err)
}

// ErrorCoder interface is for errors that can return an http status code.
type ErrorCoder interface{ ErrorCode() int }

// ErrorShower interface is for errors that can return a message intended to be shown to a user.
type ErrorShower interface{ ErrorShow() string }

// ErrorIDer interface is for errors that can return a unique ID for this chain of errors.
// This can be useful when trying to correlate errors reported by users to more detailed information in a log.
type ErrorIDer interface{ ErrorID() string }

// ErrorLocer interface is for errors that can return location (file:line) information.
type ErrorLocer interface{ ErrorLoc() string }

// mkid returns a (usually) unique identifier
func mkid() string {
	return fmt.Sprintf("%X", rand.Int63())
}

// errDetail is used internally by the Error... methods.
type errDetail struct {
	code int    // http status code
	show string // message to return in HTTP response
	loc  string // file:line info
	err  error  // underlying error
	id   string // unique id
}

// Error implements the error interface.
func (e *errDetail) Error() string {

	var buf bytes.Buffer
	errstr := e.err.Error()
	// buf.Grow(len(e.show) + len(e.loc) + len(e.id) + len(errstr) + 20)

	fmt.Fprintf(&buf, "errDetail: id=%q, code=%d, show=%q, loc=%q, err: %s", e.id, e.code, e.show, e.loc, errstr)

	return buf.String()
}

// ErrorLoc returns a string describing the location (file:line) of the error.
func (e *errDetail) ErrorLoc() string {
	return e.loc
}

// ErrorShow returns the "show" message output in an http response.
func (e *errDetail) ErrorShow() string {
	return e.show
}

// ErrorID returns the unique ID associated with this error chain.
func (e *errDetail) ErrorID() string {
	return e.id
}

// ErrorCode returns the http status code.
func (e *errDetail) ErrorCode() int {
	ret := e.code
	if ret == 0 {
		return 500
	}
	return ret
}

// Unwrap returns the underlying error
func (e *errDetail) Unwrap() error {
	return e.err
}

// Error returns an error wrapped with the calling location and a unique ID.
// If error has already been wrapped by a call from this package cause will be returned.
// Passing a nil error will return nil
func Error(cause error) error {

	// return *errDetail as-is
	if edetail, ok := cause.(*errDetail); ok {
		if edetail == nil {
			return nil
		}
		return edetail
	}

	if cause == nil {
		return nil
	}

	_, file, line, _ := runtime.Caller(1)

	ret := errDetail{
		loc: fmt.Sprintf("%s:%v", file, line),
		err: fmt.Errorf("%w", cause),
		id:  mkid(),
	}

	return &ret

}

// Errorf acts like fmt.Errorf but also records the calling locatin and a unqiue ID.
func Errorf(fmtstr string, args ...interface{}) error {

	_, file, line, _ := runtime.Caller(1)

	ret := errDetail{
		loc: fmt.Sprintf("%s:%v", file, line),
		err: fmt.Errorf(fmtstr, args...),
		id:  mkid(),
	}

	return &ret
}

// ErrorCodef is like Errorf but also includes a code.
func ErrorCodef(code int, fmtstr string, args ...interface{}) error {

	_, file, line, _ := runtime.Caller(1)

	ret := errDetail{
		code: code,
		loc:  fmt.Sprintf("%s:%v", file, line),
		err:  fmt.Errorf(fmtstr, args...),
		id:   mkid(),
	}

	return &ret
}

// ErrorShowf is like Error but also includes a formatted "show" string, which WriteError will display.
func ErrorShowf(cause error, fmtstr string, args ...interface{}) error {

	show := fmt.Sprintf(fmtstr, args...)

	if cause == nil {
		cause = errors.New(show)
	}

	_, file, line, _ := runtime.Caller(1)

	ret := errDetail{
		loc:  fmt.Sprintf("%s:%v", file, line),
		show: show,
		err:  cause,
		id:   mkid(),
	}

	return &ret

}

// ErrorCodeShowf is like ErrorShowf but also with an error code.
func ErrorCodeShowf(code int, cause error, fmtstr string, args ...interface{}) error {

	show := fmt.Sprintf(fmtstr, args...)

	if cause == nil {
		cause = errors.New(show)
	}

	_, file, line, _ := runtime.Caller(1)

	ret := errDetail{
		code: code,
		loc:  fmt.Sprintf("%s:%v", file, line),
		show: show,
		err:  cause,
		id:   mkid(),
	}

	return &ret

}

// WriteError will write an error as an HTTP response and take into account the other wrapping from this package.
func WriteError(w http.ResponseWriter, err error) error {

	if err == nil {
		return nil
	}

	log.Printf("Error: %s", err.Error())

	var ret error

	w.Header().Set("Content-Type", "text/plain")

	var ec ErrorCoder
	if errors.As(err, &ec) {
		w.WriteHeader(ec.ErrorCode())
	} else {
		w.WriteHeader(500)
	}

	var showText string
	var es ErrorShower
	if errors.As(err, &es) {
		showText = es.ErrorShow()
	}
	if showText == "" {
		showText = "internal error"
	}

	_, ret = fmt.Fprint(w, showText)

	var ei ErrorIDer
	if errors.As(err, &ei) {
		_, ret = fmt.Fprintf(w, " [ID:%s]", ei.ErrorID())
	}

	return ret

}
