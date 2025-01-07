package get

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cnoe-io/idpbuilder/pkg/entity"
	"github.com/cnoe-io/idpbuilder/pkg/printer"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/build"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	argoCDAdminUsername          = "admin"
	argoCDInitialAdminSecretName = "argocd-initial-admin-secret"
	giteaAdminSecretName         = "gitea-credential"
)

var SecretsCmd = &cobra.Command{
	Use:          "secrets",
	Short:        "retrieve secrets from the cluster",
	Long:         ``,
	RunE:         getSecretsE,
	SilenceUsage: true,
}

// well known secrets that are part of the core packages
var (
	corePkgSecrets = map[string][]string{
		"argocd": []string{argoCDInitialAdminSecretName},
		"gitea":  []string{giteaAdminSecretName},
	}
)

func getSecretsE(cmd *cobra.Command, args []string) error {
	ctx, ctxCancel := context.WithCancel(cmd.Context())
	defer ctxCancel()
	kubeConfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	opts := build.NewBuildOptions{
		KubeConfigPath: kubeConfigPath,
		Scheme:         k8s.GetScheme(),
		CancelFunc:     ctxCancel,
	}

	b := build.NewBuild(opts)

	kubeConfig, err := b.GetKubeConfig()
	if err != nil {
		return fmt.Errorf("getting kube config: %w", err)
	}

	kubeClient, err := b.GetKubeClient(kubeConfig)
	if err != nil {
		return fmt.Errorf("getting kube client: %w", err)
	}

	if len(packages) == 0 {
		return printAllPackageSecrets(ctx, os.Stdout, kubeClient, outputFormat)
	}

	return printPackageSecrets(ctx, os.Stdout, kubeClient, outputFormat)
}

func printAllPackageSecrets(ctx context.Context, outWriter io.Writer, kubeClient client.Client, format string) error {
	selector := labels.NewSelector()
	secrets := []entity.Secret{}
	secretPrinter := printer.SecretPrinter{
		Secrets:   secrets,
		OutWriter: outWriter,
	}

	for k, v := range corePkgSecrets {
		for i := range v {
			secret, sErr := getCorePackageSecret(ctx, kubeClient, k, v[i])
			if sErr != nil {
				if errors.IsNotFound(sErr) {
					continue
				}
				return fmt.Errorf("getting secret %s in %s: %w", v[i], k, sErr)
			}
			secrets = append(secrets, populateSecret(secret, true))
		}
	}

	cnoeLabelSecrets, err := getSecretsByCNOELabel(ctx, kubeClient, selector)
	if err != nil {
		return fmt.Errorf("listing secrets: %w", err)
	}

	for i := range cnoeLabelSecrets.Items {
		secrets = append(secrets, populateSecret(cnoeLabelSecrets.Items[i], false))
	}

	if len(secrets) == 0 {
		fmt.Println("no secrets found")
		return nil
	}

	secretPrinter.Secrets = secrets
	return secretPrinter.PrintOutput(format)
}

func printPackageSecrets(ctx context.Context, outWriter io.Writer, kubeClient client.Client, format string) error {
	selector := labels.NewSelector()
	secrets := []entity.Secret{}
	secretPrinter := printer.SecretPrinter{
		OutWriter: outWriter,
	}

	for i := range packages {
		p := packages[i]
		secretNames, ok := corePkgSecrets[p]
		if ok {
			for j := range secretNames {
				secret, sErr := getCorePackageSecret(ctx, kubeClient, p, secretNames[j])
				if sErr != nil {
					if errors.IsNotFound(sErr) {
						continue
					}
					return fmt.Errorf("getting secret %s in %s: %w", secretNames[j], p, sErr)
				}
				secrets = append(secrets, populateSecret(secret, true))
			}
			continue
		}

		req, pErr := labels.NewRequirement(v1alpha1.PackageNameLabelKey, selection.Equals, []string{p})
		if pErr != nil {
			return fmt.Errorf("building requirement for %s: %w", p, pErr)
		}

		pkgSelector := selector.Add(*req)

		cnoeLabelSecrets, err := getSecretsByCNOELabel(ctx, kubeClient, pkgSelector)
		if err != nil {
			return fmt.Errorf("listing secrets: %w", err)
		}

		for i := range cnoeLabelSecrets.Items {
			secrets = append(secrets, populateSecret(cnoeLabelSecrets.Items[i], false))
		}

		if len(secrets) == 0 {
			fmt.Println("no secrets found")
			return nil
		}
	}

	secretPrinter.Secrets = secrets
	return secretPrinter.PrintOutput(format)
}

func generateSecretTable(secretTable []entity.Secret) metav1.Table {
	table := &metav1.Table{}
	table.ColumnDefinitions = []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string"},
		{Name: "Namespace", Type: "string"},
		{Name: "Username", Type: "string"},
		{Name: "Password", Type: "string"},
		{Name: "Token", Type: "string"},
		{Name: "Data", Type: "string"},
	}
	for _, secret := range secretTable {
		var dataEntries []string

		if !secret.IsCore {
			for key, value := range secret.Data {
				dataEntries = append(dataEntries, fmt.Sprintf("%s=%s", key, value))
			}
		}
		dataString := strings.Join(dataEntries, ", ")
		row := metav1.TableRow{
			Cells: []interface{}{
				secret.Name,
				secret.Namespace,
				secret.Username,
				secret.Password,
				secret.Token,
				dataString,
			},
		}
		table.Rows = append(table.Rows, row)
	}
	return *table
}

func populateSecret(s v1.Secret, isCoreSecret bool) entity.Secret {
	secret := entity.Secret{
		Name:      s.Name,
		Namespace: s.Namespace,
	}

	if isCoreSecret {
		secret.IsCore = true
		secret.Username = string(s.Data["username"])
		secret.Password = string(s.Data["password"])
		secret.Token = string(s.Data["token"])
		secret.Data = nil
	} else {
		newData := make(map[string]string)
		for key, value := range s.Data {
			newData[key] = string(value)
		}
		if len(newData) > 0 {
			secret.Data = newData
		}
	}

	return secret
}

func getSecretsByCNOELabel(ctx context.Context, kubeClient client.Client, l labels.Selector) (v1.SecretList, error) {
	req, err := labels.NewRequirement(v1alpha1.CLISecretLabelKey, selection.Equals, []string{v1alpha1.CLISecretLabelValue})
	if err != nil {
		return v1.SecretList{}, fmt.Errorf("building labels with key %s and value %s : %w", v1alpha1.CLISecretLabelKey, v1alpha1.CLISecretLabelValue, err)
	}

	secrets := v1.SecretList{}
	opts := client.ListOptions{
		LabelSelector: l.Add(*req),
		Namespace:     "", // find in all namespaces
	}
	return secrets, kubeClient.List(ctx, &secrets, &opts)
}

func getCorePackageSecret(ctx context.Context, kubeClient client.Client, ns, name string) (v1.Secret, error) {
	s, err := util.GetSecretByName(ctx, kubeClient, ns, name)
	if err != nil {
		return v1.Secret{}, err
	}

	if name == argoCDInitialAdminSecretName && s.Data != nil {
		s.Data["username"] = []byte(argoCDAdminUsername)
	}
	return s, nil
}
