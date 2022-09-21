package flexibleengine

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/openstack/mrs/v2/jobs"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	// JobFlink is a type of the MRS job, which specifies the use of Flink componment.
	// The Flink is a unified computing framework that supports both batch processing and stream processing.
	JobFlink = "Flink"
	// JobHiveSQL is a type of the MRS job, which specifies the use of Hive componment by a sql command.
	// The Hive is a data warehouse infrastructure built on Hadoop.
	JobHiveSQL = "HiveSql"
	// JobHiveScript is a type of the MRS job, which specifies the use of Hive componment by a sql file.
	JobHiveScript = "HiveScript"
	// JobMapReduce is a type of the MRS job, which specifies the use of MapReduce componment.
	// MapReduce is the core of Hadoop.
	JobMapReduce = "MapReduce"
	// JobSparkSubmit is a type of the MRS job, which specifies the use of Spark componment to submit a job to MRS
	// executor.
	JobSparkSubmit = "SparkSubmit"
	// JobSparkSQL is a type of the MRS job, which specifies the use of Spark componment by a sql command.
	JobSparkSQL = "SparkSql"
	// JobSparkScript is a type of the MRS job, which specifies the use of Spark componment by a sql file.
	JobSparkScript = "SparkScript"
)

func resourceMRSJobV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceMRSJobV2Create,
		Read:   resourceMRSJobV2Read,
		Delete: resourceMRSJobV2Delete,

		Importer: &schema.ResourceImporter{
			State: resourceMRSSubResourceImportState,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile("^[A-Za-z0-9_-]{1,64}$"),
					"The name consists of 1 to 64 characters, which only letters, digits, hyphens (-) and "+
						"underscores (_) are allowed."),
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					JobFlink, JobHiveSQL, JobHiveScript, JobMapReduce, JobSparkSubmit, JobSparkSQL, JobSparkScript,
				}, false),
			},
			"program_path": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"parameters": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"program_parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"service_parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"sql": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"start_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"submit_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"finish_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func buildMRSJobProgramParameters(programs map[string]interface{}) []string {
	result := make([]string, 0)
	for k, v := range programs {
		result = append(result, k)
		result = append(result, v.(string))
	}
	log.Printf("[DEBUG] The program parameters are: %+v", result)
	return result
}

func buildMRSJobParameters(parameters string) []string {
	return strings.Split(parameters, " ")
}

// The Request arguments of the flink job is: run -d <program parameters> -m yarn-cluster <jar path> <parameters>.
func buildMRSFlinkJobRequestArguments(d *schema.ResourceData) []string {
	programsMap := d.Get("program_parameters").(map[string]interface{})
	programs := buildMRSJobProgramParameters(programsMap)
	parameters := buildMRSJobParameters(d.Get("parameters").(string))
	// The capacity of the result array is the sum of the respective lengths of 'run', '-d', '-m', 'yarn-cluster',
	// jar path, program parameters and parameters.
	result := make([]string, 0, 5+len(programs)+len(parameters))

	result = append(result, "run")
	result = append(result, "-d")
	result = append(result, programs...)
	result = append(result, "-m")
	result = append(result, "yarn-cluster")
	result = append(result, d.Get("program_path").(string))
	result = append(result, parameters...)

	return result
}

// The request arguments of the SQL job is: <program parameters> <sql (file or path)>.
func buildMRSSQLJobRequestArguments(d *schema.ResourceData) []string {
	programsMap := d.Get("program_parameters").(map[string]interface{})
	programs := buildMRSJobProgramParameters(programsMap)
	// The capacity of the result array is the sum of the respective lengths of program parameters and sql parameter.
	result := make([]string, 0, 1+len(programs))

	result = append(result, programs...)
	result = append(result, d.Get("sql").(string))

	return result
}

// The request arguments of the MapReduce job is: <jar path> <parameters>.
func buildMRSMapReduceJobRequestArguments(d *schema.ResourceData) []string {
	parameters := buildMRSJobParameters(d.Get("parameters").(string))
	// The capacity of the result array is the sum of the respective lengths of jar path (/python path) and parameter.
	result := make([]string, 0, 1+len(parameters))

	result = append(result, d.Get("program_path").(string))
	result = append(result, parameters...)
	return result
}

