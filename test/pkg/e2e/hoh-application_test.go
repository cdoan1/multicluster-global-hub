package tests

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	appsv1alpha1 "open-cluster-management.io/multicloud-operators-subscription/pkg/apis/apps/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/stolostron/multicluster-global-hub/test/pkg/utils"
)

const (
	APP_LABEL_KEY     = "app"
	APP_LABEL_VALUE   = "test"
	APP_SUB_YAML      = "../../resources/app/app-pacman-appsub.yaml"
	APP_SUB_NAME      = "pacman-appsub"
	APP_SUB_NAMESPACE = "pacman"
)

var _ = Describe("Deploy the application to the managed cluster", Label("e2e-tests-app"), Ordered, func() {
	var token string
	var httpClient *http.Client
	var managedClusterName1 string
	var managedClusterName2 string
	var appClient client.Client

	BeforeAll(func() {
		By("Get token for the non-k8s-api")
		initToken, err := utils.FetchBearerToken(testOptions)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(len(initToken)).Should(BeNumerically(">", 0))
		token = initToken

		By("Config request of the api")
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		httpClient = &http.Client{Timeout: time.Second * 10, Transport: transport}

		By("Get managed cluster name")
		Eventually(func() error {
			managedClusters := getManagedCluster(httpClient, token)
			if len(managedClusters) <= 1 {
				return fmt.Errorf("wrong number of managed cluster, should be %d, but found %d", 2, len(managedClusters))
			}
			managedClusterName1 = managedClusters[0].Name
			managedClusterName2 = managedClusters[1].Name
			return nil
		}, 5*60*time.Second, 5*1*time.Second).ShouldNot(HaveOccurred())
		Expect(len(managedClusterName1)).Should(BeNumerically(">", 0),
			"managedclustername1 should not be empty")
		Expect(len(managedClusterName2)).Should(BeNumerically(">", 0),
			"managedclustername2 should not be empty")

		By("Get the appsubreport client")
		cfg, err := clients.RestConfig(clients.HubClusterName())
		Expect(err).ShouldNot(HaveOccurred())
		scheme := runtime.NewScheme()
		appsv1alpha1.AddToScheme(scheme)
		appClient, err = client.New(cfg, client.Options{Scheme: scheme})
		Expect(err).ShouldNot(HaveOccurred())
	})

	It(fmt.Sprintf("add the app label[ %s: %s ] to the %s", APP_LABEL_KEY, APP_LABEL_VALUE, managedClusterName1), func() {
		By("Add label to the managedcluster1")
		patches := []patch{
			{
				Op:    "add",
				Path:  "/metadata/labels/" + APP_LABEL_KEY,
				Value: APP_LABEL_VALUE,
			},
		}
		updateClusterLabel(httpClient, patches, token, managedClusterName1)

		By("Check the label is added")
		Eventually(func() error {
			managedClusters := getManagedCluster(httpClient, token)
			for _, cluster := range managedClusters {
				if val, ok := cluster.Labels[APP_LABEL_KEY]; ok {
					if val == APP_LABEL_VALUE && cluster.Name == managedClusterName1 {
						return nil
					}
				}
			}
			return fmt.Errorf("the label %s: %s is not exist", APP_LABEL_KEY, APP_LABEL_VALUE)
		}, 12*60*time.Second, 5*1*time.Second).ShouldNot(HaveOccurred())

		By("Print result after adding the label")
		managedClusters := getManagedCluster(httpClient, token)
		printClusterLabel(managedClusters)
	})

	It("deploy the application/subscription", func() {
		By("Apply the appsub to labeled cluster")
		msg, err := clients.Kubectl(clients.HubClusterName(), "apply", "-f", APP_SUB_YAML)
		Expect(err).ShouldNot(HaveOccurred(), msg)

		By("Check the appsub is applied to the cluster")
		seconds := 0
		Eventually(func() error {
			seconds += 5
			if seconds > 40 {
				seconds = 0
				return checkAppsubreport(appClient, 1, []string{managedClusterName1}, true)
			} else {
				return checkAppsubreport(appClient, 1, []string{managedClusterName1}, false)
			}
		}, 12*60*time.Second, 5*1*time.Second).ShouldNot(HaveOccurred())
	})

	It(fmt.Sprintf("Add the app label[ %s: %s ] to the %s", APP_LABEL_KEY, APP_LABEL_VALUE, managedClusterName2), func() {
		By("Add the lablel to managedcluster2")
		patches := []patch{
			{
				Op:    "add",
				Path:  "/metadata/labels/" + APP_LABEL_KEY,
				Value: APP_LABEL_VALUE,
			},
		}
		updateClusterLabel(httpClient, patches, token, managedClusterName2)

		By("Check the label is added to managedcluster2")
		Eventually(func() error {
			managedClusters := getManagedCluster(httpClient, token)
			for _, cluster := range managedClusters {
				if val, ok := cluster.Labels[APP_LABEL_KEY]; ok {
					if val == APP_LABEL_VALUE && cluster.Name == managedClusterName2 {
						By("Print result after adding the label")
						managedClusters := getManagedCluster(httpClient, token)
						printClusterLabel(managedClusters)
						return nil
					}
				}
			}
			return fmt.Errorf("the label %s: %s is not exist", APP_LABEL_KEY, APP_LABEL_VALUE)
		}, 12*60*time.Second, 5*1*time.Second).ShouldNot(HaveOccurred())

		By("Check the appsub apply to the clusters")
		seconds := 0
		Eventually(func() error {
			seconds += 5
			if seconds > 40 {
				seconds = 0
				return checkAppsubreport(appClient, 2, []string{
					managedClusterName1, managedClusterName2,
				}, true)
			} else {
				return checkAppsubreport(appClient, 2, []string{
					managedClusterName1, managedClusterName2,
				}, false)
			}
		}, 12*60*time.Second, 5*1*time.Second).ShouldNot(HaveOccurred())
	})

	AfterAll(func() {
		By("Remove from clusters")
		patches := []patch{
			{
				Op:    "remove",
				Path:  "/metadata/labels/" + APP_LABEL_KEY,
				Value: APP_LABEL_VALUE,
			},
		}
		updateClusterLabel(httpClient, patches, token, managedClusterName1)
		updateClusterLabel(httpClient, patches, token, managedClusterName2)

		By("Check label is removed from clusters")
		Eventually(func() error {
			managedClusters := getManagedCluster(httpClient, token)
			for _, cluster := range managedClusters {
				if val, ok := cluster.Labels[APP_LABEL_KEY]; ok {
					if val == APP_LABEL_VALUE {
						return fmt.Errorf("the label %s: %s is not removed from %s",
							APP_LABEL_KEY, APP_LABEL_VALUE, cluster.Name)
					}
				}
			}
			return nil
		}, 12*60*time.Second, 5*1*time.Second).ShouldNot(HaveOccurred())

		By("Remove the appsub resource")
		msg, err := clients.Kubectl(clients.HubClusterName(), "delete", "-f", APP_SUB_YAML)
		Expect(err).ShouldNot(HaveOccurred(), msg)
	})
})

