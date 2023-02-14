package controllers

import (
	"fmt"
	"net/http"
)

func writeFatalError(w http.ResponseWriter, msg string, originalErr error) {
	errMsg := fmt.Sprintf("Returning error to user msg=%s err %+v", msg, originalErr)
	fmt.Println(errMsg)
	w.WriteHeader(500)
	_, err := w.Write([]byte(errMsg))
	if err != nil {
		fmt.Println(fmt.Sprintf("Fatal error occurred writing error response to users err=%+v", err))
	}
}
