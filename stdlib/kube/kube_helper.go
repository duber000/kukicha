package kube

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Connect creates a Cluster using the default kubeconfig (~/.kube/config).
func Connect() (Cluster, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Cluster{}, fmt.Errorf("kube connect: %w", err)
	}
	kubeconfig := filepath.Join(home, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return Cluster{}, fmt.Errorf("kube connect: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return Cluster{}, fmt.Errorf("kube connect: %w", err)
	}
	return Cluster{client: clientset, namespace: "default"}, nil
}

// Open creates a Cluster from the builder configuration.
func Open(cfg Config) (Cluster, error) {
	var restConfig *rest.Config
	var err error

	if cfg.inCluster {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return Cluster{}, fmt.Errorf("kube in-cluster: %w", err)
		}
	} else {
		kubeconfig := cfg.kubeconfig
		if kubeconfig == "" {
			home, homeErr := os.UserHomeDir()
			if homeErr != nil {
				return Cluster{}, fmt.Errorf("kube config: %w", homeErr)
			}
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
		loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}
		overrides := &clientcmd.ConfigOverrides{}
		if cfg.context != "" {
			overrides.CurrentContext = cfg.context
		}
		restConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides).ClientConfig()
		if err != nil {
			return Cluster{}, fmt.Errorf("kube open: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return Cluster{}, fmt.Errorf("kube open: %w", err)
	}
	return Cluster{client: clientset, namespace: "default"}, nil
}

func clientset(c Cluster) *kubernetes.Clientset {
	return c.client.(*kubernetes.Clientset)
}

// --- Pods ---

// ListPods lists all pods in the cluster's current namespace.
func ListPods(c Cluster) (PodList, error) {
	pods, err := clientset(c).CoreV1().Pods(c.namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return PodList{}, fmt.Errorf("kube list pods: %w", err)
	}
	return PodList{items: pods}, nil
}

// ListPodsLabeled lists pods matching a label selector.
func ListPodsLabeled(c Cluster, selector string) (PodList, error) {
	pods, err := clientset(c).CoreV1().Pods(c.namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return PodList{}, fmt.Errorf("kube list pods labeled: %w", err)
	}
	return PodList{items: pods}, nil
}

// GetPod retrieves a single pod by name.
func GetPod(c Cluster, name string) (Pod, error) {
	pod, err := clientset(c).CoreV1().Pods(c.namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return Pod{}, fmt.Errorf("kube get pod: %w", err)
	}
	return Pod{pod: pod}, nil
}

// DeletePod deletes a pod by name.
func DeletePod(c Cluster, name string) error {
	if err := clientset(c).CoreV1().Pods(c.namespace).Delete(context.Background(), name, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("kube delete pod: %w", err)
	}
	return nil
}

// --- Deployments ---

// ListDeployments lists all deployments in the cluster's current namespace.
func ListDeployments(c Cluster) (DeploymentList, error) {
	deps, err := clientset(c).AppsV1().Deployments(c.namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return DeploymentList{}, fmt.Errorf("kube list deployments: %w", err)
	}
	return DeploymentList{items: deps}, nil
}

// GetDeployment retrieves a single deployment by name.
func GetDeployment(c Cluster, name string) (Deployment, error) {
	dep, err := clientset(c).AppsV1().Deployments(c.namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return Deployment{}, fmt.Errorf("kube get deployment: %w", err)
	}
	return Deployment{dep: dep}, nil
}

// ScaleDeployment updates the replica count of a deployment.
func ScaleDeployment(c Cluster, name string, replicas int32) error {
	scale, err := clientset(c).AppsV1().Deployments(c.namespace).GetScale(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("kube scale get: %w", err)
	}
	scale.Spec.Replicas = replicas
	_, err = clientset(c).AppsV1().Deployments(c.namespace).UpdateScale(context.Background(), name, scale, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("kube scale update: %w", err)
	}
	return nil
}

// --- Services ---

// ListServices lists all services in the cluster's current namespace.
func ListServices(c Cluster) (ServiceList, error) {
	svcs, err := clientset(c).CoreV1().Services(c.namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return ServiceList{}, fmt.Errorf("kube list services: %w", err)
	}
	return ServiceList{items: svcs}, nil
}

