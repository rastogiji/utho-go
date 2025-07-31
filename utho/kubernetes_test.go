package utho

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestKubernetesService_Create_happyPath(t *testing.T) {
	token := "token"
	var payload CreateKubernetesParams
	_ = faker.FakeData(&payload)

	client, mux, _, teardown := setup(token)
	defer teardown()

	var dummyResponse CreateResponse
	_ = faker.FakeData(&dummyResponse)
	dummyResponse.Status = "success"

	serverResponse, _ := json.Marshal(dummyResponse)

	mux.HandleFunc("/kubernetes/deploy", func(w http.ResponseWriter, req *http.Request) {
		testHttpMethod(t, req, http.MethodPost)
		testHeader(t, req, "Authorization", "Bearer "+token)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(serverResponse))
	})

	ctx := context.Background()
	got, err := client.Kubernetes().Create(ctx, payload)

	assert.Nil(t, err)
	assert.Equal(t, dummyResponse, *got)
}

func TestKubernetesService_Create_invalidServer(t *testing.T) {
	client, _ := NewClient("token")

	ctx := context.Background()
	_, err := client.Kubernetes().Create(ctx, CreateKubernetesParams{})
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
}

func TestKubernetesService_Read_happyPath(t *testing.T) {
	client, mux, _, teardown := setup("token")
	defer teardown()

	randomInts, err := faker.RandomInt(1000, 9999)
	if err != nil || len(randomInts) == 0 {
		t.Fatalf("failed to generate random int: %v", err)
	}
	clusterId := int(randomInts[0])
	var dummyCluster KubernetesRead
	_ = faker.FakeData(&dummyCluster)
	dummyCluster.Info.Cluster.ID = clusterId

	serverResponse, _ := json.Marshal(dummyCluster)			

	URL := fmt.Sprintf("/kubernetes/%d", clusterId)
	mux.HandleFunc(URL, func(w http.ResponseWriter, req *http.Request) {
		testHttpMethod(t, req, "GET")
		testHeader(t, req, "Authorization", "Bearer token")
		fmt.Fprint(w, string(serverResponse))
	})

	ctx := context.Background()
	got, _ := client.Kubernetes().Read(ctx, clusterId)
	if !reflect.DeepEqual(*got, dummyCluster) {
		t.Errorf("Response = %v, want %v", *got, dummyCluster)
	}
}

func TestKubernetesService_Read_invalidServer(t *testing.T) {
	client, _ := NewClient("token")

	ctx := context.Background()
	cluster, err := client.Kubernetes().Read(ctx, 0000)
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
	if cluster != nil {
		t.Errorf("Was not expecting any cluster to be returned, instead got %v", cluster)
	}
}

func TestKubernetesService_List_happyPath(t *testing.T) {
	client, mux, _, teardown := setup("token")
	defer teardown()

	var dummyK8sList []K8s
	for i := 0; i < 3; i++ { // Generate a list of 3 dummy Kubernetes clusters
		var k8s K8s
		_ = faker.FakeData(&k8s)
		dummyK8sList = append(dummyK8sList, k8s)
	}

	serverResponse, _ := json.Marshal(KubernetesList{
		K8s:     dummyK8sList,
		Status:  "success",
		Message: "success",
	})

	mux.HandleFunc("/kubernetes", func(w http.ResponseWriter, req *http.Request) {
		testHttpMethod(t, req, "GET")
		testHeader(t, req, "Authorization", "Bearer token")
		fmt.Fprint(w, string(serverResponse))
	})

	ctx := context.Background()
	got, _ := client.Kubernetes().List(ctx)
	if len(got) != len(dummyK8sList) {
		t.Errorf("Was expecting %d Kubernetes clusters to be returned, instead got %d", len(dummyK8sList), len(got))
	}

	assert.Equal(t, dummyK8sList, got)
}

func TestKubernetesService_List_invalidServer(t *testing.T) {
	client, _ := NewClient("token")

	ctx := context.Background()
	stacks, err := client.Kubernetes().List(ctx)
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
	if stacks != nil {
		t.Errorf("Was not expecting any stacks to be returned, instead got %v", stacks)
	}
}

