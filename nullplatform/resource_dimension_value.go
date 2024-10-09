package nullplatform

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDimensionValue() *schema.Resource {
	return &schema.Resource{
		Description: "The dimension_value resource allows you to configure a Nullplatform Dimension Value",

		CreateContext: resourceDimensionValueCreate,
		ReadContext:   resourceDimensionValueRead,
		DeleteContext: resourceDimensionValueDelete,

		Schema: map[string]*schema.Schema{
			"dimension_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the parent dimension.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the dimension value.",
			},
			"nrn": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The NRN (Null Resource Name) of the dimension value.",
			},
			"slug": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The slug of the dimension value.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the dimension value.",
			},
		},
	}
}

func resourceDimensionValueCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(NullOps)

	dimensionValue := &DimensionValue{
		DimensionID: d.Get("dimension_id").(int),
		Name:        d.Get("name").(string),
		NRN:         d.Get("nrn").(string),
	}

	createdValue, err := c.CreateDimensionValue(dimensionValue)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(createdValue.ID))

	return resourceDimensionValueRead(ctx, d, m)
}

func resourceDimensionValueRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(NullOps)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid dimension value ID: %v", err))
	}

	dimensionID := d.Get("dimension_id").(int)

	value, err := c.GetDimensionValue(dimensionID, id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("dimension_id", value.DimensionID)
	d.Set("name", value.Name)
	d.Set("nrn", value.NRN)
	d.Set("slug", value.Slug)
	d.Set("status", value.Status)

	return nil
}

func resourceDimensionValueDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(NullOps)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid dimension value ID: %v", err))
	}

	dimensionID := d.Get("dimension_id").(int)

	err = c.DeleteDimensionValue(dimensionID, id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}