// GetService retrieves a single service by name.
func GetService(c Cluster, name string) (Service, error) {
	svc, err := clientset(c).CoreV1().Services(c.namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return Service{}, fmt.Errorf("kube get service: %w", err)
	}
	return Service{svc: svc}, nil
}

// --- Nodes ---

// ListNodes lists all nodes in the cluster.
func ListNodes(c Cluster) (NodeList, error) {
	nodes, err := clientset(c).CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return NodeList{}, fmt.Errorf("kube list nodes: %w", err)
	}
	return NodeList{items: nodes}, nil
}

// GetNode retrieves a single node by name.
func GetNode(c Cluster, name string) (Node, error) {
	node, err := clientset(c).CoreV1().Nodes().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return Node{}, fmt.Errorf("kube get node: %w", err)
	}
	return Node{node: node}, nil
}

// --- Namespaces ---

// ListNamespaces lists all namespaces in the cluster.
func ListNamespaces(c Cluster) (NamespaceList, error) {
	nsList, err := clientset(c).CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return NamespaceList{}, fmt.Errorf("kube list namespaces: %w", err)
	}
	return NamespaceList{items: nsList}, nil
}

// --- List accessors ---

// Pods returns the list of Pod items from a PodList.
func Pods(list PodList) []Pod {
	podList := list.items.(*corev1.PodList)
	result := make([]Pod, len(podList.Items))
	for i := range podList.Items {
		result[i] = Pod{pod: &podList.Items[i]}
	}
	return result
}

// Deployments returns the list of Deployment items from a DeploymentList.
func Deployments(list DeploymentList) []Deployment {
	depList := list.items.(*appsv1.DeploymentList)
	result := make([]Deployment, len(depList.Items))
	for i := range depList.Items {
		result[i] = Deployment{dep: &depList.Items[i]}
	}
	return result
}

// Services returns the list of Service items from a ServiceList.
func Services(list ServiceList) []Service {
	svcList := list.items.(*corev1.ServiceList)
	result := make([]Service, len(svcList.Items))
	for i := range svcList.Items {
		result[i] = Service{svc: &svcList.Items[i]}
	}
	return result
}

// Nodes returns the list of Node items from a NodeList.
func Nodes(list NodeList) []Node {
	nodeList := list.items.(*corev1.NodeList)
	result := make([]Node, len(nodeList.Items))
	for i := range nodeList.Items {
		result[i] = Node{node: &nodeList.Items[i]}
	}
	return result
}

// Namespaces returns the list of NamespaceItem items from a NamespaceList.
func Namespaces(list NamespaceList) []NamespaceItem {
	nsList := list.items.(*corev1.NamespaceList)
	result := make([]NamespaceItem, len(nsList.Items))
	for i := range nsList.Items {
		result[i] = NamespaceItem{ns: &nsList.Items[i]}
	}
	return result
}

// --- Pod accessors ---

func pod(p Pod) *corev1.Pod { return p.pod.(*corev1.Pod) }

// PodName returns the name of the pod.
func PodName(p Pod) string { return pod(p).Name }

// PodStatus returns the phase of the pod (Running, Pending, etc.).
func PodStatus(p Pod) string { return string(pod(p).Status.Phase) }

// PodIP returns the pod's IP address.
func PodIP(p Pod) string { return pod(p).Status.PodIP }

// PodNode returns the name of the node the pod is running on.
func PodNode(p Pod) string { return pod(p).Spec.NodeName }

