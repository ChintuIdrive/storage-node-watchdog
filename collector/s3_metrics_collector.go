package collector

import (
	"ChintuIdrive/storage-node-watchdog/clients"
	"ChintuIdrive/storage-node-watchdog/conf"
	"ChintuIdrive/storage-node-watchdog/dto"
	"log"
	"time"
)

type S3Metrics struct {
	DNS                   string                   `json:"dns"`
	BucketsCount          int                      `json:"buckets_count"`
	BucketListingDuration time.Duration            `json:"bucket_listing_duration"`
	ObjectMetricsMap      map[string]ObjectMetrics `json:"object_metrics_map"`
}

type ObjectMetrics struct {
	ObjectsCount           int           `json:"objects_count"`
	ObjecttListingDuration time.Duration `json:"object_listing_duration"`
}

type S3MetricCollector struct {
	controllerCliet *clients.ControllerClient
	config          *conf.Config
}

func NewS3MetricCollector(config *conf.Config, cc *clients.ControllerClient) *S3MetricCollector {
	return &S3MetricCollector{
		config:          config,
		controllerCliet: cc,
	}
}

func (s3mc *S3MetricCollector) CollectS3Metrics(tenat dto.Tenant) (*S3Metrics, error) {
	s3config, err := s3mc.config.GetS3Config(tenat)
	if err != nil {
		log.Printf("Error getting S3 config for tenant %s: %v", tenat.DNS, err)
		return nil, err
	}

	acckey, err := s3mc.controllerCliet.GetSavedAccessKey(tenat)
	if err != nil {
		log.Printf("Error getting saved access key for tenant %s: %v", tenat.DNS, err)
		acckey, err = s3mc.controllerCliet.GetAccessKeys(tenat)
		if err != nil {
			return nil, err
		}
	}

	ds, err := acckey.SecretKey.GetDString()
	if err != nil {
		return nil, err
	}
	acckey.SecretKey.DString = ds
	log.Print(ds)
	client := clients.NewS3Client(tenat.DNS, acckey.AccessKey, ds)
	startTime := time.Now()
	buckets, err := client.ListBuckets()
	duration := time.Since(startTime)
	if err != nil {
		return nil, err
	}
	s3metrics := &S3Metrics{
		DNS:                   tenat.DNS,
		BucketsCount:          len(buckets),
		BucketListingDuration: duration,
		ObjectMetricsMap:      make(map[string]ObjectMetrics),
	}

	if s3config.BucketSelector == 0 {
		log.Printf("No specific bucket selector configured for tenant %s, processing all buckets", tenat.DNS)
		return s3metrics, nil
	}

	bucketsToProcess := buckets // set s3config.BucketSelector = -1 to process all the buckets
	if s3config.BucketSelector > 0 && s3config.BucketSelector < len(buckets) {
		bucketsToProcess = buckets[:s3config.BucketSelector]
	}

	for _, bucket := range bucketsToProcess {
		startTime = time.Now()
		objCount, err := client.ListObjectsForBucket(*bucket.Name, s3config.PageSelector)
		duration = time.Since(startTime)
		if err != nil {
			log.Printf("Error listing objects for bucket %s: %v", *bucket.Name, err)
			continue
		}
		objMetric := ObjectMetrics{
			ObjectsCount:           objCount,
			ObjecttListingDuration: duration,
		}
		s3metrics.ObjectMetricsMap[*bucket.Name] = objMetric
		log.Print(*bucket.Name)
	}

	return s3metrics, nil
}