func TestKubernetesService_Delete_happyPath(t *testing.T) {
	token := "token"
	payload := DeleteKubernetesParams{
		ClusterId: 1111,
		Confirm:   "I am aware this action will delete data and cluster permanently",
	}

	client, mux, _, teardown := setup(token)
	defer teardown()

	URL := fmt.Sprintf("/kubernetes/%d/destroy", payload.ClusterId)
	mux.HandleFunc(URL, func(w http.ResponseWriter, req *http.Request) {
		testHttpMethod(t, req, http.MethodDelete)
		testHeader(t, req, "Authorization", "Bearer "+token)
		fmt.Fprint(w, dummyDeleteResponseJson)
	})

	want := DeleteResponse{Status: "success", Message: "success"}

	ctx := context.Background()
	got, _ := client.Kubernetes().Delete(ctx, payload)
	if !reflect.DeepEqual(*got, want) {
		t.Errorf("Response = %v, want %v", *got, want)
	}
}

func TestKubernetesService_Delete_invalidServer(t *testing.T) {
	client, _ := NewClient("token")

	ctx := context.Background()
	delResponse, err := client.Kubernetes().Delete(ctx, DeleteKubernetesParams{})
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
	if delResponse != nil {
		t.Errorf("Was not expecting any reponse to be returned, instead got %v", delResponse)
	}
}

// K8s Loadbalancer test
func TestKubernetesServices_CreateLoadbalancer_happyPath(t *testing.T) {
	token := "token"

	payload := CreateKubernetesLoadbalancerParams{
		ClusterId:      1111,
		LoadbalancerId: 2222,
	}

	client, mux, _, teardown := setup(token)
	defer teardown()

	URL := fmt.Sprintf("/kubernetes/%d/loadbalancer/%d", payload.ClusterId, payload.LoadbalancerId)
	mux.HandleFunc(URL, func(w http.ResponseWriter, req *http.Request) {
		testHttpMethod(t, req, http.MethodPost)
		testHeader(t, req, "Authorization", "Bearer "+token)
		fmt.Fprint(w, dummyCreateBasicResponseJson)
	})

	ctx := context.Background()
	got, err := client.Kubernetes().CreateLoadbalancer(ctx, payload)

	var want CreateResponse
	_ = json.Unmarshal([]byte(dummyCreateBasicResponseJson), &want)

	assert.Nil(t, err)
	assert.Equal(t, want, *got)
}

func TestKubernetesServices_CreateLoadbalancer_invalidServer(t *testing.T) {
	client, _ := NewClient("token")

	ctx := context.Background()
	_, err := client.Kubernetes().CreateLoadbalancer(ctx, CreateKubernetesLoadbalancerParams{})
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
}

func TestKubernetesServices_ReadLoadbalancer_happyPath(t *testing.T) {
	client, mux, _, teardown := setup("token")
	defer teardown()

	randomInts, err := faker.RandomInt(1000, 9999)
	if err != nil || len(randomInts) == 0 {
		t.Fatalf("failed to generate random int: %v", err)
	}
	clusterId := int(randomInts[0])
	loadbalancerId := randomInts[1]
	var dummyLoadbalancer K8sLoadbalancers
	_ = faker.FakeData(&dummyLoadbalancer)
	dummyLoadbalancer.ID = loadbalancerId

	serverResponse, _ := json.Marshal(KubernetesRead{
		LoadBalancers: []K8sLoadbalancers{dummyLoadbalancer},
		Status:        "success",
	})

	URL := fmt.Sprintf("/kubernetes/%d", clusterId)
	mux.HandleFunc(URL, func(w http.ResponseWriter, req *http.Request) {
		testHttpMethod(t, req, "GET")
		testHeader(t, req, "Authorization", "Bearer token")
		fmt.Fprint(w, string(serverResponse))
	})

	ctx := context.Background()
	got, _ := client.Kubernetes().ReadLoadbalancer(ctx, clusterId, loadbalancerId)
	assert.Equal(t, dummyLoadbalancer, *got)
}

func TestKubernetesServices_ReadLoadbalancer_invalidServer(t *testing.T) {
	client, _ := NewClient("token")

	ctx := context.Background()
	loadbalancer, err := client.Kubernetes().ReadLoadbalancer(ctx, 1111, 122134)
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
	if loadbalancer != nil {
		t.Errorf("Was not expecting any loadbalancer to be returned, instead got %v", loadbalancer)
	}
}

