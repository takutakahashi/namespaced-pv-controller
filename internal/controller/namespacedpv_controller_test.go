package controller

import (
	"context"
	"time"

	namespacedpvv1 "github.com/homirun/namespaced-pv-controller/api/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("namespaced pv controller", func() {

	ctx := context.Background()
	var stopFunc func()

	BeforeEach(func() {
		err := k8sClient.DeleteAllOf(ctx, &namespacedpvv1.NamespacedPv{}, client.InNamespace("test"))
		Expect(err).NotTo(HaveOccurred())
		pvs := &corev1.PersistentVolumeList{}
		err = k8sClient.List(ctx, pvs, client.InNamespace("test"))
		Expect(err).NotTo(HaveOccurred())
		for _, pv := range pvs.Items {
			err = k8sClient.Delete(ctx, &pv)
			Expect(err).NotTo(HaveOccurred())
		}
		time.Sleep(100 * time.Millisecond)

		mgr, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme.Scheme,
		})
		Expect(err).NotTo(HaveOccurred())

		reconciler := &NamespacedPvReconciler{
			Client: k8sClient,
			Scheme: scheme.Scheme,
		}
		err = reconciler.SetupWithManager(mgr)
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(ctx)
		stopFunc = cancel
		go func() {
			err := mgr.Start(ctx)
			if err != nil {
				panic(err)
			}
		}()
		time.Sleep(100 * time.Millisecond)

	})

	AfterEach(func() {
		stopFunc()
		time.Sleep(100 * time.Millisecond)
	})

	It("should create PersistentVolume", func() {
		namespacedPv := newNamespacedPv()
		err := k8sClient.Create(ctx, namespacedPv)
		Expect(err).NotTo(HaveOccurred())

		pv := corev1.PersistentVolume{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: "test", Name: "test-pv-test"}, &pv)
		}).Should(Succeed())
	})

})

func newNamespacedPv() *namespacedpvv1.NamespacedPv {
	return &namespacedpvv1.NamespacedPv{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-namespace",
			Namespace: "test",
		},
		Spec: namespacedpvv1.NamespacedPvSpec{
			VolumeName:       "test-pv",
			StorageClassName: "test-storageclass",
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("1Gi"),
			},
			Nfs: namespacedpvv1.NFS{
				Server:   "192.0.0.1",
				Path:     "/data/share",
				ReadOnly: false,
			},
			ReclaimPolicy: corev1.PersistentVolumeReclaimRetain,
			MountOptions:  "nolock,vers=4.1",
		},
	}
}
