package nullplatform

import (
	"log"
	"reflect"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceScope() *schema.Resource {
	return &schema.Resource{
		Create: ScopeCreate,
		Read:   ScopeRead,
		Update: ScopeUpdate,
		Delete: ScopeDelete,

		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"nrn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"scope_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"scope_type": {
				Type:     schema.TypeString,
				Default:  "serverless",
				Optional: true,
			},
			"null_application_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"s3_assets_bucket": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"scope_workflow_role": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"log_group_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"lambda_function_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"lambda_current_function_version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"lambda_function_role": {
				Type:     schema.TypeString,
				Required: true,
			},
			"lambda_function_main_alias": {
				Type:     schema.TypeString,
				Required: true,
			},
			"log_reader_role": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"lambda_function_warm_alias": {
				Type:     schema.TypeString,
				Default:  "",
				Optional: true,
			},
			"capabilities_serverless_handler_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"capabilities_serverless_timeout": {
				Type:     schema.TypeInt,
				Default:  10,
				Optional: true,
			},
			"capabilities_serverless_runtime_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"capabilities_serverless_memory": {
				Type:     schema.TypeInt,
				Default:  128,
				Optional: true,
			},
			"dimensions": {
				Type:     schema.TypeMap,
				ForceNew: true,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"runtime_configurations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
		},
	}
}

func ScopeCreate(d *schema.ResourceData, m any) error {
	nullOps := m.(NullOps)

	log.Print("--- CREATE Serverless scope ---")
	log.Printf(">>> schema.ResourceData: %+v", d)
	log.Printf(">>> meta data: %+v", m)

	applicationId := d.Get("null_application_id").(int)
	scopeName := d.Get("scope_name").(string)
	scopeType := d.Get("scope_type").(string)
	serverless_runtime := d.Get("capabilities_serverless_runtime_id").(string)
	serverless_handler := d.Get("capabilities_serverless_handler_name").(string)
	serverless_timeout := d.Get("capabilities_serverless_timeout").(int)
	serverless_memory := d.Get("capabilities_serverless_memory").(int)

	dimensionsMap := d.Get("dimensions").(map[string]interface{})
	// Convert the dimensions to a map[string]string
	dimensions := make(map[string]string)
	for key, value := range dimensionsMap {
		dimensions[key] = value.(string)
	}

	newScope := &Scope{
		Name:            scopeName,
		ApplicationId:   applicationId,
		Type:            scopeType,
		ExternalCreated: true,
		RequestedSpec: &RequestSpec{
			MemoryInGb:   0.5,
			CpuProfile:   "standard",
			LocalStorage: 8,
		},
		Capabilities: &Capability{
			Visibility: map[string]string{
				"reachability": "account",
			},
			ServerlessRuntime: map[string]string{
				"provider": "aws_lambda",
				"id":       serverless_runtime,
			},
			ServerlessHandler: map[string]string{
				"name": serverless_handler,
			},
			ServerlessTimeout: map[string]int{
				"timeout_in_seconds": serverless_timeout,
			},
			ServerlessEphemeralStorage: map[string]int{
				"memory_in_mb": 512,
			},
			ServerlessMemory: map[string]int{
				"memory_in_mb": serverless_memory,
			},
		},
		Dimensions: dimensions,
	}

	s, err := nullOps.CreateScope(newScope)

	if err != nil {
		return err
	}

	log.Print("--- BEFORE patch NRN ---")

	nrnErr := patchNrnForScope(s.Nrn, d, m)

	if nrnErr != nil {
		log.Print("--- AFTER patch NRN failed ******---")
		return nrnErr
	}

	log.Print("--- AFTER patch NRN success ---")

	d.SetId(strconv.Itoa(s.Id))

	return ScopeRead(d, m)
}

func patchNrnForScope(scopeNrn string, d *schema.ResourceData, m any) error {
	nullOps := m.(NullOps)

	s3AssetsBucket := d.Get("s3_assets_bucket").(string)
	scopeWorkflowRole := d.Get("scope_workflow_role").(string)
	logGroupName := d.Get("log_group_name").(string)
	lambdaFunctinoName := d.Get("lambda_function_name").(string)
	lambdaCurrentFunctionVersion := d.Get("lambda_current_function_version").(string)
	lambdaFunctionRole := d.Get("lambda_function_role").(string)
	lambdaFunctionMainAlias := d.Get("lambda_function_main_alias").(string)
	logReaderRole := d.Get("log_reader_role").(string)
	lambdaFunctionWarmAlias := d.Get("lambda_function_warm_alias").(string)

	nrnReq := &PatchNRN{
		AWSS3AssestBucket:               s3AssetsBucket,
		AWSScopeWorkflowRole:            scopeWorkflowRole,
		AWSLogGroupName:                 logGroupName,
		AWSLambdaFunctionName:           lambdaFunctinoName,
		AWSLambdaCurrentFunctionVersion: lambdaCurrentFunctionVersion,
		AWSLambdaFunctionRole:           lambdaFunctionRole,
		AWSLambdaFunctionMainAlias:      lambdaFunctionMainAlias,
		AWSLogReaderLog:                 logReaderRole,
		AWSLambdaFunctionWarmAlias:      lambdaFunctionWarmAlias,
	}

	return nullOps.PatchNRN(scopeNrn, nrnReq)
}

