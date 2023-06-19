package protopool_test

import (
	"math/rand"
	"sync"
	"testing"

	"github.com/jussi-kalliokoski/protopool"
	"github.com/jussi-kalliokoski/protopool/internal/enum"
)

func Test(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		t.Run("Reset", func(t *testing.T) {
			var v protopool.Value
			resettable := protopool.New[Resettable](&v)
			resettable.Value = 123
			plain := protopool.New[Plain](&v)
			plain.Value = 234

			v.Reset()
			resettable = protopool.New[Resettable](&v)
			plain = protopool.New[Plain](&v)

			assertEqual(t, 0, plain.Value)
			assertEqual(t, 0, resettable.Value)
			assertEqual(t, 1, resettable.ResetCount)
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("Reset", func(t *testing.T) {
			var v protopool.Value
			s1 := []int64(nil)
			l1 := protopool.NewList(&v, &s1)
			l1.Append(1, 2, 3, 4)
			s2 := []int64(nil)
			l2 := protopool.NewList(&v, &s2)
			l2.Append(4, 3)

			v.Reset()
			l1 = protopool.NewList(&v, &s1)
			l1.Append(9, 8)
			l2 = protopool.NewList(&v, &s2)
			l2.Append(5, 4, 3, 2)

			assertSlicesEqual(t, []int64{9, 8}, s1)
			assertSlicesEqual(t, []int64{5, 4, 3, 2}, s2)
		})
	})

	t.Run("Put", func(t *testing.T) {
		t.Run("Reset", func(t *testing.T) {
			var v protopool.Value
			var m1 map[int32]uint64
			protopool.Put(&v, &m1, 1, 2)
			var m2 map[int32]uint64
			protopool.Put(&v, &m2, 3, 4)

			v.Reset()
			m1 = nil
			protopool.Put(&v, &m1, 8, 9)
			protopool.Put(&v, &m1, 5, 6)
			m2 = nil
			protopool.Put(&v, &m2, 7, 8)
			protopool.Put(&v, &m2, 4, 5)

			assertMapsEqual(t, map[int32]uint64{8: 9, 5: 6}, m1)
			assertMapsEqual(t, map[int32]uint64{7: 8, 4: 5}, m2)
		})
	})
}

func Benchmark(b *testing.B) {
	var p sync.Pool
	src := rand.NewSource(1)
	rng := rand.New(src)
	heapAlloc := &HeapAllocator{}
	heapBuilder := &MessageBuilder{alloc: heapAlloc, rng: rng}

	b.Run("without pool", func(b *testing.B) {
		b.ReportAllocs()
		for n := 0; n < b.N; n++ {
			src.Seed(1)
			_ = heapBuilder.BuildMessage()
		}
	})

	b.Run("with pool", func(b *testing.B) {
		b.ReportAllocs()
		for n := 0; n < b.N; n++ {
			src.Seed(1)
			poolBuilder, _ := p.Get().(*MessageBuilder)
			if poolBuilder == nil {
				poolBuilder = &MessageBuilder{
					alloc: &PoolAllocator{},
					rng:   rng,
				}
			}
			_ = poolBuilder.BuildMessage()
			poolBuilder.alloc.(*PoolAllocator).v.Reset()
			p.Put(poolBuilder)
		}
	})
}

type Resettable struct {
	ResetCount int64
	Value      int64
}

func (r *Resettable) Reset() {
	*r = Resettable{ResetCount: r.ResetCount + 1}
}

type Plain struct {
	Value int64
}

type MessageBuilder struct {
	alloc Allocator
	rng   *rand.Rand
}

func (b *MessageBuilder) BuildMessage() *Message {
	msg := PopulatePointer[Message](b.alloc)
	msg.Timestamp = b.randomI64()
	msg.CorrelationID = b.randomString()
	msg.Request = b.buildRequest()
	return msg
}

