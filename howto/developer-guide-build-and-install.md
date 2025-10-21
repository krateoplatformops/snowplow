
# Developer Guide: Building and Installing `snowplow`

This guide walks you through creating a local Kubernetes cluster with [kind](https://kind.sigs.k8s.io/), building the `snowplow` image with [`ko`](https://ko.build/), setting up `jq` custom modules, deploying `snowplow`, and waiting until it’s ready.


## 1. Start a local Kind cluster

Create a local Kubernetes cluster using **Kind**.

The cluster exposes ports `30081` and `30082` on the host for easy access to `snowplow` services.

```sh {name=kind-up}
kind get kubeconfig >/dev/null 2>&1 || \
cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 30081
    hostPort: 30081
    listenAddress: "127.0.0.1"
    protocol: TCP
  - containerPort: 30082
    hostPort: 30082
    listenAddress: "127.0.0.1"
    protocol: TCP
EOF
```


## 2. Create a namespace

Create a dedicated namespace where `snowplow` and its related resources will live.

```sh {name=create-namespace depends=kind-up}
export NAMESPACE="demo-system"
kubectl create namespace ${NAMESPACE}
```


## 3. Build the `snowplow` image with `ko`

Use [`ko`](https://ko.build/) to build and push the `snowplow` Docker image directly to the **Kind internal registry** (`kind.local`).

```sh {name=build depends=kind-up}
KO_DOCKER_REPO=kind.local ko build --base-import-paths .
```


## 4. Create a ConfigMap for custom `jq` modules

`Snowplow` uses custom `jq` modules during runtime. Create a ConfigMap to store them.

```sh {name=jq-custom-modules depends=create-namespace}
cat <<'EOF' | kubectl create configmap jq-custom-modules \
  --from-file=custom.jq=/dev/stdin \
  --namespace=${NAMESPACE}
def shout($s): ($s | ascii_upcase + "!!!");

def flipchar($c):
  {
    "a": "ɐ", "b": "q", "c": "ɔ", "d": "p", "e": "ǝ", "f": "ɟ", "g": "ƃ", "h": "ɥ", "i": "ᴉ",
    "j": "ɾ", "k": "ʞ", "l": "ʃ", "m": "ɯ", "n": "u", "o": "o", "p": "d", "q": "b", "r": "ɹ",
    "s": "s", "t": "ʇ", "u": "n", "v": "ʌ", "w": "ʍ", "x": "x", "y": "ʎ", "z": "z",
    "A": "∀", "B": "𐐒", "C": "Ɔ", "D": "p", "E": "Ǝ", "F": "Ⅎ", "G": "פ", "H": "H",
    "I": "I", "J": "ſ", "K": "ʞ", "L": "˥", "M": "W", "N": "N", "O": "O", "P": "Ԁ",
    "Q": "Q", "R": "ᴚ", "S": "S", "T": "┴", "U": "∩", "V": "Λ", "W": "M", "X": "X",
    "Y": "⅄", "Z": "Z", ".": "˙", ",": "'", "'": ",", "\"": ",,", "_": "‾", "?": "¿",
    "!": "¡", "(": ")", ")": "(", "[": "]", "]": "[", "{": "}", "}": "{"
  }[$c] // $c;

def flip($s):
  $s
  | explode
  | map([.] | implode | flipchar(.))
  | reverse
  | join("");
EOF
```

## 5. Deploy `snowplow`

Deploy `snowplow` using a single manifest that includes:

* a `ServiceAccount`
* a `Service` exposed on `NodePort`
* a `Deployment` for the `snowplow` app
* RBAC roles and bindings

```sh {name=deploy depends=jq-custom-modules}
cat <<EOF | kubectl apply -f -
---
kind: ServiceAccount
apiVersion: v1
metadata:
  name: snowplow
  namespace: ${NAMESPACE}
---
apiVersion: v1
kind: Service
metadata:
  name: snowplow
  namespace: ${NAMESPACE}
spec:
  selector:
    app: snowplow
  type: NodePort
  ports:
  - name: http
    port: 8081
    targetPort: http
    protocol: TCP
    nodePort: 30081
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: snowplow
  namespace: ${NAMESPACE}
  labels:
    app: snowplow
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: snowplow
  template:
    metadata:
      labels:
        app: snowplow
    spec:
      serviceAccountName: snowplow
      volumes:
      - name: jq-modules
        configMap:
          name: jq-custom-modules
      containers:
      - name: snowplow
        image: kind.local/snowplow:latest
        imagePullPolicy: Never
        args:
          - --debug=false
          - --blizzard=false
          - --port=8081
          - --authn-namespace=${NAMESPACE}
          - --jwt-sign-key=AbbraCadabbra
          - --pretty-log=false
          - --jq-modules-path=/jq-modules
        ports:
        - name: http
          containerPort: 8081
        volumeMounts:
        - name: jq-modules
          mountPath: /jq-modules
          readOnly: true
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: snowplow
rules:
- apiGroups: ["core.krateo.io"]
  resources: ["compositiondefinitions", "schemadefinitions"]
  verbs: ["get", "list"]
- apiGroups: ["templates.krateo.io"]
  resources: ["*"]
  verbs: ["get", "list"]
- apiGroups: ["apiextensions.k8s.io"]
  resources: ["customresourcedefinitions"]
  verbs: ["get", "list"]
- apiGroups: [""]
  resources: ["namespaces", "configmaps", "secrets"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: snowplow
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: snowplow
subjects:
- kind: ServiceAccount
  name: snowplow
  namespace: ${NAMESPACE}
EOF
```


## 6. Wait until the `snowplow` deployment is ready

Finally, wait for the `snowplow` deployment to become **available**.
This ensures all pods are up and running before proceeding.

```sh {name=wait-for-snowplow depends=deploy}
kubectl wait deployment/snowplow \
  --namespace ${NAMESPACE} \
  --for=condition=available \
  --timeout=90s
```
