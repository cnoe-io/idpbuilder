package build

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/cnoe-io/idpbuilder/globals"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	certificateOrgName = "cnoe.io"
)

var (
	certificateValidLength = time.Hour * 8766 // one year
)

func createIngressCertificateSecret(ctx context.Context, kubeClient client.Client, cert []byte) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      globals.SelfSignedCertCMName,
			Namespace: corev1.NamespaceDefault,
		},
		Data: map[string][]byte{
			globals.SelfSignedCertCMKeyName: cert,
		},
	}
	err := kubeClient.Create(ctx, secret)
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return nil
		}
		return fmt.Errorf("creating configmap for certificate: %w", err)
	}
	return nil
}

func getIngressCertificateAndKey(ctx context.Context, kubeClient client.Client, name, namespace string) ([]byte, []byte, error) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: corev1.SecretTypeTLS,
	}

	err := kubeClient.Get(ctx, client.ObjectKeyFromObject(secret), secret)
	if err != nil {
		return nil, nil, err
	}
	cert, ok := secret.Data[corev1.TLSCertKey]
	if !ok {
		return nil, nil, fmt.Errorf("key %s not found in secret %s", corev1.TLSCertKey, name)
	}
	privateKey, ok := secret.Data[corev1.TLSPrivateKeyKey]
	if !ok {
		return nil, nil, fmt.Errorf("key %s not found in secret %s", corev1.TLSPrivateKeyKey, name)
	}

	return cert, privateKey, nil
}

func getOrCreateIngressCertificateAndKey(ctx context.Context, kubeClient client.Client, name, namespace string, sans []string) ([]byte, []byte, error) {
	c, p, err := getIngressCertificateAndKey(ctx, kubeClient, name, namespace)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			cert, privateKey, cErr := createSelfSignedCertificate(sans)
			if cErr != nil {
				return nil, nil, cErr
			}

			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Type: corev1.SecretTypeTLS,
				StringData: map[string]string{
					corev1.TLSPrivateKeyKey: string(privateKey),
					corev1.TLSCertKey:       string(cert),
				},
			}
			cErr = kubeClient.Create(ctx, secret)
			if cErr != nil {
				return nil, nil, fmt.Errorf("creating secret %s: %w", secret.Name, err)
			}
			return cert, privateKey, nil
		} else {
			return nil, nil, fmt.Errorf("getting secret %s: %w", name, err)
		}
	}
	return c, p, nil
}

func createSelfSignedCertificate(sans []string) ([]byte, []byte, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generating private key: %w", err)
	}

	keyUsage := x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign
	notBefore := time.Now()
	notAfter := notBefore.Add(certificateValidLength)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("generating certificate serial number: %w", err)
	}

	cert := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{certificateOrgName},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		DNSNames:              sans,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &cert, &cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("creating certificate: %w", err)
	}

	var certB bytes.Buffer
	var keyB bytes.Buffer
	err = pem.Encode(io.Writer(&certB), &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	if err != nil {
		return nil, nil, fmt.Errorf("encoding cert: %w", err)
	}

	certOut, err := io.ReadAll(&certB)
	if err != nil {
		return nil, nil, fmt.Errorf("reading buffer: %w", err)
	}

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal private key: %w", err)
	}

	err = pem.Encode(io.Writer(&keyB), &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes})
	if err != nil {
		return nil, nil, fmt.Errorf("encoding private key: %w", err)
	}
	privateKeyOut, err := io.ReadAll(&keyB)
	if err != nil {
		return nil, nil, fmt.Errorf("reading buffer: %w", err)
	}

	return certOut, privateKeyOut, nil
}

func setupSelfSignedCertificate(ctx context.Context, logger logr.Logger, kubeclient client.Client, config util.CorePackageTemplateConfig) ([]byte, error) {
	if err := k8s.EnsureNamespace(ctx, kubeclient, globals.NginxNamespace); err != nil {
		return nil, err
	}

	sans := []string{
		globals.DefaultHostName,
		globals.DefaultSANWildcard,
	}
	if config.Host != globals.DefaultHostName {
		sans = []string{
			config.Host,
			fmt.Sprintf("*.%s", config.Host),
		}
	}

	logger.V(1).Info("Creating/getting certificate", "host", config.Host, "sans", sans)
	cert, _, err := getOrCreateIngressCertificateAndKey(ctx, kubeclient, globals.SelfSignedCertSecretName, globals.NginxNamespace, sans)
	if err != nil {
		return nil, err
	}

	logger.V(1).Info("Creating secret for certificate", "host", config.Host)
	err = createIngressCertificateSecret(ctx, kubeclient, cert)
	if err != nil {
		return nil, err
	}
	return cert, nil
}