// The request arguments of the SparkSubmit job is:
//
//	<program parameters> --master yarn-cluster <jar path (/python path)> <parameters>.
func buildMRSSparkSubmitJobRequestArguments(d *schema.ResourceData) []string {
	programsMap := d.Get("program_parameters").(map[string]interface{})
	programs := buildMRSJobProgramParameters(programsMap)
	parameters := buildMRSJobParameters(d.Get("parameters").(string))

	// The capacity of the result array is the sum of the respective lengths of '--master', 'yarn-cluster',
	// jar path (/python path), program parameters and parameters.
	result := make([]string, 0, 3+len(programs)+len(parameters))

	result = append(result, programs...)
	result = append(result, "--master")
	result = append(result, "yarn-cluster")
	result = append(result, d.Get("program_path").(string))
	result = append(result, parameters...)

	return result
}

func buildMRSJobProperties(d *schema.ResourceData) map[string]string {
	result := make(map[string]string)
	properties := d.Get("service_parameters").(map[string]interface{})
	for k, v := range properties {
		result[k] = v.(string)
	}
	log.Printf("[DEBUG] The properties are: %+v", result)
	return result
}

func buildMRSJobCreateParameters(d *schema.ResourceData) jobs.CreateOpts {
	opts := jobs.CreateOpts{
		JobType:    d.Get("type").(string),
		JobName:    d.Get("name").(string),
		Properties: buildMRSJobProperties(d),
	}
	switch d.Get("type").(string) {
	case JobFlink:
		opts.Arguments = buildMRSFlinkJobRequestArguments(d)
	case JobHiveSQL, JobHiveScript, JobSparkSQL, JobSparkScript:
		opts.Arguments = buildMRSSQLJobRequestArguments(d)
	case JobMapReduce:
		opts.Arguments = buildMRSMapReduceJobRequestArguments(d)
	default:
		opts.Arguments = buildMRSSparkSubmitJobRequestArguments(d)
	}
	return opts
}

func resourceMRSJobV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	client, err := config.MrsV2Client(config.GetRegion(d))
	if err != nil {
		return fmt.Errorf("Error creating FlexibleEngine MRS V2 client: %s", err)
	}

	opts := buildMRSJobCreateParameters(d)
	clusterID := d.Get("cluster_id").(string)
	resp, err := jobs.Create(client, clusterID, opts).Extract()
	if err != nil {
		return fmt.Errorf("Error execution MapReduce job: %s", err)
	}
	d.SetId(resp.JobSubmitResult.JobId)

	stateConf := &resource.StateChangeConf{
		Pending:      []string{"NEW", "NEW_SAVING", "ACCEPTED", "SUBMITTED", "RUNNING"},
		Target:       []string{"FINISHED", "FAILED"},
		Refresh:      mrsJobStateRefreshFunc(client, clusterID, d.Id()),
		Timeout:      d.Timeout(schema.TimeoutCreate),
		Delay:        30 * time.Second,
		PollInterval: 10 * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for job (%s) to become ready: %s ", d.Id(), err)
	}

	return resourceMRSJobV2Read(d, meta)
}

func mrsJobStateRefreshFunc(client *golangsdk.ServiceClient, clusterID, jobId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := jobs.Get(client, clusterID, jobId).Extract()
		if err != nil {
			if isResourceNotFound(err) {
				return resp, "DELETED", nil
			}
			return nil, "", err
		}
		return resp, resp.JobState, nil
	}
}

// The arguments (/program parameters) of the response is a list string.
// For example: "["run", "-d", "-m", "yarn-cluster", "obs://obs-demo-analysis-tf/program/driver_behavior.jar"]".
func makeMRSArgumentsByString(str string) []string {
	regex := regexp.MustCompile(`^\[(.*)\]$`)
	result := regex.FindStringSubmatch(str)
	if len(result) > 1 {
		str := result[1]
		// Separate all elements based on commas.
		return strings.Split(str, ", ")
	}
	return []string{}
}

