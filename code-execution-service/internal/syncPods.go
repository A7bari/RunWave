package internal

import (
	"log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

// Pop returns the first standby pod for a given language and removes it from the list.
func (pm *PodManager) UsePod(language string) *corev1.Pod {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if pods, exists := pm.standByPodsByLang[language]; exists && len(pods) > 0 {
		pod := pods[0]
		pm.standByPodsByLang[language] = pods[1:]
		pm.inUsePods[pod.Name] = true
		return pod
	}
	return nil
}

func (pm *PodManager) ReleasePod(name string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	delete(pm.inUsePods, name)
}

func (pm *PodManager) SetupInformer() {
	informerFactory := informers.NewSharedInformerFactoryWithOptions(
		pm.client,
		0,
		informers.WithNamespace(pm.serviceConfig.Namespace),
	)
	podInformer := informerFactory.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    pm.onPodAdd,
		UpdateFunc: pm.onPodUpdate,
		DeleteFunc: pm.onPodDelete,
	})
	go informerFactory.Start(pm.ctx.Done())
	informerFactory.WaitForCacheSync(pm.ctx.Done())
}

func (pm *PodManager) onPodAdd(obj interface{}) {
	pod := obj.(*corev1.Pod)
	pm.handlePodSync(pod)
}

func (pm *PodManager) onPodUpdate(oldObj, newObj interface{}) {
	pod := newObj.(*corev1.Pod)
	pm.handlePodSync(pod)
}

// handlePodSync stores the full pod object for each language.
func (pm *PodManager) handlePodSync(pod *corev1.Pod) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	language, exists := pod.Labels["language"]
	if !exists {
		return
	}

	// ensure the pod is not in use
	if _, exists := pm.inUsePods[pod.Name]; exists || pod.Labels["app"] == pm.serviceConfig.InUseLabel {
		return
	}

	if _, exists := pm.standByPodsByLang[language]; !exists {
		pm.standByPodsByLang[language] = []*corev1.Pod{}
	}

	if pod.Status.Phase == corev1.PodRunning {
		pods := pm.standByPodsByLang[language]
		found := false
		for _, existingPod := range pods {
			if existingPod.Name == pod.Name {
				found = true
				break
			}
		}
		if !found {
			log.Printf("Adding pod %s to standby list for language %s", pod.Name, language)
			pm.standByPodsByLang[language] = append(pods, pod)
		}
	}
}

func (pm *PodManager) onPodDelete(obj interface{}) {
	pod := obj.(*corev1.Pod)
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if _, exists := pm.inUsePods[pod.Name]; exists {
		delete(pm.inUsePods, pod.Name)
		return
	}

	language, exists := pod.Labels["language"]
	if !exists {
		return
	}

	if pods, exists := pm.standByPodsByLang[language]; exists {
		for i, storedPod := range pods {
			if storedPod.Name == pod.Name {
				pm.standByPodsByLang[language] = append(pods[:i], pods[i+1:]...)
				break
			}
		}
	}
}
