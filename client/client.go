package client

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"voinc-backend/client/messages"
	"voinc-backend/commons"

	"github.com/golang/protobuf/proto"
	"github.com/zeromq/goczmq"
)

var lock = &sync.Mutex{}

type Client struct {
	Dealer *goczmq.Sock
	ip     string
}

func InitializeClient(addr string, port int) *Client {
	c := &Client{}

	// Create a dealer socket and connect it to the router.
	var err error
	c.Dealer, err = goczmq.NewDealer("tcp://" + addr + ":" + strconv.Itoa(port))
	if err != nil {
		log.Fatal(err)
	}
	c.ip = addr
	return c
}

func (c *Client) Listen(send func(v interface{}) error) {
	// Register a poller
	poller, err := goczmq.NewPoller()
	if err != nil {
		log.Fatal(err)
	}
	poller.Add(c.Dealer)

	for {
		socket := poller.Wait(-1)

		if socket == c.Dealer {
			log.Println("Message recieved")
			responseArr, err := c.Dealer.RecvMessage()
			if err != nil {
				log.Fatal(err)
			}
			log.Println(responseArr)
			// Get last element in responseArr[]
			// UnMarshall into
			response := responseArr[len(responseArr)-1]
			log.Printf("RESPONSE BYTE: %q", response)

			// UnMarshall main
			mainResponse := &messages.MainResp{}
			err = proto.Unmarshal(response[:], mainResponse)
			if err != nil {
				log.Printf("ERROR: %v", err)
				if e, ok := err.(*json.SyntaxError); ok {
					log.Printf("syntax error at byte offset %d", e.Offset)
				}
				log.Printf("sakura response: %q", response[:])
				send(commons.MessageToClient{
					Status:  "ERROR",
					Content: "Could not process json returned from instance",
				})
				return
			}
			log.Printf(string(mainResponse.MsgType))
			switch mainResponse.MsgType {
			case messages.MsgTypes_TYPE_HEARTBEAT:
				log.Printf("cant you heart BEAT to the beat of the asdyrum ayyy")

				ip := map[string]string{"ip": c.ip}
				ipJson, _ := json.Marshal(ip)
				send(commons.MessageToClient{
					Status:  "UPDATE",
					Content: string(ipJson),
				})
				break
			case messages.MsgTypes_TYPE_JOB:
				log.Printf("Job response received")

				response, _ := json.Marshal(mainResponse.Content)
				send(commons.MessageToClient{
					Status:  "COMPLETE",
					Content: string(response),
				})
				break
			case messages.MsgTypes_TYPE_CODE:
				send(commons.MessageToClient{
					Status:  "UPDATE",
					Content: "Received code :)",
				})
			default:
				send(commons.MessageToClient{
					Status:  "ERROR",
					Content: "Unrecognized",
				})

			}

		} else {
			log.Fatal("Unexpected socket")
		}
	}
}

func (c *Client) SendHeartbeat() {
	fmt.Println("SendHeartBeat()")
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
	fmt.Println("SendCode()")
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
	fmt.Println("SendJobs()")
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