func (b *MessageBuilder) buildRequest() *Request {
	msg := PopulatePointer[Request](b.alloc)
	msg.Context = b.buildContext()
	PopulateSlice(b.alloc, &msg.Targets, nTargets, b.buildTarget)
	PopulateSlice(b.alloc, &msg.Placements, nChildren, b.buildPlacement)
	msg.DefaultPlacement = b.buildPlacement()
	return msg
}

func (b *MessageBuilder) buildContext() *Context {
	msg := PopulatePointer[Context](b.alloc)
	msg.Server = b.buildServer()
	msg.Client = b.buildClient()
	msg.Location = b.buildLocation()
	msg.Company = b.buildCompany()
	PopulateSlice(b.alloc, &msg.Tokens, nChildren, b.randomU32)
	msg.OS = b.randomString()
	msg.TransactionID = b.randomString()
	msg.MachineID = b.randomString()
	msg.OriginalTimestamp = b.randomI64()
	msg.CompletionTimestamp = b.randomI64()
	msg.OrderNumber = b.randomU32()
	msg.TargetGroup = b.randomI64()
	msg.DestinationGroup = b.randomI64()
	msg.IntermediationGroup = b.randomI64()
	msg.IsReady = b.randomBool()
	msg.IsAvailable = b.randomBool()
	PopulateEnum(b.rng, &msg.RequestType)
	PopulateEnum(b.rng, &msg.RequestReason)
	PopulateEnum(b.rng, &msg.ResponseType)
	return msg
}

func (b *MessageBuilder) buildServer() *Server {
	msg := PopulatePointer[Server](b.alloc)
	msg.ID = b.randomString()
	msg.HasCluster = b.randomBool()
	msg.HasPersistentCluster = b.randomBool()
	msg.HasConnectionLimit = b.randomBool()
	msg.RestartCount = b.randomI32()
	msg.RestartCount24H = b.randomI32()
	msg.RestartCount7D = b.randomI32()
	msg.RebootCount = b.randomI32()
	msg.RebootCount24H = b.randomI32()
	msg.RebootCount7D = b.randomI32()
	msg.ReplicaCount = b.randomI32()
	msg.ReplicaCount24H = b.randomI32()
	msg.ReplicaCount7D = b.randomI32()
	msg.HealthyReplicaCount = b.randomI32()
	msg.HealthyReplicaCount24H = b.randomI32()
	msg.HealthyReplicaCount7D = b.randomI32()
	msg.ConnectionCount = b.randomI32()
	msg.ConnectionCount24H = b.randomI32()
	msg.ConnectionCount7D = b.randomI32()
	msg.CreationTimestamp = b.randomI64()
	PopulateSlice(b.alloc, &msg.ClusterIDs, nChildren, b.randomString)
	msg.ClusterCounters = b.buildClusterCounters()
	PopulateMap2(b.alloc, &msg.ReplicaCounters, nChildren, b.randomString, b.buildReplicaCounters)
	PopulateEnum(b.rng, &msg.TokenScope)
	PopulateSlice(b.alloc, &msg.Tokens, nChildren, b.randomString)
	PopulateSlice(b.alloc, &msg.ReplicaIDs, nChildren, b.randomString)
	PopulateSlice(b.alloc, &msg.HealthyReplicaIDs, nChildren, b.randomString)
	PopulateSlice(b.alloc, &msg.ReplicaLatestRestartTimestamps, nChildren, b.randomI64)
	PopulateSlice(b.alloc, &msg.UnhealthyReplicaIDs, nChildren, b.randomString)
	msg.OriginalID = b.randomString()
	return msg
}

func (b *MessageBuilder) buildClusterCounters() *ClusterCounters {
	msg := PopulatePointer[ClusterCounters](b.alloc)
	msg.SeedsCount = b.randomI32()
	msg.SeedsCount24H = b.randomI32()
	msg.SeedsCount7D = b.randomI32()
	msg.TreesCount = b.randomI32()
	msg.BranchesCount = b.randomI32()
	PopulateMap2(b.alloc, &msg.SeedsByBranch, nChildren, b.randomI32, b.randomI32)
	PopulateMap2(b.alloc, &msg.BranchesByTree, nChildren, b.randomI32, b.randomI32)
	return msg
}

