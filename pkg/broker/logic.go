package broker

import (
	//"github.com/kubernetes/client-go/kubernetes/typed/core/v1"
	"fmt"
	"net/http"
	"sync"

	"github.com/golang/glog"
	"github.com/pmorie/osb-broker-lib/pkg/broker"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
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
		instances: make(map[string]*instance, 10),
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
	instances map[string]*instance
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

	b.Lock()
	defer b.Unlock()

	response := broker.ProvisionResponse{}

	instance := &instance{
		ID:        request.InstanceID,
		ServiceID: request.ServiceID,
		PlanID:    request.PlanID,
		Params:    request.Parameters,
		Context:   request.Context,
	}

	glog.Infof("Checking for Params...")
	if instance.Params["namespace"] == "" {
		return nil, fmt.Errorf("Namespace Param not set or empty")
	}
	if instance.Params["podname"] == "" {
		return nil, fmt.Errorf("Podname Param not set or empty")
	}

	// Check to see if this is the same instance
	if i := b.instances[request.InstanceID]; i != nil {
		if i.Match(instance) {
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
	b.instances[request.InstanceID] = instance

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode([]byte(podjson), nil, nil)
	if err != nil {
		panic(err)
	}
	pod := obj.(*v1.Pod)
	pod.Namespace = instance.Params["namespace"].(string)
	pod.Name = instance.Params["podname"].(string)
	glog.Infof("Creating Pod %s in Namespace %s...", pod.Name, pod.Namespace)
	pod, err = clientset.CoreV1().Pods(pod.Namespace).Create(pod)
	if err != nil {
		panic(err)
	}
	glog.Infof("Pod created!")
	glog.Infof("Pod Details: \nResourcename: %v\nCreationtime: %v", pod.String(), pod.CreationTimestamp)

	glog.Infof("%v", response)
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

// instance is intended as an example of a type that holds information about a service instance
type instance struct {
	ID        string
	ServiceID string
	PlanID    string
	Params    map[string]interface{}
	Context   map[string]interface{}
}

func (i *instance) Match(other *instance) bool {
	return reflect.DeepEqual(i, other)
}

func int32Ptr(i int32) *int32 { return &i }

var podjson = `apiVersion: v1
kind: Pod
spec:
  containers:
  - name: nginx
    image: nginx
    ports:
    - containerPort: 80
    volumeMounts:
    - name: workdir
      mountPath: /usr/share/nginx/html
  # These containers are run during pod initialization
  initContainers:
  - name: install
    image: busybox
    command:
    - wget
    - "-O"
    - "/work-dir/index.html"
    - http://kubernetes.io
    volumeMounts:
    - name: workdir
      mountPath: "/work-dir"
  dnsPolicy: Default
  volumes:
  - name: workdir
    emptyDir: {}`
