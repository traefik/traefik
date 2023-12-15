package api

import (
    "net/http"
    "github.com/sirupsen/logrus"
    "github.com/gorilla/mux"
)

// SetLogLevelHandler sets the log level for the application.
func (h *Handler) putLogLevel(rw http.ResponseWriter, request *http.Request) {
    vars := mux.Vars(request)
    levelStr := vars["logLevel"]

    level, err := logrus.ParseLevel(levelStr)
    if err != nil {
        // Respond with an error if the log level is not recognized
		writeError(rw, err.Error(), http.StatusInternalServerError)
        return
    }

    logrus.SetLevel(level)
    rw.WriteHeader(http.StatusOK)
    rw.Write([]byte("Log level set to " + levelStr))
}
