package timedb

import (
    "bytes"
    "io"
    "time"
)

type Entry struct {
    Value float64
    Time time.Time
}

type blockData [][]Entry

func (self blockData) blockBase() float64 {
    sum := float64(0.0)
    for _, entries := range self {
        sum += entries[0].Value
    }
    return sum / float64(len(self))
}

type Block struct {
    buckets [][]byte
    Baseline float64
    Precision int
}

func NewBlock(start time.Time, precision int, values [][]Entry) *Block {
    data := blockData(values)
    block := &Block{
        buckets: make([][]byte, len(values)),
        Baseline: data.blockBase(),
        Precision: precision,
    }
    for i, entries := range values {
        buffer := &bytes.Buffer{}
        enc := block.createBucketEncoder(buffer)
        for _, entry := range entries {
            enc.WriteFloat64(entry.Value)
        }
        enc.Close()
        block.buckets[i] = make([]byte, len(buffer.Bytes()))
        copy(block.buckets[i], buffer.Bytes())
    }
    return block
}

func (self *Block) createBucketEncoder(out io.Writer) *bucketEncoder {
    return newBucketEncoder(self.Baseline, self.Precision, out)
}