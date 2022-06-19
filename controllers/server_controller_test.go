package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	kubeberth "github.com/kubeberth/kubeberth-operator/api/v1alpha1"
	//+kubebuilder:scaffold:imports
)

var _ = Describe("Server controller", func() {
	const (
		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When updating Server Status", func() {
		It("Should update Server Status when a new Server is created", func() {
			By("By creating a new Server")
			ctx := context.Background()
			server := newServer()
			Expect(k8sClient.Create(ctx, server)).NotTo(HaveOccurred())

			serverLookupKey := types.NamespacedName{Namespace: "kubeberth-test", Name: "server-test"}
			createdServer := &kubeberth.Server{}
			Eventually(func() error {
				return k8sClient.Get(ctx, serverLookupKey, createdServer)
			}).Should(Succeed())
			Expect(*createdServer.Spec.CPU).Should(Equal(resource.MustParse("1")))

			Eventually(func() error {
				return k8sClient.Get(ctx, serverLookupKey, createdServer)
			}).Should(Succeed())
			Expect(*createdServer.Spec.Memory).Should(Equal(resource.MustParse("1Gi")))

			Eventually(func() error {
				return k8sClient.Get(ctx, serverLookupKey, createdServer)
			}).Should(Succeed())
			Expect(createdServer.Spec.Hostname).Should(Equal("test"))

			By("By checking the Server has Status.CPU")
			Consistently(func() (string, error) {
				err := k8sClient.Get(ctx, serverLookupKey, createdServer)
				if err != nil || createdServer.Status.CPU == "" {
					return "getting error", err
				}
				return createdServer.Status.CPU, nil
			}, duration, interval).Should(Equal("1"))

			By("By checking the Server has Status.Memory")
			Consistently(func() (string, error) {
				err := k8sClient.Get(ctx, serverLookupKey, createdServer)
				if err != nil || createdServer.Status.Memory == "" {
					return "getting error", err
				}
				return createdServer.Status.Memory, nil
			}, duration, interval).Should(Equal("1Gi"))

			By("By checking the Server has Status.Hostname")
			Consistently(func() (string, error) {
				err := k8sClient.Get(ctx, serverLookupKey, createdServer)
				if err != nil || createdServer.Status.Hostname == "" {
					return "getting error", err
				}
				return createdServer.Status.Hostname, nil
			}, duration, interval).Should(Equal("test"))

		})
	})
})

func newServer() *kubeberth.Server {
	running := true
	cpu := resource.MustParse("1")
	memory := resource.MustParse("1Gi")

	return &kubeberth.Server{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "server-test",
			Namespace: "kubeberth-test",
		},
		Spec: kubeberth.ServerSpec{
			Running:    &running,
			CPU:        &cpu,
			Memory:     &memory,
			MACAddress: "52:42:00:00:00:00",
			Hostname:   "test",
			Disk: &kubeberth.AttachedDisk{
				Namespace: "kubeberth-test",
				Name:      "archive-test",
			},
			CloudInit: &kubeberth.AttachedCloudInit{
				Namespace: "kubeberth-test",
				Name:      "cloudinit-test",
			},
		},
	}
}