func TestKubernetesServices_ListLoadbalancer_happyPath(t *testing.T) {
	client, mux, _, teardown := setup("token")
	defer teardown()

	randomInts, err := faker.RandomInt(1000, 9999)
	if err != nil || len(randomInts) == 0 {
		t.Fatalf("failed to generate random int: %v", err)
	}
	clusterId := int(randomInts[0])
	var dummyLoadbalancers []K8sLoadbalancers
	for i := 0; i < 3; i++ {
		var loadbalancer K8sLoadbalancers
		_ = faker.FakeData(&loadbalancer)
		dummyLoadbalancers = append(dummyLoadbalancers, loadbalancer)
	}

	serverResponse, _ := json.Marshal(KubernetesRead{
		LoadBalancers: dummyLoadbalancers,
		Status:        "success",
	})

	URL := fmt.Sprintf("/kubernetes/%d", clusterId)
	mux.HandleFunc(URL, func(w http.ResponseWriter, req *http.Request) {
		testHttpMethod(t, req, "GET")
		testHeader(t, req, "Authorization", "Bearer token")
		fmt.Fprint(w, string(serverResponse))
	})

	ctx := context.Background()
	got, _ := client.Kubernetes().ListLoadbalancers(ctx, clusterId)
	assert.Equal(t, dummyLoadbalancers, got)
}

func TestKubernetesServices_ListLoadbalancer_invalidServer(t *testing.T) {
	client, _ := NewClient("token")

	ctx := context.Background()
	loadbalancer, err := client.Kubernetes().ListLoadbalancers(ctx, 1111)
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
	if loadbalancer != nil {
		t.Errorf("Was not expecting any kubernetes loadbalancer to be returned, instead got %v", loadbalancer)
	}
}

func TestKubernetesServices_DeleteLoadbalancer_happyPath(t *testing.T) {
	token := "token"
	clusterId := 1111
	loadbalancerId := 22222

	client, mux, _, teardown := setup(token)
	defer teardown()

	URL := fmt.Sprintf("/kubernetes/%d/loadbalancer/%d", clusterId, loadbalancerId)
	mux.HandleFunc(URL, func(w http.ResponseWriter, req *http.Request) {
		testHttpMethod(t, req, http.MethodDelete)
		testHeader(t, req, "Authorization", "Bearer "+token)
		fmt.Fprint(w, dummyDeleteResponseJson)
	})

	want := DeleteResponse{Status: "success", Message: "success"}

	ctx := context.Background()
	got, _ := client.Kubernetes().DeleteLoadbalancer(ctx, clusterId, loadbalancerId)
	if !reflect.DeepEqual(*got, want) {
		t.Errorf("Response = %v, want %v", *got, want)
	}
}

func TestKubernetesServices_DeleteLoadbalancer_invalidServer(t *testing.T) {
	client, _ := NewClient("token")

	ctx := context.Background()
	delResponse, err := client.Kubernetes().DeleteLoadbalancer(ctx, 1111, 123543)
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
	if delResponse != nil {
		t.Errorf("Was not expecting any reponse to be returned, instead got %v", delResponse)
	}
}

// K8s SecurityGroup test
func TestKubernetesServices_CreateSecurityGroup_happyPath(t *testing.T) {
	token := "token"

	payload := CreateKubernetesSecurityGroupParams{
		ClusterId:                 1111,
		KubernetesSecurityGroupId: 44444,
	}

	client, mux, _, teardown := setup(token)
	defer teardown()

	URL := fmt.Sprintf("/kubernetes/%d/securitygroup/%d", payload.ClusterId, payload.KubernetesSecurityGroupId)
	mux.HandleFunc(URL, func(w http.ResponseWriter, req *http.Request) {
		testHttpMethod(t, req, http.MethodPost)
		testHeader(t, req, "Authorization", "Bearer "+token)
		fmt.Fprint(w, dummyCreateBasicResponseJson)
	})

	ctx := context.Background()
	got, err := client.Kubernetes().CreateSecurityGroup(ctx, payload)

	var want CreateResponse
	_ = json.Unmarshal([]byte(dummyCreateBasicResponseJson), &want)

	assert.Nil(t, err)
	assert.Equal(t, want, *got)
}

func TestKubernetesServices_CreateSecurityGroup_invalidServer(t *testing.T) {
	client, _ := NewClient("token")

	ctx := context.Background()
	_, err := client.Kubernetes().CreateSecurityGroup(ctx, CreateKubernetesSecurityGroupParams{})
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
}