// PodAge returns a human-readable age string for the pod.
func PodAge(p Pod) string {
	d := time.Since(pod(p).CreationTimestamp.Time)
	if d.Hours() >= 24 {
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
	if d.Hours() >= 1 {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dm", int(d.Minutes()))
}

// PodReady returns whether all containers in the pod are ready.
func PodReady(p Pod) bool {
	for _, cond := range pod(p).Status.Conditions {
		if cond.Type == corev1.PodReady {
			return cond.Status == corev1.ConditionTrue
		}
	}
	return false
}

// PodRestarts returns the total restart count across all containers.
func PodRestarts(p Pod) int32 {
	var total int32
	for _, cs := range pod(p).Status.ContainerStatuses {
		total += cs.RestartCount
	}
	return total
}

// PodLabels returns the pod's labels.
func PodLabels(p Pod) map[string]string { return pod(p).Labels }

// --- Deployment accessors ---

func deployment(d Deployment) *appsv1.Deployment { return d.dep.(*appsv1.Deployment) }

// DeploymentName returns the deployment name.
func DeploymentName(d Deployment) string { return deployment(d).Name }

// DeploymentReplicas returns the desired replica count.
func DeploymentReplicas(d Deployment) int32 {
	if deployment(d).Spec.Replicas != nil {
		return *deployment(d).Spec.Replicas
	}
	return 1
}

// DeploymentReady returns the number of ready replicas.
func DeploymentReady(d Deployment) int32 { return deployment(d).Status.ReadyReplicas }

// DeploymentImage returns the image of the first container.
func DeploymentImage(d Deployment) string {
	containers := deployment(d).Spec.Template.Spec.Containers
	if len(containers) > 0 {
		return containers[0].Image
	}
	return ""
}

// --- Service accessors ---

func service(s Service) *corev1.Service { return s.svc.(*corev1.Service) }

// ServiceName returns the service name.
func ServiceName(s Service) string { return service(s).Name }

// ServiceType returns the service type (ClusterIP, NodePort, LoadBalancer).
func ServiceType(s Service) string { return string(service(s).Spec.Type) }

// ServiceClusterIP returns the cluster IP address.
func ServiceClusterIP(s Service) string { return service(s).Spec.ClusterIP }

// ServicePorts returns a list of port descriptions like "80/TCP".
func ServicePorts(s Service) []string {
	ports := service(s).Spec.Ports
	result := make([]string, len(ports))
	for i, p := range ports {
		result[i] = fmt.Sprintf("%d/%s", p.Port, p.Protocol)
	}
	return result
}

// --- Node accessors ---

func node(n Node) *corev1.Node { return n.node.(*corev1.Node) }

// NodeName returns the node name.
func NodeName(n Node) string { return node(n).Name }

// NodeReady returns whether the node is in Ready condition.
func NodeReady(n Node) bool {
	for _, cond := range node(n).Status.Conditions {
		if cond.Type == corev1.NodeReady {
			return cond.Status == corev1.ConditionTrue
		}
	}
	return false
}

// NodeRoles returns the roles of the node (e.g., "control-plane", "worker").
func NodeRoles(n Node) []string {
	var roles []string
	for label := range node(n).Labels {
		if strings.HasPrefix(label, "node-role.kubernetes.io/") {
			role := strings.TrimPrefix(label, "node-role.kubernetes.io/")
			if role != "" {
				roles = append(roles, role)
			}
		}
	}
	if len(roles) == 0 {
		roles = append(roles, "<none>")
	}
	return roles
}

// NodeVersion returns the kubelet version of the node.
func NodeVersion(n Node) string { return node(n).Status.NodeInfo.KubeletVersion }

// --- Namespace accessors ---

func nsItem(n NamespaceItem) *corev1.Namespace { return n.ns.(*corev1.Namespace) }

// NamespaceName returns the namespace name.
func NamespaceName(n NamespaceItem) string { return nsItem(n).Name }

// --- Logs ---

// PodLogs retrieves the full log output from a pod's first container.
func PodLogs(c Cluster, name string) (string, error) {
	req := clientset(c).CoreV1().Pods(c.namespace).GetLogs(name, &corev1.PodLogOptions{})
	stream, err := req.Stream(context.Background())
	if err != nil {
		return "", fmt.Errorf("kube pod logs: %w", err)
	}
	defer stream.Close()
	data, err := io.ReadAll(stream)
	if err != nil {
		return "", fmt.Errorf("kube pod logs read: %w", err)
	}
	return string(data), nil
}

// PodLogsTail retrieves the last N lines of log output from a pod.
func PodLogsTail(c Cluster, name string, lines int64) (string, error) {
	req := clientset(c).CoreV1().Pods(c.namespace).GetLogs(name, &corev1.PodLogOptions{
		TailLines: &lines,
	})
	stream, err := req.Stream(context.Background())
	if err != nil {
		return "", fmt.Errorf("kube pod logs tail: %w", err)
	}
	defer stream.Close()
	data, err := io.ReadAll(stream)
	if err != nil {
		return "", fmt.Errorf("kube pod logs tail read: %w", err)
	}
	return string(data), nil
}
