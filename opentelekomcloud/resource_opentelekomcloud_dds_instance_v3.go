package opentelekomcloud

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/huaweicloud/golangsdk"
	"github.com/huaweicloud/golangsdk/openstack/dds/v3/instances"
	"log"
	"time"
)

func resourceDdsInstanceV3() *schema.Resource {
	return &schema.Resource{
		Create: resourceDdsInstanceV3Create,
		Read:   resourceDdsInstanceV3Read,
		Update: resourceDdsInstanceV3Update,
		Delete: resourceDdsInstanceV3Delete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"datastore": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"DDS-Community",
							}, true),
						},
						"version": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"3.4", "3.2",
							}, true),
						},
						"storage_engine": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"wiredTiger",
							}, true),
						},
					},
				},
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Type:      schema.TypeString,
				Sensitive: true,
				Required:  true,
			},
			"disk_encryption_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"mode": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Sharding", "ReplicaSet",
				}, true),
			},
			"flavor": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"mongos", "shard", "config", "replica",
							}, true),
						},
						"num": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"storage": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"ULTRAHIGH",
							}, true),
						},
						"size": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
						"spec_code": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"backup_strategy": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_time": {
							Type:     schema.TypeString,
							Required: true,
						},
						"keep_days": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDdsDataStore(d *schema.ResourceData) instances.DataStore {
	var dataStore instances.DataStore
	dataStoreRaw := d.Get("datastore").([]interface{})
	log.Printf("[DEBUG] dataStoreRaw: %+v", dataStoreRaw)
	if len(dataStoreRaw) == 1 {
		dataStore.Type = dataStoreRaw[0].(map[string]interface{})["type"].(string)
		dataStore.Version = dataStoreRaw[0].(map[string]interface{})["version"].(string)
		dataStore.StorageEngine = dataStoreRaw[0].(map[string]interface{})["storage_engine"].(string)
	}
	log.Printf("[DEBUG] datastore: %+v", dataStore)
	return dataStore
}

func resourceDdsFlavors(d *schema.ResourceData) []instances.Flavor {
	var flavors []instances.Flavor
	flavorRaw := d.Get("flavor").([]interface{})
	log.Printf("[DEBUG] flavorRaw: %+v", flavorRaw)
	for i := range flavorRaw {
		flavor := flavorRaw[i].(map[string]interface{})
		flavorReq := instances.Flavor{
			Type:     flavor["type"].(string),
			Num:      flavor["num"].(int),
			Storage:  flavor["storage"].(string),
			Size:     flavor["size"].(int),
			SpecCode: flavor["spec_code"].(string),
		}
		flavors = append(flavors, flavorReq)
	}
	log.Printf("[DEBUG] flavors: %+v", flavors)
	return flavors
}

func resourceDdsBackupStrategy(d *schema.ResourceData) instances.BackupStrategy {
	var backupStrategy instances.BackupStrategy
	backupStrategyRaw := d.Get("backup_strategy").([]interface{})
	log.Printf("[DEBUG] backupStrategyRaw: %+v", backupStrategyRaw)
	if len(backupStrategyRaw) == 1 {
		backupStrategy.StartTime = backupStrategyRaw[0].(map[string]interface{})["start_time"].(string)
		backupStrategy.KeepDays = backupStrategyRaw[0].(map[string]interface{})["keep_days"].(int)
	} else {
		backupStrategy.StartTime = "00:00-01:00"
		backupStrategy.KeepDays = 7
	}
	log.Printf("[DEBUG] backupStrategy: %+v", backupStrategy)
	return backupStrategy
}

func DdsInstanceStateRefreshFunc(client *golangsdk.ServiceClient, instanceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		opts := instances.ListInstanceOpts{
			Id: instanceID,
		}
		allPages, err := instances.List(client, &opts).AllPages()
		if err != nil {
			return nil, "", err
		}
		instancesList, err := instances.ExtractInstances(allPages)
		if err != nil {
			return nil, "", err
		}

		if instancesList.TotalCount == 0 {
			var instance instances.InstanceResponse
			return instance, "deleted", nil
		}
		ddsInstances := instancesList.Instances

		status := ddsInstances[0].Status
		// wait for updating
		if status == "normal" && len(ddsInstances[0].Actions) > 0 {
			status = "updating"
		}
		return ddsInstances[0], status, nil
	}
}

func resourceDdsInstanceV3Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	client, err := config.ddsV3Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenTelekomCloud DDS client: %s ", err)
	}

	createOpts := instances.CreateOpts{
		Name:             d.Get("name").(string),
		DataStore:        resourceDdsDataStore(d),
		Region:           GetRegion(d, config),
		AvailabilityZone: d.Get("availability_zone").(string),
		VpcId:            d.Get("vpc_id").(string),
		SubnetId:         d.Get("subnet_id").(string),
		SecurityGroupId:  d.Get("security_group_id").(string),
		Password:         d.Get("password").(string),
		DiskEncryptionId: d.Get("disk_encryption_id").(string),
		Mode:             d.Get("mode").(string),
		Flavor:           resourceDdsFlavors(d),
		BackupStrategy:   resourceDdsBackupStrategy(d),
	}
	log.Printf("[DEBUG] Create Options: %#v", createOpts)

	instance, err := instances.Create(client, createOpts).Extract()
	if err != nil {
		return fmt.Errorf("Error getting instance from result: %s ", err)
	}
	log.Printf("[DEBUG] Create : instance %s: %#v", instance.Id, instance)

	d.SetId(instance.Id)
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"creating"},
		Target:     []string{"normal"},
		Refresh:    DdsInstanceStateRefreshFunc(client, instance.Id),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      120 * time.Second,
		MinTimeout: 20 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for instance (%s) to become ready: %s ", instance.Id, err)
	}

	return resourceDdsInstanceV3Read(d, meta)
}

