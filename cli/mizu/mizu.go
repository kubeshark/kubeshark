package mizu

import (
	"context"
	"fmt"
	"github.com/up9inc/mizu/cli/kubernetes"
	"regexp"
)

func Run(podRegex *regexp.Regexp) {
	kubernetesProvider := kubernetes.New("")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	added, removed := kubernetes.FilteredWatch(ctx, kubernetesProvider.GetPodWatcher(ctx), podRegex, nil, nil)

	func() {
		for {
			select {
				case newTarget := <- added:
					fmt.Printf("new pod %s\n", newTarget.Pod)

				case removedTarget := <- removed:
					fmt.Printf("removed pod %s\n", removedTarget.Pod)

				case <- ctx.Done():
					return
			}
		}
	}()
}