func (b *MessageBuilder) buildReplicaCounters() *ReplicaCounters {
	msg := PopulatePointer[ReplicaCounters](b.alloc)
	msg.RebootCount = b.randomI32()
	msg.RebootCount24H = b.randomI32()
	msg.RebootCount7D = b.randomI32()
	msg.RestartCount = b.randomI32()
	msg.RestartCount24H = b.randomI32()
	msg.RestartCount7D = b.randomI32()
	return msg
}

func (b *MessageBuilder) buildClient() *Client {
	msg := PopulatePointer[Client](b.alloc)
	msg.SDKVersion = b.randomString()
	msg.ID = b.randomString()
	msg.ConnectionType = b.randomString()
	msg.Type = b.randomString()
	msg.GPUCount = b.randomI64()
	msg.DPI = b.randomI64()
	msg.Resolution = b.randomString()
	msg.GPU = b.randomString()
	msg.CPU = b.randomString()
	msg.CPUCount = b.randomI64()
	PopulateSlice(b.alloc, &msg.Peers, nChildren, b.randomString)
	msg.Mass = b.randomU64()
	msg.SoftwareVersion = b.randomString()
	msg.AcceptLanguage = b.randomString()
	msg.Language = b.randomString()
	return msg
}

func (b *MessageBuilder) buildLocation() *Location {
	msg := PopulatePointer[Location](b.alloc)
	msg.Country = b.randomString()
	msg.City = b.randomString()
	msg.CityID = b.randomI64()
	msg.SubCountry = b.randomString()
	msg.SubCountryID = b.randomI64()
	msg.Timezone = b.randomString()
	return msg
}

func (b *MessageBuilder) buildCompany() *Company {
	msg := PopulatePointer[Company](b.alloc)
	msg.CustomerID = b.randomI64()
	msg.IsLarge = b.randomBool()
	msg.BillingID = b.randomString()
	msg.ProjectID = b.randomI64()
	msg.IsActive = b.randomBool()
	return msg
}

func (b *MessageBuilder) buildTarget() *Target {
	msg := PopulatePointer[Target](b.alloc)
	msg.ID = b.randomI64()
	msg.PatchID = b.randomString()
	PopulateSlice(b.alloc, &msg.Peers, nChildren, b.buildPeer)
	msg.RestartCount = b.randomI32()
	msg.RestartCount24H = b.randomI32()
	msg.RestartCount7D = b.randomI32()
	msg.RebootCount = b.randomI32()
	msg.RebootCount24H = b.randomI32()
	msg.RebootCount7D = b.randomI32()
	msg.ReplicaCount = b.randomI32()
	msg.ReplicaCount24H = b.randomI32()
	msg.ReplicaCount7D = b.randomI32()
	msg.ConnectionCount = b.randomI32()
	msg.ConnectionCount24H = b.randomI32()
	msg.ConnectionCount7D = b.randomI32()
	msg.CreationTimestamp = b.randomI64()
	return msg
}

