package hit_tracker

import (
	"context"
	"time"

	"github.com/buildbuddy-io/buildbuddy/server/environment"
	"github.com/buildbuddy-io/buildbuddy/server/interfaces"
	"github.com/buildbuddy-io/buildbuddy/server/metrics"
	"github.com/buildbuddy-io/buildbuddy/server/tables"
	"github.com/buildbuddy-io/buildbuddy/server/util/bazel_request"
	"github.com/buildbuddy-io/buildbuddy/server/util/log"
	"github.com/prometheus/client_golang/prometheus"

	capb "github.com/buildbuddy-io/buildbuddy/proto/cache"
	repb "github.com/buildbuddy-io/buildbuddy/proto/remote_execution"
)

type CacheMode int
type counterType int

const (
	// Prometheus CacheEventTypeLabel values

	hitLabel    = "hit"
	missLabel   = "miss"
	uploadLabel = "upload"

	// Prometheus CacheTypeLabel values

	actionCacheLabel = "action_cache"
	casLabel         = "cas"

	CAS         CacheMode = iota // CAS cache
	ActionCache                  // Action cache

	Hit counterType = iota
	Miss
	Upload

	DownloadSizeBytes
	UploadSizeBytes

	DownloadUsec
	UploadUsec

	CachedActionExecUsec

	// New counter types go here!
)

func cacheTypePrefix(actionCache bool, name string) string {
	if actionCache {
		return "action-cache-" + name
	} else {
		return "cas-" + name
	}
}

func counterKey(iid string) string {
	return "hit_tracker/" + iid
}

func counterField(actionCache bool, ct counterType) string {
	switch ct {
	case Hit:
		return cacheTypePrefix(actionCache, "hits")
	case Miss:
		return cacheTypePrefix(actionCache, "misses")
	case Upload:
		return cacheTypePrefix(actionCache, "uploads")
	case DownloadSizeBytes:
		return "download-size-bytes"
	case UploadSizeBytes:
		return "upload-size-bytes"
	case DownloadUsec:
		return "download-usec"
	case UploadUsec:
		return "upload-usec"
	case CachedActionExecUsec:
		return "cached-action-exec-usec"
	default:
		return "UNKNOWN-COUNTER-TYPE"
	}
}

type HitTracker struct {
	c           interfaces.MetricsCollector
	usage       interfaces.UsageTracker
	ctx         context.Context
	iid         string
	actionCache bool
}

func NewHitTracker(ctx context.Context, env environment.Env, actionCache bool) *HitTracker {
	return &HitTracker{
		c:           env.GetMetricsCollector(),
		usage:       env.GetUsageTracker(),
		ctx:         ctx,
		iid:         bazel_request.GetInvocationID(ctx),
		actionCache: actionCache,
	}
}

func (h *HitTracker) counterKey() string {
	return counterKey(h.iid)
}

func (h *HitTracker) counterField(ct counterType) string {
	return counterField(h.actionCache, ct)
}

func (h *HitTracker) cacheTypeLabel() string {
	if h.actionCache {
		return actionCacheLabel
	}
	return casLabel
}

// Example Usage:
//
// ht := NewHitTracker(env, invocationID, false /*=actionCache*/)
// if err := ht.TrackMiss(); err != nil {
//   log.Printf("Error counting cache miss.")
// }
func (h *HitTracker) TrackMiss(d *repb.Digest) error {
	if h.c == nil || h.iid == "" {
		return nil
	}
	metrics.CacheEvents.With(prometheus.Labels{
		metrics.CacheTypeLabel:      h.cacheTypeLabel(),
		metrics.CacheEventTypeLabel: missLabel,
	}).Inc()
	return h.c.IncrementCount(h.ctx, h.counterKey(), h.counterField(Miss), 1)
}

func (h *HitTracker) TrackEmptyHit() error {
	metrics.CacheEvents.With(prometheus.Labels{
		metrics.CacheTypeLabel:      h.cacheTypeLabel(),
		metrics.CacheEventTypeLabel: hitLabel,
	}).Inc()
	if h.c == nil || h.iid == "" {
		return nil
	}
	return h.c.IncrementCount(h.ctx, h.counterKey(), h.counterField(Hit), 1)
}

type closeFunction func() error

type transferTimer struct {
	closeFn closeFunction
}

func (t *transferTimer) Close() error {
	return t.closeFn()
}

func cacheEventTypeLabel(c counterType) string {
	if c == Hit {
		return hitLabel
	}
	if c == Miss {
		return missLabel
	}
	return uploadLabel
}

func sizeMetric(ct counterType) *prometheus.HistogramVec {
	if ct == UploadSizeBytes {
		return metrics.CacheUploadSizeBytes
	}
	return metrics.CacheDownloadSizeBytes
}

func durationMetric(ct counterType) *prometheus.HistogramVec {
	if ct == UploadUsec {
		return metrics.CacheUploadDurationUsec
	}
	return metrics.CacheDownloadDurationUsec
}

