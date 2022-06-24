# kubeberth-operator

[KubeVirt](https://kubevirt.io/)のカスタムリソースをラップしてクラウドリソースっぽいものをつくるオペレーター。



## 動作環境(検証済み)
- Kubernetes v1.23.1
- [KubeVirt](https://github.com/kubevirt/kubevirt) v0.51.0

  以下のfeatureGatesを有効にしてください
  - BlockVolume
  - DataVolumes
  - ExpandDisks
  - HotplugVolumes
  - LiveMigration

- [Containerized Data Importer](https://github.com/kubevirt/containerized-data-importer) v1.38.1
- Dynamic Provisioning用のStorageClass
- [cert-manager](https://github.com/cert-manager/cert-manager)
- [external-dns](https://github.com/kubernetes-sigs/external-dns) (任意)



## 構築手順

オペレーターをデプロイします。

```
$ git clone https://github.com/kubeberth/kubeberth-operator.git
$ cd kubeberth-operator
$ make deploy
```

namespaceと設定のリソースを作成します。

```
$ cat > config.yaml <<EOF
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
  # [要修正]
  # Dynamic Provisioning用のStorageClassを記載
  storageClassName: "longhorn"
  # [要修正]
  # external-dnsの環境がある場合、利用したいドメインを記載
  externalDNSDomainName: "kubeberth.k8s.arpa"

---
EOF
$ kubectl apply -f config.yaml
```

これで準備完了です。

```
$ kubectl -n kubeberth get kubeberth
NAME        STORAGECLASS   EXTERNALDNSDOMAIN
kubeberth   longhorn       kubeberth.k8s.arpa
```



## リソース

### アーカイブ

```
$ cat config/samples/berth_v1alpha1_archive.yaml 
apiVersion: berth.kubeberth.io/v1alpha1
kind: Archive
metadata:
  name: archive-sample
  namespace: kubeberth
spec:
  # Local ISO Image Repository
  url: "http://minio.home.arpa:9000/kubevirt/images/ubuntu-20.04-server-cloudimg-arm64.img"
  # Public ISO Image Repository
  # url: "https://cloud-images.ubuntu.com/focal/current/focal-server-cloudimg-arm64.img"
```


### cloud-init

```
$ cat config/samples/berth_v1alpha1_cloudinit.yaml 
apiVersion: berth.kubeberth.io/v1alpha1
kind: CloudInit
metadata:
  name: cloudinit-sample
  namespace: kubeberth
spec:
  userData: |
    #cloud-config
    timezone: Asia/Tokyo
    ssh_pwauth: True
    password: ubuntu
    chpasswd: { expire: False }
    disable_root: false
    #ssh_authorized_keys:
    #- ssh-rsa XXXXXXXXXXXXXXXXXXXXXXXXX
```

### ディスク

```
$ cat config/samples/berth_v1alpha1_disk.yaml 
apiVersion: berth.kubeberth.io/v1alpha1
kind: Disk
metadata:
  name: disk-sample
  namespace: kubeberth
spec:
  size: 32Gi
  source:
    archive:
      name: archive-sample
      namespace: kubeberth
```


### サーバ

```
$ cat config/samples/berth_v1alpha1_server.yaml 
apiVersion: berth.kubeberth.io/v1alpha1
kind: Server
metadata:
  name: server-sample
  namespace: kubeberth
spec:
  running: true
  cpu: 2
  memory: 2Gi
  macAddress: 52:42:00:c7:90:10
  hostname: ubuntu-server
  disk:
    name: disk-sample
    namespace: kubeberth
  cloudInit:
    name: cloudinit-sample
    namespace: kubeberth
```



## 開発環境

- go version v1.17.6
- kubebuilder v3.3.0