func (b *MessageBuilder) buildPeer() *Peer {
	msg := PopulatePointer[Peer](b.alloc)
	msg.ID = b.randomString()
	msg.RestartCount = b.randomI32()
	msg.RestartCount24H = b.randomI32()
	msg.RestartCount7D = b.randomI32()
	msg.RebootCount = b.randomI32()
	msg.RebootCount24H = b.randomI32()
	msg.RebootCount7D = b.randomI32()
	msg.ReplicaCount = b.randomI32()
	msg.ReplicaCount24H = b.randomI32()
	msg.ReplicaCount7D = b.randomI32()
	msg.ConnectionCount = b.randomI32()
	msg.ConnectionCount24H = b.randomI32()
	msg.ConnectionCount7D = b.randomI32()
	msg.AverageRestartCount = b.randomI32()
	msg.AverageRestartCount24H = b.randomI32()
	msg.AverageRestartCount7D = b.randomI32()
	msg.AverageRebootCount = b.randomI32()
	msg.AverageRebootCount24H = b.randomI32()
	msg.AverageRebootCount7D = b.randomI32()
	msg.AverageReplicaCount = b.randomI32()
	msg.AverageReplicaCount24H = b.randomI32()
	msg.AverageReplicaCount7D = b.randomI32()
	msg.AverageConnectionCount = b.randomI32()
	msg.AverageConnectionCount24H = b.randomI32()
	msg.AverageConnectionCount7D = b.randomI32()
	msg.MedianRestartCount = b.randomI32()
	msg.MedianRestartCount24H = b.randomI32()
	msg.MedianRestartCount7D = b.randomI32()
	msg.MedianRebootCount = b.randomI32()
	msg.MedianRebootCount24H = b.randomI32()
	msg.MedianRebootCount7D = b.randomI32()
	msg.MedianReplicaCount = b.randomI32()
	msg.MedianReplicaCount24H = b.randomI32()
	msg.MedianReplicaCount7D = b.randomI32()
	msg.MedianConnectionCount = b.randomI32()
	msg.MedianConnectionCount24H = b.randomI32()
	msg.MedianConnectionCount7D = b.randomI32()
	msg.IsReady = b.randomBool()
	msg.IsAvailable = b.randomBool()
	msg.IsPossible = b.randomBool()
	// msg.GlobalAverageCounters = b.buildUptimeCounters()
	// msg.GlobalMedianCounters = b.buildUptimeCounters()
	return msg
}

func (b *MessageBuilder) buildUptimeCounters() *UptimeCounters { //nolint: unused
	msg := PopulatePointer[UptimeCounters](b.alloc)
	msg.RebootCount = b.randomI32()
	msg.RestartCount = b.randomI32()
	return msg
}

func (b *MessageBuilder) buildPlacement() *Placement {
	msg := PopulatePointer[Placement](b.alloc)
	msg.ID = b.randomString()
	msg.Toleration = b.randomF64()
	msg.TolerationDelta = b.randomF64()
	return msg
}

func (b *MessageBuilder) randomI64() int64 {
	return b.rng.Int63()
}

func (b *MessageBuilder) randomI32() int32 {
	return b.rng.Int31()
}

func (b *MessageBuilder) randomU64() uint64 {
	return b.rng.Uint64()
}

func (b *MessageBuilder) randomU32() uint32 {
	return b.rng.Uint32()
}

func (b *MessageBuilder) randomF64() float64 {
	return b.rng.Float64()
}

func (b *MessageBuilder) randomString() string {
	s := randomStringBuffer[b.rng.Intn(len(randomStringBuffer)-2):]
	return s[:rand.Intn(len(s)-1)+1]
}

var randomStringBuffer = func() string {
	rng := rand.New(rand.NewSource(1))
	dst := make([]byte, 1024)
	for i := range dst {
		if v := byte(rng.Intn(16)); v < 10 {
			dst[i] = '0' + v
		} else {
			dst[i] = 'a' + v
		}
	}
	return string(dst)
}()

func (b *MessageBuilder) randomBool() bool {
	return b.rng.Intn(2) == 0
}

type Allocator interface{}

type Message struct {
	Timestamp     int64    `json:"timestamp"`
	CorrelationID string   `json:"correlation_id"`
	Request       *Request `json:"request"`
}

func (v *Message) Reset() {
	*v = Message{}
}

type Request struct {
	Context          *Context     `json:"context"`
	Targets          []*Target    `json:"targets"`
	Placements       []*Placement `json:"placements"`
	DefaultPlacement *Placement   `json:"default_placement"`
}

func (v *Request) Reset() {
	*v = Request{}
}

