package export

import (
	"context"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/build"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/util/homedir"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var SecretsCmd = &cobra.Command{
	Use:   "secret",
	Short: "retrieve secrets from the cluster",
	Long:  ``,
	RunE:  exportSecretsE,
}

//go:embed templates
var templates embed.FS

type TemplateData struct {
	Name      string
	Namespace string
	Data      map[string]string
}

func exportSecretsE(cmd *cobra.Command, args []string) error {
	ctx, ctxCancel := context.WithCancel(ctrl.SetupSignalHandler())
	defer ctxCancel()
	kubeConfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	b := build.NewBuild("", "", kubeConfigPath, "", "", util.CorePackageTemplateConfig{}, []string{}, false, k8s.GetScheme(), ctxCancel)

	kubeConfig, err := b.GetKubeConfig()
	if err != nil {
		return fmt.Errorf("getting kube config: %w", err)
	}

	kubeClient, err := b.GetKubeClient(kubeConfig)
	if err != nil {
		return fmt.Errorf("getting kube client: %w", err)
	}
	// secrets that are part of the core packages
	corePkgSecrets := map[string]string{
		"argocd": "argocd-initial-admin-secret",
		"gitea":  "gitea-admin-secret",
	}
	secretsToPrint := make([]any, 0, len(corePkgSecrets))

	for k, v := range corePkgSecrets {
		secret, sErr := getSecretByName(ctx, kubeClient, k, v)
		if sErr != nil {
			if errors.IsNotFound(sErr) {
				continue
			}
			return fmt.Errorf("getting secret %s in %s: %w", v, k, sErr)
		}
		secretsToPrint = append(secretsToPrint, secretToTemplateData(secret))
	}

	secrets, err := getSecretsByExportLabel(ctx, kubeClient)
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

	return renderTemplate("templates/secrets.tmpl", os.Stdout, secretsToPrint)
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

func getSecretsByExportLabel(ctx context.Context, kubeClient client.Client) (v1.SecretList, error) {
	l := labels.NewSelector()

	req, err := labels.NewRequirement(v1alpha1.ExportLabelKey, selection.Equals, []string{v1alpha1.ExportLabelValue})
	if err != nil {
		return v1.SecretList{}, fmt.Errorf("building labels with key %s and value %s : %w", v1alpha1.ExportLabelKey, v1alpha1.ExportLabelValue, err)
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
