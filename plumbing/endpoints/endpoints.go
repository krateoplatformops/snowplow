package endpoints

import (
	"context"
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

func FromSecret(ctx context.Context, rc *rest.Config, name, namespace string) (Endpoint, error) {
	cli, err := newSecretsRESTClient(rc)
	if err != nil {
		return Endpoint{}, err
	}

	sec, err := getSecret(ctx, getSecretOptions{
		cli:       cli,
		name:      name,
		namespace: namespace,
	})
	if err != nil {
		return Endpoint{}, err
	}

	res := Endpoint{}
	if v, ok := sec.Data["server-url"]; ok {
		res.ServerURL = string(v)
	} else {
		return res, fmt.Errorf("missed required attribute for endpoint: server-url")
	}

	if v, ok := sec.Data["proxy-url"]; ok {
		res.ProxyURL = string(v)
	}

	if v, ok := sec.Data["token"]; ok {
		res.Token = string(v)
	}

	if v, ok := sec.Data["username"]; ok {
		res.Username = string(v)
	}

	if v, ok := sec.Data["password"]; ok {
		res.Password = string(v)
	}

	if v, ok := sec.Data["certificate-authority-data"]; ok {
		res.CertificateAuthorityData = string(v)
	}

	if v, ok := sec.Data["client-key-data"]; ok {
		res.ClientKeyData = string(v)
	}

	if v, ok := sec.Data["client-certificate-data"]; ok {
		res.ClientCertificateData = string(v)
	}

	if v, ok := sec.Data["debug"]; ok {
		res.Debug, _ = strconv.ParseBool(string(v))
	}

	return res, nil
}

func newSecretsRESTClient(rc *rest.Config) (*rest.RESTClient, error) {
	gv := schema.GroupVersion{
		Group:   "",
		Version: "v1",
	}

	sb := runtime.NewSchemeBuilder(
		func(reg *runtime.Scheme) error {
			reg.AddKnownTypes(
				gv,
				&corev1.Secret{},
				&corev1.SecretList{},
				&metav1.ListOptions{},
				&metav1.GetOptions{},
				&metav1.DeleteOptions{},
				&metav1.CreateOptions{},
				&metav1.UpdateOptions{},
				&metav1.PatchOptions{},
				&metav1.Status{},
			)
			return nil
		})

	s := runtime.NewScheme()
	sb.AddToScheme(s)

	config := *rc
	config.APIPath = "/api"
	config.GroupVersion = &gv
	config.NegotiatedSerializer = serializer.NewCodecFactory(s).
		WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	cli, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	//pc := runtime.NewParameterCodec(s)

	return cli, nil
}

type getSecretOptions struct {
	cli       *rest.RESTClient
	name      string
	namespace string
}

func getSecret(ctx context.Context, opts getSecretOptions) (result *corev1.Secret, err error) {
	result = &corev1.Secret{}
	err = opts.cli.Get().
		Namespace(opts.namespace).
		Resource("secrets").
		Name(opts.name).
		Do(ctx).
		Into(result)
	return
}