type Context struct {
	Server              *Server       `json:"server"`
	Client              *Client       `json:"client"`
	Location            *Location     `json:"location"`
	Company             *Company      `json:"company"`
	Tokens              []uint32      `json:"tokens"`
	OS                  string        `json:"os"`
	TransactionID       string        `json:"transaction_id"`
	MachineID           string        `json:"machine_id"`
	OriginalTimestamp   int64         `json:"original_timestamp"`
	CompletionTimestamp int64         `json:"completion_timestamp"`
	OrderNumber         uint32        `json:"order_number"`
	TargetGroup         int64         `json:"target_group"`
	DestinationGroup    int64         `json:"destination_group"`
	IntermediationGroup int64         `json:"intermediation_group"`
	IsReady             bool          `json:"is_ready"`
	IsAvailable         bool          `json:"is_available"`
	RequestType         RequestType   `json:"request_type"`
	RequestReason       RequestReason `json:"request_reason"`
	ResponseType        ResponseType  `json:"response_type"`
}

type Server struct {
	ID                             string                      `json:"string"`
	HasCluster                     bool                        `json:"has_cluster"`
	HasPersistentCluster           bool                        `json:"has_persistent_cluster"`
	HasConnectionLimit             bool                        `json:"has_connection_limit"`
	RestartCount                   int32                       `json:"restart_count"`
	RestartCount24H                int32                       `json:"restart_count_24h"`
	RestartCount7D                 int32                       `json:"restart_count_7d"`
	RebootCount                    int32                       `json:"reboot_count"`
	RebootCount24H                 int32                       `json:"reboot_count_24h"`
	RebootCount7D                  int32                       `json:"reboot_count_7d"`
	ReplicaCount                   int32                       `json:"replica_count"`
	ReplicaCount24H                int32                       `json:"replica_count_24h"`
	ReplicaCount7D                 int32                       `json:"replica_count_7d"`
	HealthyReplicaCount            int32                       `json:"healthy_replica_count"`
	HealthyReplicaCount24H         int32                       `json:"healthy_replica_count_24h"`
	HealthyReplicaCount7D          int32                       `json:"healthy_replica_count_7d"`
	ConnectionCount                int32                       `json:"connection_count"`
	ConnectionCount24H             int32                       `json:"connection_count_24h"`
	ConnectionCount7D              int32                       `json:"connection_count_7d"`
	CreationTimestamp              int64                       `json:"creation_timestamp"`
	ClusterIDs                     []string                    `json:"cluster_ids"`
	HasCrashed                     bool                        `json:"has_crashed"`
	ClusterCounters                *ClusterCounters            `json:"cluster_counters"`
	ReplicaCounters                map[string]*ReplicaCounters `json:"replica_counters"`
	TokenScope                     TokenScope                  `json:"token_scope"`
	Tokens                         []string                    `json:"tokens"`
	ReplicaIDs                     []string                    `json:"replica_ids"`
	HealthyReplicaIDs              []string                    `json:"healthy_replica_ids"`
	ReplicaLatestRestartTimestamps []int64                     `json:"replica_latest_restart_timestamps"`
	UnhealthyReplicaIDs            []string                    `json:"unhealthy_replica_ids"`
	OriginalID                     string                      `json:"original_id"`
}

func (v *Server) Reset() {
	*v = Server{}
}

type ClusterCounters struct {
	SeedsCount     int32           `json:"seeds_count"`
	SeedsCount24H  int32           `json:"seeds_count_24h"`
	SeedsCount7D   int32           `json:"seeds_count_7d"`
	TreesCount     int32           `json:"trees_count"`
	BranchesCount  int32           `json:"branches_count"`
	SeedsByBranch  map[int32]int32 `json:"seeds_by_branch"`
	BranchesByTree map[int32]int32 `json:"branches_by_tree"`
}

type ReplicaCounters struct {
	RebootCount     int32 `json:"reboot_count"`
	RebootCount24H  int32 `json:"reboot_count_24h"`
	RebootCount7D   int32 `json:"reboot_count_7d"`
	RestartCount    int32 `json:"restart_count"`
	RestartCount24H int32 `json:"restart_count_24h"`
	RestartCount7D  int32 `json:"restart_count_7d"`
}

