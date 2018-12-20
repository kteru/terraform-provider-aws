package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsSecurityHubProductSubscription() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSecurityHubProductSubscriptionCreate,
		Read:   resourceAwsSecurityHubProductSubscriptionRead,
		Delete: resourceAwsSecurityHubProductSubscriptionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"product_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
		},
	}
}

func resourceAwsSecurityHubProductSubscriptionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Printf("[DEBUG] Enabling Security Hub product subscription for product %s", d.Get("product_arn"))

	resp, err := conn.EnableImportFindingsForProduct(&securityhub.EnableImportFindingsForProductInput{
		ProductArn: aws.String(d.Get("product_arn").(string)),
	})

	if err != nil {
		return fmt.Errorf("Error enabling Security Hub product subscription for product %s: %s", d.Get("product_arn"), err)
	}

	d.SetId(*resp.ProductSubscriptionArn)

	return resourceAwsSecurityHubProductSubscriptionRead(d, meta)
}

func resourceAwsSecurityHubProductSubscriptionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn

	log.Printf("[DEBUG] Reading Security Hub product subscriptions to find %s", d.Id())

	exists, err := resourceAwsSecurityHubProductSubscriptionCheckExists(conn, d.Id())

	if err != nil {
		return fmt.Errorf("Error reading Security Hub product subscriptions to find %s: %s", d.Id(), err)
	}

	if !exists {
		log.Printf("[WARN] Security Hub product subscriptions (%s) not found, removing from state", d.Id())
		d.SetId("")
	}

	return nil
}

func resourceAwsSecurityHubProductSubscriptionCheckExists(conn *securityhub.SecurityHub, productSubscriptionArn string) (bool, error) {
	input := &securityhub.ListEnabledProductsForImportInput{}
	exists := false

	err := conn.ListEnabledProductsForImportPages(input, func(page *securityhub.ListEnabledProductsForImportOutput, lastPage bool) bool {
		for _, readProductSubscriptionArn := range page.ProductSubscriptions {
			if aws.StringValue(readProductSubscriptionArn) == productSubscriptionArn {
				exists = true
				return false
			}
		}
		return !lastPage
	})

	if err != nil {
		return false, err
	}

	return exists, nil
}

func resourceAwsSecurityHubProductSubscriptionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Printf("[DEBUG] Disabling Security Hub product subscription %s", d.Id())

	_, err := conn.DisableImportFindingsForProduct(&securityhub.DisableImportFindingsForProductInput{
		ProductSubscriptionArn: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("Error disabling Security Hub product subscription %s: %s", d.Id(), err)
	}

	return nil
}
