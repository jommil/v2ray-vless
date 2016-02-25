package protocol_test

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
	"time"

	v2net "github.com/v2ray/v2ray-core/common/net"
	proto "github.com/v2ray/v2ray-core/common/protocol"
	"github.com/v2ray/v2ray-core/common/uuid"
	. "github.com/v2ray/v2ray-core/proxy/vmess/protocol"
	protocoltesting "github.com/v2ray/v2ray-core/proxy/vmess/protocol/testing"
	v2testing "github.com/v2ray/v2ray-core/testing"
	"github.com/v2ray/v2ray-core/testing/assert"
)

func newStaticTimestampGenerator(t proto.Timestamp) proto.TimestampGenerator {
	return func() proto.Timestamp {
		return t
	}
}

func TestVMessSerialization(t *testing.T) {
	v2testing.Current(t)

	id, err := uuid.ParseString("2b2966ac-16aa-4fbf-8d81-c5f172a3da51")
	assert.Error(err).IsNil()

	userId := proto.NewID(id)

	testUser := &proto.User{
		ID: userId,
	}

	userSet := protocoltesting.MockUserSet{[]*proto.User{}, make(map[string]int), make(map[string]proto.Timestamp)}
	userSet.Add(testUser)

	request := new(VMessRequest)
	request.Version = byte(0x01)
	request.User = testUser

	randBytes := make([]byte, 36)
	_, err = rand.Read(randBytes)
	assert.Error(err).IsNil()
	request.RequestIV = randBytes[:16]
	request.RequestKey = randBytes[16:32]
	request.ResponseHeader = randBytes[32]

	request.Command = byte(0x01)
	request.Address = v2net.DomainAddress("v2ray.com")
	request.Port = v2net.Port(80)

	mockTime := proto.Timestamp(1823730)

	buffer, err := request.ToBytes(newStaticTimestampGenerator(mockTime), nil)
	if err != nil {
		t.Fatal(err)
	}

	userSet.UserHashes[string(buffer.Value[:16])] = 0
	userSet.Timestamps[string(buffer.Value[:16])] = mockTime

	requestReader := NewVMessRequestReader(&userSet)
	actualRequest, err := requestReader.Read(bytes.NewReader(buffer.Value))
	if err != nil {
		t.Fatal(err)
	}

	assert.Byte(actualRequest.Version).Named("Version").Equals(byte(0x01))
	assert.String(actualRequest.User.ID).Named("UserId").Equals(request.User.ID.String())
	assert.Bytes(actualRequest.RequestIV).Named("RequestIV").Equals(request.RequestIV[:])
	assert.Bytes(actualRequest.RequestKey).Named("RequestKey").Equals(request.RequestKey[:])
	assert.Byte(actualRequest.ResponseHeader).Named("ResponseHeader").Equals(request.ResponseHeader)
	assert.Byte(actualRequest.Command).Named("Command").Equals(request.Command)
	assert.String(actualRequest.Address).Named("Address").Equals(request.Address.String())
}

func TestReadSingleByte(t *testing.T) {
	v2testing.Current(t)

	reader := NewVMessRequestReader(nil)
	_, err := reader.Read(bytes.NewReader(make([]byte, 1)))
	assert.Error(err).Equals(io.ErrUnexpectedEOF)
}

func BenchmarkVMessRequestWriting(b *testing.B) {
	id, err := uuid.ParseString("2b2966ac-16aa-4fbf-8d81-c5f172a3da51")
	assert.Error(err).IsNil()

	userId := proto.NewID(id)
	userSet := protocoltesting.MockUserSet{[]*proto.User{}, make(map[string]int), make(map[string]proto.Timestamp)}

	testUser := &proto.User{
		ID: userId,
	}
	userSet.Add(testUser)

	request := new(VMessRequest)
	request.Version = byte(0x01)
	request.User = testUser

	randBytes := make([]byte, 36)
	rand.Read(randBytes)
	request.RequestIV = randBytes[:16]
	request.RequestKey = randBytes[16:32]
	request.ResponseHeader = randBytes[32]

	request.Command = byte(0x01)
	request.Address = v2net.DomainAddress("v2ray.com")
	request.Port = v2net.Port(80)

	for i := 0; i < b.N; i++ {
		request.ToBytes(proto.NewTimestampGenerator(proto.Timestamp(time.Now().Unix()), 30), nil)
	}
}
