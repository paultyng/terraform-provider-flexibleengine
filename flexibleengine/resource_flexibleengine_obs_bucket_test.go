package flexibleengine

import (
	"fmt"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils/fmtp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccObsBucket_basic(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "flexibleengine_obs_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckS3(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckObsBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObsBucket_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket", testAccObsBucketName(rInt)),
					resource.TestCheckResourceAttr(resourceName, "bucket_domain_name", testAccObsBucketDomainName(rInt)),
					resource.TestCheckResourceAttr(resourceName, "acl", "private"),
					resource.TestCheckResourceAttr(resourceName, "storage_class", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, "encryption", "false"),
					resource.TestCheckResourceAttr(resourceName, "multi_az", "false"),
					resource.TestCheckResourceAttr(resourceName, "region", OS_REGION_NAME),
				),
			},
			{
				Config: testAccObsBucket_basic_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", "public-read"),
					resource.TestCheckResourceAttr(resourceName, "storage_class", "GLACIER"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"acl",
					"force_destroy",
				},
			},
		},
	})
}

func TestAccObsBucket_multiAZ(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "flexibleengine_obs_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheckS3(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccObsBucketConfig_multiAZ(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket", testAccObsBucketName(rInt)),
					resource.TestCheckResourceAttr(resourceName, "acl", "public-read"),
					resource.TestCheckResourceAttr(resourceName, "storage_class", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, "multi_az", "true"),
				),
			},
		},
	})
}

/*
func TestAccObsBucket_tags(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "flexibleengine_obs_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheckS3(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccObsBucketConfigWithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "tags.name", testAccObsBucketName(rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}
*/

func TestAccObsBucket_versioning(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "flexibleengine_obs_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckS3(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckObsBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObsBucketConfigWithVersioning(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "versioning", "true"),
				),
			},
			{
				Config: testAccObsBucketConfigWithDisableVersioning(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "versioning", "false"),
				),
			},
		},
	})
}

func TestAccObsBucket_logging(t *testing.T) {
	rInt := acctest.RandInt()
	target_bucket := fmt.Sprintf("tf-test-log-bucket-%d", rInt)
	resourceName := "flexibleengine_obs_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckS3(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckObsBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObsBucketConfigWithLogging(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					testAccCheckObsBucketLogging(resourceName, target_bucket, "log/"),
				),
			},
		},
	})
}

func TestAccObsBucket_lifecycle(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "flexibleengine_obs_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckS3(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckObsBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObsBucketConfigWithLifecycle(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.name", "rule1"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.1.name", "rule2"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.1.prefix", "path2/"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.2.name", "rule3"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.2.prefix", "path3/"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.1.transition.0.days", "30"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.1.transition.1.days", "180"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.2.noncurrent_version_transition.0.days", "60"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.2.noncurrent_version_transition.1.days", "180"),
				),
			},
		},
	})
}

func TestAccObsBucket_website(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "flexibleengine_obs_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckS3(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckObsBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObsBucketWebsiteConfigWithRoutingRules(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "website.0.index_document", "index.html"),
					resource.TestCheckResourceAttr(
						resourceName, "website.0.error_document", "error.html"),
				),
			},
		},
	})
}

func TestAccObsBucket_cors(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "flexibleengine_obs_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckS3(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckObsBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObsBucketConfigWithCORS(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "cors_rule.0.allowed_origins.0", "https://www.example.com"),
					resource.TestCheckResourceAttr(
						resourceName, "cors_rule.0.allowed_methods.0", "PUT"),
					resource.TestCheckResourceAttr(
						resourceName, "cors_rule.0.allowed_headers.0", "*"),
					resource.TestCheckResourceAttr(
						resourceName, "cors_rule.0.expose_headers.1", "ETag"),
					resource.TestCheckResourceAttr(
						resourceName, "cors_rule.0.max_age_seconds", "3000"),
				),
			},
		},
	})
}

