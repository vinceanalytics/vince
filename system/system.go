package system

// histograms
var (
	SaveDuration      = &histogramMetric{name: "mike_save_duration"}
	MergeDuration     = &histogramMetric{name: "bob_merge_duration"}
	SiteCacheDuration = &histogramMetric{name: "sites_cache_update"}
)

// counters
var (
	DataPointReceived           = &counterMetric{name: "data_point_received"}
	DataPointAccepted           = &counterMetric{name: "data_point_accepted"}
	DataPointRejected           = &counterMetric{name: "data_point_rejected"}
	DataPointRejectedBadRequest = &counterMetric{name: "data_point_rejected_bad_request"}
	DataPointDropped            = &counterMetric{name: "data_point_dropped"}
)
