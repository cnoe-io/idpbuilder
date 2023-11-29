package controllers

import (
	"context"

	"github.com/cnoe-io/idpbuilder/pkg/controllers/gitrepository"
	"github.com/cnoe-io/idpbuilder/pkg/controllers/localbuild"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func RunControllers(ctx context.Context, mgr manager.Manager, exitCh chan error, ctxCancel context.CancelFunc) error {
	log := log.FromContext(ctx)

	// Run Localbuild controller
	if err := (&localbuild.LocalbuildReconciler{
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		CancelFunc: ctxCancel,
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create localbuild controller")
		return err
	}

	err := (&gitrepository.RepositoryReconciler{
		Client:          mgr.GetClient(),
		Scheme:          mgr.GetScheme(),
		Recorder:        mgr.GetEventRecorderFor("gitrepository-controller"),
		GiteaClientFunc: gitrepository.NewGiteaClient,
	}).SetupWithManager(mgr, nil)
	if err != nil {
		log.Error(err, "unable to create repo controller")
	}

	// Start our manager in another goroutine
	log.Info("starting manager")
	go func() {
		if err := mgr.Start(ctx); err != nil {
			log.Error(err, "problem running manager")
			exitCh <- err
		}
		exitCh <- nil
	}()

	return nil
}
