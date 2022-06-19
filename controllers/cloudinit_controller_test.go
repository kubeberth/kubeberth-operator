package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	kubeberth "github.com/kubeberth/kubeberth-operator/api/v1alpha1"
	//+kubebuilder:scaffold:imports
)

var _ = Describe("CloudInit controller", func() {
	const (
		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When updating CloudInit Status", func() {
		It("Should update CloudInit Status when a new CloudInit is created", func() {
			By("By creating a new CloudInit")
			ctx := context.Background()
			cloudinit := newCloudInit()
			Expect(k8sClient.Create(ctx, cloudinit)).NotTo(HaveOccurred())

			cloudinitLookupKey := types.NamespacedName{Namespace: "kubeberth-test", Name: "cloudinit-test"}
			createdCloudInit := &kubeberth.CloudInit{}
			Eventually(func() error {
				return k8sClient.Get(ctx, cloudinitLookupKey, createdCloudInit)
			}).Should(Succeed())
			Expect(createdCloudInit.Spec.UserData).Should(Equal("userdata-test"))
			Expect(createdCloudInit.Spec.NetworkData).Should(Equal("networkdata-test"))

			By("By checking the CloudInit has Status.UserData")
			Consistently(func() (string, error) {
				err := k8sClient.Get(ctx, cloudinitLookupKey, createdCloudInit)
				if err != nil || createdCloudInit.Status.UserData == "" {
					return "getting error", err
				}
				return createdCloudInit.Status.UserData, nil
			}, duration, interval).Should(Equal("userdata-test"))

			By("By checking the CloudInit has Status.NetworkData")
			Consistently(func() (string, error) {
				err := k8sClient.Get(ctx, cloudinitLookupKey, createdCloudInit)
				if err != nil || createdCloudInit.Status.NetworkData == "" {
					return "getting error", err
				}
				return createdCloudInit.Status.NetworkData, nil
			}, duration, interval).Should(Equal("networkdata-test"))
		})
	})
})

func newCloudInit() *kubeberth.CloudInit {
	return &kubeberth.CloudInit{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cloudinit-test",
			Namespace: "kubeberth-test",
		},
		Spec: kubeberth.CloudInitSpec{
			UserData:    "userdata-test",
			NetworkData: "networkdata-test",
		},
	}
}
