#!/bin/bash

if [ $0 != "./scripts/quick-start.sh" ]; then
  exit 1
fi

set -eu
OS=$(go env GOOS)
ARCH=$(go env GOARCH)

function install_kubectl {
  local VERSION="v1.24.2"
  echo -n "Install kubectl ${VERSION} ... "
  curl -sfLO "https://dl.k8s.io/release/${VERSION}/bin/${OS}/${ARCH}/kubectl"
  chmod +x ./kubectl
  mv ./kubectl tools
  echo "Done!"
}

function install_kubebuilder {
    local VERSION="v3.3.0"
    echo -n "Install kubebuilder ${VERSION} ... "
    curl -sL -o kubebuilder "https://github.com/kubernetes-sigs/kubebuilder/releases/download/${VERSION}/kubebuilder_${OS}_${ARCH}"
    chmod +x ./kubebuilder
    mv ./kubebuilder tools
    echo "Done!"
}

function install_kind {
    local VERSION="v0.14.0"
    echo -n "Install kind ${VERSION} ... "
    curl -sL -o kind "https://kind.sigs.k8s.io/dl/${VERSION}/kind-${OS}-$ARCH"
    chmod +x ./kind
    mv ./kind tools
    echo "Done!"
}

function install_kustomize {
    local VERSION="v4.5.2"
    local TARNAME="kustomize_${VERSION}_${OS}_${ARCH}.tar.gz"
    local DOWNLOAD_URL="https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/${VERSION}/${TARNAME}"
    echo -n "Install kustomize ${VERSION} ... "
    curl -sL -o "${TARNAME}" "${DOWNLOAD_URL}"
    tar xf "${TARNAME}"
    rm "${TARNAME}"
    chmod +x ./kustomize
    mv ./kustomize tools
    echo "Done!"
}

function install_virtctl {
  local VERSION="v0.54.0"
  echo -n "Install virtctl ${VERSION} ... "
  curl -sL -o virtctl https://github.com/kubevirt/kubevirt/releases/download/${VERSION}/virtctl-${VERSION}-${ARCH}
  chmod +x ./virtctl
  mv ./virtctl tools
  echo "Done!"
}

function install_mc {
  echo -n "Install mc ... "
  curl -sL -o mc "https://dl.min.io/client/mc/release/${OS}-${ARCH}/mc"
  chmod +x mc
  mv ./mc tools
  echo "Done!"
}

function create_kind_cluster {
  ./tools/kind create cluster --config ./hack/kind-kubeberth-dev.yaml
  echo -n "Updating node ... "
  docker exec -it kubeberth-dev-worker sh -c "apt update; apt install -y qemu-kvm libvirt-daemon" > /dev/null
  echo "Done!"
}

function deploy_calico {
  echo -n "Deploy calico ... "
  ./tools/kubectl apply -f ./hack/calico-setup.yaml > /dev/null
  sleep 3
  ./tools/kubectl apply -f ./hack/calico-kube-controllers.yaml > /dev/null
  sleep 3
  ./tools/kubectl apply -f ./hack/calico-node.yaml > /dev/null
  sleep 3
  echo "Done!"
}


function deploy_certmanager {
  local VERSION="v1.8.2"
  echo -n "Deploy cert-manager ${VERSION} ... "
  ./tools/kubectl apply -f "https://github.com/cert-manager/cert-manager/releases/download/${VERSION}/cert-manager.yaml" > /dev/null
  echo "Done!"
}

function deploy_kubevirt {
  local VERSION="v0.54.0"
  echo "Deploy kubevirt ${VERSION} ... "
  ./tools/kubectl apply -f "https://github.com/kubevirt/kubevirt/releases/download/${VERSION}/kubevirt-operator.yaml" > /dev/null
  sleep 3
  ./tools/kubectl apply -f hack/kubevirt-cr.yaml > /dev/null
  echo -n " Wait for 60 seconds ... "
  sleep 60
  echo "Done!"
}

function deploy_cdi {
  local VERSION="v1.51.0"
  echo "Deploy cdi ${VERSION} ... "
  ./tools/kubectl apply -f "https://github.com/kubevirt/containerized-data-importer/releases/download/${VERSION}/cdi-operator.yaml" > /dev/null
  sleep 3
  ./tools/kubectl apply -f hack/cdi-cr.yaml > /dev/null
  echo -n " Wait for 60 seconds ... "
  sleep 60
  echo "Done!"
}

function deploy_metallb {
  local VERSION="v0.12.1"
  echo -n "Deploy metallb ${VERSION} ... "
  local NETWORK=`docker network inspect kind | jq '.[].IPAM.Config[0].Gateway' | tr -d '"' | awk -F '.' '{print $1"."$2}'`
  ./tools/kubectl apply -f "https://raw.githubusercontent.com/metallb/metallb/${VERSION}/manifests/namespace.yaml" > /dev/null 
  ./tools/kubectl apply -f "https://raw.githubusercontent.com/metallb/metallb/${VERSION}/manifests/metallb.yaml" > /dev/null 2>&1
  cat <<EOF | ./tools/kubectl apply -f - > /dev/null
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: metallb-system
  name: config
data:
  config: |
    address-pools:
    - name: default
      protocol: layer2
      addresses:
      - ${NETWORK}.255.1-${NETWORK}.255.254
EOF
  echo "Done!"
}

mkdir -p tools
install_kubectl
install_kubebuilder
install_kind
install_kustomize
install_virtctl
install_mc
create_kind_cluster
deploy_calico
deploy_certmanager
deploy_kubevirt
deploy_cdi
deploy_metallb

echo "Deploy kubeberth-operator ... "
make deploy > /dev/null
echo -n " Wait for 90 seconds ... "
sleep 90
echo "Done!"

echo "Deploy minio for Archive Repository ..."
mkdir -p data
docker run -d \
  --network kind \
  -p 9000:9000 \
  -p 9001:9001 \
  --user $(id -u):$(id -g) \
  --name minio \
  -e "MINIO_ROOT_USER=minio" \
  -e "MINIO_ROOT_PASSWORD=miniominio" \
  -v ${PWD}/data:/data \
  quay.io/minio/minio server /data --console-address ":9001"
sleep 3
./tools/mc alias set local http://`docker inspect minio | jq '.[].NetworkSettings.Networks.kind.IPAddress' | tr -d '"'`:9000 minio miniominio
./tools/mc mb local/kubeberth/archives
./tools/mc policy set public local/kubeberth/archives
echo "Done!"

MINIO_IP_ADDRESS=`docker inspect minio | jq '.[].NetworkSettings.Networks.kind.IPAddress' | tr -d '"'`
cat <<EOF | ./tools/kubectl apply -f - > /dev/null
---

apiVersion: v1
kind: Namespace
metadata:
  name: kubeberth

---

apiVersion: berth.kubeberth.io/v1alpha1
kind: KubeBerth
metadata:
  name: kubeberth
  namespace: kubeberth
spec:
  archiveRepositoryURL: "http://${MINIO_IP_ADDRESS}:9000"
  accessKey: "minio"
  secretKey: "miniominio"
  archiveRepositoryTarget: "local/kubeberth/archives"
  storageClassName: standard

---
EOF

./tools/kubectl config set-context $(./tools/kubectl config current-context) --namespace=kubeberth > /dev/null
./tools/kubectl -n kubevirt wait kubevirt kubevirt --for condition=Available
./tools/kubectl -n cdi wait cdi cdi --for condition=Available

echo
echo "ALL DONE!!!"
echo
echo "You need to execute below command for developing kubeberth-operator."
echo "$ export PATH=\$PWD/tools:\$PATH"

exit 0
