package main

import (
	"context"
	"fmt"
	"log"
	"encoding/base64"
	"k8s.io/client-go/kubernetes"
	"github.com/spf13/viper"
	"google.golang.org/api/container/v1"
	"k8s.io/api"
	"k8s.io/apimachinery"
	"k8s.io/client-go"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// Message is the payload of a Pub/Sub event.  Initial build is Stackdriver alert -> pub/sub ->
type Message struct {
	Data []byte `json:"data"`
}

func main() {
	// Read in all ENVs
	viper.AutomaticEnv()

	// Expect that these are set via environmental variable.  If not then fail with log message
	if viper.Get("PROJECT_ID") == "" {
		log.Fatal("PROJECT_ID must be set")
	}

	if viper.Get("CLUSTER_NAME") == "" {
		log.Fatal("CLUSTER_NAME must be set")
	}

	if viper.Get("CLUSTER_NAMESPACE") == "" {
		log.Fatal("CLUSTER_NAMESPACE must be set")
	}

	if viper.Get("POD_LABEL") == "" {
		log.Fatal("POD_LABEL must be set")
	}

	var projectID string = viper.Get("PROJECT_ID")
	var clusterName string = viper.Get("CLUSTER_NAME")

}

func RemovePod(ctx context.Context, prjectId string, labelSelector metav1.LabelSelector ){
	kubeConfig, err := AcquireContext(ctx, projectId)
	if err != nil {
		return err
	}

	for clusterName := range kubeConfig.Clusters {
		cfg, err := clientcmd.NewNonInteractiveClientConfig(*kubeConfig, clusterName, &clientcmd.ConfigOverrides{CurrentContext: clusterName}, nil).ClientConfig()
		if err != nil {
			return fmt.Errorf("failed to create k8s config: %w", err)
		}
		k8s, err := kubernetes.NewConfigFor(ctx)
		if err != nil {
			return fmt.Errorf("failed to init k8s client for cluster: %w", err)
		}

		// podList, err := k8s.CoreV1().Pods.List(metav1.ListOptions{})
		ns, err := k8s.CoreV1().Namespaces().List(metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list namespaces cluster=%s: %w", clusterName, err)
		}

		log.Printf("Namespaces found in cluster=%s", clusterName)

		for _, item := range ns.Items {
			log.Println(item.Name)
		}

	}
}

// AcquireContext is used to setup GKE context and allow interacting with the clusters apis
func AcquireContext(ctx context.Context, projectID string) (*api.Config, error) {
	svc, err := container.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("container.NewService: %w", err)
	}

	ret := api.Config{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters:   map[string]*api.Cluster{},  //reference names to cluster configs
		AuthInfos:  map[string]*api.AuthInfo{}, //ref names to user configs
		Contexts:   map[string]*api.Context{},  //ref name to context config
	}


	// Get the Cluster based on the CLUSTER_NAME variable being passed
	resp, err := svc.Projects.Zones.Clusters.Get(projectId string, zone string, clusterId string //call via CLUSTER_NAME var for now)
	if err != nil {
		return nil, fmt.Errorf("cluster get error project %s %w", prjectId, err)
	}

	for _, f := range resp.Clusters {
		name := fmt.Sprintf(f.Name)
		location := fmt.Sprintf(f.Location)
		cert, err := base64.StdEncoding.DecodeString(f.masterAuth.clusterCaCertificate)
		if err != nil {
			return nil, fmt.Errorf("invalid cert for cluster=%s cert=%s: %w" name, f.masterAuth.clusterCaCertificate, err)
		}
		ret.Clusters[name] = &api.Cluster{
			CertificateAuthorityData: cert,
			Server: "https://" + f.Endpoint,
		}
		ret.Contexts[name] = &api.Contexts{
			Cluster: name,
			AuthInfo: name,
		}
		ret.AuthInfos[name] = &api.AuthInfo{
			AuthProvider: &api.AuthProviderConfig{
				Name: "gcp",
				Config: map[string]string{
					"scopes": "https://www.googleapis.com/auth/cloud-platform",
				}
			}
		}
	}
}
