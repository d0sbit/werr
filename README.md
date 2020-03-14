# werr
werr is a Go library that makes web error handling simpler

# Usage

```go

// ServeHTTP implements http.Handler and thus returns no error
func (h *SomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

    // wrap your handler in a WriteError which does sensible things with errors
    werr.WriteError(w, func() error {

        

        // WriteError does nothing
        return nil 
    }())
}

```