package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type ServerParameters struct {
	port     int    // webhook server port
	certFile string // path to the x509 certificate for https
	keyFile  string // path to the x509 private key matching Certfile
}

type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

type AdmissionReviewResponse struct {
	ApiVersion string   `json:"apiVersion"`
	Kind       string   `json:"kind"`
	Response   Response `json:"response"`
}

type Response struct {
	UID       string `json:"uid"`
	Allowed   bool   `json:"allowed"`
	PatchType string `json:"patchType"`
	Patch     string `json:"patch"`
}

var (
	globalDeserializer = serializer.NewCodecFactory(runtime.NewScheme())
	config             *rest.Config
	clientSet          *kubernetes.Clientset
	kubeConfig         string
	serverParameters   ServerParameters
)

func main() {
	useKubeConfig := os.Getenv("USE_KUBECONFIG") // In production, it does not exist. It will be ServiceAccount token in production
	kubeConfigFilePath := os.Getenv("KUBECONFIG")

	// Changing these parameters by users from Command line.
	flag.IntVar(&serverParameters.port, "port", 8443, "webhook server port")
	flag.StringVar(&serverParameters.certFile, "tlsCertFile", "/etc/webhook/certs/tls.crt", "File containing the x509 certificate")
	flag.StringVar(&serverParameters.keyFile, "tlsKeyFile", "/etc/webhook/certs/tls.key", "File containing the x509 private key")
	flag.Parse()

	// When talking to k8s, One way is the use ServiceAccount token
	// The other way is local ~/.kube/config file.
	// This statement decide that.
	if len(useKubeConfig) == 0 {
		clusterConfig, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		config = clusterConfig
	} else {
		if kubeConfigFilePath == "" {
			if home := homedir.HomeDir(); home != "" {
				kubeConfig = filepath.Join(home, ".kube", "config")
			} else {
				kubeConfig = kubeConfigFilePath
			}
		}

		fmt.Println("kubeConfig " + kubeConfig)

		flags, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			panic(err.Error())
		}
		config = flags
	}

	//clientSet is to talk to K8s API
	forConfig, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	clientSet = forConfig

	test()
	http.HandleFunc("/", HandleRoot)
	http.HandleFunc("/mutate", HandleMutate)
	log.Fatal(http.ListenAndServeTLS(":"+strconv.Itoa(serverParameters.port), serverParameters.certFile, serverParameters.keyFile, nil))
}

func HandleMutate(writer http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err.Error())
	}
	err = ioutil.WriteFile("/tmp/request", body, 0644)
	if err != nil {
		panic(err.Error())
	}
	var admissionReviewReq v1beta1.AdmissionReview
	_, _, err = globalDeserializer.UniversalDeserializer().Decode(body, nil, &admissionReviewReq)
	bytes, err := json.Marshal(&admissionReviewReq)
	if err != nil {
		panic(err.Error())
	}
	log.Printf("%v\n, admissionReviewReq")
	ioutil.WriteFile("/tmp/admission", bytes, 0644)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Errorf("Could not deserialize request: %v", err)
	} else if admissionReviewReq.Request == nil {
		writer.WriteHeader(http.StatusBadRequest)
		errors.New("malformed admission review: request is nil")
	}

	fmt.Printf("Type: %v\tEvent: %v\tName: %v\n",
		admissionReviewReq.Request.Kind,
		admissionReviewReq.Request.Operation,
		admissionReviewReq.Request.Name)

	var pod v1.Pod
	err = json.Unmarshal(admissionReviewReq.Request.Object.Raw, &pod)
	fmt.Printf("Pod is: %+v\n", pod)
	if err != nil {
		panic(err.Error())
	}

	labels := pod.ObjectMeta.Labels
	labels["example-webhook"] = "worked-like-a-charm"

	var patches []PatchOperation
	patches = append(patches, PatchOperation{
		Op:    "add",
		Path:  "/metadata/labels",
		Value: labels,
	})

	patchesBytes, _ := json.Marshal(patches)

	admissionReviewResponse := v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			UID:     admissionReviewReq.Request.UID,
			Allowed: true,
			Patch:   patchesBytes,
		},
	}

	fmt.Printf("admissionReviewResponse is: %+v\n", admissionReviewResponse)
	responseByte, err := json.Marshal(&admissionReviewResponse)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("marshalled admission review is: %+v\n", responseByte)
	writer.Write(responseByte)
}

func HandleRoot(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("Handle Root"))
}
