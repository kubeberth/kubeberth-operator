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

var _ = Describe("KubeBerth controller", func() {
	const (
		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When updating KubeBerth Status", func() {
		It("Should update KubeBerth Status when a new KubeBerth is created", func() {
			By("By creating a new KubeBerth")
			ctx := context.Background()
			kb := newKubeBerth()
			Expect(k8sClient.Create(ctx, kb)).NotTo(HaveOccurred())

			kubeberthLookupKey := types.NamespacedName{Namespace: "kubeberth-test", Name: "kubeberth-test"}
			createdKubeBerth := &kubeberth.KubeBerth{}
			Eventually(func() error {
				return k8sClient.Get(ctx, kubeberthLookupKey, createdKubeBerth)
			}).Should(Succeed())
			Expect(createdKubeBerth.Spec.StorageClassName).Should(Equal("test"))

			By("By checking the KubeBerth has Status.StorageClass")
			Consistently(func() (string, error) {
				err := k8sClient.Get(ctx, kubeberthLookupKey, createdKubeBerth)
				if err != nil || createdKubeBerth.Status.StorageClass == "" {
					return "getting error", err
				}
				return createdKubeBerth.Status.StorageClass, nil
			}, duration, interval).Should(Equal("test"))
		})
	})
})

func newKubeBerth() *kubeberth.KubeBerth {
	return &kubeberth.KubeBerth{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kubeberth-test",
			Namespace: "kubeberth-test",
		},
		Spec: kubeberth.KubeBerthSpec{
			StorageClassName: "test",
		},
	}
}