func (h *HitTracker) makeCloseFunc(d *repb.Digest, start time.Time, actionCounter, sizeCounter, timeCounter counterType) closeFunction {
	return func() error {
		dur := time.Since(start)

		et := cacheEventTypeLabel(actionCounter)
		ct := h.cacheTypeLabel()
		metrics.CacheEvents.With(prometheus.Labels{
			metrics.CacheTypeLabel:      ct,
			metrics.CacheEventTypeLabel: et,
		}).Inc()
		sizeMetric(sizeCounter).With(prometheus.Labels{
			metrics.CacheTypeLabel: ct,
		}).Observe(float64(d.GetSizeBytes()))
		durationMetric(timeCounter).With(prometheus.Labels{
			metrics.CacheTypeLabel: ct,
		}).Observe(float64(dur.Microseconds()))

		if err := h.recordCacheUsage(d, actionCounter); err != nil {
			return err
		}

		if h.c == nil || h.iid == "" {
			return nil
		}

		if err := h.c.IncrementCount(h.ctx, h.counterKey(), h.counterField(actionCounter), 1); err != nil {
			return err
		}
		if err := h.c.IncrementCount(h.ctx, h.counterKey(), h.counterField(sizeCounter), d.GetSizeBytes()); err != nil {
			return err
		}
		if err := h.c.IncrementCount(h.ctx, h.counterKey(), h.counterField(timeCounter), dur.Microseconds()); err != nil {
			return err
		}

		return nil
	}
}

// Example Usage:
//
// ht := NewHitTracker(env, invocationID, false /*=actionCache*/)
// dlt := ht.TrackDownload(d)
// defer dlt.Close()
// ... body of download logic ...
func (h *HitTracker) TrackDownload(d *repb.Digest) *transferTimer {
	start := time.Now()
	return &transferTimer{
		closeFn: h.makeCloseFunc(d, start, Hit, DownloadSizeBytes, DownloadUsec),
	}
}

// Example Usage:
//
// ht := NewHitTracker(env, invocationID, false /*=actionCache*/)
// ult := ht.TrackUpload(d)
// defer ult.Close()
// ... body of download logic ...
func (h *HitTracker) TrackUpload(d *repb.Digest) *transferTimer {
	start := time.Now()
	return &transferTimer{
		closeFn: h.makeCloseFunc(d, start, Upload, UploadSizeBytes, UploadUsec),
	}
}

func (h *HitTracker) recordCacheUsage(d *repb.Digest, actionCounter counterType) error {
	if h.usage == nil || actionCounter != Hit {
		return nil
	}
	c := &tables.UsageCounts{
		TotalDownloadSizeBytes: d.GetSizeBytes(),
	}
	if h.actionCache {
		c.ActionCacheHits = 1
	} else {
		c.CASCacheHits = 1
	}
	return h.usage.Increment(h.ctx, c)
}

func computeThroughputBytesPerSecond(sizeBytes, durationUsec int64) int64 {
	if durationUsec == 0 {
		return 0
	}
	dur := time.Duration(durationUsec * int64(time.Microsecond))
	return int64(float64(sizeBytes) / dur.Seconds())
}

func CollectCacheStats(ctx context.Context, env environment.Env, iid string) *capb.CacheStats {
	c := env.GetMetricsCollector()
	if c == nil || iid == "" {
		return nil
	}
	cs := &capb.CacheStats{}

	counts, err := c.ReadCounts(ctx, counterKey(iid))
	if err != nil {
		log.Warningf("Failed to collect cache stats: %s", err)
		return cs
	}
	if counts == nil {
		counts = make(map[string]int64)
	}

	cs.ActionCacheHits = counts[counterField(true, Hit)]
	cs.ActionCacheMisses = counts[counterField(true, Miss)]
	cs.ActionCacheUploads = counts[counterField(true, Upload)]

	cs.CasCacheHits = counts[counterField(false, Hit)]
	cs.CasCacheMisses = counts[counterField(false, Miss)]
	cs.CasCacheUploads = counts[counterField(false, Upload)]

	cs.TotalDownloadSizeBytes = counts[counterField(false, DownloadSizeBytes)]
	cs.TotalUploadSizeBytes = counts[counterField(false, UploadSizeBytes)]
	cs.TotalDownloadUsec = counts[counterField(false, DownloadUsec)]
	cs.TotalUploadUsec = counts[counterField(false, UploadUsec)]

	cs.DownloadThroughputBytesPerSecond = computeThroughputBytesPerSecond(cs.TotalDownloadSizeBytes, cs.TotalDownloadUsec)
	cs.UploadThroughputBytesPerSecond = computeThroughputBytesPerSecond(cs.TotalUploadSizeBytes, cs.TotalUploadUsec)

	cs.TotalCachedActionExecUsec = counts[counterField(false, CachedActionExecUsec)]

	return cs
}

func CleanupCacheStats(ctx context.Context, env environment.Env, iid string) {
	c := env.GetMetricsCollector()
	if c == nil || iid == "" {
		return
	}

	if err := c.Delete(ctx, counterKey(iid)); err != nil {
		log.Warningf("Failed to clean up cache stats: %s", err)
	}
}
