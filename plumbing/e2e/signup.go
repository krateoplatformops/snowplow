package e2e

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"hash/fnv"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/krateoplatformops/snowplow/plumbing/endpoints"
	"github.com/krateoplatformops/snowplow/plumbing/kubeutil"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type signupHandler struct {
	caData       string
	proxyURL     string
	serverURL    string
	certDuration time.Duration
	restconfig   *rest.Config
	namespace    string
}

func (g *signupHandler) SignUp(user string, groups []string) (endpoints.Endpoint, error) {
	if len(g.caData) == 0 {
		caCrt, err := kubeutil.CACrt(context.Background(), g.restconfig)
		if err != nil {
			return endpoints.Endpoint{}, err
		}
		g.caData = caCrt
	}

	ep, err := g.generateEndpoint(user, groups)
	if err != nil {
		return endpoints.Endpoint{}, err
	}
	ep.Username = user

	err = endpoints.Store(context.TODO(), g.restconfig, g.namespace, ep)
	return ep, err
}

func (g *signupHandler) generateEndpoint(user string, groups []string) (ep endpoints.Endpoint, err error) {
	if len(g.serverURL) == 0 {
		host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
		if len(host) == 0 || len(port) == 0 {
			return ep, rest.ErrNotInCluster
		}
		g.serverURL = "https://" + net.JoinHostPort(host, port)
	}

	cli, err := kubernetes.NewForConfig(g.restconfig)
	if err != nil {
		return ep, err
	}

	cert, key, err := generateClientCertAndKey(cli, generateClientCertAndKeyOpts{
		userID:   mkID(fmt.Sprintf("%s@%s", user, strings.Join(groups, ","))),
		username: user,
		groups:   groups,
		duration: g.certDuration,
	})
	if err != nil {
		return ep, err
	}

	ep.ServerURL = g.serverURL
	ep.CertificateAuthorityData = g.caData
	ep.ClientCertificateData = cert
	ep.ClientKeyData = key

	return
}

type generateClientCertAndKeyOpts struct {
	duration time.Duration
	userID   string
	username string
	groups   []string
}

func generateClientCertAndKey(client kubernetes.Interface, o generateClientCertAndKeyOpts) (string, string, error) {
	key, err := newPrivateKey()
	if err != nil {
		return "", "", err
	}

	req, err := newCertificateRequest(key, o.username, o.groups)
	if err != nil {
		return "", "", err
	}

	csr := newCertificateSigningRequest(req, o.duration, o.userID, o.username)

	err = createCertificateSigningRequests(client, csr)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return "", "", fmt.Errorf("creating CSR kubernetes object: %w", err)
		}

		if err := deleteCertificateSigningRequest(client, csr.Name); err != nil {
			return "", "", fmt.Errorf("deleting existing CSR kubernetes object: %w", err)
		}

		if err := createCertificateSigningRequests(client, csr); err != nil {
			return "", "", fmt.Errorf("creating CSR kubernetes object: %w", err)
		}
	}

	err = approveCertificateSigningRequest(client, csr)
	if err != nil {
		return "", "", err
	}

	err = waitForCertificate(client, csr.Name)
	if err != nil {
		return "", "", err
	}

	crt, err := certificate(client, csr.Name)
	if err != nil {
		return "", "", err
	}

	crtStr := base64.StdEncoding.EncodeToString(crt)
	keyStr := base64.StdEncoding.EncodeToString(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}))
	return crtStr, keyStr, nil
}

func mkID(in string) string {
	hash := fnv.New64a()
	hash.Write([]byte(in))
	return strconv.FormatUint(hash.Sum64(), 16)
}
