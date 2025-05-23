package ims

import (
	"context"
	"log"
	"regexp"
	"sort"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/ims/v2/images"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func DataSourceImagesImageV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceImagesImageV2Read,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
			},
			"visibility": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"size_min": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"size_max": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"sort_key": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"sort_direction": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "asc",
				ValidateFunc: validation.StringInSlice([]string{
					"asc", "desc",
				}, false),
			},
			"tag": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"most_recent": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"container_format": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disk_format": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protected": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"size_bytes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"schema": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"image_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_registered": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"system_cmk_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"os_bit": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"os_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"support_disk_intensive": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"support_high_performance": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"support_kvm": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"support_kvm_gpu_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"support_kvm_infiniband": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"support_large_memory": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"support_xen": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"support_xen_gpu_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"support_xen_hana": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"min_disk": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"min_ram": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"virtual_env_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"backup_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_origin": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_source_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"original_image_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"os_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"login_user": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hw_firmware_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceImagesImageV2Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.ImageV2Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud IMSv2 client: %w", err)
	}

	listOpts := images.ListImagesOpts{
		Name:       d.Get("name").(string),
		Visibility: d.Get("visibility").(string),
		Owner:      d.Get("owner").(string),
		Status:     "active",
		SizeMin:    int64(d.Get("size_min").(int)),
		SizeMax:    int64(d.Get("size_max").(int)),
		SortKey:    d.Get("sort_key").(string),
		SortDir:    d.Get("sort_direction").(string),
		Tag:        d.Get("tag").(string),
	}

	log.Printf("[DEBUG] List Options: %#v", listOpts)

	var img images.ImageInfo
	ims, err := images.ListImages(client, listOpts)
	if err != nil {
		return fmterr.Errorf("unable to query images: %s", err)
	}

	var filteredImages []images.ImageInfo
	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		for _, image := range ims {
			if r.MatchString(image.Name) {
				filteredImages = append(filteredImages, image)
			}
		}
		ims = filteredImages
	}

	if len(ims) < 1 {
		return fmterr.Errorf("your query returned no results. " +
			"Please change your search criteria and try again")
	}

	if len(ims) > 1 {
		recent := d.Get("most_recent").(bool)
		log.Printf("[DEBUG] Multiple results found and `most_recent` is set to: %t", recent)
		if recent {
			img = mostRecentImage(ims)
		} else {
			return fmterr.Errorf("your query returned more than one result. Please try a more " +
				"specific search criteria, or set `most_recent` attribute to true")
		}
	} else {
		img = ims[0]
	}

	log.Printf("[DEBUG] Single Image found: %s", img.Id)
	d.SetId(img.Id)

	mErr := multierror.Append(nil,
		d.Set("name", img.Name),
		d.Set("tags", img.Tags),
		d.Set("container_format", img.ContainerFormat),
		d.Set("disk_format", img.DiskFormat),
		d.Set("min_disk", img.MinDisk),
		d.Set("min_ram", img.MinRam),
		d.Set("owner", img.Owner),
		d.Set("protected", img.Protected),
		d.Set("visibility", img.Visibility),
		d.Set("checksum", img.Checksum),
		d.Set("size_bytes", img.Size),
		d.Set("created_at", img.CreatedAt.Format(time.RFC3339)),
		d.Set("updated_at", img.UpdatedAt.Format(time.RFC3339)),
		d.Set("file", img.File),
		d.Set("schema", img.Schema),
		d.Set("id", img.Id),
		d.Set("status", img.Status),
		d.Set("os_type", img.OsType),
		d.Set("platform", img.Platform),
		d.Set("backup_id", img.BackupId),
		d.Set("data_origin", img.DataOrigin),
		d.Set("description", img.Description),
		d.Set("image_source_type", img.ImageSourceType),
		d.Set("is_registered", img.Isregistered),
		d.Set("original_image_name", img.Originalimagename),
		d.Set("os_bit", img.OsBit),
		d.Set("os_version", img.OsVersion),
		d.Set("virtual_env_type", img.VirtualEnvType),
		d.Set("login_user", img.ImageLoginUser),
		d.Set("image_type", img.Imagetype),
		d.Set("system_cmk_id", img.SystemCmkid),
		d.Set("support_disk_intensive", img.SupportDiskintensive),
		d.Set("support_high_performance", img.SupportHighperformance),
		d.Set("support_kvm", img.SupportKvm),
		d.Set("support_kvm_gpu_type", img.SupportKvmGpuType),
		d.Set("support_kvm_infiniband", img.SupportKvmInfiniband),
		d.Set("support_large_memory", img.SupportLargememory),
		d.Set("support_xen", img.SupportXen),
		d.Set("support_xen_gpu_type", img.SupportXenGpuType),
		d.Set("support_xen_hana", img.SupportXenHana),
		d.Set("hw_firmware_type", img.HwFirmwareType),
	)

	if mErr.ErrorOrNil() != nil {
		return diag.FromErr(mErr)
	}

	return nil
}

type imageSort []images.ImageInfo

func (a imageSort) Len() int {
	return len(a)
}

func (a imageSort) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a imageSort) Less(i, j int) bool {
	itime := a[i].CreatedAt
	jtime := a[j].CreatedAt
	return itime.Unix() < jtime.Unix()
}

// Returns the most recent Image out of a slice of images.
func mostRecentImage(images []images.ImageInfo) images.ImageInfo {
	sortedImages := images
	sort.Sort(imageSort(sortedImages))
	return sortedImages[len(sortedImages)-1]
}
