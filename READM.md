# 简单的controller

    也可以参考官方的sample：https://github.com/kubernetes/sample-controller
    
# 实现步骤

## 1.向k8s集群提交CRD模板和其实例化对象，使k8s能识别
    1 官方文档了解下CRD 
    https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#create-a-customresourcedefinition
    
    2 登录可以执行kubectl命令的机器，创建student.yaml文件
    [root@localhost student]# kubectl apply -f student.yaml
    customresourcedefinition.apiextensions.k8s.io/students.stable.k8s.io created
    [root@localhost student]# kubectl get crd
    NAME                          CREATED AT
    crontabs.stable.example.com   2019-03-26T01:48:32Z
    students.stable.k8s.io        2019-04-12T02:42:08Z
    
    3 通过模板student.yaml来实例一个Student对象，创建test1.yaml，同理实例test2.yaml
    kubectl apply -f fanlin.yaml
    [root@localhost student]# kubectl apply -f test1.yaml
    student.stable.k8s.io/test1 created
    
    4 kubectl describe std test1
## 2.自动生成代码前的准备工作
    1 在samplecontroller目录下新建目录，并添加以下文件
    [root@localhost studentcontroller]# tree pkg
    pkg
    ├── apis
    │   └── stable
    │       ├── register.go
    │       └── v1
    │           ├── doc.go
    │           ├── register.go
    │           ├── types.go
    主要是为代码生成工具准备好资源对象的声明和注册接口
    
    2 下载依赖
    go get -u k8s.io/apimachinery/pkg/apis/meta/v1
    go get -u k8s.io/code-generator/...
## 3.自动生成Client、Informer、WorkQueue相关的代码
    [root@localhost code-generator]# export GOPATH=/root/mygolang
    [root@localhost code-generator]# ./generate-groups.sh all "github.com/gfandada/samplecontroller/pkg/client" "github.com/gfandada/samplecontroller/pkg/apis" "stable:v1"
    Generating deepcopy funcs
    Generating clientset for stable:v1 at github.com/gfandada/samplecontroller/pkg/client/clientset
    Generating listers for stable:v1 at github.com/gfandada/samplecontroller/pkg/client/listers
    Generating informers for stable:v1 at github.com/gfandada/samplecontroller/pkg/client/informers
    [root@localhost samplecontroller]# tree pkg
    pkg
    ├── apis
    │   └── stable
    │       ├── register.go
    │       └── v1
    │           ├── doc.go
    │           ├── register.go
    │           ├── types.go
    │           └── zz_generated.deepcopy.go
    ├── client
    │   ├── clientset
    │   │   └── versioned
    │   │       ├── clientset.go
    │   │       ├── doc.go
    │   │       ├── fake
    │   │       │   ├── clientset_generated.go
    │   │       │   ├── doc.go
    │   │       │   └── register.go
    │   │       ├── scheme
    │   │       │   ├── doc.go
    │   │       │   └── register.go
    │   │       └── typed
    │   │           └── stable
    │   │               └── v1
    │   │                   ├── doc.go
    │   │                   ├── fake
    │   │                   │   ├── doc.go
    │   │                   │   ├── fake_stable_client.go
    │   │                   │   └── fake_student.go
    │   │                   ├── generated_expansion.go
    │   │                   ├── stable_client.go
    │   │                   └── student.go
    │   ├── informers
    │   │   └── externalversions
    │   │       ├── factory.go
    │   │       ├── generic.go
    │   │       ├── internalinterfaces
    │   │       │   └── factory_interfaces.go
    │   │       └── stable
    │   │           ├── interface.go
    │   │           └── v1
    │   │               ├── interface.go
    │   │               └── student.go
    │   └── listers
    │       └── stable
    │           └── v1
    │               ├── expansion_generated.go
    │               └── student.go
## 4.编写controller的业务逻辑
    参考sample-controller工程，写的比较简单
## 5.启动controller    
    [root@localhost samplecontroller]# ./samplecontroller -kubeconfig=$你的kubeconfig路径 -alsologtostderr=true       
    I0415 11:39:34.898050    3501 samplecontroller.go:59] 创建事件广播器
    I0415 11:39:34.900428    3501 samplecontroller.go:76] 监听student的add/update/delete事件
    I0415 11:39:34.900459    3501 samplecontroller.go:102] 开始controller业务，开始一次缓存数据同步
    I0415 11:39:35.000893    3501 samplecontroller.go:107] 启动10个worker
    I0415 11:39:35.000973    3501 samplecontroller.go:112] worker已经全部启动
    ......
## 6.修改crd实例文件观察controller
    ........
    kubectl apply -f test1.yaml
    kubectl describe std test1
    
