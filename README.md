### Create a kind cluster
```shell
kind create cluster --name webhook --image kindest/node:v1.23.0
```
### TLS certificate for Web Hook
- In order to be invoked our web hook by K8s, we need a TLS certificate.  
```shell
docker run -it --rm -v ${PWD}:/work -w /work debian /bin/bash
```
```shell
apt update && \
apt install golang-cfssl -y && \
cfssl gencert -initca tls/ca-csr.json | cfssljson -bare /tmp/ca && \
cfssl gencert \
-ca=/tmp/ca.pem \
-ca-key=/tmp/ca-key.pem \
-config=/work/tls/ca-config.json \
-hostname="example-webhook.default.svc,example-webhook.default.svc.cluster.local,localhost,127.0.0.1" \
-profile=default /work/tls/ca-csr.json | cfssljson -bare /tmp/example-webhook
```

- Create tls.key and tls.crt secret
```yaml
cat << EOF > example-webhook-tls.yaml
apiVersion: v1
kind: Secret
metadata:
  name: example-webhook-tls
data:
  tls.crt: $(cat /tmp/example-webhook.pem | base64 | tr -d '\n')
  tls.key: $(cat /tmp/example-webhook-key.pem | base64 | tr -d '\n')
EOF
```

- webhook.yaml caBundle substitution
````shell
ca_pem=$(openssl base64 -A <"/tmp/ca.pem") && \
sed -e 's/${CA_BUNDLE}/'"$ca_pem"'/g' <"webhook-template.yaml" > webhook.yaml
````