// The string arguments of the flink job is: 'run -d <program parameters> -m yarn-cluster <parameters>'.
func makeMRSFlinkJobParameters(job *jobs.Job) (string, string, map[string]interface{}, error) {
	programs := make(map[string]interface{})
	arguments := makeMRSArgumentsByString(job.Arguments)
	// The arguments must contain a head of 'run' and '-d'.
	if len(arguments) < 2 {
		return "", "", programs, fmt.Errorf("Wrong flink arguments length of the API response")
	}
	arguments = arguments[2:] // remove 'run -d' program argument.
	for arguments[0] != "-m" && len(arguments) > 1 {
		programs[arguments[0]] = arguments[1]
		arguments = arguments[2:]
	}
	// The remaining elements of arguments must contain '-m', 'yarn-cluster' and jar path.
	if len(arguments) < 3 {
		return "", "", programs, fmt.Errorf("Wrong flink arguments length of the API response")
	}
	arguments = arguments[2:] // remove '-m yarn-cluster'.
	// get jar path and remove it from argument list.
	jarPath := arguments[0]
	arguments = arguments[1:] // remove jar path.
	parameters := strings.Join(arguments, " ")
	return jarPath, parameters, programs, nil
}

// The string arguments of the flink job is: '<program parameters> <sql statement/file path>'.
func makeMRSSQLJobParameters(job *jobs.Job) (string, map[string]interface{}, error) {
	programs := make(map[string]interface{})
	arguments := makeMRSArgumentsByString(job.Arguments)
	for len(arguments) > 1 {
		// The program parameters in the state is a map.
		programs[arguments[0]] = arguments[1]
		arguments = arguments[2:]
	}
	if len(arguments) < 1 {
		return "", programs, fmt.Errorf("The arguments of the API response has not contain statement of SQL file")
	}
	return arguments[0], programs, nil
}

// The string arguments of the flink job is: '<jar path (/python path)> <parameters>'.
func makeMRSMapReduceJobParameters(job *jobs.Job) (string, string, error) {
	arguments := makeMRSArgumentsByString(job.Arguments)
	// The arguments must contain jar path.
	if len(arguments) < 1 {
		return "", "", fmt.Errorf("Wrong arguments length of the API response")
	}
	// get jar path and remove it from argument list.
	jarPath := arguments[0]
	arguments = arguments[1:]
	// get parameters string.
	parameters := strings.Join(arguments, " ")

	return jarPath, parameters, nil
}

// The string arguments of the flink job is:
//
//	'<program parameters> --master yarn-cluster <jar path (/python path)> <parameters>'.
func makeMRSSparkSubmitJobParameters(job *jobs.Job) (string, string, map[string]interface{}, error) {
	programs := make(map[string]interface{})
	arguments := makeMRSArgumentsByString(job.Arguments)

	for arguments[0] != "--master" && len(arguments) > 1 {
		// The program parameters in the state is a map.
		programs[arguments[0]] = arguments[1]
		arguments = arguments[2:]
	}
	// The remaining elements of arguments must contain '--master', 'yarn-cluster' and jar path (/python path).
	if len(arguments) < 3 {
		return "", "", programs, fmt.Errorf("Wrong arguments length of the API response")
	}
	arguments = arguments[2:] // remove '--master' and 'yarn-clsuter' program arguments.
	// get jar path (/python path) and remove it from argument list.
	jarPath := arguments[0]
	arguments = arguments[1:]
	// get parameters string.
	parameters := strings.Join(arguments, " ")

	return jarPath, parameters, programs, nil
}

func setMRSFlinkJob(d *schema.ResourceData, resp *jobs.Job) error {
	jarPath, parameters, programs, err := makeMRSFlinkJobParameters(resp)
	if err != nil {
		return err
	}

	mErr := multierror.Append(
		d.Set("program_path", jarPath),
		d.Set("parameters", parameters),
		d.Set("program_parameters", programs),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return fmt.Errorf("Error setting job fields(: jar path, parameters or program parameters): %s", err)
	}

	return nil
}

func setMRSSQLJob(d *schema.ResourceData, resp *jobs.Job) error {
	statement, programs, err := makeMRSSQLJobParameters(resp)
	if err != nil {
		return err
	}

	mErr := multierror.Append(
		d.Set("sql", statement),
		d.Set("program_parameters", programs),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return fmt.Errorf("Error setting job fields(: sql or program parameters): %s", err)
	}

	return nil
}

func setMRSMapReduceSubmitJob(d *schema.ResourceData, resp *jobs.Job) error {
	jarPath, parameters, err := makeMRSMapReduceJobParameters(resp)
	if err != nil {
		return err
	}

	mErr := multierror.Append(
		d.Set("program_path", jarPath),
		d.Set("parameters", parameters),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return fmt.Errorf("Error setting job fields (jar path or parameters): %s", err)
	}

	return nil
}