type Client struct {
	SDKVersion      string   `json:"sdk_version"`
	ID              string   `json:"id"`
	ConnectionType  string   `json:"connection_type"`
	Type            string   `json:"type"`
	GPUCount        int64    `json:"gpu_count"`
	DPI             int64    `json:"dpi"`
	Resolution      string   `json:"resolution"`
	GPU             string   `json:"gpu"`
	CPU             string   `json:"cpu"`
	CPUCount        int64    `json:"cpu_count"`
	Peers           []string `json:"peers"`
	Mass            uint64   `json:"mass"`
	SoftwareVersion string   `json:"software_version"`
	AcceptLanguage  string   `json:"accept_language"`
	Language        string   `json:"language"`
}

func (v *Client) Reset() {
	*v = Client{}
}

type Location struct {
	Country      string `json:"country"`
	City         string `json:"city"`
	CityID       int64  `json:"city_id"`
	SubCountry   string `json:"sub_country"`
	SubCountryID int64  `json:"sub_country_id"`
	Timezone     string `json:"timezone"`
}

func (v *Location) Reset() {
	*v = Location{}
}

type Company struct {
	CustomerID int64  `json:"customer_id"`
	IsLarge    bool   `json:"is_large"`
	BillingID  string `json:"billing_id"`
	ProjectID  int64  `json:"project_id"`
	IsActive   bool   `json:"is_active"`
}

func (v *Company) Reset() {
	*v = Company{}
}

type Target struct {
	ID                 int64   `json:"id"`
	PatchID            string  `json:"patch_id"`
	Peers              []*Peer `json:"peers"`
	RestartCount       int32   `json:"restart_count"`
	RestartCount24H    int32   `json:"restart_count_24h"`
	RestartCount7D     int32   `json:"restart_count_7d"`
	RebootCount        int32   `json:"reboot_count"`
	RebootCount24H     int32   `json:"reboot_count_24h"`
	RebootCount7D      int32   `json:"reboot_count_7d"`
	ReplicaCount       int32   `json:"replica_count"`
	ReplicaCount24H    int32   `json:"replica_count_24h"`
	ReplicaCount7D     int32   `json:"replica_count_7d"`
	ConnectionCount    int32   `json:"connection_count"`
	ConnectionCount24H int32   `json:"connection_count_24h"`
	ConnectionCount7D  int32   `json:"connection_count_7d"`
	CreationTimestamp  int64   `json:"creation_timestamp"`
}

func (v *Target) Reset() {
	*v = Target{}
}

type Peer struct {
	ID                        string          `json:"id"`
	RestartCount              int32           `json:"restart_count"`
	RestartCount24H           int32           `json:"restart_count_24h"`
	RestartCount7D            int32           `json:"restart_count_7d"`
	RebootCount               int32           `json:"reboot_count"`
	RebootCount24H            int32           `json:"reboot_count_24h"`
	RebootCount7D             int32           `json:"reboot_count_7d"`
	ReplicaCount              int32           `json:"replica_count"`
	ReplicaCount24H           int32           `json:"replica_count_24h"`
	ReplicaCount7D            int32           `json:"replica_count_7d"`
	ConnectionCount           int32           `json:"connection_count"`
	ConnectionCount24H        int32           `json:"connection_count_24h"`
	ConnectionCount7D         int32           `json:"connection_count_7d"`
	AverageRestartCount       int32           `json:"average_restart_count"`
	AverageRestartCount24H    int32           `json:"average_restart_count_24h"`
	AverageRestartCount7D     int32           `json:"average_restart_count_7d"`
	AverageRebootCount        int32           `json:"average_reboot_count"`
	AverageRebootCount24H     int32           `json:"average_reboot_count_24h"`
	AverageRebootCount7D      int32           `json:"average_reboot_count_7d"`
	AverageReplicaCount       int32           `json:"average_replica_count"`
	AverageReplicaCount24H    int32           `json:"average_replica_count_24h"`
	AverageReplicaCount7D     int32           `json:"average_replica_count_7d"`
	AverageConnectionCount    int32           `json:"average_connection_count"`
	AverageConnectionCount24H int32           `json:"average_connection_count_24h"`
	AverageConnectionCount7D  int32           `json:"average_connection_count_7d"`
	MedianRestartCount        int32           `json:"median_restart_count"`
	MedianRestartCount24H     int32           `json:"median_restart_count_24h"`
	MedianRestartCount7D      int32           `json:"median_restart_count_7d"`
	MedianRebootCount         int32           `json:"median_reboot_count"`
	MedianRebootCount24H      int32           `json:"median_reboot_count_24h"`
	MedianRebootCount7D       int32           `json:"median_reboot_count_7d"`
	MedianReplicaCount        int32           `json:"median_replica_count"`
	MedianReplicaCount24H     int32           `json:"median_replica_count_24h"`
	MedianReplicaCount7D      int32           `json:"median_replica_count_7d"`
	MedianConnectionCount     int32           `json:"median_connection_count"`
	MedianConnectionCount24H  int32           `json:"median_connection_count_24h"`
	MedianConnectionCount7D   int32           `json:"median_connection_count_7d"`
	IsReady                   bool            `json:"is_ready"`
	IsAvailable               bool            `json:"is_available"`
	IsPossible                bool            `json:"is_possible"`
	GlobalAverageCounters     *UptimeCounters `json:"global_average_counters"`
	GlobalMedianCounters      *UptimeCounters `json:"global_median_counters"`
}

