package controllers

import (
	"context"

	"github.com/cnoe-io/idpbuilder/pkg/controllers/custompackage"
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
	cfg util.CorePackageTemplateConfig,
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

	// Start our manager in another goroutine
	logger.V(1).Info("starting manager")
	go func() {
		if err := mgr.Start(ctx); err != nil {
			logger.Error(err, "problem running manager")
			exitCh <- err
		}
		exitCh <- nil
	}()

	return nil
}
