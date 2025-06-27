package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GroupVersion is group version used to register these objects
var GroupVersion = schema.GroupVersion{Group: "springboot.tutorial.example.com", Version: "v1"}

// ConfigSpec 定义配置管理相关的字段
type ConfigSpec struct {
    // ConfigMapRef 引用的 ConfigMap
    // +optional
    ConfigMapRef *corev1.LocalObjectReference `json:"configMapRef,omitempty"`
    
    // MountPath 配置文件挂载路径
    // +kubebuilder:default="/app/config"
    // +optional
    MountPath string `json:"mountPath,omitempty"`
    
    // Env 环境变量配置
    // +optional
    Env []corev1.EnvVar `json:"env,omitempty"`
    
    // EnvFrom 从 ConfigMap 或 Secret 导入环境变量
    // +optional
    EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`
}

// ServiceSpec 定义服务暴露相关的字段
type ServiceSpec struct {
    // Type 服务类型
    // +kubebuilder:default=ClusterIP
    // +optional
    Type corev1.ServiceType `json:"type,omitempty"`
    
    // NodePort 当 Type 为 NodePort 时的端口
    // +optional
    NodePort int32 `json:"nodePort,omitempty"`
    
    // Ports 额外的端口配置
    // +optional
    Ports []corev1.ServicePort `json:"ports,omitempty"`
    
    // Annotations 服务注解
    // +optional
    Annotations map[string]string `json:"annotations,omitempty"`
}

// TLSSpec 定义 TLS 配置
type TLSSpec struct {
    // Enabled 是否启用 TLS
    Enabled bool `json:"enabled"`
    
    // SecretName TLS 证书 Secret 名称
    // +optional
    SecretName string `json:"secretName,omitempty"`
}

// IngressSpec 定义 Ingress 相关的字段
type IngressSpec struct {
    // Enabled 是否启用 Ingress
    Enabled bool `json:"enabled"`
    
    // ClassName Ingress 类名
    // +optional
    ClassName *string `json:"className,omitempty"`
    
    // Host 主机名
    // +optional
    Host string `json:"host,omitempty"`
    
    // Path 路径
    // +kubebuilder:default="/"
    // +optional
    Path string `json:"path,omitempty"`
    
    // PathType 路径类型
    // +kubebuilder:default=Prefix
    // +optional
    PathType string `json:"pathType,omitempty"`
    
    // TLS TLS 配置
    // +optional
    TLS *TLSSpec `json:"tls,omitempty"`
    
    // Annotations Ingress 注解
    // +optional
    Annotations map[string]string `json:"annotations,omitempty"`
}

// SpringBootAppSpec defines the desired state of SpringBootApp
type SpringBootAppSpec struct {
    // Image is the container image for the Spring Boot application
    Image string `json:"image"`
    
    // Replicas is the number of desired replicas
    // +kubebuilder:default=1
    // +optional
    Replicas *int32 `json:"replicas,omitempty"`
    
    // Port is the port that the application listens on
    // +kubebuilder:default=8080
    // +optional
    Port int32 `json:"port,omitempty"`
    
    // Config 配置管理
    // +optional
    Config *ConfigSpec `json:"config,omitempty"`
    
    // Service 服务配置
    // +optional
    Service *ServiceSpec `json:"service,omitempty"`
    
    // Ingress Ingress 配置
    // +optional
    Ingress *IngressSpec `json:"ingress,omitempty"`
}

// SpringBootAppStatus defines the observed state of SpringBootApp
type SpringBootAppStatus struct {
    // Replicas is the current number of replicas
    Replicas int32 `json:"replicas"`
    
    // ReadyReplicas is the number of ready replicas
    ReadyReplicas int32 `json:"readyReplicas"`
    
    // Phase represents the current phase of the application
    Phase string `json:"phase,omitempty"`
    
    // Conditions represent the latest available observations
    Conditions []metav1.Condition `json:"conditions,omitempty"`
    
    // ConfigHash 配置文件的哈希值，用于检测配置变更
    ConfigHash string `json:"configHash,omitempty"`
    
    // ServiceURL 服务访问地址
    ServiceURL string `json:"serviceURL,omitempty"`
    
    // IngressURL Ingress 访问地址
    IngressURL string `json:"ingressURL,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`
// +kubebuilder:printcolumn:name="Ready",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Service",type=string,JSONPath=`.status.serviceURL`
// +kubebuilder:printcolumn:name="Ingress",type=string,JSONPath=`.status.ingressURL`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// SpringBootApp is the Schema for the springbootapps API
type SpringBootApp struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   SpringBootAppSpec   `json:"spec,omitempty"`
    Status SpringBootAppStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SpringBootAppList contains a list of SpringBootApp
type SpringBootAppList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []SpringBootApp `json:"items"`
}

// NOTE: 在实际的Operator项目中，SchemeBuilder会由operator-sdk自动生成
// 这里仅作为示例代码展示API结构