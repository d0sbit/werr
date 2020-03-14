# werr
werr is a Go library that makes web error handling simpler. 

This library uses Go modules and is follows the Go 1.13 convention of error values implementing `Unwrap() error` to express one error wrapping another.

See godoc at https://godoc.org/github.com/d0sbit/werr

# Usage

```go

// ServeHTTP implements http.Handler and thus returns no error
func (h *SomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

    // wrap your handler in a WriteError which does sensible things with errors
    werr.WriteError(w, func() error {

        err := doSomeInternalThing()
        if err != nil {
            // you can return errors as-is and WriteError will
            // send a generic 500 response and log err
            return err
        }

        err = doSomeOtherThing()
        if err != nil {
            // wrapping with Error() will record the file and line number
            return werr.Error(err)
        }

        err = doSomeOtherThing2()
        if err != nil {
            // Errorf is like fmt.Errorf but automtically includes an ID and file:line number in the log
            return werr.Errorf("doSomeOtherThing2 failed: %w", err) // error only shows in log, not response
        }

        err = someThingThatMeans400IfItFails()
        if err != nil {
            // ErrorCodef is like Errorf but allows you to set an HTTP status code
            return werr.ErrorCodef(400, "bad input: %w", err) // error only shows in log, not response
        }

        err = someReallyInternalThing()
        if err != nil {
            // ErrorShowf can be used to provide an error message that shows in the response
            return werr.ErrorShowf(err, "something internal went awry") // message is sent in response and log, err shows in log
        }

        err = someReallyInternalThing2()
        if err != nil {
            // ErrorCodeShowf is like ErrorShowf but with an http response code
            return werr.ErrorCodeShowf(504, err, "something internal went awry")
        }
        
        // TODO: write successful response

        // WriteError does nothing if passed nil
        return nil 
    }())
}

```