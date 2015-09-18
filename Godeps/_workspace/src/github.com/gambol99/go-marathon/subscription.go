/*
Copyright 2014 Rohith All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package marathon

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

type Subscriptions struct {
	CallbackURLs []string `json:"callbackUrls"`
}

// Retrieve a list of registered subscriptions
func (client *Client) Subscriptions() (*Subscriptions, error) {
	subscriptions := new(Subscriptions)
	if err := client.apiGet(MARATHON_API_SUBSCRIPTION, nil, subscriptions); err != nil {
		return nil, err
	} else {
		return subscriptions, nil
	}
}

// Add your self as a listener to events from Marathon
//		channel:	a EventsChannel used to receive event on
func (client *Client) AddEventsListener(channel EventsChannel, filter int) error {
	client.Lock()
	defer client.Unlock()
	// step: someone has asked to start listening to event, we need to register for events
	// if we haven't done so already
	if err := client.RegisterSubscription(); err != nil {
		return err
	}

	if _, found := client.listeners[channel]; !found {
		client.log("AddEventsListener() Adding a watch for events: %d, channel: %v", filter, channel)
		client.listeners[channel] = filter
	}
	return nil
}

// Remove the channel from the events listeners
//		channel:	the channel you are removing
func (client *Client) RemoveEventsListener(channel EventsChannel) {
	client.Lock()
	defer client.Unlock()
	if _, found := client.listeners[channel]; found {
		delete(client.listeners, channel)
		/* step: if there is no one listening anymore, lets remove our self
		from the events callback */
		if len(client.listeners) <= 0 {
			client.UnSubscribe()
		}
	}
}

// Retrieve the subscription call back URL used when registering
func (client *Client) SubscriptionURL() string {
	return fmt.Sprintf("http://%s:%d%s", client.ipaddress, client.config.EventsPort, DEFAULT_EVENTS_URL)
}

// Register ourselves with Marathon to receive events from it's callback facility
func (client *Client) RegisterSubscription() error {
	if client.events_http == nil {
		if ip_address, err := getInterfaceAddress(client.config.EventsInterface); err != nil {
			return errors.New(fmt.Sprintf("Unable to get the ip address from the interface: %s, error: %s",
				client.config.EventsInterface, err))
		} else {
			// step: set the ip address
			client.ipaddress = ip_address
			binding := fmt.Sprintf("%s:%d", ip_address, client.config.EventsPort)
			// step: register the handler
			http.HandleFunc(DEFAULT_EVENTS_URL, client.HandleMarathonEvent)
			// step: create the http server
			client.events_http = &http.Server{
				Addr:           binding,
				Handler:        nil,
				ReadTimeout:    10 * time.Second,
				WriteTimeout:   10 * time.Second,
				MaxHeaderBytes: 1 << 20,
			}
			client.log("RegisterSubscription() Attempting to listen on binding: %s", binding)

			// @todo need to add a timeout value here
			listener, err := net.Listen("tcp", binding)
			if err != nil {
				return nil
			}

			client.log("RegisterSubscription() Starting to listen on http events service")
			go func() {
				for {
					client.events_http.Serve(listener)
					client.log("RegisterSubscription() Exitted the http events service")
				}
			}()
		}
	}

	// step: get the callback url
	callback := client.SubscriptionURL()

	// step: check if the callback is registered
	client.log("RegisterSubscription() Checking if we already have a subscription for callback %s", callback)
	if found, err := client.HasSubscription(callback); err != nil {
		return err
	} else if !found {
		client.log("RegisterSubscription() Registering a subscription with Marathon: callback: %s", callback)
		// step: we need to register our self
		uri := fmt.Sprintf("%s?callbackUrl=%s", MARATHON_API_SUBSCRIPTION, callback)
		if err := client.apiPost(uri, "", nil); err != nil {
			return err
		}
	} else {
		client.log("RegisterSubscription() A subscription already exists for this callback: %s", callback)
	}
	return nil
}

// Remove ourselves from Marathon's callback facility
func (client *Client) UnSubscribe() error {
	/* step: remove from the list of subscriptions */
	return client.apiDelete(fmt.Sprintf("%s?callbackUrl=%s", MARATHON_API_SUBSCRIPTION, client.SubscriptionURL()), nil, nil)
}

