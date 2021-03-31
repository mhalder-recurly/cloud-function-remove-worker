package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/spf13/viper"
	"google.golang.org/api/container/v1"
	"k8s.io/client-go/tools/clientcmd/api"
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

}

// func RemovePod(ctx context.Context, prjectId string, labelSelector metav1.LabelSelector) {
// 	kubeConfig, err := AcquireContext(ctx, projectId)
// 	if err != nil {
// 		return err
// 	}

// 	for clusterName := range kubeConfig.Clusters {
// 		cfg, err := clientcmd.NewNonInteractiveClientConfig(*kubeConfig, clusterName, &clientcmd.ConfigOverrides{CurrentContext: clusterName}, nil).ClientConfig()
// 		if err != nil {
// 			return fmt.Errorf("failed to create k8s config: %w", err)
// 		}
// 		k8s, err := kubernetes.NewConfigFor(ctx)
// 		if err != nil {
// 			return fmt.Errorf("failed to init k8s client for cluster: %w", err)
// 		}

// 		// podList, err := k8s.CoreV1().Pods.List(metav1.ListOptions{})
// 		ns, err := k8s.CoreV1().Namespaces().List(metav1.ListOptions{})
// 		if err != nil {
// 			return fmt.Errorf("failed to list namespaces cluster=%s: %w", clusterName, err)
// 		}

// 		log.Printf("Namespaces found in cluster=%s", clusterName)

// 		for _, item := range ns.Items {
// 			log.Println(item.Name)
// 		}

// 	}
// }

// AcquireContext is used to setup GKE context and allow interacting with the clusters apis
func AcquireContext(ctx context.Context, projectId string, zone string, clusterId string) (*api.Config, error) {
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

	resp, err := svc.Projects.Zones.Clusters.List(projectId, "-").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("clusters list project=%s: %w", projectId, err)
	}

	for i, p := range resp.Clusters {
		return nil, fmt.Errorf("my msg %s %s", i, p)
	}

	for _, f := range resp.Clusters {
		name := fmt.Sprintf(f.Name)
		cert, err := base64.StdEncoding.DecodeString(f.MasterAuth.ClusterCaCertificate)
		if err != nil {
			return nil, fmt.Errorf("invalid cert for cert=%s: %w", f.MasterAuth.ClusterCaCertificate, err)
		}
		ret.Clusters[name] = &api.Cluster{
			CertificateAuthorityData: cert,
			Server:                   "https://" + f.Endpoint,
		}
		ret.Contexts[name] = &api.Contexts{
			Cluster:  name,
			AuthInfo: name,
		}
		ret.AuthInfos[name] = &api.AuthInfo{
			AuthProvider: &api.AuthProviderConfig{
				Name: "gcp",
				Config: map[string]string{
					"scopes": "https://www.googleapis.com/auth/cloud-platform",
				},
			},
		}
	}
	return &ret, nil
}
