## cloud-function-remove-worker
This cloud function code acquires context from a GKE hosted k8s control node and then performs a delete pod operations based on the configuration.

#### Configurations
Configurations are done via ENVs:
PROJECT_ID, CLUSTER_NAME, NAMESPACE, and POD_LABEL
Example:
PROJECT_ID=ext-staging-1-ba731dde
CLUSTER_NAME=ext-staging-gke-cluster
NAMESPACE=ext-staging
POD_LABEL=component=recurly-worker


###### To Do
* There is no longing outside of failures
* Accept single and multiple POD_LABEL configs
* Should each func be broken out to its own package?
* Make the four envs required to be set