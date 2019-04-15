package app

import (
	"strings"
	"time"

	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	clientset "github.com/gfandada/samplecontroller/pkg/client/clientset/versioned"
	informers "github.com/gfandada/samplecontroller/pkg/client/informers/externalversions"
	"github.com/gfandada/samplecontroller/pkg/signals"
)

func getKubeCfg(args []string) string {
	for _, v := range args {
		strs := strings.Split(v, "=")
		if strs[0] == "config" {
			return strs[1]
		}
	}
	return ""
}

func getKubeMasterUrl(args []string) string {
	for _, v := range args {
		strs := strings.Split(v, "=")
		if strs[0] == "master" {
			return strs[1]
		}
	}
	return ""
}

func NewApp(args []string) {
	// 处理信号量
	stopCh := signals.SetupSignalHandler()

	// 处理入参
	cfg, err := clientcmd.BuildConfigFromFlags(getKubeMasterUrl(args), getKubeCfg(args))
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	// 创建标准的client
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	// 创建student资源的client
	studentClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building example clientset: %s", err.Error())
	}

	// 创建informer
	studentInformerFactory := informers.NewSharedInformerFactory(studentClient, time.Second*30)

	controller := NewSampleController(kubeClient, studentClient,
		studentInformerFactory.Stable().V1().Students())

	// 启动informer
	go studentInformerFactory.Start(stopCh)

	// 启动controller
	if err = controller.Run(10, stopCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}
}
