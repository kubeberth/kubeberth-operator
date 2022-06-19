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

var _ = Describe("Disk controller", func() {
	const (
		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When updating Disk Status", func() {
		It("Should update Disk Status when a new Disk is created", func() {
			By("By creating a new Disk")
			ctx := context.Background()
			disk := newDisk()
			Expect(k8sClient.Create(ctx, disk)).NotTo(HaveOccurred())

			diskLookupKey := types.NamespacedName{Namespace: "kubeberth-test", Name: "disk-test"}
			createdDisk := &kubeberth.Disk{}
			Eventually(func() error {
				return k8sClient.Get(ctx, diskLookupKey, createdDisk)
			}).Should(Succeed())
			Expect(createdDisk.Spec.Size).Should(Equal("1Gi"))

			By("By checking the Disk has Status.Size")
			Consistently(func() (string, error) {
				err := k8sClient.Get(ctx, diskLookupKey, createdDisk)
				if err != nil || createdDisk.Status.Size == "" {
					return "getting error", err
				}
				return createdDisk.Status.Size, nil
			}, duration, interval).Should(Equal("1Gi"))
		})
	})
})

func newDisk() *kubeberth.Disk {
	return &kubeberth.Disk{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "disk-test",
			Namespace: "kubeberth-test",
		},
		Spec: kubeberth.DiskSpec{
			Size: "1Gi",
		},
	}
}
