package main

import (
	"log"
	"net/http"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	globalDeserializer = serializer.NewCodecFactory(runtime.NewScheme())
)

func main() {
	http.HandleFunc("/", HandleRoot)
	http.HandleFunc("/mutate", HandleMutate)
	log.Fatal(http.ListenAndServe(":80", nil))
}

func HandleMutate(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("Mutate"))
}

func HandleRoot(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("Handle Root"))
}