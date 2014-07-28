package timedb

import (
	"bytes"
	"code.google.com/p/go-uuid/uuid"
	"github.com/FlukeNetworks/timedb/bucket"
	"time"
)

type bucketStoreContext struct {
	buffer, lastBuffer *bytes.Buffer
	encoder            *bucket.BucketEncoder
}

// Destroys the back buffer, replaces it with the front buffer,
// then creates a new front buffer
func (self *bucketStoreContext) swapBuffers() {
	self.lastBuffer = self.buffer
	self.buffer = &bytes.Buffer{}
}

type BucketRepository interface {
	Put(contexts map[string]*bucketStoreContext, store *BucketStore) error
}

type BucketStore struct {
	Duration   time.Duration
	Multiplier float64
	Repository BucketRepository
	contexts   map[string]map[string]*bucketStoreContext
	endTimes   map[string]*time.Time
}

func (self *BucketStore) Insert(series uuid.UUID, entry Entry) error {
	var err error
	seriesStr := series.String()
	contexts := self.contexts[seriesStr]
	if contexts == nil {
		contexts = map[string]*bucketStoreContext{}
		self.contexts[seriesStr] = contexts
	}
	if self.endTimes[seriesStr] == nil {
		tmp := entry.Timestamp.Truncate(self.Duration)
		self.endTimes[seriesStr] = &tmp
	} else if entry.Timestamp.After(*self.endTimes[seriesStr]) {
		for _, ctx := range contexts {
			ctx.swapBuffers()
			ctx.encoder.Close()
		}
		err = self.Repository.Put(contexts, self)
		self.endTimes[seriesStr] = nil
	}
	// Write all attributes to their encoders
	for k, v := range entry.Attributes {
		ctx := contexts[k]
		if ctx == nil {
			ctx = &bucketStoreContext{
				buffer: &bytes.Buffer{},
			}
			ctx.encoder = bucket.NewBucketEncoder(marshalFloat64(entry.Attributes[k], self.Multiplier), ctx.buffer)
			contexts[k] = ctx
		}
		contexts[k].encoder.WriteInt(marshalFloat64(v, self.Multiplier))
	}
	return err
}

func marshalFloat64(v float64, multiplier float64) int64 {
	return int64(v * multiplier)
}