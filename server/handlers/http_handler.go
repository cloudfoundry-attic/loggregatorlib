package handlers

import (
	"mime/multipart"
	"net/http"
)

type httpHandler struct {
	messages <-chan []byte
}

func NewHttpHandler(m <-chan []byte) *httpHandler {
	return &httpHandler{messages: m}
}

func (h *httpHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	mp := multipart.NewWriter(rw)
	defer mp.Close()

	rw.Header().Set("Content-Type", `multipart/x-protobuf; boundary=`+mp.Boundary())

	for message := range h.messages {
		partWriter, _ := mp.CreatePart(nil)
		partWriter.Write(message)
	}
}