func resourceDdsInstanceV3Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	client, err := config.ddsV3Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenTelekomCloud DDS client: %s", err)
	}

	instanceID := d.Id()
	opts := instances.ListInstanceOpts{
		Id: instanceID,
	}
	allPages, err := instances.List(client, &opts).AllPages()
	if err != nil {
		return fmt.Errorf("error fetching DDS instance: %s", err)
	}
	ddsInstancesPage, err := instances.ExtractInstances(allPages)
	if err != nil {
		return fmt.Errorf("error extracting DDS instance: %s", err)
	}
	if ddsInstancesPage.TotalCount == 0 {
		return fmt.Errorf("error fetching DDS instance: deleted")
	}
	ddsInstances := ddsInstancesPage.Instances
	instance := ddsInstances[0]

	log.Printf("[DEBUG] Retrieved instance %s: %#v", instanceID, instance)

	me := &multierror.Error{}
	me = multierror.Append(me,
		d.Set("region", instance.Region),
		d.Set("name", instance.Name),
		d.Set("vpc_id", instance.VpcId),
		d.Set("subnet_id", instance.SubnetId),
		d.Set("security_group_id", instance.SecurityGroupId),
		d.Set("disk_encryption_id", instance.DiskEncryptionId),
		d.Set("mode", instance.Mode),
		d.Set("status", instance.Status),
	)

	dataStoreList := make([]map[string]interface{}, 0, 1)
	dataStore := map[string]interface{}{
		"type":           instance.DataStore.Type,
		"version":        instance.DataStore.Version,
		"storage_engine": instance.Engine,
	}
	dataStoreList = append(dataStoreList, dataStore)
	if err = d.Set("datastore", dataStoreList); err != nil {
		return fmt.Errorf("[DEBUG] Error saving datastore to state for OpenTelekomCloud DDS Instance(%s): %s", d.Id(), err)
	}

	backupStrategyList := make([]map[string]interface{}, 0, 1)
	backupStrategy := map[string]interface{}{
		"start_time": instance.BackupStrategy.StartTime,
		"keep_days":  instance.BackupStrategy.KeepDays,
	}
	backupStrategyList = append(backupStrategyList, backupStrategy)
	if err = d.Set("backup_strategy", backupStrategyList); err != nil {
		return fmt.Errorf("[DEBUG] Error saving backup_strategy to state for OpenTelekomCloud DDS Instance(%s): %s", d.Id(), err)
	}

	return nil
}

func resourceDdsInstanceV3Update(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	client, err := config.ddsV3Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenTelekomCloud DDS client: %s ", err)
	}

	var opts []instances.UpdateOpt
	if d.HasChange("name") {
		opt := instances.UpdateOpt{
			Param:  "new_instance_name",
			Value:  d.Get("name").(string),
			Action: "modify-name",
			Method: "put",
		}
		opts = append(opts, opt)
	}

	if d.HasChange("password") {
		opt := instances.UpdateOpt{
			Param:  "user_pwd",
			Value:  d.Get("password").(string),
			Action: "reset-password",
			Method: "put",
		}
		opts = append(opts, opt)
	}

	if d.HasChange("security_group_id") {
		opt := instances.UpdateOpt{
			Param:  "security_group_id",
			Value:  d.Get("security_group_id").(string),
			Action: "modify-security-group",
			Method: "post",
		}
		opts = append(opts, opt)
	}

	r := instances.Update(client, d.Id(), opts)
	if r.Err != nil {
		return fmt.Errorf("Error updating instance from result: %s ", r.Err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"updating"},
		Target:     []string{"normal"},
		Refresh:    DdsInstanceStateRefreshFunc(client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      15 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for instance (%s) to become ready: %s ", d.Id(), err)
	}

	return resourceDdsInstanceV3Read(d, meta)
}

func resourceDdsInstanceV3Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	client, err := config.ddsV3Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenTelekomCloud DDS client: %s ", err)
	}

	instanceId := d.Id()
	result := instances.Delete(client, instanceId)
	if result.Err != nil {
		return err
	}
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"normal", "abnormal", "frozen", "createfail", "enlargefail", "data_disk_full"},
		Target:     []string{"deleted"},
		Refresh:    DdsInstanceStateRefreshFunc(client, instanceId),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      15 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for instance (%s) to be deleted: %s ", instanceId, err)
	}
	log.Printf("[DEBUG] Successfully deleted instance %s", instanceId)
	return nil
}