func setMRSSparkSubmitJob(d *schema.ResourceData, resp *jobs.Job) error {
	jarPath, parameters, programs, err := makeMRSSparkSubmitJobParameters(resp)
	if err != nil {
		return err
	}

	mErr := multierror.Append(
		d.Set("program_path", jarPath),
		d.Set("parameters", parameters),
		d.Set("program_parameters", programs),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return fmt.Errorf("Error setting job fields(: jar path, parameters or program parameters): %s", err)
	}

	return nil
}

func setMRSJobParametersByArguments(d *schema.ResourceData, job *jobs.Job) error {
	switch job.JobType {
	case JobHiveSQL, JobHiveScript, JobSparkSQL, JobSparkScript:
		return setMRSSQLJob(d, job)
	case JobMapReduce:
		return setMRSMapReduceSubmitJob(d, job)
	case JobFlink:
		return setMRSFlinkJob(d, job)
	default:
		return setMRSSparkSubmitJob(d, job)
	}
}

// The properties of the response is a map string, and the separator between key and value is the equal sign (=).
// For example: "{fs.obs.access.key=xxx, fs.obs.secret.key=xxx}".
func setMRSJobProperties(d *schema.ResourceData, resp string) error {
	properties := make(map[string]interface{})
	// Remove the braces around the map string.
	regex := regexp.MustCompile(`^{(.*)}$`)
	result := regex.FindStringSubmatch(resp)
	if len(result) > 1 {
		str := result[1]
		if str == "" {
			return nil
		}
		// Separate all key-value pairs based on commas.
		elements := strings.Split(str, ", ")
		for _, element := range elements {
			property := strings.SplitN(element, "=", 2)
			if len(property) == 2 {
				properties[property[0]] = property[1]
				continue
			}
			return fmt.Errorf("The property (%s) of the MRS job is invalid", element)
		}
	}

	return d.Set("service_parameters", properties)
}

func setMRSTimeProperties(d *schema.ResourceData, key string, value int) error {
	keyTime := time.Unix(int64(value), 0)
	return d.Set(key, keyTime.Format(RFC3339ZNoTNoZ))
}

func resourceMRSJobV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	region := config.GetRegion(d)
	client, err := config.MrsV2Client(region)
	if err != nil {
		return fmt.Errorf("Error creating FlexibleEngine MRS client: %s", err)
	}

	clusterID := d.Get("cluster_id").(string)
	resp, err := jobs.Get(client, clusterID, d.Id()).Extract()
	if err != nil {
		return fmt.Errorf("Error getting MRS job form server: %s", err)
	}

	log.Printf("[DEBUG] Retrieved MRS job (%s): %+v", d.Id(), resp)
	d.SetId(resp.JobId)
	mErr := multierror.Append(
		d.Set("region", region),
		d.Set("type", resp.JobType),
		d.Set("name", resp.JobName),
		d.Set("status", resp.JobState),
		setMRSJobParametersByArguments(d, resp),
		setMRSJobProperties(d, resp.Properties),
		setMRSTimeProperties(d, "start_time", resp.StartedTime/1000),
		setMRSTimeProperties(d, "submit_time", resp.SubmittedTime/1000),
		setMRSTimeProperties(d, "finish_time", resp.FinishedTime/1000),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return fmt.Errorf("Error setting MRS job fields: %s", err)
	}

	return nil
}

func resourceMRSJobV2Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	client, err := config.MrsV2Client(config.GetRegion(d))
	if err != nil {
		return fmt.Errorf("Error creating FlexibleEngine MRS client: %s", err)
	}

	clusterID := d.Get("cluster_id").(string)
	opts := jobs.DeleteOpts{
		JobIds: []string{d.Id()},
	}
	err = jobs.Delete(client, clusterID, opts).ExtractErr()
	if err != nil {
		return fmt.Errorf("Error deleting FlexibleEngine MRS job: %s", err)
	}

	d.SetId("")
	return nil
}

func resourceMRSSubResourceImportState(d *schema.ResourceData,
	meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid format specified for import IDs, must be <cluster_id>/<id>")
	}

	d.SetId(parts[1])
	d.Set("cluster_id", parts[0])
	return []*schema.ResourceData{d}, nil
}
