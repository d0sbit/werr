package werr

import "net/http"

func ExampleFull() {

	// placeholder
	something := func() error { return nil }

	// handlers implements http.Handler and thus return no error
	_ = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// wrap your handler in a WriteError which does sensible things with errors
		WriteError(w, func() error {

			err := something()
			if err != nil {
				// you can return errors as-is and WriteError will
				// send a generic 500 response and log err
				return err
			}

			err = something()
			if err != nil {
				// wrapping with Error() will record the file and line number
				return Error(err)
			}

			err = something()
			if err != nil {
				// Errorf is like fmt.Errorf but automtically includes an ID and file:line number in the log
				return Errorf("something failed: %w", err) // error only shows in log, not response
			}

			err = something()
			if err != nil {
				// ErrorCodef is like Errorf but allows you to set an HTTP status code
				return ErrorCodef(400, "bad input: %w", err) // error only shows in log, not response
			}

			err = something()
			if err != nil {
				// ErrorShowf can be used to provide an error message that shows in the response
				return ErrorShowf(err, "something internal went awry") // message is sent in response and log, err shows in log
			}

			err = something()
			if err != nil {
				// ErrorCodeShowf is like ErrorShowf but with an http response code
				return ErrorCodeShowf(504, err, "something internal went awry")
			}

			// TODO: write successful response

			// WriteError does nothing if passed nil
			return nil
		}())

	})

}
