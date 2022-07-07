#!/bin/bash

set -eu
OS=$(go env GOOS)
ARCH=$(go env GOARCH)

function install_kubectl {
  local VERSION="v1.23.1"
  local DOWNLOAD_URL="https://dl.k8s.io/release/$VERSION/bin/$OS/$ARCH/kubectl"
  echo -n "Install kubectl $VERSION ... "
  curl -sfLO "$DOWNLOAD_URL"
  chmod +x ./kubectl
  mv ./kubectl tools
  echo "Done!"
}

function install_kubebuilder {
    local VERSION="v3.3.0"
    local FILENAME="kubebuilder_${OS}_$ARCH"
    local DOWNLOAD_URL="https://github.com/kubernetes-sigs/kubebuilder/releases/download/$VERSION/$FILENAME"
    echo -n "Install kubebuilder $VERSION ... "
    curl -sL -o kubebuilder "$DOWNLOAD_URL"
    chmod +x ./kubebuilder
    mv ./kubebuilder tools
    echo "Done!"
}

function install_kind {
    local VERSION="v0.14.0"
    local FILENAME="kind-${OS}-$ARCH"
    local DOWNLOAD_URL="https://kind.sigs.k8s.io/dl/$VERSION/$FILENAME"
    echo -n "Install kind $VERSION ... "
    curl -sL -o kind "$DOWNLOAD_URL"
    chmod +x ./kind
    mv ./kind tools
    echo "Done!"
}

function install_clusterctl {
    local VERSION="v1.1.4"
    local FILENAME="clusterctl-${OS}-$ARCH"
    local DOWNLOAD_URL="https://github.com/kubernetes-sigs/cluster-api/releases/download/$VERSION/$FILENAME"
    echo -n "Install clusterctl $VERSION ... "
    curl -sL -o clusterctl "$DOWNLOAD_URL"
    chmod +x ./clusterctl
    mv ./clusterctl tools
    echo "Done!"
}

function install_kustomize {
    local VERSION="v4.5.2"
    local TARNAME="kustomize_${VERSION}_${OS}_${ARCH}.tar.gz"
    local DOWNLOAD_URL="https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/$VERSION/$TARNAME"
    echo -n "Install kustomize $VERSION ... "
    curl -sL -o "$TARNAME" "$DOWNLOAD_URL"
    tar xf "$TARNAME"
    rm "$TARNAME"
    chmod +x ./kustomize
    mv ./kustomize tools
    echo "Done!"
}

mkdir -p tools
install_kubectl
install_kubebuilder
install_kind
install_clusterctl
install_kustomize
echo
echo "ALL DONE!"
echo "You need to execute below command for developing kubeberth-operator."
echo "$ export PATH=\$PWD/tools:\$PATH"

exit 0