func TestKubernetesServices_ReadSecurityGroup_happyPath(t *testing.T) {
	client, mux, _, teardown := setup("token")
	defer teardown()

	randomInts, err := faker.RandomInt(1000, 9999)
	if err != nil || len(randomInts) == 0 {
		t.Fatalf("failed to generate random int: %v", err)
	}
	clusterId := randomInts[0]
	securityGroupId := randomInts[1]
	var dummySecurityGroup K8sSecurityGroups
	_ = faker.FakeData(&dummySecurityGroup)
	dummySecurityGroup.ID = securityGroupId

	serverResponse, _ := json.Marshal(KubernetesRead{
		SecurityGroups: []K8sSecurityGroups{dummySecurityGroup},
		Status:         "success",
	})

	URL := fmt.Sprintf("/kubernetes/%d", clusterId)
	mux.HandleFunc(URL, func(w http.ResponseWriter, req *http.Request) {
		testHttpMethod(t, req, "GET")
		testHeader(t, req, "Authorization", "Bearer token")
		fmt.Fprint(w, string(serverResponse))
	})

	ctx := context.Background()
	got, _ := client.Kubernetes().ReadSecurityGroup(ctx, clusterId, securityGroupId)
	assert.Equal(t, dummySecurityGroup, *got)
}

func TestKubernetesServices_ReadSecurityGroup_invalidServer(t *testing.T) {
	client, _ := NewClient("token")

	ctx := context.Background()
	securityGroup, err := client.Kubernetes().ReadSecurityGroup(ctx, 1111, 122134)
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
	if securityGroup != nil {
		t.Errorf("Was not expecting any securityGroup to be returned, instead got %v", securityGroup)
	}
}

func TestKubernetesServices_ListSecurityGroup_happyPath(t *testing.T) {
	client, mux, _, teardown := setup("token")
	defer teardown()

	randomInts, err := faker.RandomInt(1000, 9999)
	if err != nil || len(randomInts) == 0 {
		t.Fatalf("failed to generate random int: %v", err)
	}
	clusterId := randomInts[0]
	var dummySecurityGroups []K8sSecurityGroups
	for i := 0; i < 3; i++ {
		var securityGroup K8sSecurityGroups
		_ = faker.FakeData(&securityGroup)
		dummySecurityGroups = append(dummySecurityGroups, securityGroup)
	}

	serverResponse, _ := json.Marshal(KubernetesRead{
		SecurityGroups: dummySecurityGroups,
		Status:         "success",
	})

	URL := fmt.Sprintf("/kubernetes/%d", clusterId)
	mux.HandleFunc(URL, func(w http.ResponseWriter, req *http.Request) {
		testHttpMethod(t, req, "GET")
		testHeader(t, req, "Authorization", "Bearer token")
		fmt.Fprint(w, string(serverResponse))
	})

	ctx := context.Background()
	got, _ := client.Kubernetes().ListSecurityGroups(ctx, clusterId)
	assert.Equal(t, dummySecurityGroups, got)
}

func TestKubernetesServices_ListSecurityGroup_invalidServer(t *testing.T) {
	client, _ := NewClient("token")

	ctx := context.Background()
	securitygroup, err := client.Kubernetes().ListSecurityGroups(ctx, 1111)
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
	if securitygroup != nil {
		t.Errorf("Was not expecting any kubernetes securitygroup to be returned, instead got %v", securitygroup)
	}
}

func TestKubernetesServices_DeleteSecurityGroup_happyPath(t *testing.T) {
	token := "token"
	clusterId := 1111
	securityGroupId := 44444

	client, mux, _, teardown := setup(token)
	defer teardown()

	URL := fmt.Sprintf("/kubernetes/%d/securitygroup/%d", clusterId, securityGroupId)
	mux.HandleFunc(URL, func(w http.ResponseWriter, req *http.Request) {
		testHttpMethod(t, req, http.MethodDelete)
		testHeader(t, req, "Authorization", "Bearer "+token)
		fmt.Fprint(w, dummyDeleteResponseJson)
	})

	want := DeleteResponse{Status: "success", Message: "success"}

	ctx := context.Background()
	got, _ := client.Kubernetes().DeleteSecurityGroup(ctx, clusterId, securityGroupId)
	if !reflect.DeepEqual(*got, want) {
		t.Errorf("Response = %v, want %v", *got, want)
	}
}