type UptimeCounters struct {
	RestartCount int32 `json:"restart_count"`
	RebootCount  int32 `json:"reboot_count"`
}

type Placement struct {
	ID              string  `json:"id"`
	Toleration      float64 `json:"toleration"`
	TolerationDelta float64 `json:"toleration_delta"`
}

func (v *Placement) Reset() {
	*v = Placement{}
}

type RequestType int

const (
	RequestTypeInvalid RequestType = iota
	RequestTypeBroken
	RequestTypeCorrupted
	RequestTypeSalty
)

var internalRequestType = enum.New([]enum.Value[RequestType]{
	{Value: RequestTypeInvalid, String: "invalid"},
	{Value: RequestTypeBroken, String: "broken"},
	{Value: RequestTypeCorrupted, String: "corrupted"},
	{Value: RequestTypeSalty, String: "salty"},
})

func (RequestType) Descriptor() enum.Descriptor {
	return internalRequestType.Descriptor
}

func (v RequestType) MarshalText() ([]byte, error) {
	return internalRequestType.MarshalText(v)
}

func (v *RequestType) UnmarshalText(data []byte) error {
	return internalRequestType.UnmarshalText(v, data)
}

type RequestReason int

const (
	RequestReasonUnreasonable RequestReason = iota
	RequestReasonOutrageous
	RequestReasonUnknown
	RequestReasonReaper
)

var internalRequestReason = enum.New([]enum.Value[RequestReason]{
	{Value: RequestReasonUnreasonable, String: "unreasonable"},
	{Value: RequestReasonOutrageous, String: "outrageous"},
	{Value: RequestReasonUnknown, String: "unknown"},
	{Value: RequestReasonReaper, String: "reaper"},
})

func (RequestReason) Descriptor() enum.Descriptor {
	return internalRequestReason.Descriptor
}

func (v RequestReason) MarshalText() ([]byte, error) {
	return internalRequestReason.MarshalText(v)
}

func (v *RequestReason) UnmarshalText(data []byte) error {
	return internalRequestReason.UnmarshalText(v, data)
}

type ResponseType int

const (
	ResponseTypeBad ResponseType = iota
	ResponseTypeUgly
	ResponseTypeGood
)

var internalResponseType = enum.New([]enum.Value[ResponseType]{
	{Value: ResponseTypeBad, String: "bad"},
	{Value: ResponseTypeUgly, String: "ugly"},
	{Value: ResponseTypeGood, String: "good"},
})

func (ResponseType) Descriptor() enum.Descriptor {
	return internalResponseType.Descriptor
}

func (v ResponseType) MarshalText() ([]byte, error) {
	return internalResponseType.MarshalText(v)
}