// Check to see a subscription already exists with Marathon
//		callback:	the url of the callback
func (client *Client) HasSubscription(callback string) (bool, error) {
	client.log("HasSubscription() Checking for subscription: %s", callback)
	/* step: generate our events callback */
	if subscriptions, err := client.Subscriptions(); err != nil {
		return false, err
	} else {
		for _, subscription := range subscriptions.CallbackURLs {
			if callback == subscription {
				return true, nil
			}
		}
	}
	return false, nil
}

func (client *Client) HandleMarathonEvent(writer http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		client.log("HandleMarathonEvent() Failed to decode the event type, content: %s, error: %s")
		return
	}

	// step: process the event and decode the event
	content := string(body[:])
	event_type := new(EventType)
	err = json.NewDecoder(strings.NewReader(content)).Decode(event_type)
	if err != nil {
		client.log("HandleMarathonEvent() Failed to decode the event type, content: %s, error: %s", content, err)
		return
	}

	client.log("HandleMarathonEvent() Recieved marathon event, %s", event_type.EventType)

	// step: check the type is handled
	event, err := client.GetEvent(event_type.EventType)
	if err != nil {
		client.log("HandleMarathonEvent() Unable to retrieve the event, type: %s", event_type.EventType)
		return
	}

	// step: lets decode message
	err = json.NewDecoder(strings.NewReader(content)).Decode(event.Event)
	if err != nil {
		client.log("HandleMarathonEvent() Failed to decode the event type: %d, name: %s error: %s", event.ID, err)
		return
	}

	client.log("HandleMarathonEvent() Decoded the marathon event, %s", event)

	client.RLock()
	defer client.RUnlock()

	// step: check if anyone is listen for this event
	for channel, filter := range client.listeners {
		// step: check if this listener wants this event type
		client.log("HandleMarathonEvent() checking: channel: %v, type: %d, filter: %d", channel, event.ID, filter)
		if event.ID&filter != 0 {
			client.log("HandleMarathonEvent() Event type: %d being listened to, sending to listener: %v", event.ID, channel)
			go func(ch EventsChannel, e *Event) {
				ch <- e
			}(channel, event)
		} else {
			client.log("HandleMarathonEvent() Event type: %d is not being listened to by listener: %v", event.ID, channel)
		}
	}
}

func (client *Client) GetEvent(name string) (*Event, error) {
	// step: check it's supported
	if id, found := Events[name]; found {
		event := new(Event)
		event.ID = id
		event.Name = name
		switch name {
		case "api_post_event":
			event.Event = new(EventAPIRequest)
		case "status_update_event":
			event.Event = new(EventStatusUpdate)
		case "framework_message_event":
			event.Event = new(EventFrameworkMessage)
		case "subscribe_event":
			event.Event = new(EventSubscription)
		case "unsubscribe_event":
			event.Event = new(EventUnsubscription)
		case "add_health_check_event":
			event.Event = new(EventAddHealthCheck)
		case "remove_health_check_event":
			event.Event = new(EventRemoveHealthCheck)
		case "failed_health_check_event":
			event.Event = new(EventFailedHealthCheck)
		case "health_status_changed_event":
			event.Event = new(EventHealthCheckChanged)
		case "group_change_success":
			event.Event = new(EventGroupChangeSuccess)
		case "group_change_failed":
			event.Event = new(EventGroupChangeFailed)
		case "deployment_success":
			event.Event = new(EventDeploymentSuccess)
		case "deployment_failed":
			event.Event = new(EventDeploymentFailed)
		case "deployment_info":
			event.Event = new(EventDeploymentInfo)
		case "deployment_step_success":
			event.Event = new(EventDeploymentStepSuccess)
		case "deployment_step_failure":
			event.Event = new(EventDeploymentStepFailure)
		case "app_terminated_event":
			event.Event = new(EventAppTerminated)
		}
		return event, nil
	} else {
		return nil, errors.New(fmt.Sprintf("The event type: %d was not found or supported", name))
	}
}
