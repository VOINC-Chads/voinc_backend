package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"voinc-backend/websocket"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
)

var lock = &sync.Mutex{}

type terraform struct {
	ctx        context.Context
	tf         *tfexec.Terraform
	execPath   string
	workingDir string
}

var terraformInstance *terraform

func GetInstance() *terraform {
	if terraformInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		if terraformInstance == nil {
			fmt.Println("Creating single instance now.")
			terraformInstance = &terraform{}
			fmt.Println("Created instance, now initialize")
			terraformInstance.Initialize()
		} else {
			fmt.Println("Single instance already created.")
		}
	} else {
		fmt.Println("Single instance already created.")
	}

	return terraformInstance
}

func (t *terraform) Initialize() {
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("1.0.6")),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		log.Fatalf("error installing Terraform: %s", err)
	}
	t.execPath = execPath

	t.workingDir = "./infra"
	tf, err := tfexec.NewTerraform(t.workingDir, t.execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}

	t.tf = tf

	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}

	state, err := tf.Show(context.Background())
	if err != nil {
		log.Fatalf("error running Show: %s", err)
	}

	fmt.Println("Terraform version:", state.FormatVersion) // "0.1"
}

func (t *terraform) Apply() {
	ctx := context.Background()

	// Run "terraform apply" to apply the changes
	err := t.tf.Apply(ctx)
	if err != nil {
		log.Fatalf("Error running terraform apply: %s", err)
	}

	// Print the Terraform output
	output, err := t.tf.Output(ctx)
	if err != nil {
		log.Fatalf("Error getting terraform output: %s", err)
	}
	fmt.Printf("Terraform output: %s\n", output)
	var ipMaps interface{}
	errJson := json.Unmarshal([]byte(output["public-ip"].Value), &ipMaps)
	if errJson != nil {
		// TODO: Fix this?
		//panic(errJson)
	}

	// Navigate the interface using type assertions.
	for uuid, ip := range ipMaps.(map[string]interface{}) {
		ip = ip.(map[string]interface{})["public_ip"]
		fmt.Printf("UUID: %s, IP: %s\n", uuid, ip)
		(*websocket.Sessions)[uuid] = ip.(string) // Update ip of each uuid
	}
}
