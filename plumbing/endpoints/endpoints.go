package endpoints

import (
	"context"
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
)

func FromSecret(ctx context.Context, rc *rest.Config, name, namespace string) (Endpoint, error) {
	cli, err := newSecretsRESTClient(rc)
	if err != nil {
		return Endpoint{}, err
	}

	sec, err := getSecret(ctx, clientOptions{
		cli:       cli,
		name:      name,
		namespace: namespace,
	})
	if err != nil {
		return Endpoint{}, err
	}

	res := Endpoint{}
	if v, ok := sec.Data[serverUrlLabel]; ok {
		res.ServerURL = string(v)
	} else {
		return res, fmt.Errorf("missed required attribute for endpoint: server-url")
	}

	if v, ok := sec.Data[proxyUrlLabel]; ok {
		res.ProxyURL = string(v)
	}

	if v, ok := sec.Data[tokenLabel]; ok {
		res.Token = string(v)
	}

	if v, ok := sec.Data[usernameLabel]; ok {
		res.Username = string(v)
	}

	if v, ok := sec.Data[passwordLabel]; ok {
		res.Password = string(v)
	}

	if v, ok := sec.Data[caLabel]; ok {
		res.CertificateAuthorityData = string(v)
	}

	if v, ok := sec.Data[clientKeyLabel]; ok {
		res.ClientKeyData = string(v)
	}

	if v, ok := sec.Data[clientCertLabel]; ok {
		res.ClientCertificateData = string(v)
	}

	if v, ok := sec.Data[debugLabel]; ok {
		res.Debug, _ = strconv.ParseBool(string(v))
	}

	return res, nil
}

func Store(ctx context.Context, rc *rest.Config, ns string, ep Endpoint) error {
	sec := corev1.Secret{}
	sec.SetName(fmt.Sprintf(secretNameFmt, ep.Username))
	sec.SetNamespace(ns)
	sec.StringData = map[string]string{
		usernameLabel:   ep.Username,
		passwordLabel:   ep.Password,
		tokenLabel:      ep.Token,
		caLabel:         ep.CertificateAuthorityData,
		clientCertLabel: ep.ClientCertificateData,
		clientKeyLabel:  ep.ClientKeyData,
		serverUrlLabel:  ep.ServerURL,
		proxyUrlLabel:   ep.ProxyURL,
		debugLabel:      strconv.FormatBool(ep.Debug),
	}

	cli, err := newSecretsRESTClient(rc)
	if err != nil {
		return err
	}
	err = createSecret(ctx, &sec, clientOptions{
		cli:       cli,
		name:      sec.Name,
		namespace: ns,
	})
	if err == nil {
		return nil
	}

	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
	}

	return updateSecret(ctx, &sec, clientOptions{
		cli:       cli,
		name:      sec.Name,
		namespace: ns,
	})
}

const (
	clientCertLabel = "client-certificate-data"
	clientKeyLabel  = "client-key-data"
	caLabel         = "certificate-authority-data"
	proxyUrlLabel   = "proxy-url"
	serverUrlLabel  = "server-url"
	debugLabel      = "debug"
	passwordLabel   = "password"
	usernameLabel   = "username"
	tokenLabel      = "token"

	secretNameFmt = "%s-clientconfig"
)
