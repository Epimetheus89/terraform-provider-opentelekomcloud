package vbs

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/vbs/v2/policies"
	vbsTags "github.com/opentelekomcloud/gophertelekomcloud/openstack/vbs/v2/tags"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceVBSBackupPolicyV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVBSBackupPolicyV2Create,
		ReadContext:   resourceVBSBackupPolicyV2Read,
		UpdateContext: resourceVBSBackupPolicyV2Update,
		DeleteContext: resourceVBSBackupPolicyV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		DeprecationMessage: "Please use `opentelekomcloud_cbr_policy_v3` resource instead.",

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: common.ValidateVBSPolicyName,
			},

			"resources": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"start_time": {
				Type:     schema.TypeString,
				Required: true,
			},
			"frequency": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"week_frequency"},
				ValidateFunc:  common.ValidateVBSPolicyFrequency,
			},
			"week_frequency": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 7,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"rentention_num": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"rentention_day"},
				ValidateFunc:  common.ValidateVBSPolicyRetentionNum,
			},
			"rentention_day": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: common.ValidateVBSPolicyRetentionNum,
			},
			"retain_first_backup": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: common.ValidateVBSPolicyRetainBackup,
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "ON",
				ValidateFunc: common.ValidateVBSPolicyStatus,
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     false,
							ValidateFunc: common.ValidateVBSTagKey,
						},
						"value": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     false,
							ValidateFunc: common.ValidateVBSTagValue,
						},
					},
				},
			},
			"policy_resource_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceVBSBackupPolicyV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	vbsClient, err := config.VbsV2Client(config.GetRegion(d))

	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud VBS Client: %s", err)
	}

	_, isExist1 := d.GetOk("frequency")
	_, isExist2 := d.GetOk("week_frequency")
	if !isExist1 && !isExist2 {
		return fmterr.Errorf("either frequency or week_frequency must be specified")
	}

	_, isExist1 = d.GetOk("rentention_num")
	_, isExist2 = d.GetOk("rentention_day")
	if !isExist1 && !isExist2 {
		return fmterr.Errorf("either rentention_num or rentention_day must be specified")
	}

	weeks, err := buildWeekFrequencyResource(d)
	if err != nil {
		return diag.FromErr(err)
	}

	createOpts := policies.CreateOpts{
		Name: d.Get("name").(string),
		ScheduledPolicy: policies.ScheduledPolicy{
			StartTime:         d.Get("start_time").(string),
			Frequency:         d.Get("frequency").(int),
			WeekFrequency:     weeks,
			RententionNum:     d.Get("rentention_num").(int),
			RententionDay:     d.Get("rentention_day").(int),
			RemainFirstBackup: d.Get("retain_first_backup").(string),
			Status:            d.Get("status").(string),
		},
		Tags: resourceVBSTagsV2(d),
	}

	create, err := policies.Create(vbsClient, createOpts).Extract()

	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud Backup Policy: %s", err)
	}
	d.SetId(create.ID)

	// associate volumes to backup policy
	resources := buildAssociateResource(d.Get("resources").([]interface{}))
	if len(resources) > 0 {
		opts := policies.AssociateOpts{
			PolicyID:  d.Id(),
			Resources: resources,
		}

		_, err := policies.Associate(vbsClient, opts).ExtractResource()
		if err != nil {
			return fmterr.Errorf("error associate volumes to VBS backup policy %s: %s",
				d.Id(), err)
		}
	}

	return resourceVBSBackupPolicyV2Read(ctx, d, meta)
}

