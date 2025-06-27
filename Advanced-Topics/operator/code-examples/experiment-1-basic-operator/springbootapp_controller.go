package controllers

import (
    "context"
    "fmt"
    
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/util/intstr"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/log"
    
    springbootv1 "github.com/example/springboot-operator/api/v1"
)

// SpringBootAppReconciler reconciles a SpringBootApp object
type SpringBootAppReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=springboot.tutorial.example.com,resources=springbootapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=springboot.tutorial.example.com,resources=springbootapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=springboot.tutorial.example.com,resources=springbootapps/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

func (r *SpringBootAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)
    
    // 获取 SpringBootApp 实例
    var springBootApp springbootv1.SpringBootApp
    if err := r.Get(ctx, req.NamespacedName, &springBootApp); err != nil {
        if errors.IsNotFound(err) {
            log.Info("SpringBootApp resource not found. Ignoring since object must be deleted")
            return ctrl.Result{}, nil
        }
        log.Error(err, "Failed to get SpringBootApp")
        return ctrl.Result{}, err
    }
    
    // 创建或更新 Deployment
    if err := r.reconcileDeployment(ctx, &springBootApp); err != nil {
        log.Error(err, "Failed to reconcile Deployment")
        return ctrl.Result{}, err
    }
    
    // 创建或更新 Service
    if err := r.reconcileService(ctx, &springBootApp); err != nil {
        log.Error(err, "Failed to reconcile Service")
        return ctrl.Result{}, err
    }
    
    // 更新状态
    if err := r.updateStatus(ctx, &springBootApp); err != nil {
        log.Error(err, "Failed to update SpringBootApp status")
        return ctrl.Result{}, err
    }
    
    return ctrl.Result{}, nil
}

func (r *SpringBootAppReconciler) reconcileDeployment(ctx context.Context, app *springbootv1.SpringBootApp) error {
    deployment := &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:      app.Name,
            Namespace: app.Namespace,
        },
    }
    
    _, err := ctrl.CreateOrUpdate(ctx, r.Client, deployment, func() error {
        // 设置 Owner Reference
        if err := ctrl.SetControllerReference(app, deployment, r.Scheme); err != nil {
            return err
        }
        
        // 设置默认值
        replicas := int32(1)
        if app.Spec.Replicas != nil {
            replicas = *app.Spec.Replicas
        }
        
        port := int32(8080)
        if app.Spec.Port != 0 {
            port = app.Spec.Port
        }
        
        // 配置 Deployment Spec
        deployment.Spec = appsv1.DeploymentSpec{
            Replicas: &replicas,
            Selector: &metav1.LabelSelector{
                MatchLabels: map[string]string{
                    "app": app.Name,
                },
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: map[string]string{
                        "app": app.Name,
                    },
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Name:  "spring-boot-app",
                            Image: app.Spec.Image,
                            Ports: []corev1.ContainerPort{
                                {
                                    ContainerPort: port,
                                    Protocol:      corev1.ProtocolTCP,
                                },
                            },
                            LivenessProbe: &corev1.Probe{
                                ProbeHandler: corev1.ProbeHandler{
                                    HTTPGet: &corev1.HTTPGetAction{
                                        Path: "/actuator/health",
                                        Port: intstr.FromInt(int(port)),
                                    },
                                },
                                InitialDelaySeconds: 30,
                                PeriodSeconds:       10,
                            },
                            ReadinessProbe: &corev1.Probe{
                                ProbeHandler: corev1.ProbeHandler{
                                    HTTPGet: &corev1.HTTPGetAction{
                                        Path: "/actuator/health",
                                        Port: intstr.FromInt(int(port)),
                                    },
                                },
                                InitialDelaySeconds: 5,
                                PeriodSeconds:       5,
                            },
                        },
                    },
                },
            },
        }
        
        return nil
    })
    
    return err
}

func (r *SpringBootAppReconciler) reconcileService(ctx context.Context, app *springbootv1.SpringBootApp) error {
    service := &corev1.Service{
        ObjectMeta: metav1.ObjectMeta{
            Name:      app.Name,
            Namespace: app.Namespace,
        },
    }
    
    _, err := ctrl.CreateOrUpdate(ctx, r.Client, service, func() error {
        // 设置 Owner Reference
        if err := ctrl.SetControllerReference(app, service, r.Scheme); err != nil {
            return err
        }
        
        port := int32(8080)
        if app.Spec.Port != 0 {
            port = app.Spec.Port
        }
        
        // 配置 Service Spec
        service.Spec = corev1.ServiceSpec{
            Selector: map[string]string{
                "app": app.Name,
            },
            Ports: []corev1.ServicePort{
                {
                    Port:       port,
                    TargetPort: intstr.FromInt(int(port)),
                    Protocol:   corev1.ProtocolTCP,
                },
            },
            Type: corev1.ServiceTypeClusterIP,
        }
        
        return nil
    })
    
    return err
}

func (r *SpringBootAppReconciler) updateStatus(ctx context.Context, app *springbootv1.SpringBootApp) error {
    // 获取对应的 Deployment
    deployment := &appsv1.Deployment{}
    err := r.Get(ctx, client.ObjectKey{
        Namespace: app.Namespace,
        Name:      app.Name,
    }, deployment)
    
    if err != nil {
        return err
    }
    
    // 更新状态
    app.Status.Replicas = deployment.Status.Replicas
    app.Status.ReadyReplicas = deployment.Status.ReadyReplicas
    
    if deployment.Status.ReadyReplicas == deployment.Status.Replicas {
        app.Status.Phase = "Running"
    } else {
        app.Status.Phase = "Pending"
    }
    
    return r.Status().Update(ctx, app)
}

// SetupWithManager sets up the controller with the Manager.
func (r *SpringBootAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&springbootv1.SpringBootApp{}).
        Owns(&appsv1.Deployment{}).
        Owns(&corev1.Service{}).
        Complete(r)
}