package snowflake

import (
	"errors"
	"sync/atomic"
	"time"
)

// These constants are the bit lengths of snowflake ID parts.
const (
	TimestampLength uint8  = 43
	MachineIDLength uint8  = 9
	SequenceLength  uint8  = 12
	MaxSequence     uint16 = 1<<SequenceLength - 1
	MaxTimestamp    uint64 = 1<<TimestampLength - 1
	MaxMachineID    uint16 = 1<<MachineIDLength - 1

	machineIDMoveLength = SequenceLength
	timestampMoveLength = MachineIDLength + SequenceLength
)

// SequenceResolver the snowflake sequence resolver.
//
// When you want to use the snowflake algorithm to generate unique ID, You must ensure: The sequence-number generated in the same millisecond of the same node is unique.
// Based on this, we create this interface provide following resolver:
//
//	AtomicResolver : base sync/atomic (by default).
type SequenceResolver func(ms int64) (uint16, error)

// default start time is 2008-11-10 23:00:00 UTC, why ? In the playground the time begins at 2009-11-10 23:00:00 UTC.
// It can run on golang playground.
// default machineID is 0
// default resolver is AtomicResolver
var (
	resolver      SequenceResolver
	machineID     uint64 = 0
	startTime            = time.Date(2008, 11, 10, 23, 0, 0, 0, time.UTC)
	lastTimestamp int64  = 0 // 👈 新增：记录上一次生成 ID 的毫秒时间（相对于 Unix）
)

// ID use ID to generate snowflake id, and it will ignore error. if you want error info, you need use NextID method.
// This function is thread safe.
func ID() uint64 {
	id, _ := NextID()
	return id
}

// NextID use NextID to generate snowflake id and return an error.
// This function is thread safe.
// NextID use NextID to generate snowflake id and return an error.
// This function is thread safe.
func NextID() (uint64, error) {
	now := currentMillis()
	last := atomic.LoadInt64(&lastTimestamp)

	// ⏰ 时钟回拨检测
	if now < last {
		backward := last - now
		// 🛡️ 最大容忍回拨：5000 毫秒（5秒）
		if backward > 5000 {
			return 0, errors.New("clock moved backward too much (>5s), refusing to generate ID")
		}
		// 在容忍范围内，等待时间追上
		time.Sleep(time.Duration(backward) * time.Millisecond)
		now = currentMillis()
	}

	// 获取序列号
	seqResolver := callSequenceResolver()
	seq, err := seqResolver(now)
	if err != nil {
		return 0, err
	}

	// 序列号溢出：等待下一毫秒
	for seq >= MaxSequence {
		now = waitForNextMillis(now)
		seq, err = seqResolver(now)
		if err != nil {
			return 0, err
		}
	}

	// 更新 lastTimestamp（必须在生成 ID 前完成）
	atomic.StoreInt64(&lastTimestamp, now)

	// 计算相对于 startTime 的偏移
	df := elapsedTime(now, startTime)
	if df < 0 || uint64(df) > MaxTimestamp {
		return 0, errors.New("the maximum life cycle of the snowflake algorithm is 2^43-1(millis), please check start-time")
	}

	id := (uint64(df) << uint64(timestampMoveLength)) |
		(machineID << uint64(machineIDMoveLength)) |
		uint64(seq)
	return id, nil
}

// SetStartTime set the start time for snowflake algorithm.
//
// It will panic when:
//
//	s IsZero
//	s > current millisecond,
//	current millisecond - s > 2^41(69 years).
//
// This function is thread-unsafe, recommended you call him in the main function.
func SetStartTime(s time.Time) {
	s = s.UTC()

	if s.IsZero() {
		panic("The start time cannot be a zero value")
	}

	if s.After(time.Now().UTC()) {
		panic("The s cannot be greater than the current millisecond")
	}

	// since we check the current millisecond is greater than s, so we don't need to check the overflow.
	df := elapsedTime(currentMillis(), s)
	if uint64(df) > MaxTimestamp {
		panic("The maximum life cycle of the snowflake algorithm is 279 years")
	}

	startTime = s
}

// SetMachineID specify the machine ID. It will panic when machined > max limit for 2^9-1.
// This function is thread-unsafe, recommended you call him in the main function.
func SetMachineID(m uint16) {
	if m > MaxMachineID {
		panic("The machineID cannot be greater than 511")
	}
	machineID = uint64(m)
}

// SetSequenceResolver set a custom sequence resolver.
// This function is thread-unsafe, recommended you call him in the main function.
func SetSequenceResolver(seq SequenceResolver) {
	if seq != nil {
		resolver = seq
	}
}

// SID snowflake id
type SID struct {
	Sequence  uint64
	MachineID uint64
	Timestamp uint64
	ID        uint64
}

// GenerateTime snowflake generate at, return a UTC time.
func (id *SID) GenerateTime() time.Time {
	ms := startTime.UTC().UnixNano()/1e6 + int64(id.Timestamp)

	return time.Unix(0, ms*int64(time.Millisecond)).UTC()
}

// ParseID parse snowflake it to SID struct.
func ParseID(id uint64) SID {
	t := id >> uint64(SequenceLength+MachineIDLength)
	sequence := id & uint64(MaxSequence)
	mID := (id & (uint64(MaxMachineID) << SequenceLength)) >> SequenceLength

	return SID{
		ID:        id,
		Sequence:  sequence,
		MachineID: mID,
		Timestamp: t,
	}
}

//--------------------------------------------------------------------
// private function defined.
//--------------------------------------------------------------------

func waitForNextMillis(last int64) int64 {
	for {
		now := currentMillis()
		if now > last {
			return now
		}
		// 避免 CPU 空转，微小休眠
		time.Sleep(1 * time.Nanosecond)
	}
}
func callSequenceResolver() SequenceResolver {
	if resolver == nil {
		return AtomicResolver
	}

	return resolver
}

func elapsedTime(noms int64, s time.Time) int64 {
	return noms - s.UTC().UnixNano()/1e6
}

// currentMillis get current millisecond.
func currentMillis() int64 {
	return time.Now().UTC().UnixNano() / 1e6
}