func resourceVBSBackupPolicyV2Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	vbsClient, err := config.VbsV2Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud VBS Client: %s", err)
	}

	PolicyOpts := policies.ListOpts{ID: d.Id()}
	policyList, err := policies.List(vbsClient, PolicyOpts)
	if err != nil {
		if _, ok := err.(golangsdk.ErrDefault404); ok {
			d.SetId("")
			return nil
		}

		return fmterr.Errorf("error retrieving OpenTelekomCloud Backup Policy: %s", err)
	}

	if len(policyList) == 0 {
		d.SetId("")
		return nil
	}

	policy := policyList[0]

	mErr := multierror.Append(
		d.Set("name", policy.Name),
		d.Set("start_time", policy.ScheduledPolicy.StartTime),
		d.Set("frequency", policy.ScheduledPolicy.Frequency),
		d.Set("week_frequency", policy.ScheduledPolicy.WeekFrequency),
		d.Set("rentention_num", policy.ScheduledPolicy.RententionNum),
		d.Set("rentention_day", policy.ScheduledPolicy.RententionDay),
		d.Set("retain_first_backup", policy.ScheduledPolicy.RemainFirstBackup),
		d.Set("status", policy.ScheduledPolicy.Status),
		d.Set("policy_resource_count", policy.ResourceCount),
	)
	if mErr.ErrorOrNil() != nil {
		return fmterr.Errorf("error setting policy fields: %s", mErr)
	}

	tags, err := vbsTags.Get(vbsClient, d.Id()).Extract()

	if err != nil {
		if _, ok := err.(golangsdk.ErrDefault404); ok {
			return nil
		}
		return fmterr.Errorf("error retrieving OpenTelekomCloud Backup Policy Tags: %s", err)
	}
	var tagList []map[string]interface{}
	for _, v := range tags.Tags {
		tag := make(map[string]interface{})
		tag["key"] = v.Key
		tag["value"] = v.Value

		tagList = append(tagList, tag)
	}
	if err := d.Set("tags", tagList); err != nil {
		return fmterr.Errorf("[DEBUG] Error saving tags to state for OpenTelekomCloud backup policy (%s): %s", d.Id(), err)
	}
	return nil
}

func resourceVBSBackupPolicyV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	vbsClient, err := config.VbsV2Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error updating OpenTelekomCloud VBS client: %s", err)
	}

	_, isExist1 := d.GetOk("frequency")
	_, isExist2 := d.GetOk("week_frequency")
	if !isExist1 && !isExist2 {
		return fmterr.Errorf("either frequency or week_frequency must be specified")
	}

	_, isExist1 = d.GetOk("rentention_num")
	_, isExist2 = d.GetOk("rentention_day")
	if !isExist1 && !isExist2 {
		return fmterr.Errorf("either rentention_num or rentention_day must be specified")
	}

	frequency := d.Get("frequency").(int)
	weeks, err := buildWeekFrequencyResource(d)
	if err != nil {
		return diag.FromErr(err)
	}

	var updateOpts policies.UpdateOpts
	if frequency != 0 {
		updateOpts.ScheduledPolicy.Frequency = frequency
	} else {
		updateOpts.ScheduledPolicy.WeekFrequency = weeks
	}

	if d.HasChange("name") || d.HasChange("start_time") || d.HasChange("retain_first_backup") ||
		d.HasChange("rentention_num") || d.HasChange("rentention_day") || d.HasChange("status") ||
		d.HasChange("frequency") || d.HasChange("week_frequency") {
		if d.HasChange("name") {
			updateOpts.Name = d.Get("name").(string)
		}
		if d.HasChange("start_time") {
			updateOpts.ScheduledPolicy.StartTime = d.Get("start_time").(string)
		}
		if d.HasChange("rentention_num") {
			updateOpts.ScheduledPolicy.RententionNum = d.Get("rentention_num").(int)
		}
		if d.HasChange("rentention_day") {
			updateOpts.ScheduledPolicy.RententionDay = d.Get("rentention_day").(int)
		}
		if d.HasChange("retain_first_backup") {
			updateOpts.ScheduledPolicy.RemainFirstBackup = d.Get("retain_first_backup").(string)
		}
		if d.HasChange("status") {
			updateOpts.ScheduledPolicy.Status = d.Get("status").(string)
		}

		_, err = policies.Update(vbsClient, d.Id(), updateOpts).Extract()
		if err != nil {
			return fmterr.Errorf("error updating OpenTelekomCloud backup policy: %s", err)
		}
	}
	if d.HasChange("tags") {
		oldTags, _ := vbsTags.Get(vbsClient, d.Id()).Extract()
		deleteopts := vbsTags.BatchOpts{Action: vbsTags.ActionDelete, Tags: oldTags.Tags}
		deleteTags := vbsTags.BatchAction(vbsClient, d.Id(), deleteopts)
		if deleteTags.Err != nil {
			return fmterr.Errorf("error updating OpenTelekomCloud backup policy tags: %s", deleteTags.Err)
		}

		createTags := vbsTags.BatchAction(vbsClient, d.Id(), vbsTags.BatchOpts{Action: vbsTags.ActionCreate, Tags: resourceVBSUpdateTagsV2(d)})
		if createTags.Err != nil {
			return fmterr.Errorf("error updating OpenTelekomCloud backup policy tags: %s", createTags.Err)
		}
	}

	if d.HasChange("resources") {
		old, new := d.GetChange("resources")

		// disassociate old volumes from backup policy
		removeResources := buildDisassociateResource(old.([]interface{}))
		if len(removeResources) > 0 {
			opts := policies.DisassociateOpts{
				Resources: removeResources,
			}

			_, err := policies.Disassociate(vbsClient, d.Id(), opts).ExtractResource()
			if err != nil {
				return fmterr.Errorf("error disassociate volumes from VBS backup policy %s: %s",
					d.Id(), err)
			}
		}

		// associate new volumes to backup policy
		addResources := buildAssociateResource(new.([]interface{}))
		if len(addResources) > 0 {
			opts := policies.AssociateOpts{
				PolicyID:  d.Id(),
				Resources: addResources,
			}

			_, err := policies.Associate(vbsClient, opts).ExtractResource()
			if err != nil {
				return fmterr.Errorf("error associate volumes to VBS backup policy %s: %s",
					d.Id(), err)
			}
		}
	}

	return resourceVBSBackupPolicyV2Read(ctx, d, meta)
}

