package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var (
	globalDeserializer = serializer.NewCodecFactory(runtime.NewScheme())
	config             *rest.Config
	clientSet          *kubernetes.Clientset
	kubeConfig         string
)

func main() {
	useKubeConfig := os.Getenv("USE_KUBECONFIG") // In production, it does not exist. It will be ServiceAccount token in production
	kubeConfigFilePath := os.Getenv("KUBECONFIG")

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
	log.Fatal(http.ListenAndServe(":80", nil))
}

func HandleMutate(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("Mutate"))
}

func HandleRoot(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("Handle Root"))
}
