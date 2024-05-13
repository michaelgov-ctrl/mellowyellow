package worker

import (
	"encoding/json"
	"net/http"
)

type envelope map[string]any

func (a *Api) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	js, err := json.MarshalIndent(b, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for k, v := range headers {
		w.Header()[k] = v
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (a *Api) logError(r *http.Request, err error) {
	method, uri := r.Method, r.URL.RequestURI()

	a.Logger.Error(err.Error(), "method", method, "uri", uri)
}

func (a *Api) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := envelope{"error": message}

	if err := a.writeJSON(w, status, env, nil); err != nil {
		a.logError(r, err)
		w.WriteHeader(500)
	}
}

func (a *Api) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	a.logError(r, err)

	message := "the server encountered a problem and could not process your request"
	a.errorResponse(w, r, http.StatusInternalServerError, message)
}
