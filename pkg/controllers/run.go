package controllers

import (
	"context"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/controllers/custompackage"
	"github.com/cnoe-io/idpbuilder/pkg/controllers/gatewayprovider"
	"github.com/cnoe-io/idpbuilder/pkg/controllers/gitopsprovider"
	"github.com/cnoe-io/idpbuilder/pkg/controllers/gitprovider"
	"github.com/cnoe-io/idpbuilder/pkg/controllers/platform"
	"github.com/cnoe-io/idpbuilder/pkg/util"

	"github.com/cnoe-io/idpbuilder/pkg/controllers/gitrepository"
	"github.com/cnoe-io/idpbuilder/pkg/controllers/localbuild"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func RunControllers(
	ctx context.Context,
	mgr manager.Manager,
	exitCh chan error,
	ctxCancel context.CancelFunc,
	exitOnSync bool,
	cfg v1alpha1.BuildCustomizationSpec,
	tmpDir string,
) error {
	logger := log.FromContext(ctx)

	repoMap := util.NewRepoLock()

	// Run Localbuild controller
	if err := (&localbuild.LocalbuildReconciler{
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		ExitOnSync: exitOnSync,
		CancelFunc: ctxCancel,
		Config:     cfg,
		TempDir:    tmpDir,
		RepoMap:    repoMap,
	}).SetupWithManager(mgr); err != nil {
		logger.Error(err, "unable to create localbuild controller")
		return err
	}

	err := (&gitrepository.RepositoryReconciler{
		Client:          mgr.GetClient(),
		Scheme:          mgr.GetScheme(),
		Recorder:        mgr.GetEventRecorderFor("gitrepository-controller"),
		Config:          cfg,
		GitProviderFunc: gitrepository.GetGitProvider,
		TempDir:         tmpDir,
		RepoMap:         repoMap,
	}).SetupWithManager(mgr, nil)
	if err != nil {
		logger.Error(err, "unable to create repo controller")
	}

	err = (&custompackage.Reconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("custompackage-controller"),
		TempDir:  tmpDir,
		RepoMap:  repoMap,
	}).SetupWithManager(mgr)
	if err != nil {
		logger.Error(err, "unable to create custom package controller")
	}

	// Run Platform controller
	if err := (&platform.PlatformReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		logger.Error(err, "unable to create platform controller")
		return err
	}

	// Run GiteaProvider controller
	if err := (&gitprovider.GiteaProviderReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Config: cfg,
	}).SetupWithManager(mgr); err != nil {
		logger.Error(err, "unable to create GiteaProvider controller")
		return err
	}

	// Run NginxGateway controller
	if err := (&gatewayprovider.NginxGatewayReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Config: cfg,
	}).SetupWithManager(mgr); err != nil {
		logger.Error(err, "unable to create nginxgateway controller")
		return err
	}

	// Run ArgoCDProvider controller (Phase 1.3)
	if err := (&gitopsprovider.ArgoCDProviderReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		logger.Error(err, "unable to create argocdprovider controller")
		return err
	}

	// Start our manager in another goroutine
	logger.V(1).Info("starting manager")

	go func() {
		exitCh <- mgr.Start(ctx)
		close(exitCh)
	}()

	return nil
}