### Webhook Configuration
````yaml
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: example-webhook
webhooks:
  - admissionReviewVersions:  # What type of reviews we accept
      - "v1"
      - "v1beta1"
    timeoutSeconds: 30 # How long we are planning to run the code
    clientConfig:
      caBundle: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUZXakNDQTBJQ0NRRHFra2loNUR5YTBUQU5CZ2txaGtpRzl3MEJBUXNGQURCdk1Rc3dDUVlEVlFRR0V3SlUKVWpFUE1BMEdBMVVFQ0F3R1FXNXJZWEpoTVE4d0RRWURWUVFIREFaQmJtdGhjbUV4Q3pBSkJnTlZCQW9NQWxSWgpNUXN3Q1FZRFZRUUxEQUpVV1RFa01DSUdBMVVFQXd3YlpYaGhiWEJzWlMxM1pXSm9iMjlyTG1SbFptRjFiSFF1CmMzWmpNQjRYRFRJeE1USXhNVEUyTWpnek9Wb1hEVEl5TVRJeE1URTJNamd6T1Zvd2J6RUxNQWtHQTFVRUJoTUMKVkZJeER6QU5CZ05WQkFnTUJrRnVhMkZ5WVRFUE1BMEdBMVVFQnd3R1FXNXJZWEpoTVFzd0NRWURWUVFLREFKVQpXVEVMTUFrR0ExVUVDd3dDVkZreEpEQWlCZ05WQkFNTUcyVjRZVzF3YkdVdGQyVmlhRzl2YXk1a1pXWmhkV3gwCkxuTjJZekNDQWlJd0RRWUpLb1pJaHZjTkFRRUJCUUFEZ2dJUEFEQ0NBZ29DZ2dJQkFNaGZ2MjE4NDJUcUZSSGoKQWRpU0VQZmZkUVNGd0Y2b2t3RmxMRWRVOTZKamhBVHhsalpqdzdYM25ZVjd4WDlENzV2SjJDY0xzeW51c1E2OApKMEV2U1VqZ0NnL3QrYjJBVW5zdmNCbW1IaFcza2RoeGJqMVhaTXplc2hvUEtBTWxIT05ObkljVnF2VXVDckI0CmRZSlkyckxGdzI4MVdxbm1paktUczIwS1kyTy9XUG1yN0dTenlpK0IrdnNmeHhlNXdYNFJwY2xMMHRlMkZnWGwKVkRnSmZpeEJTNE1GTWhwSFA4ekVadWJxUldNYVE2NTc0VVNQUjlONlVWbmZQUit6QzVmd0dHYVFqcHJmLzdmKwpnY0hyeTl3dEtuWDlNQVpvRzhtMVYwVk51KzNBUEJQa3JFMjdYdXNGYnFrejdUcFlFVUVLOE1BdkdveU5xZis0CnlRNVVRTTdSNG5uN1hiWCtPNEtGeEZ1MGJIeUY2NzYxUHlBMzJCOVdjUjlpMDZ2RW5ybHlVc1cvekU2SEcyNlYKWlR3OFd6SGdFQmpGTEdXRWJzVmF2b1MveFhWbDNINWRTMVpUQTRTdmRYR3ZQMHo2bzc4MTF3ZmZIU1ZTNnBhUgo1SVZhM0JHZ2ZlK1JIczJZbHp0dUNvdkNraU8zSHY5NjE1Rk44SHVPU21YRXI2bVgyaEh1TmxUbStpSGJ5R1hUCm5tSG10YTNxMzlXUzU2eVI5UlMzVW5qYzcxWmJZakpsM0hEWmtXd01sNzFHV042eGNqK0RJMGtUdEhSY2wvZ2wKalFjaTR1blZkc2VDZmJBK3g5bU9aRElMRkVRVlV6NzhzMDBJOStGU0E2MTVkUnBjOXZvVG00L2RRbGVPNVRnWgoyRmZ6Zng1K2d0TE1HRUJQckJKYXQzdStub1ZwQWdNQkFBRXdEUVlKS29aSWh2Y05BUUVMQlFBRGdnSUJBQ2p2CmZPVVRhMERyUUt4NUhRb3A2cEpwRklmMHdOeXdDVDVkbHppa1V6ZEVzVEtqVDMzTCtmclJGUS9aRVB2cnNraHMKMHZJa1BDL1gwM0I3T3JLS05pUXpPcnBCRk1YOTJ0WE9kd0ptaitQTy85bk5OL01wNzV0L1pZcTVLOWp2NVZxMApway94STROVWs4bU1rZ2JoUWwwYXRMd1prdHdWSzNTZndQakMxQ3FmNDRmYnlmQ1hYZ1FRYm9MTGt2Mlk3NDF3CjBDYUJpeUdseG8vSFhZS1phWUFmRkNvZGVWN0pZeGJkRnROVVJBV0IxRjVFWjBkci9kcTIrV2xqSEVIN0NjSmcKSFlEOTJzUVE4Y0dZa2RwTktCWmEvNWpyNXJRN3FySUhDYkNkTzlUaFJjVWxpL2RCb0FGQys3dlh0QmVIRzFjNApaa0E4Sy9wa1RpRTVTbDVnWmNtaTlMVGg4b0NZSGlzM0MvWE1qTk53WGFteFZwRk1xYmcwQnBJZG5xRitZazVrCmFqbjRFQmh0QTQzTUVHM040cFVXTWtaUDdzOWl0RmVkL0gwY25hT2tBcHhoR1FoWVdaU0ZwTVhJMnU2cERJWlAKbjRDT2tJS2c5UUtBYU1oVTBUWXkwQ1RlOEZ5NkR1UGRhOHdvbmFGQ0ZhTkcyOFZmM05JbE5vUEtKY29vQkhwYgpzWE81dkYvT0dUWmJOK1JYMEZRK3Z2bFBXRTZVdTNTbGYxamw0MEd3T2NjQytHTkxVWXR4dnN5dFZXc21mY0FICjhQVkNyeWE3MXZKU284R3N5enZCaE1iWlNlVGdsUFYwbnM3OHA1ODNEZldFQWlJV2pteWlTOEFHNVJoRmVGa1IKTUhLczVZa2lrZTE3ZDN3YXRYNFBvQU9aaTVnY0kveHFpazVnblkyMAotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
      service:
        name: example-webhook # What k8s resource to call for webhook
        namespace: default
        path: "/mutate" # Handler endpoint for this webhook
    name: example-webhook.default.svc  #DNS qualified name
    sideEffects: None
    objectSelector: # The resource that needs to have this label, which resources that qualifies to request to web hook
      matchLabels:
        example-webhook-enabled: "true"
    rules:
      - apiGroups: [""]
      - apiVersions: ["v1"]
      - resources: ["pods"]
      - operations: ["CREATE", "UPDATE"]
````

#### Write some code for webhook to operate
To create local development environment:
````dockerfile
FROM golang:1.17-alpine as dev-env
WORKDIR /app
````
````shell
docker build . -t webhook && \
docker run -it -p 80:80 -v ${PWD}:/app webhook sh
````

````go
package main

import (
	"log"
	"net/http"
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
````

```shell
go build -o webhook && \
./webhook
```

### Accessing  K8s from code
- Since the code runs in a container and our K8s cluster also in a container, we need to communicate them.
- We can run our development environment with --net host command.
````shell
docker run -it --rm --net host -v ${HOME}/.kube/:/root/.kube -v ${PWD}:/app webhook sh
````
```
apk add --no-cache curl && \
chmod +x kubectl && \
mv ./kubectl /usr/local/bin/kubectl
```
- We need global se/deserializer for k8s objects. Therefore, we import: 
```
"k8s.io/apimachinery/pkg/runtime"
"k8s.io/apimachinery/pkg/runtime/serializer"
```

and as global variable:
```
var (
	globalDeserializer = serializer.NewCodecFactory(runtime.NewScheme())
)
```

### Authentication
- In production environment we can use K8s ServiceAccount token to authenticate with the API server.
- When we are in development, we can use ~/.kube/config

````shell
go build -o webhook && \
export USE_KUBECONFIG=true && \
./webhook 
````
- To test functionality of connection and fetching resources from k8s cluster
    - Don't forget to set USE_KUBECONFIG property to true
````
func test() {
	pods, err := clientSet.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return
	}

	fmt.Printf("Number of pods %d\n", len(pods.Items))
}
````

