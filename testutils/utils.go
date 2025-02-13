package testutils

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	. "github.com/onsi/gomega"
	promv1a1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func WaitForResource(k8sClient client.Client, ctx context.Context, obj runtime.Object, timeout time.Duration, interval time.Duration) {
	key, err := client.ObjectKeyFromObject(obj)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())

	EventuallyWithOffset(1, func() bool {
		err := k8sClient.Get(ctx, key, obj)
		return err == nil
	}, timeout, interval).Should(BeTrue())
}

func EnsureNoResource(k8sClient client.Client, ctx context.Context, obj runtime.Object, timeout time.Duration, interval time.Duration) {
	key, err := client.ObjectKeyFromObject(obj)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())

	ConsistentlyWithOffset(1, func() bool {
		return errors.IsNotFound((k8sClient.Get(ctx, key, obj)))
	}, timeout, interval).Should(BeTrue())
}

func EnsureNoResources(k8sClient client.Client, ctx context.Context, list []runtime.Object, timeout time.Duration, interval time.Duration) {
	ConsistentlyWithOffset(1, func() bool {
		for i := range list {
			key, err := client.ObjectKeyFromObject(list[i])
			ExpectWithOffset(2, err).ToNot(HaveOccurred())
			if !errors.IsNotFound(k8sClient.Get(ctx, key, list[i])) {
				return false
			}
		}
		return true
	}, timeout, interval).Should(BeTrue())
}

func GetResourceKey(obj runtime.Object) client.ObjectKey {
	key, err := client.ObjectKeyFromObject(obj)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	return key
}

func ResourceHasLabel(k8sClient client.Client, ctx context.Context, obj runtime.Object, labelKey string, labelValue string) bool {
	accessor, err := meta.Accessor(obj)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())

	key := client.ObjectKey{Namespace: accessor.GetNamespace(), Name: accessor.GetName()}
	if err := k8sClient.Get(ctx, key, obj); err != nil {
		return false
	}

	value, ok := accessor.GetLabels()[labelKey]
	return ok && value == labelValue
}

func ProbeReadiness() (int, error) {
	resp, err := http.Get("http://localhost:8081/readyz")
	if err != nil {
		return 0, err
	}
	return resp.StatusCode, nil
}

func ToJsonOrDie(value interface{}) []byte {
	if bytes, err := json.Marshal(value); err == nil {
		return bytes
	} else {
		panic(err)
	}
}

func WaitForAlertManagerSMTPReceiverEmailConfigToUpdate(
	k8sClient client.Client,
	ctx context.Context,
	amconfigKey client.ObjectKey,
	emailadresses []string,
	receiverName string,
	timeout time.Duration,
	interval time.Duration,
) {
	EventuallyWithOffset(1, func() string {
		amconfig := &promv1a1.AlertmanagerConfig{}
		ExpectWithOffset(2, k8sClient.Get(ctx, amconfigKey, amconfig)).Should(Succeed())
		for i := range amconfig.Spec.Receivers {
			reciever := &amconfig.Spec.Receivers[i]
			if reciever.Name == receiverName {
				if len(reciever.EmailConfigs) > 0 {
					return reciever.EmailConfigs[0].To
				}
				return ""
			}
		}
		return ""
	}, timeout, interval).Should(Equal(strings.Join(emailadresses, ", ")))
}
