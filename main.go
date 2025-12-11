package main

import (
  "os"
  ctrl "sigs.k8s.io/controller-runtime"
  "sigs.k8s.io/controller-runtime/pkg/log/zap"
  "k8s.io/apimachinery/pkg/runtime"
  corev1 "k8s.io/api/core/v1"
	"custom_controller/controllers"
)

func main() {
  ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

  scheme := runtime.NewScheme()
  _ = corev1.AddToScheme(scheme)

  mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
    Scheme: scheme,
  })
  if err != nil {
    os.Exit(1)
  }

  if err = (&controllers.PodReconciler{
    Client: mgr.GetClient(),
  }).SetupWithManager(mgr); err != nil {
    os.Exit(1)
  }

  if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
    os.Exit(1)
  }
}