func TestAccObsBucket_notification(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "flexibleengine_obs_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckS3(t)
			testAccPreCheckOBSNotification(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckObsBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObsBucketConfigWithNotification(rInt, OS_OBS_URN_SMN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "topic_configurations.0.topic_id", "001"),
					resource.TestCheckResourceAttr(
						resourceName, "topic_configurations.0.topic", OS_OBS_URN_SMN),
					resource.TestCheckResourceAttr(
						resourceName, "topic_configurations.0.events.0", "ObjectCreated:*"),
					resource.TestCheckResourceAttr(
						resourceName, "topic_configurations.0.filter_rules.0.name", "prefix"),
					resource.TestCheckResourceAttr(
						resourceName, "topic_configurations.0.filter_rules.0.value", "tf"),
					resource.TestCheckResourceAttr(
						resourceName, "topic_configurations.1.topic_id", "002"),
					resource.TestCheckResourceAttr(
						resourceName, "topic_configurations.1.topic", OS_OBS_URN_SMN),
					resource.TestCheckResourceAttr(
						resourceName, "topic_configurations.1.events.0", "ObjectCreated:Put"),
					resource.TestCheckResourceAttr(
						resourceName, "topic_configurations.1.filter_rules.0.name", "prefix"),
					resource.TestCheckResourceAttr(
						resourceName, "topic_configurations.1.filter_rules.0.value", "demo"),
				),
			},
			{
				Config: testAccObsBucketConfigWithNotificationDelete(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObsBucketExists(resourceName),
					testAccCheckObsBucketNotificationDelete(resourceName),
				),
			},
		},
	})
}

func testAccCheckObsBucketDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	obsClient, err := config.ObjectStorageClient(OS_REGION_NAME)
	if err != nil {
		return fmt.Errorf("Error creating FlexibleEngine OBS client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "flexibleengine_obs_bucket" {
			continue
		}

		_, err := obsClient.HeadBucket(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("FlexibleEngine OBS Bucket %s still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckObsBucketExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		obsClient, err := config.ObjectStorageClient(OS_REGION_NAME)
		if err != nil {
			return fmt.Errorf("Error creating FlexibleEngine OBS client: %s", err)
		}

		_, err = obsClient.HeadBucket(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("FlexibleEngine OBS Bucket not found: %v", err)
		}
		return nil
	}
}

func testAccCheckObsBucketLogging(name, target, prefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		config := testAccProvider.Meta().(*Config)
		obsClient, err := config.ObjectStorageClient(OS_REGION_NAME)
		if err != nil {
			return fmt.Errorf("Error creating FlexibleEngine OBS client: %s", err)
		}

		output, err := obsClient.GetBucketLoggingConfiguration(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error getting logging configuration of OBS bucket: %s", err)
		}

		if output.TargetBucket != target {
			return fmt.Errorf("%s.logging: Attribute 'target_bucket' expected %s, got %s",
				name, output.TargetBucket, target)
		}
		if output.TargetPrefix != prefix {
			return fmt.Errorf("%s.logging: Attribute 'target_prefix' expected %s, got %s",
				name, output.TargetPrefix, prefix)
		}

		return nil
	}
}

func testAccCheckObsBucketNotificationDelete(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		conf := testAccProvider.Meta().(*Config)
		obsClient, err := conf.ObjectStorageClient(OS_REGION_NAME)
		if err != nil {
			return fmt.Errorf("error creating FlexibleEngine OBS client: %s", err)
		}

		if rs.Primary.ID == "" {
			return fmtp.Errorf("No ID is set. ")
		}

		notifications, err := obsClient.GetBucketNotification(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error getting notification of OBS bucket: %s, err: %s", rs.Primary.ID, err)
		}
		if len(notifications.TopicConfigurations) > 0 {
			return fmt.Errorf("error deleteing notification of OBS bucket: %s, err: %s", rs.Primary.ID, err)
		}
		return nil
	}
}

// These need a bit of randomness as the name can only be used once globally
func testAccObsBucketName(randInt int) string {
	return fmt.Sprintf("tf-test-bucket-%d", randInt)
}

func testAccObsBucketDomainName(randInt int) string {
	return fmt.Sprintf("tf-test-bucket-%d.oss.%s.prod-cloud-ocb.orange-business.com", randInt, OS_REGION_NAME)
}

func testAccObsBucket_basic(randInt int) string {
	return fmt.Sprintf(`
resource "flexibleengine_obs_bucket" "bucket" {
  bucket        = "tf-test-bucket-%d"
  storage_class = "STANDARD"
  acl           = "private"
}
`, randInt)
}

func testAccObsBucket_basic_update(randInt int) string {
	return fmt.Sprintf(`
resource "flexibleengine_obs_bucket" "bucket" {
  bucket        = "tf-test-bucket-%d"
  storage_class = "GLACIER"
  acl           = "public-read"
}
`, randInt)
}

