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
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -sha256 -days 365
```
Hostname will be:
```
example-webhook.default.svc
```

- Create tls.key and tls.crt secret
```yaml
cat << EOF > example-webhook-tls.yaml
apiVersion: v1
kind: Secret
metadata:
  name: example-webhook-tls
data:
  tls.crt: $(cat cert.pem | base64 | tr -d '\n')
  tls.key: $(cat key.pem | base64 | tr -d '\n')
EOF
```

- webhook.yaml caBundle substitution
````shell
ca_pem=$(cat tls/cert.pem | base64 | tr -d '\n') && \
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
      - operations: ["CREATE"]
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