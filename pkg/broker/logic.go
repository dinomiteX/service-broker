package broker

import (
	//"github.com/kubernetes/client-go/kubernetes/typed/core/v1"
	"net/http"
	"sync"

	"github.com/golang/glog"
	"github.com/pmorie/osb-broker-lib/pkg/broker"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"reflect"
	//rest "k8s.io/client-go/rest"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

// NewBusinessLogic is a hook that is called with the Options the program is run
// with. NewBusinessLogic is the place where you will initialize your
// BusinessLogic the parameters passed in.
func NewBusinessLogic(o Options) (*BusinessLogic, error) {
	// For example, if your BusinessLogic requires a parameter from the command
	// line, you would unpack it from the Options and set it on the
	// BusinessLogic here.
	return &BusinessLogic{
		async:     o.Async,
		instances: make(map[string]*exampleInstance, 10),
	}, nil
}

// BusinessLogic provides an implementation of the broker.BusinessLogic
// interface.
type BusinessLogic struct {
	// Indicates if the broker should handle the requests asynchronously.
	async bool
	// Synchronize go routines.
	sync.RWMutex
	// Add fields here! These fields are provided purely as an example
	instances map[string]*exampleInstance
}

var _ broker.Interface = &BusinessLogic{}

func truePtr() *bool {
	b := true
	return &b
}

func (b *BusinessLogic) GetCatalog(c *broker.RequestContext) (*broker.CatalogResponse, error) {
	// Your catalog business logic goes here
	glog.Infof("Starting Business Logic...")

	service1 := new(osb.Service)
	service1.Name = "Service-1"
	service1.ID = "service-1-id"
	service1.Description = "First Service Made by Dino"
	service1.Bindable = true
	service1.PlanUpdatable = truePtr()
	service1.Plans = []osb.Plan{
		{
			Name:        "plan-1",
			ID:          "service-1-plan-1-id",
			Description: "Description Plan 1",
			Free:        truePtr(),
			Schemas: &osb.Schemas{
				ServiceInstance: &osb.ServiceInstanceSchema{
					Create: &osb.InputParametersSchema{
						Parameters: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"color": map[string]interface{}{
									"type":    "string",
									"default": "Clear",
									"enum": []string{
										"Clear",
										"Beige",
										"Grey",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Name:        "plan-2",
			ID:          "service-1-plan-2-id",
			Description: "Description plan 2",
			Free:        truePtr(),
			Schemas: &osb.Schemas{
				ServiceInstance: &osb.ServiceInstanceSchema{
					Create: &osb.InputParametersSchema{
						Parameters: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"color": map[string]interface{}{
									"type":    "string",
									"default": "Clear",
									"enum": []string{
										"Clear",
										"Beige",
										"Grey",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	glog.Infof("Service 1 created...")
	service2 := new(osb.Service)
	service2.Name = "Service-2"
	service2.ID = "service-2-id"
	service2.Description = "Second Service Made by Dino"
	service2.Bindable = true
	service2.PlanUpdatable = truePtr()
	service2.Plans = []osb.Plan{
		{
			Name:        "plan-1",
			ID:          "service-2-plan-1-id",
			Description: "Description Plan 1",
			Free:        truePtr(),
			Schemas: &osb.Schemas{
				ServiceInstance: &osb.ServiceInstanceSchema{
					Create: &osb.InputParametersSchema{
						Parameters: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"color": map[string]interface{}{
									"type":    "string",
									"default": "Clear",
									"enum": []string{
										"Clear",
										"Beige",
										"Grey",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Name:        "plan-2",
			ID:          "service-2-plan-2-id",
			Description: "Description plan 2",
			Free:        truePtr(),
			Schemas: &osb.Schemas{
				ServiceInstance: &osb.ServiceInstanceSchema{
					Create: &osb.InputParametersSchema{
						Parameters: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"color": map[string]interface{}{
									"type":    "string",
									"default": "Clear",
									"enum": []string{
										"Clear",
										"Beige",
										"Grey",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	glog.Infof("Service 2 created...")
	services := []osb.Service{*service1, *service2}
	glog.Infof("Service List created...")
	response := &broker.CatalogResponse{}
	osbResponse := &osb.CatalogResponse{
		Services: services,
	}

	//test := &osb.CatalogResponse{}

	glog.Infof("catalog response: %#+v", osbResponse)

	response.CatalogResponse = *osbResponse

	return response, nil
}

func (b *BusinessLogic) Provision(request *osb.ProvisionRequest, c *broker.RequestContext) (*broker.ProvisionResponse, error) {
	// Your provision business logic goes here
	// example implementation:
	glog.Infof("Starting Provisioning...")
	glog.Infof("%d", request.Parameters)
	b.Lock()
	defer b.Unlock()

	response := broker.ProvisionResponse{}

	exampleInstance := &exampleInstance{
		ID:        request.InstanceID,
		ServiceID: request.ServiceID,
		PlanID:    request.PlanID,
		Params:    request.Parameters,
		Context:   request.Context,
	}

	// Check to see if this is the same instance
	if i := b.instances[request.InstanceID]; i != nil {
		if i.Match(exampleInstance) {
			response.Exists = true
			return &response, nil
		} else {
			// Instance ID in use, this is a conflict.
			description := "InstanceID in use"
			return nil, osb.HTTPStatusCodeError{
				StatusCode:  http.StatusConflict,
				Description: &description,
			}
		}
	}
	b.instances[request.InstanceID] = exampleInstance

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	deploymentsClient := clientset.AppsV1().Deployments("test-ns")

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []v1.ContainerPort{
								{
									Name:          "http",
									Protocol:      v1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := deploymentsClient.Create(deployment)
	if err != nil {
		panic(err)
	}
	glog.Infof("Deployment created!")
	glog.Infof("Result: %s", result.String())
	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *BusinessLogic) Deprovision(request *osb.DeprovisionRequest, c *broker.RequestContext) (*broker.DeprovisionResponse, error) {
	// Your deprovision business logic goes here

	// example implementation:
	b.Lock()
	defer b.Unlock()

	response := broker.DeprovisionResponse{}

	delete(b.instances, request.InstanceID)

	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *BusinessLogic) LastOperation(request *osb.LastOperationRequest, c *broker.RequestContext) (*broker.LastOperationResponse, error) {
	// Your last-operation business logic goes here

	return nil, nil
}

func (b *BusinessLogic) Bind(request *osb.BindRequest, c *broker.RequestContext) (*broker.BindResponse, error) {
	// Your bind business logic goes here

	// example implementation:
	b.Lock()
	defer b.Unlock()

	instance, ok := b.instances[request.InstanceID]
	if !ok {
		return nil, osb.HTTPStatusCodeError{
			StatusCode: http.StatusNotFound,
		}
	}

	response := broker.BindResponse{
		BindResponse: osb.BindResponse{
			Credentials: instance.Params,
		},
	}
	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *BusinessLogic) Unbind(request *osb.UnbindRequest, c *broker.RequestContext) (*broker.UnbindResponse, error) {
	// Your unbind business logic goes here
	return &broker.UnbindResponse{}, nil
}

func (b *BusinessLogic) Update(request *osb.UpdateInstanceRequest, c *broker.RequestContext) (*broker.UpdateInstanceResponse, error) {
	// Your logic for updating a service goes here.
	response := broker.UpdateInstanceResponse{}
	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *BusinessLogic) ValidateBrokerAPIVersion(version string) error {
	return nil
}

// example types

// exampleInstance is intended as an example of a type that holds information about a service instance
type exampleInstance struct {
	ID        string
	ServiceID string
	PlanID    string
	Params    map[string]interface{}
	Context   map[string]interface{}
}

func (i *exampleInstance) Match(other *exampleInstance) bool {
	return reflect.DeepEqual(i, other)
}

func int32Ptr(i int32) *int32 { return &i }
