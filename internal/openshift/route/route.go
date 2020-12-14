package route

import (
	"context"
	routev1 "github.com/openshift/api/route/v1"
	clientset "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func List(master, configPath, namespace string) ([]routev1.Route, error) {
	config, err := clientcmd.BuildConfigFromFlags(master, configPath)
	if err != nil {
		return nil, err
	}

	client, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	routes, err := client.Routes(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return routes.Items, nil
}