func testAccObsBucketConfig_multiAZ(randInt int) string {
	return fmt.Sprintf(`
resource "flexibleengine_obs_bucket" "bucket" {
  bucket   = "tf-test-bucket-%d"
  acl      = "public-read"
  multi_az = true
}
`, randInt)
}

func testAccObsBucketConfigWithTags(randInt int) string {
	return fmt.Sprintf(`
resource "flexibleengine_obs_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl = "private"

  tags = {
    name = "tf-test-bucket-%d"
    foo = "bar"
    key1 = "value1"
  }
}
`, randInt, randInt)
}

func testAccObsBucketConfigWithVersioning(randInt int) string {
	return fmt.Sprintf(`
resource "flexibleengine_obs_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl = "private"
  versioning = true
}
`, randInt)
}

func testAccObsBucketConfigWithDisableVersioning(randInt int) string {
	return fmt.Sprintf(`
resource "flexibleengine_obs_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl = "private"
  versioning = false
}
`, randInt)
}

func testAccObsBucketConfigWithLogging(randInt int) string {
	return fmt.Sprintf(`
resource "flexibleengine_obs_bucket" "log_bucket" {
  bucket = "tf-test-log-bucket-%d"
  acl = "log-delivery-write"
  force_destroy = "true"
}
resource "flexibleengine_obs_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl = "private"

  logging {
    target_bucket = flexibleengine_obs_bucket.log_bucket.id
    target_prefix = "log/"
  }
}
`, randInt, randInt)
}

func testAccObsBucketConfigWithLifecycle(randInt int) string {
	return fmt.Sprintf(`
resource "flexibleengine_obs_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl = "private"
  versioning = true

  lifecycle_rule {
    name = "rule1"
    prefix = "path1/"
    enabled = true

    expiration {
      days = 365
    }
  }
  lifecycle_rule {
    name = "rule2"
    prefix = "path2/"
    enabled = true

    expiration {
      days = 365
    }

    transition {
      days = 30
      storage_class = "STANDARD_IA"
    }
    transition {
      days = 180
      storage_class = "GLACIER"
    }
  }
  lifecycle_rule {
    name = "rule3"
    prefix = "path3/"
    enabled = true

    noncurrent_version_expiration {
      days = 365
    }

    noncurrent_version_transition {
      days = 60
      storage_class = "STANDARD_IA"
    }
    noncurrent_version_transition {
      days = 180
      storage_class = "GLACIER"
    }
  }
}
`, randInt)
}

func testAccObsBucketWebsiteConfigWithRoutingRules(randInt int) string {
	return fmt.Sprintf(`
resource "flexibleengine_obs_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl = "public-read"

  website {
    index_document = "index.html"
    error_document = "error.html"
    routing_rules = <<EOF
[{
	"Condition": {
		"KeyPrefixEquals": "docs/"
	},
	"Redirect": {
		"ReplaceKeyPrefixWith": "documents/"
	}
}]
EOF
  }
}
`, randInt)
}

func testAccObsBucketConfigWithCORS(randInt int) string {
	return fmt.Sprintf(`
resource "flexibleengine_obs_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl = "public-read"

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT","POST"]
    allowed_origins = ["https://www.example.com"]
    expose_headers  = ["x-amz-server-side-encryption","ETag"]
    max_age_seconds = 3000
  }
}
`, randInt)
}

func testAccObsBucketConfigWithNotification(randInt int, urnSmn string) string {
	return fmt.Sprintf(`
resource "flexibleengine_obs_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl = "public-read"

  topic_configurations {
    topic_id = "001"
    topic    = "%s"
    events   = ["ObjectCreated:*"]

	filter_rules {
      name  = "prefix"
      value = "tf"
	}

	filter_rules {
      name  = "suffix"
      value = ".jpg"
	}
  }

  topic_configurations {
    topic_id = "002"
    topic    = "%s"
    events   = ["ObjectCreated:Put"]

	filter_rules {
      name  = "prefix"
      value = "demo"
	}

	filter_rules {
      name  = "suffix"
      value = ".png"
	}
  }

}
`, randInt, urnSmn, urnSmn)
}

func testAccObsBucketConfigWithNotificationDelete(randInt int) string {
	return fmt.Sprintf(`
resource "flexibleengine_obs_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl = "public-read"
  
}
`, randInt)
}
