package get

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/build"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/util/homedir"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const (
	secretTemplatePath           = "templates/secrets.tmpl"
	argoCDAdminUsername          = "admin"
	argoCDInitialAdminSecretName = "argocd-initial-admin-secret"
	giteaAdminSecretName         = "gitea-credential"
)

//go:embed templates
var templates embed.FS

var SecretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "retrieve secrets from the cluster",
	Long:  ``,
	RunE:  getSecretsE,
}

// well known secrets that are part of the core packages
var corePkgSecrets = map[string][]string{
	"argocd": []string{argoCDInitialAdminSecretName},
	"gitea":  []string{giteaAdminSecretName},
}

type TemplateData struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Data      map[string]string `json:"data"`
}

func getSecretsE(cmd *cobra.Command, args []string) error {
	ctx, ctxCancel := context.WithCancel(ctrl.SetupSignalHandler())
	defer ctxCancel()
	kubeConfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	opts := build.NewBuildOptions{}
	opts.KubeConfigPath = kubeConfigPath
	opts.Scheme = k8s.GetScheme()
	opts.CancelFunc = ctxCancel

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
	secretsToPrint := make([]any, 0, 2)

	for k, v := range corePkgSecrets {
		for i := range v {
			secret, sErr := getCorePackageSecret(ctx, kubeClient, k, v[i])
			if sErr != nil {
				if errors.IsNotFound(sErr) {
					continue
				}
				return fmt.Errorf("getting secret %s in %s: %w", v[i], k, sErr)
			}
			secretsToPrint = append(secretsToPrint, secretToTemplateData(secret))
		}
	}

	secrets, err := getSecretsByCNOELabel(ctx, kubeClient, selector)
	if err != nil {
		return fmt.Errorf("listing secrets: %w", err)
	}

	for i := range secrets.Items {
		secretsToPrint = append(secretsToPrint, secretToTemplateData(secrets.Items[i]))
	}

	if len(secretsToPrint) == 0 {
		fmt.Println("no secrets found")
		return nil
	}
	return printOutput(secretTemplatePath, outWriter, secretsToPrint, format)
}

func printPackageSecrets(ctx context.Context, outWriter io.Writer, kubeClient client.Client, format string) error {
	selector := labels.NewSelector()
	secretsToPrint := make([]any, 0, 2)

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
				secretsToPrint = append(secretsToPrint, secretToTemplateData(secret))
			}
			continue
		}

		req, pErr := labels.NewRequirement(v1alpha1.PackageNameLabelKey, selection.Equals, []string{p})
		if pErr != nil {
			return fmt.Errorf("building requirement for %s: %w", p, pErr)
		}

		pkgSelector := selector.Add(*req)

		secrets, pErr := getSecretsByCNOELabel(ctx, kubeClient, pkgSelector)
		if pErr != nil {
			return fmt.Errorf("listing secrets: %w", pErr)
		}

		for j := range secrets.Items {
			secretsToPrint = append(secretsToPrint, secretToTemplateData(secrets.Items[j]))
		}
	}

	return printOutput(secretTemplatePath, outWriter, secretsToPrint, format)
}

func renderTemplate(templatePath string, outWriter io.Writer, data []any) error {
	tmpl, err := templates.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	t, err := template.New("secrets").Parse(string(tmpl))
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}
	for i := range data {
		tErr := t.Execute(outWriter, data[i])
		if tErr != nil {
			return fmt.Errorf("executing template for data %s : %w", data[i], tErr)
		}
	}
	return nil
}

func printOutput(templatePath string, outWriter io.Writer, data []any, format string) error {
	switch format {
	case "json":
		b, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
		b = append(b, []byte("\n")...)
		_, err = outWriter.Write(b)
		return err
	case "yaml":
		b, err := yaml.Marshal(data)
		if err != nil {
			return err
		}
		_, err = outWriter.Write(b)
		return err
	case "":
		return renderTemplate(templatePath, outWriter, data)
	default:
		return fmt.Errorf("output format %s is not supported", format)
	}
}

func secretToTemplateData(s v1.Secret) TemplateData {
	data := TemplateData{
		Name:      s.Name,
		Namespace: s.Namespace,
		Data:      make(map[string]string),
	}
	for k, v := range s.Data {
		data.Data[k] = string(v)
	}
	return data
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

func getSecretByName(ctx context.Context, kubeClient client.Client, ns, name string) (v1.Secret, error) {
	s := v1.Secret{}
	return s, kubeClient.Get(ctx, client.ObjectKey{Name: name, Namespace: ns}, &s)
}

func getCorePackageSecret(ctx context.Context, kubeClient client.Client, ns, name string) (v1.Secret, error) {
	s, err := getSecretByName(ctx, kubeClient, ns, name)
	if err != nil {
		return v1.Secret{}, err
	}

	if name == argoCDInitialAdminSecretName && s.Data != nil {
		s.Data["username"] = []byte(argoCDAdminUsername)
	}
	return s, nil
}
