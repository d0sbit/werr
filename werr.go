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

type ErrorCoder interface{ ErrorCode() int }
type ErrorShower interface{ ErrorShow() string }
type ErrorIDer interface{ ErrorID() string }
type ErrorLocer interface{ ErrorLoc() string }

func mkid() string {
	return fmt.Sprintf("%X", rand.Int63())
}

type errDetail struct {
	code int    // http status code
	show string // message to return in HTTP response
	loc  string // file:line info
	err  error  // underlying error
	id   string // unique id
}

func (e *errDetail) Error() string {

	var buf bytes.Buffer
	errstr := e.err.Error()
	// buf.Grow(len(e.show) + len(e.loc) + len(e.id) + len(errstr) + 20)

	fmt.Fprintf(&buf, "errDetail: id=%q, code=%d, show=%q, loc=%q, err: %s", e.id, e.code, e.show, e.loc, errstr)

	return buf.String()
}

func (e *errDetail) ErrorLoc() string {
	return e.loc
}

func (e *errDetail) ErrorShow() string {
	return e.show
}

func (e *errDetail) ErrorID() string {
	return e.id
}

func (e *errDetail) ErrorCode() int {
	ret := e.code
	if ret == 0 {
		return 500
	}
	return ret
}

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
