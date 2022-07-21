#!/bin/bash

if [ $0 != "./scripts/devenv.sh" ]; then
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

mkdir -p tools
install_kubectl
install_kubebuilder
install_kind
install_kustomize
install_virtctl
install_mc

echo
echo "ALL DONE!!!"
echo
echo "You need to execute below command for developing kubeberth-operator."
echo "$ export PATH=\$PWD/tools:\$PATH"

exit 0
