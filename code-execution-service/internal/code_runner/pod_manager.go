package coderunner

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	inUsePodLabel = "in-use-pod"
)

type RunOutput struct {
	Output  string
	IsError bool
}

// PodManager is an interface for pod manager
type PodManager interface {

	// Run executes a command in a pod
	Run(commade []string, podname string, timeout time.Duration) (*RunOutput, error)

	// Get a pod to run a command
	// and Lock the pod for the caller
	ConsumePod(lang string) string
}

// PodManagerImp is the implementation of PodManager
type PodManagerImp struct {
	client    kubernetes.Interface
	config    *rest.Config
	namespace string
	inUsePods map[string]*corev1.Pod
	mu        sync.Mutex
	ctx       context.Context

	// freePods is a map of pods that are not in use
	// for each language which is set as label in the pod
	// the key is the language
	freePods map[string][]*corev1.Pod

	// cond is used to signal waiting goroutines
	// when a pod is added to the freePods list
	cond *sync.Cond
}

// NewPodManager creates a new PodManager
func NewPodManager(client kubernetes.Interface, config *rest.Config, namespace string) PodManager {
	pm := &PodManagerImp{
		client:    client,
		config:    config,
		namespace: namespace,
		freePods:  make(map[string][]*corev1.Pod, 0),
		inUsePods: make(map[string]*corev1.Pod),
		ctx:       context.Background(),
	}

	pm.cond = sync.NewCond(&pm.mu)
	go pm.setupInformer()

	return pm
}

// ConsumePod gets a available pod for a language
// blocks until a pod is available
// once a pod is returned, can't be consumed by other callers
func (p *PodManagerImp) ConsumePod(lang string) string {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Block until a pod is available for the given language
	for len(p.freePods[lang]) == 0 {
		// Wait until a pod is available, possibly for another language
		p.cond.Wait()
	}

	pod := p.freePods[lang][0]
	p.freePods[lang] = p.freePods[lang][1:]
	p.inUsePods[pod.Name] = pod

	return pod.Name
}

// Run executes a command in a pod
// the pod should be consumed first
// the pod will be released after the command is executed
func (p *PodManagerImp) Run(command []string, podname string, timeout time.Duration) (*RunOutput, error) {
	pod, ok := p.inUsePods[podname]
	if !ok {
		return nil, fmt.Errorf("pod manager: pod not found, pod should be consumed first")
	}
	// change the pods label to in-use-pod
	// so K8b will automatically create a fresh replica
	// to be used in the next requests
	go p.setLabel(pod, inUsePodLabel)

	// Prepare the request for executing the command in the pod
	req := p.client.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(p.namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: pod.Spec.Containers[0].Name,
			Command:   command,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	// Create an executor for the command
	url := req.URL()
	exec, err := remotecommand.NewSPDYExecutor(p.config, "POST", url)
	if err != nil {
		log.Printf("Pod Manager: Failed to create executor for pod %s: %v", podname, err)
		return nil, err
	}

	// Buffers to capture stdout and stderr
	var stdout, stderr bytes.Buffer

	// Execute the command
	ctx, cancel := context.WithTimeout(p.ctx, timeout)
	defer cancel()
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})
	if err != nil {
		log.Printf("Pod Manager: Failed to execute command: %s in pod %s with err: %v", command, podname, err)
		return nil, err
	}

	// release at the end of execution
	defer func() {
		go p.releasePod(podname)
	}()

	// Check if there is any error output
	if stderr.Len() > 0 {
		return &RunOutput{
			Output:  stderr.String(),
			IsError: true,
		}, nil
	}

	// Return the standard output
	return &RunOutput{
		Output:  stdout.String(),
		IsError: false,
	}, nil
}

// ReleasePod releases a pod to the queue
func (p *PodManagerImp) releasePod(podname string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	err := p.client.CoreV1().
		Pods(p.namespace).
		Delete(p.ctx, podname, metav1.DeleteOptions{})

	if err != nil {
		log.Printf("Pod Manager: Failed to delete pod %s: %v", podname, err)
	}

	delete(p.inUsePods, podname)
}

func (p *PodManagerImp) setLabel(pod *corev1.Pod, label string) error {
	// Update the pod's labels to mark it as in use
	pod.Labels["app"] = label
	_, err := p.client.CoreV1().
		Pods(p.namespace).
		Update(p.ctx, pod, metav1.UpdateOptions{})

	if err != nil {
		return err
	}

	return nil
}

// AddPod adds a pod to the freePods list and signals waiting goroutines
func (p *PodManagerImp) addPod(pod *corev1.Pod) {
	p.mu.Lock()
	defer p.mu.Unlock()

	lang := pod.Labels["language"]

	p.freePods[lang] = append(p.freePods[lang], pod)
	p.cond.Signal() // Wake up one waiting goroutine
}

// SetupInformer sets up the informer for the created pod
// each pod represents a worker that can execute code
// the informer listens for pod events and updates the scheduler accordingly
func (p *PodManagerImp) setupInformer() {
	informerFactory := informers.NewSharedInformerFactoryWithOptions(
		p.client,
		0,
		informers.WithNamespace(p.namespace),
	)
	podInformer := informerFactory.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				return
			}

			// Only process pods in the Running state
			if pod.Status.Phase != corev1.PodRunning {
				log.Printf("Pod %s is not in Running state (current state: %s); skipping add", pod.Name, pod.Status.Phase)
				return
			}

			if pod.Labels["app"] != inUsePodLabel {
				p.addPod(pod)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldPod, okOld := oldObj.(*corev1.Pod)
			newPod, okNew := newObj.(*corev1.Pod)
			if !okOld || !okNew {
				return
			}

			// Check if the pod transitioned to the Running state
			if oldPod.Status.Phase != corev1.PodRunning && newPod.Status.Phase == corev1.PodRunning {
				log.Printf("Pod %s transitioned to Running state", newPod.Name)
				if newPod.Labels["app"] != inUsePodLabel {
					p.addPod(newPod)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			if pod.Labels["app"] == inUsePodLabel {
				return
			}
			p.mu.Lock()
			defer p.mu.Unlock()

			// remove the pod from the free pods list
			// if it is deleted
			lang := pod.Labels["language"]

			for i, fp := range p.freePods[lang] {
				if fp.Name == pod.Name {
					p.freePods[lang] = append(p.freePods[lang][:i], p.freePods[lang][i+1:]...)
					return
				}
			}

		},
	})

	go informerFactory.Start(p.ctx.Done())
	informerFactory.WaitForCacheSync(p.ctx.Done())
}
