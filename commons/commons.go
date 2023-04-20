package commons

// MessageToClient
//- Status: 		The type of response
//- Content: Content of the response (not always there)
type MessageToClient struct {
	Status  string `json:"status"`
	Content string `json:"content"`
}

// enum MsgTypes
type MsgTypes int32

const (
	MsgTypes_TYPE_UNKNOWN   MsgTypes = 0
	MsgTypes_TYPE_REGISTER  MsgTypes = 1
	MsgTypes_TYPE_JOB       MsgTypes = 2
	MsgTypes_TYPE_CODE      MsgTypes = 3
	MsgTypes_TYPE_HEARTBEAT MsgTypes = 4
)

// message Heartbeat
type Heartbeat struct{}

// message JobResult
type JobResult struct {
	Value  string `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	Result string `protobuf:"bytes,2,opt,name=result,proto3" json:"result,omitempty"`
}

// message JobResp
type JobResp struct {
	Status  *Status      `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	Results []*JobResult `protobuf:"bytes,2,rep,name=results,proto3" json:"results,omitempty"`
}

// message MainResp
type MainResp struct {
	MsgType MsgTypes           `protobuf:"varint,1,opt,name=msg_type,json=msgType,proto3,enum=MsgTypes" json:"msg_type,omitempty"`
	Content isMainResp_Content `protobuf_oneof:"content"`
}

type isMainResp_Content interface {
	ismainrespContent()
}

// message MainResp.JobResp
type MainResp_JobResp struct {
	JobResp *JobResp `protobuf:"bytes,2,opt,name=job_resp,json=jobResp,proto3,oneof"`
}

func (*MainResp_JobResp) ismainrespContent() {}

// message MainResp.Heartbeat
type MainResp_Heartbeat struct {
	Heartbeat *Heartbeat `protobuf:"bytes,3,opt,name=heartbeat,proto3,oneof"`
}

func (*MainResp_Heartbeat) ismainrespContent() {}

// enum Status
type Status int32

const (
	Status_SUCCESS Status = 0
	Status_ERROR   Status = 1
)
