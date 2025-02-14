package collector

import (
	"ChintuIdrive/storage-node-watchdog/clients"
	"ChintuIdrive/storage-node-watchdog/dto"
	"log"
	"time"
)

type S3Metrics struct {
	DNS                      string                   `json:"dns"`
	BucketsCount             int                      `json:"buckets_count"`
	BucketListingDuration    time.Duration            `json:"bucket_listing_duration"`
	ObjectListingDurationMap map[string]time.Duration `json:"object_listing_duration_map"`
}

type S3MetricCollector struct {
	controllerCliet *clients.ControllerClient
}

func NewS3MetricCollector(cc *clients.ControllerClient) *S3MetricCollector {
	return &S3MetricCollector{
		controllerCliet: cc,
	}
}

func (s3mc *S3MetricCollector) CollectS3Metrics(tenat dto.Tenant) (*S3Metrics, error) {
	acckey, err := s3mc.controllerCliet.GetAccessKeys(tenat)
	if err != nil {
		return nil, err
	}
	ds, err := acckey.SecretKey.GetDString()
	if err != nil {
		return nil, err
	}
	log.Print(ds)
	client := clients.NewS3Client(tenat.DNS, acckey.AccessKey, ds)
	startTime := time.Now()
	buckets, err := client.ListBuckets()
	if err != nil {
		return nil, err
	}
	duration := time.Since(startTime)
	s3metrics := &S3Metrics{
		DNS:                      tenat.DNS,
		BucketsCount:             len(buckets),
		BucketListingDuration:    duration,
		ObjectListingDurationMap: make(map[string]time.Duration),
	}
	for _, bucket := range buckets {
		startTime = time.Now()
		client.ListObjectsForBucket(*bucket.Name)
		duration = time.Since(startTime)
		s3metrics.ObjectListingDurationMap[*bucket.Name] = duration
		log.Print(bucket.Name)
	}

	return s3metrics, nil
}
