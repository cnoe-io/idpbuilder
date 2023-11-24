package localbuild

import (
	argov1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SetProjectSpec(project *argov1alpha1.AppProject) {
	project.Spec.Description = "IDP Builder Project"

	project.Spec.ClusterResourceWhitelist = []v1.GroupKind{{
		Group: "*",
		Kind:  "*",
	}}
	project.Spec.NamespaceResourceWhitelist = []v1.GroupKind{{
		Group: "*",
		Kind:  "*",
	}}

	project.Spec.Destinations = []argov1alpha1.ApplicationDestination{{
		Name:      "*",
		Namespace: "*",
		Server:    "*",
	}}

	project.Spec.SourceRepos = []string{
		"*",
	}
}

func SetApplicationSpec(app *argov1alpha1.Application, repoUrl, path, project, dstNS string, targetRevision *string) {
	headRev := "HEAD"
	if targetRevision == nil {
		targetRevision = &headRev
	}

	app.Spec.Destination = argov1alpha1.ApplicationDestination{
		Server:    "https://kubernetes.default.svc",
		Namespace: dstNS,
	}

	app.Spec.Project = project

	app.Spec.Source = &argov1alpha1.ApplicationSource{
		Path:           path,
		RepoURL:        repoUrl,
		TargetRevision: *targetRevision,
	}

	app.Spec.SyncPolicy = &argov1alpha1.SyncPolicy{
		Automated: &argov1alpha1.SyncPolicyAutomated{
			SelfHeal: true,
		},
		SyncOptions: argov1alpha1.SyncOptions{
			"CreateNamespace=true",
		},
	}
}