func ScopeRead(d *schema.ResourceData, m any) error {
	nullOps := m.(NullOps)
	scopeID := d.Id()

	log.Print("--- Terraform 'read resource Scope' operation begin ---")
	s, err := nullOps.GetScope(scopeID)

	if err != nil {
		d.SetId("")
		return err
	}

	log.Printf(">>> schema.ResourceData: %+v", d)
	log.Printf(">>> meta data: %+v", m)

	if err := d.Set("scope_name", s.Name); err != nil {
		return err
	}

	if err := d.Set("null_application_id", s.ApplicationId); err != nil {
		return err
	}

	if err := d.Set("nrn", s.Nrn); err != nil {
		return err
	}

	//if err := d.Set("runtime_configurations", s.RuntimeConfigurations); err != nil {
	//	return err
	//}

	//if err := d.Set("dimensions", s.Dimensions); err != nil {
	//	return err
	//}

	log.Print("--- Terraform 'read resource Scope' operation ends ---")

	d.Set("last_updated", time.Now().Format(time.RFC850))

	return nil
}

func getNrnForScope(scopeNrn string, nullOps NullOps) (*NRN, error) {
	nrn, err := nullOps.GetNRN(scopeNrn)

	if err != nil {
		return nil, err
	}

	return nrn, nil
}

func ScopeUpdate(d *schema.ResourceData, m any) error {
	nullOps := m.(NullOps)

	log.Print("--- Terraform 'update resource Scope' operation begin  ---")
	log.Printf(">>> schema.ResourceData: %+v", d)
	log.Printf(">>> meta data: %+v", m)

	scopeID := d.Id()

	log.Println("scopeID:", scopeID)

	ps := &Scope{}

	if d.HasChange("scope_name") {
		ps.Name = d.Get("scope_name").(string)
	}

	if d.HasChange("dimensions") {
		dimensionsMap := d.Get("dimensions").(map[string]interface{})

		// Convert the dimensions to a map[string]string
		dimensions := make(map[string]string)
		for key, value := range dimensionsMap {
			dimensions[key] = value.(string)
		}

		ps.Dimensions = dimensions
	}

	caps := &Capability{}

	if d.HasChange("capabilities_serverless_runtime_id") {
		caps.ServerlessRuntime = map[string]string{
			"provider": "aws_lambda",
			"id":       d.Get("capabilities_serverless_runtime_id").(string),
		}
	}

	if d.HasChange("capabilities_serverless_handler_name") {
		caps.ServerlessHandler = map[string]string{
			"name": d.Get("capabilities_serverless_handler_name").(string),
		}
	}

	if d.HasChange("capabilities_serverless_timeout") {
		caps.ServerlessTimeout = map[string]int{
			"timeout_in_seconds": d.Get("capabilities_serverless_timeout").(int),
		}
	}

	if d.HasChange("capabilities_serverless_memory") {
		caps.ServerlessMemory = map[string]int{
			"memory_in_mb": d.Get("capabilities_serverless_memory").(int),
		}
	}

	if !reflect.DeepEqual(caps, Capability{}) {
		ps.Capabilities = caps
	}

	log.Print("--- Scope updated ---")
	log.Printf(">>> schema.ResourceData: %+v", d)
	log.Printf(">>> meta data: %+v", m)

	if !reflect.DeepEqual(*ps, Scope{}) {
		err := nullOps.PatchScope(scopeID, ps)
		if err != nil {
			return err
		}
	}

	d.Set("last_updated", time.Now().Format(time.RFC850))

	log.Print("--- Terraform 'update resource Scope' operation ends ---")

	return nil
}

func ScopeDelete(d *schema.ResourceData, m any) error {
	nullOps := m.(NullOps)

	scopeID := d.Id()

	log.Print("--- Terraform 'delete resource Scope' operation begin ---")
	log.Printf(">>> schema.ResourceData: %+v", d)
	log.Printf(">>> meta data: %+v", m)

	pScope := &Scope{
		Status: "deleting",
	}

	log.Print("--- Scope on: 'deleting' ---")
	err := nullOps.PatchScope(scopeID, pScope)
	if err != nil {
		return err
	}

	pScope.Status = "deleted"

	log.Print("--- Scope on: 'deleted' ---")

	err = nullOps.PatchScope(scopeID, pScope)
	if err != nil {
		return err
	}

	log.Printf(">>> schema.ResourceData: %+v", d)
	log.Printf(">>> meta data: %+v", m)

	log.Println(">>> scopeID:", scopeID)

	log.Print("--- Terraform 'delete resource Scope' operation ends ---")

	d.SetId("")

	return nil
}
