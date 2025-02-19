package util

import (
	"code.gitea.io/sdk/gitea"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/util/idp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const (
	// hardcoded values from what we have in the yaml installation file.
	GiteaNamespace           = "gitea"
	GiteaAdminSecret         = "gitea-credential"
	GiteaAdminName           = "giteaAdmin"
	GiteaAdminTokenName      = "admin"
	GiteaAdminTokenFieldName = "token"
	// this is the URL accessible outside cluster. resolves to localhost
	GiteaIngressURL = "%s://gitea.cnoe.localtest.me:%s"
	// this is the URL accessible within cluster for ArgoCD to fetch resources.
	// resolves to cluster ip
	GiteaSvcURL = "%s://%s%s:%s%s"
)

func GiteaAdminSecretObject() corev1.Secret {
	return corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      GiteaAdminSecret,
			Namespace: GiteaNamespace,
		},
	}
}

func PatchPasswordSecret(ctx context.Context, kubeClient client.Client, config v1alpha1.BuildCustomizationSpec, ns string, secretName string, username string, pass string) error {
	sec, err := GetSecretByName(ctx, kubeClient, ns, secretName)
	if err != nil {
		return fmt.Errorf("getting secret to patch fails: %w", err)
	}
	u := unstructured.Unstructured{}
	u.SetName(sec.GetName())
	u.SetNamespace(sec.GetNamespace())
	u.SetGroupVersionKind(sec.GetObjectKind().GroupVersionKind())

	err = unstructured.SetNestedField(u.Object, base64.StdEncoding.EncodeToString([]byte(pass)), "data", "password")
	if err != nil {
		return fmt.Errorf("setting password field: %w", err)
	}

	if strings.Contains(secretName, "gitea") {
		// We should recreate a token as user/password changed
		giteaUrl, err := GiteaBaseUrl(ctx)
		if err != nil {
			return fmt.Errorf("getting giteaurl: %w", err)
		}

		t, err := GetGiteaToken(ctx, giteaUrl, string(username), string(pass))
		if err != nil {
			return fmt.Errorf("getting gitea token: %w", err)
		}

		token := base64.StdEncoding.EncodeToString([]byte(t))
		err = unstructured.SetNestedField(u.Object, token, "data", GiteaAdminTokenFieldName)
		if err != nil {
			return fmt.Errorf("setting gitea token field: %w", err)
		}
	}

	return kubeClient.Patch(ctx, &u, client.Apply, client.ForceOwnership, client.FieldOwner(v1alpha1.FieldManager))
}

func GetGiteaToken(ctx context.Context, baseUrl, username, password string) (string, error) {
	giteaClient, err := gitea.NewClient(baseUrl, gitea.SetHTTPClient(GetHttpClient()),
		gitea.SetBasicAuth(username, password), gitea.SetContext(ctx),
	)
	if err != nil {
		return "", fmt.Errorf("creating gitea client: %w", err)
	}
	tokens, resp, err := giteaClient.ListAccessTokens(gitea.ListAccessTokensOptions{})
	if err != nil {
		return "", fmt.Errorf("listing gitea access tokens. status: %s error : %w", resp.Status, err)
	}

	for i := range tokens {
		if tokens[i].Name == GiteaAdminTokenName {
			resp, err := giteaClient.DeleteAccessToken(tokens[i].ID)
			if err != nil {
				return "", fmt.Errorf("deleting gitea access tokens. status: %s error : %w", resp.Status, err)
			}
			break
		}
	}

	token, resp, err := giteaClient.CreateAccessToken(gitea.CreateAccessTokenOption{
		Name: GiteaAdminTokenName,
		Scopes: []gitea.AccessTokenScope{
			gitea.AccessTokenScopeAll,
		},
	})
	if err != nil {
		return "", fmt.Errorf("deleting gitea access tokens. status: %s error : %w", resp.Status, err)
	}

	return token.Token, nil
}

func GiteaBaseUrl(ctx context.Context) (string, error) {
	idpConfig, err := idp.GetConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("error fetching idp config: %s", err)
	}
	return fmt.Sprintf(GiteaIngressURL, idpConfig.Protocol, idpConfig.Port), nil
}
