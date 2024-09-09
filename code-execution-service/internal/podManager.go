package internal

import (
	"context"
	"log"
	"os/exec"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// PodManager is a struct that manages pods
type PodManager struct {
	ctx               context.Context
	client            *kubernetes.Clientset
	config            *rest.Config
	standByPodsByLang map[string][]*corev1.Pod
	inUsePods         map[string]bool
	mutex             sync.Mutex
	serviceConfig     *ServiceConfig
}

// NewPodManager creates a new PodManager
func NewPodManager(config *rest.Config, serviceConf *ServiceConfig) *PodManager {
	// Create a clientset from the config
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	manager := &PodManager{
		ctx:               context.Background(),
		client:            clientset,
		config:            config,
		standByPodsByLang: make(map[string][]*corev1.Pod),
		inUsePods:         make(map[string]bool),
		serviceConfig:     serviceConf,
	}
	manager.SetupInformer()
	return manager
}

func (pm *PodManager) GetStandByPod(lang string) (*corev1.Pod, error) {
	pod := pm.UsePod(lang)

	if pod == nil {
		return nil, ErrNoPodsAvailable
	}

	// udpate pod label to mark it as in use
	pod.Labels["app"] = pm.serviceConfig.InUseLabel
	pod, err := pm.client.CoreV1().
		Pods(pm.serviceConfig.Namespace).
		Update(
			pm.ctx,
			pod,
			metav1.UpdateOptions{},
		)
	if err != nil {
		log.Printf("Error updating pod: %v", err)
		pm.ReleasePod(pod.Name)
		return nil, err
	}

	return pod, nil
}

func (pm *PodManager) GetStandByPodRetry(lang string) (*corev1.Pod, error) {
	for i := 1; i <= pm.serviceConfig.GetPodRetries; i++ {
		pod, err := pm.GetStandByPod(lang)
		if pod != nil && err == nil {
			log.Printf("Using pod: %s", pod.Name)
			return pod, nil
		}
		log.Printf("No standby pods available, retrying in %d seconds", i)
		time.Sleep(time.Duration(i) * time.Second)
	}
	return nil, ErrNoPodsAvailable
}

func (pm *PodManager) deletePod(name string) error {
	pm.ReleasePod(name)
	return pm.client.CoreV1().
		Pods(pm.serviceConfig.Namespace).
		Delete(
			pm.ctx,
			name,
			metav1.DeleteOptions{},
		)
}

func (pm *PodManager) ExecuteCode(code string, language string) (string, error) {
	// get a standby pod and retry if all busy
	pod, err := pm.GetStandByPodRetry(language)

	if err != nil {
		return "", err
	}
	defer pm.deletePod(pod.Name)

	timeout := pm.serviceConfig.Timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	output, err := pm.executeCodeCmd(ctx, pod, code, language)

	// Check if the context was canceled due to timeout
	if ctx.Err() == context.DeadlineExceeded {
		return "", NewErrorf(ErrorCodeTimeout, "exection timed out after %s", timeout)
	}

	if err != nil {
		return "", NewErrorf(ErrorCodeExecutionErr, "execution error: %v, output: %s", err.Error(), string(output))
	}

	return output, err
}

// func (pm *PodManager) executeCodeCmd(ctx context.Context, pod *corev1.Pod, code string, language string) (string, error) {
// 	cmd, err := pm.serviceConfig.GetCmd(language)
// 	if err != nil {
// 		return "", err
// 	}
// 	cmd = append(cmd, "'"+code+"'")

// 	req := pm.client.CoreV1().RESTClient().
// 		Post().
// 		Resource("pods").
// 		Name(pod.Name).
// 		Namespace(pm.serviceConfig.Namespace).
// 		SubResource("exec").
// 		Param("stdin", "true").
// 		Param("stdout", "true").
// 		Param("stderr", "true").
// 		Param("tty", "false")

// 	for _, c := range cmd {
// 		req = req.Param("command", c)
// 	}

// 	exec, err := remotecommand.NewSPDYExecutor(
// 		pm.config,
// 		"POST",
// 		req.URL(),
// 	)
// 	if err != nil {
// 		return "", err
// 	}
// 	var stdout, stderr bytes.Buffer
// 	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
// 		Stdin:  nil, // No input needed
// 		Stdout: &stdout,
// 		Stderr: &stderr,
// 	})
// 	if err != nil {
// 		return "", err
// 	}

// 	// Check if there's any error output
// 	if stderr.Len() > 0 {
// 		return stdout.String(), fmt.Errorf(stderr.String())
// 	}

// 	// Return stdout
// 	return stdout.String(), nil
// }

func (pm *PodManager) executeCodeCmd(ctx context.Context, pod *corev1.Pod, code string, language string) (string, error) {
	var cmd *exec.Cmd

	switch language {
	case "python":
		cmd = exec.Command("kubectl", "exec", "-n", "code-exec-system", pod.Name, "--", "python", "-c", code)
	default:
		return "", NewErrorf(ErrorCodeUnsupportLanguage, "unsupported language: %s", language)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", WrapErrorf(err, ErrorCodeExecutionErr, "Execution error: %s", string(output))
	}

	return string(output), nil
}

func (pm *PodManager) HealthzCheck() HealthzResp {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	totale := 0

	StandbyPods := map[string][]string{}
	for lang, pods := range pm.standByPodsByLang {
		StandbyPods[lang] = make([]string, len(pods))
		totale += len(pods)
		for i, pod := range pods {
			StandbyPods[lang][i] = pod.Name
		}
	}
	inuse := []string{}
	for pod := range pm.inUsePods {
		totale++
		inuse = append(inuse, pod)
	}

	return HealthzResp{
		StandbyPods: StandbyPods,
		InUsePods:   inuse,
		TotalPods:   totale,
	}
}