func checkAppsubreport(appClient client.Client, expectDeployNum int, expectClusterNames []string, retry bool) error {
	appsubreport := &appsv1alpha1.SubscriptionReport{}
	err := appClient.Get(context.TODO(), types.NamespacedName{Namespace: APP_SUB_NAMESPACE, Name: APP_SUB_NAME}, appsubreport)
	if err != nil {
		return err
	}
	deployNum, err := strconv.Atoi(appsubreport.Summary.Deployed)
	if err != nil {
		return err
	}
	clusterNum, err := strconv.Atoi(appsubreport.Summary.Clusters)
	if err != nil {
		return err
	}
	if deployNum == expectDeployNum && clusterNum >= len(expectClusterNames) {
		matchedClusterNum := 0
		for _, expectClusterName := range expectClusterNames {
			for _, res := range appsubreport.Results {
				if res.Result == "deployed" && res.Source == expectClusterName {
					matchedClusterNum++
				}
			}
		}
		if matchedClusterNum == len(expectClusterNames) {
			report := &appsv1alpha1.SubscriptionReport{
				Summary: appsubreport.Summary,
				Results: appsubreport.Results,
			}
			appsubreportStr, _ := json.MarshalIndent(report, "", "  ")
			klog.V(5).Info("Appsubreport: ", string(appsubreportStr))
			return nil
		}
		return fmt.Errorf("deploy results isn't correct %v", appsubreport.Results)
	}
	appsubreportStr, _ := json.MarshalIndent(appsubreport, "", "  ")
	klog.V(6).Info("Appsubreport: ", string(appsubreportStr))
	if retry {
		msg, err := clients.Kubectl(clients.HubClusterName(), "delete", "-f", APP_SUB_YAML)
		Expect(err).ShouldNot(HaveOccurred(), msg)
		msg, err = clients.Kubectl(clients.HubClusterName(), "apply", "-f", APP_SUB_YAML)
		Expect(err).ShouldNot(HaveOccurred(), msg)
	}
	return fmt.Errorf("the appsub %s: %s hasn't deplyed to the cluster: %s", APP_SUB_NAMESPACE,
		APP_SUB_NAME, strings.Join(expectClusterNames, ","))
}