func (v *ResponseType) UnmarshalText(data []byte) error {
	return internalResponseType.UnmarshalText(v, data)
}

type TokenScope int

const (
	TokenScopeRedDot TokenScope = iota
	TokenScopeHolo
	TokenScopeIron
)

var internalTokenScope = enum.New([]enum.Value[TokenScope]{
	{Value: TokenScopeRedDot, String: "red_dot"},
	{Value: TokenScopeHolo, String: "holo"},
	{Value: TokenScopeIron, String: "iron"},
})

func (TokenScope) Descriptor() enum.Descriptor {
	return internalTokenScope.Descriptor
}

func (v TokenScope) MarshalText() ([]byte, error) {
	return internalTokenScope.MarshalText(v)
}

func (v *TokenScope) UnmarshalText(data []byte) error {
	return internalTokenScope.UnmarshalText(v, data)
}

type HeapAllocator struct{}

type PoolAllocator struct {
	v protopool.Value
}

func PopulatePointer[T any](a Allocator) *T {
	switch alloc := a.(type) {
	case *HeapAllocator:
		var empty T
		return &empty
	case *PoolAllocator:
		return protopool.New[T](&alloc.v)
	default:
		panic("unknown allocator")
	}
}

func PopulateSlice[T any](a Allocator, slice *[]T, count int, populate func() T) {
	switch alloc := a.(type) {
	case *HeapAllocator:
		for i := 0; i < count; i++ {
			*slice = append(*slice, populate())
		}
	case *PoolAllocator:
		list := protopool.NewList(&alloc.v, slice)
		for i := 0; i < count; i++ {
			list.Append(populate())
		}
	default:
		panic("unknown allocator")
	}
}

func PopulateMap[K comparable, V any](a Allocator, m *map[K]V, count int, populate func() (K, V)) {
	switch alloc := a.(type) {
	case *HeapAllocator:
		mm := map[K]V{}
		for i := 0; i < count; i++ {
			k, v := populate()
			mm[k] = v
		}
		*m = mm
	case *PoolAllocator:
		for i := 0; i < count; i++ {
			k, v := populate()
			protopool.Put(&alloc.v, m, k, v)
		}
	default:
		panic("unknown allocator")
	}
}

func PopulateMap2[K comparable, V any](a Allocator, m *map[K]V, count int, populateKey func() K, populateValue func() V) {
	PopulateMap(a, m, count, func() (K, V) {
		return populateKey(), populateValue()
	})
}

func PopulateEnum[T enum.Enum](rng *rand.Rand, v *T) {
	*v = randomEnum[T](rng)
}

func randomEnum[T enum.Enum](rng *rand.Rand) T {
	var v T
	vals := v.Descriptor().Values()
	num := vals.Get(rng.Intn(vals.Len())).Number()
	return T(num)
}

const (
	nChildren = 15
	nTargets  = 1000
)

func assertEqual[T comparable](tb testing.TB, expected, received T) {
	tb.Helper()
	if expected != received {
		tb.Fatalf("expected %##v, received %##v", expected, received)
	}
}

func assertSlicesEqual[T comparable](tb testing.TB, expected, received []T) {
	tb.Helper()
	if len(expected) != len(received) {
		tb.Fatalf("expected %d elements, received %d", len(expected), len(received))
	}
	for i := range expected {
		if expected[i] != received[i] {
			tb.Fatalf("expected %##v at index %d, received %##v", expected[i], i, received[i])
		}
	}
}

func assertMapsEqual[K, V comparable](tb testing.TB, expected, received map[K]V) {
	tb.Helper()
	for k, ev := range expected {
		rv, ok := received[k]
		if !ok {
			tb.Fatalf("expected %##v for key %##v, found none", ev, k)
		}
		if ev != rv {
			tb.Fatalf("expected %##v for key %##v, found %##v", ev, k, rv)
		}
	}
	for k := range received {
		if _, ok := expected[k]; !ok {
			tb.Fatalf("unexpected value for key %##v", k)
		}
	}
}
