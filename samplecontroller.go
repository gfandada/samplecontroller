package main

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	stablev1 "github.com/gfandada/samplecontroller/pkg/apis/stable/v1"
	clientset "github.com/gfandada/samplecontroller/pkg/client/clientset/versioned"
	studentscheme "github.com/gfandada/samplecontroller/pkg/client/clientset/versioned/scheme"
	informers "github.com/gfandada/samplecontroller/pkg/client/informers/externalversions/stable/v1"
	listers "github.com/gfandada/samplecontroller/pkg/client/listers/stable/v1"
)

const controllerAgentName = "sample-controller"

const (
	SuccessSynced         = "Synced"
	MessageResourceSynced = "Student对象同步成功"
	MessageTest           = "分布式系统中有人改了email，由" + controllerAgentName + "自动完成了向终态的演进"
)

// Controller is the controller implementation for Student resources
type SampleController struct {
	// k8s标准的clientset
	kubeclientset kubernetes.Interface
	// our own API group的clientset
	studentclientset clientset.Interface

	studentsLister listers.StudentLister
	studentsSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface

	syncHandler func(dKey string) error

	recorder record.EventRecorder
}

// 构造一个studentcontroller
func NewSampleController(
	kubeclientset kubernetes.Interface,
	studentclientset clientset.Interface,
	studentInformer informers.StudentInformer) *SampleController {

	utilruntime.Must(studentscheme.AddToScheme(scheme.Scheme))
	glog.Info("创建事件广播器")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})

	// FIXME 创建一个来源是controllerAgentName事件记录器，本身也是审计/日志的一部分
	// kubectl describe crd test1
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	controller := &SampleController{
		kubeclientset:    kubeclientset,
		studentclientset: studentclientset,
		studentsLister:   studentInformer.Lister(),
		studentsSynced:   studentInformer.Informer().HasSynced,
		workqueue:        workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Students"),
		recorder:         recorder,
	}

	glog.Info("监听student的add/update/delete事件")

	studentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.addStudentHandler,
		UpdateFunc: func(old, new interface{}) {
			oldStudent := old.(*stablev1.Student)
			newStudent := new.(*stablev1.Student)
			// 版本一致，就表示没有实际更新的操作
			if oldStudent.ResourceVersion == newStudent.ResourceVersion {
				return
			}
			controller.addStudentHandler(new)
		},
		DeleteFunc: controller.deleteStudentHandler,
	})

	controller.syncHandler = controller.syncStudent

	return controller
}

// 在此处开始controller的业务
func (c *SampleController) Run(max int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	glog.Info("开始controller业务，开始一次缓存数据同步")
	if ok := cache.WaitForCacheSync(stopCh, c.studentsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Infof("启动%d个worker", max)
	for i := 0; i < max; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("worker已经全部启动")

	<-stopCh

	glog.Info("worker已经结束")
	glog.Infof("%s已经结束", controllerAgentName)

	return nil
}

func (c *SampleController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// worker取数据处理
func (c *SampleController) processNextWorkItem() bool {
	// 从队列中取一个student
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool

		if key, ok = obj.(string); !ok {

			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// 在syncHandler中处理业务
		if err := c.syncStudent(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}

		c.workqueue.Forget(obj)
		glog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// 具体的处理过程
func (c *SampleController) syncStudent(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// 从缓存中取对象
	student, err := c.studentsLister.Students(namespace).Get(name)
	if err != nil {
		// 如果Student对象被删除了，就会走到这里，所以应该在这里加入执行
		if errors.IsNotFound(err) {
			glog.Infof("Student对象被删除，请在这里执行实际的删除业务: %s/%s ...", namespace, name)

			return nil
		}

		runtime.HandleError(fmt.Errorf("failed to list student by: %s/%s", namespace, name))

		return err
	}

	glog.Infof("这里是student对象的实际状态: %v ...", student)
	studentCopy := student.DeepCopy()

	// FIXME 这里是模拟的终态业务
	if studentCopy.Spec.Name == "gfandada" && studentCopy.Spec.Email != "gfandada@gmail.com" {
		glog.Infof("===========================================================================================================")
		studentCopy.Spec.Email = "gfandada@gmail.com"
		glog.Infof("期望的状态%v", studentCopy)
		glog.Infof("name=%v resourceVersion=%v ns=%v 类型=%v owner=%v uid=%v", studentCopy.Name, studentCopy.ResourceVersion,
			studentCopy.Namespace, studentCopy.Kind, studentCopy.OwnerReferences, studentCopy.UID)

		// FIXME  crd的curd一般都需要封装下，可以参考Deployment的封装
		result := &stablev1.Student{}
		c.studentclientset.StableV1().RESTClient().Put().
			Namespace(studentCopy.Namespace).
			Resource("students").Name(studentCopy.Name).Body(studentCopy).Do().Into(result)
		c.recorder.Event(student, corev1.EventTypeWarning, SuccessSynced, MessageTest)
		glog.Infof("===========================================================================================================")
		return nil
	}
	c.recorder.Event(student, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

// 添加student
func (c *SampleController) addStudentHandler(obj interface{}) {
	var key string
	var err error

	// 将对象放入缓存
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	// 再将key放入队列
	c.workqueue.AddRateLimited(key)
}

// 删除student
func (c *SampleController) deleteStudentHandler(obj interface{}) {
	var key string
	var err error

	// 从缓存中删除指定对象
	key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	// 再将key放入队列
	c.workqueue.AddRateLimited(key)
}