func TestKubernetesServices_DeleteSecurityGroup_invalidServer(t *testing.T) {
	client, _ := NewClient("token")

	ctx := context.Background()
	delResponse, err := client.Kubernetes().DeleteSecurityGroup(ctx, 1111, 123543)
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
	if delResponse != nil {
		t.Errorf("Was not expecting any reponse to be returned, instead got %v", delResponse)
	}
}

// K8s Targetgroup test
func TestKubernetesServices_CreateTargetgroup_happyPath(t *testing.T) {
	token := "token"

	payload := CreateKubernetesTargetgroupParams{
		ClusterId:               1111,
		KubernetesTargetgroupId: 33333,
	}

	client, mux, _, teardown := setup(token)
	defer teardown()

	URL := fmt.Sprintf("/kubernetes/%d/targetgroup/%d", payload.ClusterId, payload.KubernetesTargetgroupId)
	mux.HandleFunc(URL, func(w http.ResponseWriter, req *http.Request) {
		testHttpMethod(t, req, http.MethodPost)
		testHeader(t, req, "Authorization", "Bearer "+token)
		fmt.Fprint(w, dummyCreateBasicResponseJson)
	})

	ctx := context.Background()
	got, err := client.Kubernetes().CreateTargetgroup(ctx, payload)

	var want CreateResponse
	_ = json.Unmarshal([]byte(dummyCreateBasicResponseJson), &want)

	assert.Nil(t, err)
	assert.Equal(t, want, *got)
}

func TestKubernetesServices_CreateTargetgroup_invalidServer(t *testing.T) {
	client, _ := NewClient("token")

	ctx := context.Background()
	_, err := client.Kubernetes().CreateTargetgroup(ctx, CreateKubernetesTargetgroupParams{})
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
}

func TestKubernetesServices_ReadTargetgroup_happyPath(t *testing.T) {
	client, mux, _, teardown := setup("token")
	defer teardown()

	randomInts, err := faker.RandomInt(1000, 9999)
	if err != nil || len(randomInts) == 0 {
		t.Fatalf("failed to generate sufficient random ints: %v", err)
	}
	clusterId := randomInts[0]
	targetgroupId := randomInts[1]

	var dummyTargetGroup K8sTargetGroups
	if err := faker.FakeData(&dummyTargetGroup); err != nil {
		t.Fatalf("failed to generate fake data for target group: %v", err)
	}
	dummyTargetGroup.ID = targetgroupId

	serverResponse, _ := json.Marshal(KubernetesRead{
		Info:         KubernetesClusterInfo{Cluster: KubernetesClusterMetadata{ID: clusterId}},
		TargetGroups: []K8sTargetGroups{dummyTargetGroup},
		Status:       "success",
	})

	URL := fmt.Sprintf("/kubernetes/%d", clusterId)
	mux.HandleFunc(URL, func(w http.ResponseWriter, req *http.Request) {
		testHttpMethod(t, req, "GET")
		testHeader(t, req, "Authorization", "Bearer token")
		fmt.Fprint(w, string(serverResponse))
	})

	ctx := context.Background()
	got, err := client.Kubernetes().ReadTargetgroup(ctx, clusterId, targetgroupId)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assert.Equal(t, dummyTargetGroup, *got)
}

func TestKubernetesServices_ReadTargetgroup_invalidServer(t *testing.T) {
	client, _ := NewClient("token")

	ctx := context.Background()
	targetGroup, err := client.Kubernetes().ReadTargetgroup(ctx, 1111, 122134)
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
	if targetGroup != nil {
		t.Errorf("Was not expecting any targetGroup to be returned, instead got %v", targetGroup)
	}
}

func TestKubernetesServices_ListTargetgroup_happyPath(t *testing.T) {
	client, mux, _, teardown := setup("token")
	defer teardown()

	randomInts, err := faker.RandomInt(1000, 9999)
	if err != nil || len(randomInts) == 0 {
		t.Fatalf("failed to generate random int: %v", err)
	}
	clusterId := randomInts[0]
	var dummyTargetGroups []K8sTargetGroups
	for i := 0; i < 3; i++ {
		var targetGroup K8sTargetGroups
		_ = faker.FakeData(&targetGroup)
		dummyTargetGroups = append(dummyTargetGroups, targetGroup)
	}

	serverResponse, _ := json.Marshal(KubernetesRead{
		TargetGroups: dummyTargetGroups,
		Status:       "success",
	})

	URL := fmt.Sprintf("/kubernetes/%d", clusterId)
	mux.HandleFunc(URL, func(w http.ResponseWriter, req *http.Request) {
		testHttpMethod(t, req, "GET")
		testHeader(t, req, "Authorization", "Bearer token")
		fmt.Fprint(w, string(serverResponse))
	})

	ctx := context.Background()
	got, _ := client.Kubernetes().ListTargetgroups(ctx, clusterId)
	assert.Equal(t, dummyTargetGroups, got)
}