### Expose an Endpoint to enable TLS
- Changing these parameters by users from the Command line.
```
flag.IntVar(&serverParameters.port, "port", 8443, "webhook server port")
flag.StringVar(&serverParameters.certFile, "tlsCertFile", "/etc/webhook/certs/tls.crt", "File containing the x509 certificate")
flag.StringVar(&serverParameters.keyFile, "tlsKeyFile", "/etc/webhook/certs/tls.key", "File containing the x509 private key")
flag.Parse()
...
log.Fatal(http.ListenAndServeTLS(":" + strconv.Itoa(serverParameters.port), serverParameters.certFile, serverParameters.keyFile, nil))
```

- Write the incoming api request and write them into a file 
```
func HandleMutate(writer http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err.Error())
	}
	err = ioutil.WriteFile("/tmp/request", body, 0644)
	if err != nil {
		panic(err.Error())
	}
}
```
### Deploying the Kubernetes
- Extend the docker file for new environments
````dockerfile
FROM golang:1.17-alpine as dev-env
WORKDIR /app

FROM dev-env as build-env
COPY go.mod /app
RUN go mod download
COPY . /app
RUN CGO_ENABLED=0 go build -o /webhook

FROM alpine:3.10 as runtime
COPY --from=build-env /webhook /usr/local/bin/webhook
RUN chmod +x /usr/local/bin/webhook
CMD ["webhook"]
````
- Build and push it to dockerhub
- Apply the ./tls/example-webhook-tls.yaml to kubernetes

#### Create RBAC
````yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: example-webhook
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: example-webhook
  namespace: default
rules:
  - apiGroups:
      - ""
    resources:
      - pods
    verbs:
      - get
      - list
      - watch
---
apiVersion: rbac.authorization.k8s.io/v1 
kind: ClusterRoleBinding
metadata:
  name: example-webhook
roleRef:
  apiGroup: rbac.authorization.k8s.io/v1
  kind: ClusterRole
  name: example-webhook
subjects:
  - kind: ServiceAccount
    name: example-webhook
````
- Deployment is created
```yaml
apiVersion: v1
kind: Service
metadata:
  name: example-webhook
  namespace: default
spec:
  selector:
    app: example-webhook
  ports:
    - port: 443
      targetPort: tls
      name: application
    - port: 80
      targetPort: metrics
      name: metrics
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: example-webhook
  name: example-webhook
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: example-webhook
  template:
    metadata:
      labels:
        app: example-webhook
    spec:
      serviceAccountName: example-webhook
      volumes:
        - name: webhook-tls-certs
          secret:
            secretName: example-webhook-tls
      containers:
        - image: sprayo7/example-webhook
          name: server
          command:
            - sh
          ports:
            - containerPort: 8443
              name: tls
            - containerPort: 80
              name: metrics
          volumeMounts:
            - mountPath: /etc/webhook/certs
              name: webhook-tls-certs
```

- Apply RBAC and Deployment
```
kubectl apply -f deployment.yaml && \
kubectl apply -f rbac.yaml
```

- After checking pods are OK, apply the webhook.
```
kubectl apply -f webhook.yaml
```
### Trying the mutate endpoint
- Create a dummy pod
- It is an unsuccessful operation. Because we didn't finish the mutate process.

### Admission Review
````shell
go get k8s.io/api/admission/v1beta1
````

````
	var admissionReviewReq v1beta1.AdmissionReview
	_, _, err = globalDeserializer.UniversalDeserializer().Decode(body, nil, &admissionReviewReq)
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
````
- to unmarshal the pod
````shell
go get k8s.io/api/core/v1
````

```
	var pod v1.Pod
	err := json.Unmarshal(admissionReviewReq.Request.Object.Raw, &pod)
	if err != nil {
		panic(err.Error())
	}
```

- To JsonPatch operations
```
	patches := `[{"op": "add", "path": "/metadata/labels/example-webhook", "value": "it-worked"}]`
	patchEnc := base64.StdEncoding.EncodeToString([]byte(patches))

	admissionReviewResponse := AdmissionReviewResponse{
		ApiVersion: "admission.k8s.io/v1",
		Kind:       "AdmissionReview",
		Response: Response{
			UID:     string(admissionReviewReq.Request.UID),
			Allowed: true,
			PatchType: "JSONPatch",
			Patch: patchEnc,
		},
	}

	fmt.Printf("admissionReviewResponse is: %+v\n", admissionReviewResponse)
	marshal, err := json.Marshal(&admissionReviewResponse)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("marshalled admission review is: %+v\n", marshal)
	writer.Write(marshal)
```