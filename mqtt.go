package main

import (
	"encoding/json"
	emitter "github.com/emitter-io/go"
	"log"
)

func initEmitter(bc *Broker) emitter.Emitter {
	log.Println("Begin with emitter broker init...")

	// Create the options with default values
	options := emitter.NewClientOptions()
	// Broker for exchanging current energy consumptions (lifetime is set to one year)
	options.AddBroker(bc.BaseAddress + bc.TargetChannel1) //Broker for sending changes of users
	options.AddBroker(bc.BaseAddress + bc.TargetChannel2) //Broker for sending loading data (send data of the battery modules)
	options.AddBroker(bc.BaseAddress + bc.TargetChannel3) //Broker for receiving action data (actions like relay on)

	// set the message handler
	options.SetOnMessageHandler(messageHandler)

	// Create a new emitter client and connect to the broker
	client := emitter.NewClient(options)
	sToken := client.Connect()
	if sToken.Wait() && sToken.Error() != nil {
		panic("Error on Client.Connect() to broker: " + sToken.Error().Error())
	}

	//Subscribe to channel cq/actions
	sToken = client.Subscribe(bc.KeyChannel3, bc.TargetChannel3)
	if sToken.Wait() && sToken.Error() != nil {
		panic("Error on clientC.Subscribe(): " + sToken.Error().Error())
	}

	//Publish to the channels /changes & /battery
	//sToken = client.Publish(bc.KeyChannel1, bc.TargetChannel1, "test changes channel")
	//if sToken.Wait() && sToken.Error() != nil {
	//	panic("Error on Client.Publish(): " + sToken.Error().Error())
	//}
	return client
}

// messageHandler stores incoming actions into the global slice actions
func messageHandler(client emitter.Emitter, msg emitter.Message) {
	log.Println("Received message for actions: %s\n", msg.Topic(), string(msg.Payload()))
	//buffer message into action variable
	var temp Action
	err := json.Unmarshal(msg.Payload(), &temp)
	if err != nil {
		log.Println("Error when unmarshal received object from channel" + msg.Topic())
		panic(err)
	}
	actions = append(actions, temp)
}

// sendChanges publishes the ChangeStatus struct via json encoding to the /qc/changes/ channel
func sendChanges(client emitter.Emitter, b *Broker, changes ChangeStatus) error {
	var msg, err = json.Marshal(changes)
	if err != nil {
		log.Println("Error encoding struct ChangeStatus as string.")
		return err
	}
	sToken := client.Publish(b.KeyChannel1, b.TargetChannel1, string(msg))
	if sToken.Wait() && sToken.Error() != nil {
		log.Println("Error on Client.Publish().")
	}
	return sToken.Error()
}

// sendFeederStatus publishes the BatteryStatus struct via json encoding to the /cq/battery/ channel
func sendFeederStatus(client emitter.Emitter, b *Broker, battery FeederStatus) error {
	var msg, err = json.Marshal(battery)
	if err != nil {
		log.Println("Error encoding struct BatteryStatus as string.")
		return err
	}
	sToken := client.Publish(b.KeyChannel2, b.TargetChannel2, msg)
	if sToken.Wait() && sToken.Error() != nil {
		log.Println("Error on Client.Publish().")
	}
	return err
}
