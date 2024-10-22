package err

import "net/http"

type Error struct {
	Error string `json:"error"`
}

func InternalServerError(w http.ResponseWriter, err []byte) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(err)
}

func BadRequest(w http.ResponseWriter, err []byte) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write(err)
}

func Unauthorized(w http.ResponseWriter, err []byte) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write(err)
}

func MethodNotAllowed(w http.ResponseWriter, err []byte) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write(err)
}

func UnprocessableEntity(w http.ResponseWriter, err []byte) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	w.Write(err)
}

func FailedDependency(w http.ResponseWriter, err []byte) {
	w.WriteHeader(http.StatusFailedDependency)
	w.Write(err)
}
