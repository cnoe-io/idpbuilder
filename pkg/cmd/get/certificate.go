package get

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cnoe-io/idpbuilder/globals"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	certOutputPath string
	setupDocker    bool
)

var CertificateCmd = &cobra.Command{
	Use:   "certificate",
	Short: "Export the TLS certificate from the cluster",
	Long: `Export the self-signed TLS certificate from the cluster.

By default, the certificate is printed to stdout. Use the --docker flag to automatically
configure Docker's per-registry certificate directory, which allows Docker to trust the
Gitea container registry without requiring a Docker daemon restart.

Examples:
  # Print certificate to stdout
  idpbuilder get certificate

  # Export to Docker's registry certificate directory (no Docker restart needed)
  idpbuilder get certificate --docker

  # Export to a custom file
  idpbuilder get certificate --output ~/my-cert.crt`,
	RunE:         getCertificateE,
	SilenceUsage: true,
}

func init() {
	CertificateCmd.Flags().StringVarP(&certOutputPath, "output", "o", "", "Custom output path for the certificate file")
	CertificateCmd.Flags().BoolVar(&setupDocker, "docker", false, "Setup Docker registry certificate directory")
}

func getCertificateE(cmd *cobra.Command, args []string) error {
	ctx, ctxCancel := context.WithCancel(cmd.Context())
	defer ctxCancel()

	kubeConfig, err := util.GetKubeConfig()
	if err != nil {
		return fmt.Errorf("getting kube config: %w", err)
	}

	kubeClient, err := util.GetKubeClient(kubeConfig)
	if err != nil {
		return fmt.Errorf("getting kube client: %w", err)
	}

	// Get the certificate from the cluster
	cert, err := getCertificateFromCluster(ctx, kubeClient)
	if err != nil {
		return err
	}

	// Get the build configuration to determine registry host
	config, err := util.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("getting idp config: %w", err)
	}

	var registryHost string
	if config.UsePathRouting {
		registryHost = fmt.Sprintf("%s:%s", config.Host, config.Port)
	} else {
		registryHost = fmt.Sprintf("gitea.%s:%s", config.Host, config.Port)
	}

	// Determine output behavior
	if certOutputPath != "" {
		// Custom output path specified
		if err := os.WriteFile(certOutputPath, cert, 0644); err != nil {
			return fmt.Errorf("writing certificate to %s: %w", certOutputPath, err)
		}
		fmt.Printf("Certificate exported to: %s\n", certOutputPath)
		return nil
	} else if setupDocker {
		// Docker's per-registry certificate directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting user home directory: %w", err)
		}

		dockerCertsDir := filepath.Join(homeDir, ".docker", "certs.d", registryHost)
		if err := os.MkdirAll(dockerCertsDir, 0755); err != nil {
			return fmt.Errorf("creating Docker certificate directory: %w", err)
		}

		outputPath := filepath.Join(dockerCertsDir, "ca.crt")
		if err := os.WriteFile(outputPath, cert, 0644); err != nil {
			return fmt.Errorf("writing certificate to %s: %w", outputPath, err)
		}

		fmt.Printf("Certificate exported successfully to: %s\n", outputPath)
		fmt.Printf("Registry host: %s\n\n", registryHost)
		printPostInstallInstructions(registryHost)
		return nil
	}

	// Default: print to stdout
	fmt.Print(string(cert))
	return nil
}

func getCertificateFromCluster(ctx context.Context, kubeClient client.Client) ([]byte, error) {
	secret := &corev1.Secret{}
	secretKey := client.ObjectKey{
		Name:      globals.SelfSignedCertSecretName,
		Namespace: globals.NginxNamespace,
	}

	if err := kubeClient.Get(ctx, secretKey, secret); err != nil {
		return nil, fmt.Errorf("getting certificate secret from cluster: %w. Make sure the cluster is running.", err)
	}

	cert, ok := secret.Data[corev1.TLSCertKey]
	if !ok {
		return nil, fmt.Errorf("certificate not found in secret %s/%s", globals.NginxNamespace, globals.SelfSignedCertSecretName)
	}

	return cert, nil
}

func printPostInstallInstructions(registryHost string) {
	fmt.Println("Next steps:")
	fmt.Println("  1. The certificate has been configured for Docker's per-registry trust")
	fmt.Println("  2. No Docker restart is required - the certificate should work immediately")
	fmt.Println()
	fmt.Println("To verify Docker can access the registry:")
	fmt.Printf("  docker pull %s/test-image\n", registryHost)
	fmt.Println()

	// Platform-specific instructions for system-wide trust (optional)
	switch runtime.GOOS {
	case "darwin":
		fmt.Println("Optional: To trust this certificate system-wide (for browsers, curl, etc.):")
		fmt.Println("  1. Find the certificate file:")
		homeDir, _ := os.UserHomeDir()
		fmt.Printf("     %s\n", filepath.Join(homeDir, ".docker", "certs.d", registryHost, "ca.crt"))
		fmt.Println("  2. Double-click the certificate file to open Keychain Access")
		fmt.Println("  3. Select 'System' keychain and click 'Add'")
		fmt.Println("  4. Double-click the imported certificate")
		fmt.Println("  5. Expand 'Trust' section and set 'When using this certificate' to 'Always Trust'")
		fmt.Println()
		fmt.Println("  Or use the command line:")
		certPath := filepath.Join(homeDir, ".docker", "certs.d", registryHost, "ca.crt")
		fmt.Printf("  sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain %s\n", certPath)

	case "linux":
		fmt.Println("Optional: To trust this certificate system-wide (for browsers, curl, etc.):")
		homeDir, _ := os.UserHomeDir()
		certPath := filepath.Join(homeDir, ".docker", "certs.d", registryHost, "ca.crt")
		fmt.Printf("  sudo cp %s /usr/local/share/ca-certificates/idpbuilder-ca.crt\n", certPath)
		fmt.Println("  sudo update-ca-certificates")

	case "windows":
		fmt.Println("Optional: To trust this certificate system-wide:")
		fmt.Println("  1. Open the certificate file in Explorer")
		fmt.Println("  2. Click 'Install Certificate'")
		fmt.Println("  3. Select 'Local Machine' and click Next")
		fmt.Println("  4. Select 'Place all certificates in the following store'")
		fmt.Println("  5. Browse to 'Trusted Root Certification Authorities'")
		fmt.Println("  6. Click OK and Finish")
	}
	fmt.Println()
}
