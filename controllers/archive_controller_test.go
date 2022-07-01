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

var _ = Describe("Archive controller", func() {
	const (
		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When updating Archive Status", func() {
		It("Should update Archive Status.URL when a new Archive is created", func() {
			By("By creating a new Archive")
			ctx := context.Background()
			archive := newArchive()
			Expect(k8sClient.Create(ctx, archive)).Should(Succeed())

			archiveLookupKey := types.NamespacedName{Namespace: "kubeberth-test", Name: "archive-test"}
			createdArchive := &kubeberth.Archive{}
			Eventually(func() error {
				return k8sClient.Get(ctx, archiveLookupKey, createdArchive)
			}).Should(Succeed())
			Expect(createdArchive.Spec.Repository).Should(Equal("test"))

			By("By checking the Archive has Status.URL")
			Consistently(func() (string, error) {
				err := k8sClient.Get(ctx, archiveLookupKey, createdArchive)
				if err != nil || createdArchive.Status.Repository == "" {
					return "getting error", err
				}
				return createdArchive.Status.Repository, nil
			}, duration, interval).Should(Equal("test"))
		})
	})
})

func newArchive() *kubeberth.Archive {
	return &kubeberth.Archive{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "berth.kubeberth.io/v1alpha1",
			Kind:       "Archive",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "archive-test",
			Namespace: "kubeberth-test",
		},
		Spec: kubeberth.ArchiveSpec{
			Repository: "test",
		},
	}
}
