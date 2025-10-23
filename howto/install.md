# Installing `snowplow` on [Kind][kind]

If you have any Docker-compatible container runtime installed (including native Docker, Docker Desktop, or OrbStack), you can easily launch a disposable cluster just for this quickstart using [Kind][kind].

```sh {name=kind-up}
kind get kubeconfig >/dev/null 2>&1 || \
cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 30081 # Krateo Snowplow
    hostPort: 30081
    listenAddress: "127.0.0.1"
    protocol: TCP
  - containerPort: 30082 # Krateo AuthN Service
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

## 3. Create the JWT Secret

```sh {name=create-jwt-secret depends=create-namespace}
export JWT_SECRET=AbbraCadabbra
kubectl create secret generic jwt-sign-key \
  --from-literal=key=${JWT_SECRET} -n ${NAMESPACE}
```

## 4. Create a Krateo PlatformOps User

To quickly create a Krateo PlatformOps user, install [`krateoctl`][krateoctl] and run the following command:

```sh {name=create-krateo-user depends=create-jwt-secret}
export KRATEO_USER=cyberjoker
export KRATEO_ACCESS_TOKEN=$(krateoctl add-user -k "${JWT_SECRET}" -n "${NAMESPACE}" "${KRATEO_USER}")

echo "KRATEO_USER=${KRATEO_USER}" > .env
echo "KRATEO_ACCESS_TOKEN=${KRATEO_ACCESS_TOKEN}" >> .env
```

## 5. Deploy snowplow

Finally, install `snowplow` using the Helm chart:

```sh {name=install-snowplow depends=create-jwt-secret}
helm install snowplow https://github.com/krateoplatformops/helm-charts/raw/gh-pages/snowplow-0.20.2.tgz \
  --namespace ${NAMESPACE}
```

You are now ready to move on to the next steps. From here, you can start testing the [RESTActions][restactions] to see how the different use cases work in practice. 

Experiment with creating, updating, and querying resources to get a hands-on understanding of the platform's capabilities.

## Related ADRs

- [Decoupling `authn` from `snowplow` for Testing and Operations](./decoupling-authn-from-snowplow-for-testing.md)




[kind]: https://kind.sigs.k8s.io/
[krateoctl]: https://github.com/krateoplatformops/krateoctl/releases
[restactions]: restactions.md
