package transport

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	fuzz "github.com/google/gofuzz"
	"github.com/hashicorp/raft"
	"go.uber.org/goleak"
)

func fuzzLogType(lt *raft.LogType, c fuzz.Continue) {
	*lt = raft.LogType(c.Intn(int(raft.LogConfiguration + 1)))
}

func doFuzz(rm interface{}, i int) {
	f := fuzz.NewWithSeed(int64(i)).NilChance(0).Funcs(fuzzLogType)
	f.Fuzz(rm)
}

func verify(t *testing.T, rm1, rm2 interface{}) {
	t.Helper()
	if diff := cmp.Diff(rm1, rm2); diff != "" {
		t.Errorf("encode+decode return another value: %s", diff)
	}
}

func TestAppendEntriesRequest(t *testing.T) {
	defer goleak.VerifyNone(t)

	for i := 0; 1000 > i; i++ {
		rm := raft.AppendEntriesRequest{}
		doFuzz(&rm, i)
		pm := encodeAppendEntriesRequest(&rm)
		rm2 := decodeAppendEntriesRequest(pm)
		verify(t, &rm, rm2)
	}
}

func TestAppendEntriesResponse(t *testing.T) {
	defer goleak.VerifyNone(t)

	for i := 0; 1000 > i; i++ {
		rm := raft.AppendEntriesResponse{}
		doFuzz(&rm, i)
		pm := encodeAppendEntriesResponse(&rm)
		rm2 := decodeAppendEntriesResponse(pm)
		verify(t, &rm, rm2)
	}
}

func TestRequestVoteRequest(t *testing.T) {
	for i := 0; 1000 > i; i++ {
		rm := raft.RequestVoteRequest{}
		doFuzz(&rm, i)
		pm := encodeRequestVoteRequest(&rm)
		rm2 := decodeRequestVoteRequest(pm)
		verify(t, &rm, rm2)
	}
}

func TestRequestVoteResponse(t *testing.T) {
	defer goleak.VerifyNone(t)

	for i := 0; 1000 > i; i++ {
		rm := raft.RequestVoteResponse{}
		doFuzz(&rm, i)
		pm := encodeRequestVoteResponse(&rm)
		rm2 := decodeRequestVoteResponse(pm)
		verify(t, &rm, rm2)
	}
}

func TestInstallSnapshotRequest(t *testing.T) {
	defer goleak.VerifyNone(t)

	for i := 0; 1000 > i; i++ {
		rm := raft.InstallSnapshotRequest{}
		doFuzz(&rm, i)
		pm := encodeInstallSnapshotRequest(&rm)
		rm2 := decodeInstallSnapshotRequest(pm)
		verify(t, &rm, rm2)
	}
}

func TestInstallSnapshotResponse(t *testing.T) {
	defer goleak.VerifyNone(t)

	for i := 0; 1000 > i; i++ {
		rm := raft.InstallSnapshotResponse{}
		doFuzz(&rm, i)
		pm := encodeInstallSnapshotResponse(&rm)
		rm2 := decodeInstallSnapshotResponse(pm)
		verify(t, &rm, rm2)
	}
}

func TestTimeoutNowRequest(t *testing.T) {
	defer goleak.VerifyNone(t)

	for i := 0; 1000 > i; i++ {
		rm := raft.TimeoutNowRequest{}
		doFuzz(&rm, i)
		pm := encodeTimeoutNowRequest(&rm)
		rm2 := decodeTimeoutNowRequest(pm)
		verify(t, &rm, rm2)
	}
}

func TestTimeoutNowResponse(t *testing.T) {
	defer goleak.VerifyNone(t)

	for i := 0; 1000 > i; i++ {
		rm := raft.TimeoutNowResponse{}
		doFuzz(&rm, i)
		pm := encodeTimeoutNowResponse(&rm)
		rm2 := decodeTimeoutNowResponse(pm)
		verify(t, &rm, rm2)
	}
}
