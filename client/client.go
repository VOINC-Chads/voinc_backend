package client

import (
	"log"
	"strconv"
	"sync"

	"voinc-backend/client/messages"

	"github.com/golang/protobuf/proto"
	"github.com/zeromq/goczmq"
)

var lock = &sync.Mutex{}

type Client struct {
	Dealer *goczmq.Sock
}

func InitializeClient(addr string, port int) *Client {
	c := &Client{}

	// Create a dealer socket and connect it to the router.
	var err error
	c.Dealer, err = goczmq.NewDealer("tcp://" + addr + ":" + strconv.Itoa(port))
	if err != nil {
		log.Fatal(err)
	}

	return c
}

func (c *Client) Listen(send func(v interface{})) {
	// Register a poller
	poller, err := goczmq.NewPoller()
	if err != nil {
		log.Fatal(err)
	}
	poller.Add(c.Dealer)

	for {
		socket := poller.Wait(-1)

		if socket == c.Dealer {
			request, err := c.Dealer.RecvMessage()
			if err != nil {
				log.Fatal(err)
			}


			log.Printf("dealer received")
			send("dealer received")
			log.Println(request)
		} else {
			log.Fatal("Unexpected socket")
		}
	}
}

func (c *Client) SendHeartbeat() {
	heartbeatMsg := &messages.MainReq{
		MsgType: messages.MsgTypes_TYPE_HEARTBEAT,
		Content: &messages.MainReq_Heartbeat{
			Heartbeat: &messages.Heartbeat{},
		},
	}

	protoMsg, err := proto.Marshal(heartbeatMsg)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	err = c.Dealer.SendFrame(protoMsg, goczmq.FlagNone)
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Client) SendCode(requirements string, processCode string, executeCode string) {
	mainMsg := &messages.MainReq{
		MsgType: messages.MsgTypes_TYPE_CODE,
		Content: &messages.MainReq_CodeMsg{
			CodeMsg: &messages.CodeMsg{
				Requirements: requirements,
				ProcessCode:  processCode,
				ExecuteCode:  executeCode,
			},
		},
	}

	protoMsg, err := proto.Marshal(mainMsg)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	err = c.Dealer.SendFrame(protoMsg, goczmq.FlagNone)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("dealer sent '" + string(protoMsg) + "'")
}

func (c *Client) SendJobs(jobs []string) {
	mainMsg := &messages.MainReq{
		MsgType: messages.MsgTypes_TYPE_JOB,
		Content: &messages.MainReq_JobMsg{
			JobMsg: &messages.JobMsg{
				Jobs: jobs,
			},
		},
	}

	protoMsg, err := proto.Marshal(mainMsg)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	err = c.Dealer.SendFrame(protoMsg, goczmq.FlagNone)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("dealer sent '" + string(protoMsg) + "'")
}