func resourceVBSBackupPolicyV2Delete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	vbsClient, err := config.VbsV2Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud VBS client: %s", err)
	}

	err = policies.Delete(vbsClient, d.Id()).Err
	if err != nil {
		if _, ok := err.(golangsdk.ErrDefault404); ok {
			log.Printf("[INFO] Successfully deleted OpenTelekomCloud VBS Backup Policy %s", d.Id())
		}
		if errCode, ok := err.(golangsdk.ErrUnexpectedResponseCode); ok {
			if errCode.Actual == 409 {
				log.Printf("[INFO] Error deleting OpenTelekomCloud VBS Backup Policy %s", d.Id())
			}
		}
		log.Printf("[INFO] Successfully deleted OpenTelekomCloud VBS Backup Policy %s", d.Id())
	}

	d.SetId("")
	return nil
}

func resourceVBSTagsV2(d *schema.ResourceData) []policies.Tag {
	rawTags := d.Get("tags").(*schema.Set).List()
	tags := make([]policies.Tag, len(rawTags))
	for i, raw := range rawTags {
		rawMap := raw.(map[string]interface{})
		tags[i] = policies.Tag{
			Key:   rawMap["key"].(string),
			Value: rawMap["value"].(string),
		}
	}
	return tags
}

func resourceVBSUpdateTagsV2(d *schema.ResourceData) []vbsTags.Tag {
	rawTags := d.Get("tags").(*schema.Set).List()
	tagList := make([]vbsTags.Tag, len(rawTags))
	for i, raw := range rawTags {
		rawMap := raw.(map[string]interface{})
		tagList[i] = vbsTags.Tag{
			Key:   rawMap["key"].(string),
			Value: rawMap["value"].(string),
		}
	}
	return tagList
}

func buildAssociateResource(raw []interface{}) []policies.AssociateResource {
	resources := make([]policies.AssociateResource, len(raw))
	for i, v := range raw {
		resources[i] = policies.AssociateResource{
			ResourceID:   v.(string),
			ResourceType: "volume",
		}
	}
	return resources
}

func buildDisassociateResource(raw []interface{}) []policies.DisassociateResource {
	resources := make([]policies.DisassociateResource, len(raw))
	for i, v := range raw {
		resources[i] = policies.DisassociateResource{
			ResourceID: v.(string),
		}
	}
	return resources
}

func buildWeekFrequencyResource(d *schema.ResourceData) ([]string, error) {
	validateList := []string{"SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"}
	weeks := []string{}

	weekRaws := d.Get("week_frequency").([]interface{})
	for _, wf := range weekRaws {
		found := false
		for _, value := range validateList {
			if wf.(string) == value {
				found = true
				break
			}
		}

		if found {
			weeks = append(weeks, wf.(string))
		} else {
			return nil, fmt.Errorf("expected item of week_frequency to be one of %v, got %s",
				validateList, wf.(string))
		}
	}
	return weeks, nil
}
