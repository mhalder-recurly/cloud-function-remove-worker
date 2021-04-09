package main

// import (
// 	"fmt"
// 	"context"
// 	"os"
// 	"path/filepath"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
// 	"k8s.io/client-go/kubernetes"
// 	"k8s.io/client-go/tools/clientcmd"
// )

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/spf13/viper"
	"google.golang.org/api/container/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Message is the payload of a Pub/Sub event.  Initial build is Stackdriver alert -> pub/sub ->
type Message struct {
	Data []byte `json:"data"`
}

func main() {
	v := viper.New()
	// Read in all ENVs
	v.AutomaticEnv()

	project := v.Get("PROJECT_ID")
	cluster := v.Get("CLUSTER_NAME")
	namespace := v.Get("NAMESPACE")
	podlabel := v.Get("POD_LABEL")

	projectID := fmt.Sprint(project)
	clusterName := fmt.Sprint(cluster)
	nameSpace := fmt.Sprint(namespace)
	podLabel := fmt.Sprint(podlabel)

	if err := RemovePod(context.Background(), projectID, clusterName, nameSpace, podLabel); err != nil {
		log.Fatal(err)
	}

}

// RemovePod used the auth and provided variables to the remove nonfunctioning pod
func RemovePod(ctx context.Context, projectID string, clusterID string, nameSpace string, podLabel string) error {
	kubeConfig, err := AcquireContext(ctx, projectID, clusterID)
	if err != nil {
		return err
	}

	for clusterName := range kubeConfig.Clusters {
		cfg, err := clientcmd.NewNonInteractiveClientConfig(*kubeConfig, clusterName, &clientcmd.ConfigOverrides{CurrentContext: clusterName}, nil).ClientConfig()
		if err != nil {
			return fmt.Errorf("failed to create k8s cluster=%s config: %s", clusterName, err)
		}

		k8s, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			return err
		}

		opts := metav1.ListOptions{
			LabelSelector: podLabel,
		}


	  // delPod, err := k8s.CoreV1().Pods(nameSpace).Delete()
		// if err != nil {
		// 	return fmt.Errorf("failed to list pods cluster=%s: %s", clusterName, err)
		// }


		podList,err := k8s.CoreV1().Pods(nameSpace).List(ctx, opts)
		if err != nil {
			return fmt.Errorf("failed to list pods cluster=%s: %s", clusterName, err)
		}

		for _, podInfo := range (*podList).Items {
			err := k8s.CoreV1().Pods(podInfo.Namespace).Delete(ctx, podInfo.Name, metav1.DeleteOptions{})
			if err != nil {
				log.Fatal(err)
			}
			// _, err := fmt.Printf("%v\n", podInfo.Name)
			// if err != nil {
			// 	return fmt.Errorf("something went wrong %v", err)
			// }
		}
	}
	return nil
}

//AcquireContext gotten from here https://bionic.fullstory.com/connect-to-google-kubernetes-with-gcp-credentials-and-pure-golang/
//Used to setup GKE context and allow interacting with the clusters apis
func AcquireContext(ctx context.Context, projectID string, clusterID string) (*api.Config, error) {
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

	resp, err := svc.Projects.Zones.Clusters.List(projectID, "-").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("clusters list project=%s: %w", projectID, err)
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
		ret.Contexts[name] = &api.Context{
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
	return &ret, err
}