func TestKubernetesServices_ListTargetgroup_invalidServer(t *testing.T) {
	client, _ := NewClient("token")

	ctx := context.Background()
	targetgroup, err := client.Kubernetes().ListTargetgroups(ctx, 1111)
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
	if targetgroup != nil {
		t.Errorf("Was not expecting any kubernetes targetgroup to be returned, instead got %v", targetgroup)
	}
}

func TestKubernetesServices_DeleteTargetgroup_happyPath(t *testing.T) {
	token := "token"
	clusterId := 1111
	targetgroupId := 33333

	client, mux, _, teardown := setup(token)
	defer teardown()

	URL := fmt.Sprintf("/kubernetes/%d/targetgroup/%d", clusterId, targetgroupId)
	mux.HandleFunc(URL, func(w http.ResponseWriter, req *http.Request) {
		testHttpMethod(t, req, http.MethodDelete)
		testHeader(t, req, "Authorization", "Bearer "+token)
		fmt.Fprint(w, dummyDeleteResponseJson)
	})

	want := DeleteResponse{Status: "success", Message: "success"}

	ctx := context.Background()
	got, _ := client.Kubernetes().DeleteTargetgroup(ctx, clusterId, targetgroupId)
	if !reflect.DeepEqual(*got, want) {
		t.Errorf("Response = %v, want %v", *got, want)
	}
}

func TestKubernetesServices_DeleteTargetgroup_invalidServer(t *testing.T) {
	client, _ := NewClient("token")

	ctx := context.Background()
	delResponse, err := client.Kubernetes().DeleteTargetgroup(ctx, 1111, 123543)
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
	if delResponse != nil {
		t.Errorf("Was not expecting any reponse to be returned, instead got %v", delResponse)
	}
}

func TestKubernetesService_PowerOn_happyPath(t *testing.T) {
	token := "token"
	clusterId := 0000

	client, mux, _, teardown := setup(token)
	defer teardown()

	URL := fmt.Sprintf("/kubernetes/%d/start", clusterId)
	mux.HandleFunc(URL, func(w http.ResponseWriter, req *http.Request) {
		testHttpMethod(t, req, http.MethodPost)
		testHeader(t, req, "Authorization", "Bearer "+token)
		fmt.Fprint(w, dummyCreateBasicResponseJson)
	})

	ctx := context.Background()
	got, err := client.Kubernetes().PowerOn(ctx, clusterId)

	var want BasicResponse
	_ = json.Unmarshal([]byte(dummyCreateBasicResponseJson), &want)

	assert.Nil(t, err)
	assert.Equal(t, want, *got)
}

func TestKubernetesService_PowerOn_invalidServer(t *testing.T) {
	client, _ := NewClient("token")

	ctx := context.Background()
	_, err := client.Kubernetes().PowerOn(ctx, 11111)
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
}

func TestKubernetesService_PowerOff_happyPath(t *testing.T) {
	token := "token"
	clusterId := 1111

	client, mux, _, teardown := setup(token)
	defer teardown()

	URL := fmt.Sprintf("/kubernetes/%d/stop", clusterId)
	mux.HandleFunc(URL, func(w http.ResponseWriter, req *http.Request) {
		testHttpMethod(t, req, http.MethodPost)
		testHeader(t, req, "Authorization", "Bearer "+token)
		fmt.Fprint(w, dummyCreateBasicResponseJson)
	})

	ctx := context.Background()
	got, err := client.Kubernetes().PowerOff(ctx, clusterId)

	var want BasicResponse
	_ = json.Unmarshal([]byte(dummyCreateBasicResponseJson), &want)

	assert.Nil(t, err)
	assert.Equal(t, want, *got)
}

func TestKubernetesService_PowerOff_invalidServer(t *testing.T) {
	client, _ := NewClient("token")

	ctx := context.Background()
	_, err := client.Kubernetes().PowerOff(ctx, 11111)
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
}
