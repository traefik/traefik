package linodego

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// WaitForInstanceStatus waits for the Linode instance to reach the desired state
// before returning. It will timeout with an error after timeoutSeconds.
func (client Client) WaitForInstanceStatus(ctx context.Context, instanceID int, status InstanceStatus, timeoutSeconds int) (*Instance, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			instance, err := client.GetInstance(ctx, instanceID)
			if err != nil {
				return instance, err
			}
			complete := (instance.Status == status)

			if complete {
				return instance, nil
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("Error waiting for Instance %d status %s: %s", instanceID, status, ctx.Err())
		}
	}
}

// WaitForVolumeStatus waits for the Volume to reach the desired state
// before returning. It will timeout with an error after timeoutSeconds.
func (client Client) WaitForVolumeStatus(ctx context.Context, volumeID int, status VolumeStatus, timeoutSeconds int) (*Volume, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			volume, err := client.GetVolume(ctx, volumeID)
			if err != nil {
				return volume, err
			}
			complete := (volume.Status == status)

			if complete {
				return volume, nil
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("Error waiting for Volume %d status %s: %s", volumeID, status, ctx.Err())
		}
	}
}

// WaitForSnapshotStatus waits for the Snapshot to reach the desired state
// before returning. It will timeout with an error after timeoutSeconds.
func (client Client) WaitForSnapshotStatus(ctx context.Context, instanceID int, snapshotID int, status InstanceSnapshotStatus, timeoutSeconds int) (*InstanceSnapshot, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			snapshot, err := client.GetInstanceSnapshot(ctx, instanceID, snapshotID)
			if err != nil {
				return snapshot, err
			}
			complete := (snapshot.Status == status)

			if complete {
				return snapshot, nil
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("Error waiting for Instance %d Snapshot %d status %s: %s", instanceID, snapshotID, status, ctx.Err())
		}
	}
}

// WaitForVolumeLinodeID waits for the Volume to match the desired LinodeID
// before returning. An active Instance will not immediately attach or detach a volume, so the
// the LinodeID must be polled to determine volume readiness from the API.
// WaitForVolumeLinodeID will timeout with an error after timeoutSeconds.
func (client Client) WaitForVolumeLinodeID(ctx context.Context, volumeID int, linodeID *int, timeoutSeconds int) (*Volume, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			volume, err := client.GetVolume(ctx, volumeID)
			if err != nil {
				return volume, err
			}

			if linodeID == nil && volume.LinodeID == nil {
				return volume, nil
			} else if linodeID == nil || volume.LinodeID == nil {
				// continue waiting
			} else if *volume.LinodeID == *linodeID {
				return volume, nil
			}

		case <-ctx.Done():
			return nil, fmt.Errorf("Error waiting for Volume %d to have Instance %v: %s", volumeID, linodeID, ctx.Err())
		}
	}
}

// WaitForEventFinished waits for an entity action to reach the 'finished' state
// before returning. It will timeout with an error after timeoutSeconds.
// If the event indicates a failure both the failed event and the error will be returned.
func (client Client) WaitForEventFinished(ctx context.Context, id interface{}, entityType EntityType, action EventAction, minStart time.Time, timeoutSeconds int) (*Event, error) {
	titledEntityType := strings.Title(string(entityType))
	filter, _ := json.Marshal(map[string]interface{}{
		// Entity is not filtered by the API
		// Perhaps one day they will permit Entity ID/Type filtering.
		// We'll have to verify these values manually, for now.
		//"entity": map[string]interface{}{
		//	"id":   fmt.Sprintf("%v", id),
		//	"type": entityType,
		//},

		// Nor is action
		//"action": action,

		// Created is not correctly filtered by the API
		// We'll have to verify these values manually, for now.
		//"created": map[string]interface{}{
		//	"+gte": minStart.Format(time.RFC3339),
		//},

		// With potentially 1000+ events coming back, we should filter on something
		"seen": false,

		// Float the latest events to page 1
		"+order_by": "created",
		"+order":    "desc",
	})

	// Optimistically restrict results to page 1.  We should remove this when more
	// precise filtering options exist.
	listOptions := NewListOptions(1, string(filter))

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	if deadline, ok := ctx.Deadline(); ok {
		duration := time.Until(deadline)
		log.Printf("[INFO] Waiting %d seconds for %s events since %v for %s %v", int(duration.Seconds()), action, minStart, titledEntityType, id)
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:

			events, err := client.ListEvents(ctx, listOptions)
			if err != nil {
				return nil, err
			}

			// If there are events for this instance + action, inspect them
			for _, event := range events {
				if event.Action != action {
					// log.Println("action mismatch", event.Action, action)
					continue
				}
				if event.Entity.Type != entityType {
					// log.Println("type mismatch", event.Entity.Type, entityType)
					continue
				}

				var entID string

				switch event.Entity.ID.(type) {
				case float64, float32:
					entID = fmt.Sprintf("%.f", event.Entity.ID)
				case int:
					entID = strconv.Itoa(event.Entity.ID.(int))
				default:
					entID = fmt.Sprintf("%v", event.Entity.ID)
				}

				var findID string
				switch id.(type) {
				case float64, float32:
					findID = fmt.Sprintf("%.f", id)
				case int:
					findID = strconv.Itoa(id.(int))
				default:
					findID = fmt.Sprintf("%v", id)
				}

				if entID != findID {
					// log.Println("id mismatch", entID, findID)
					continue
				}

				// @TODO(displague) This event.Created check shouldn't be needed, but it appears
				// that the ListEvents method is not populating it correctly
				if event.Created == nil {
					log.Printf("[WARN] event.Created is nil when API returned: %#+v", event.CreatedStr)
				} else if *event.Created != minStart && !event.Created.After(minStart) {
					// Not the event we were looking for
					// log.Println(event.Created, "is not >=", minStart)
					continue

				}

				if event.Status == EventFailed {
					return &event, fmt.Errorf("%s %v action %s failed", titledEntityType, id, action)
				} else if event.Status == EventScheduled {
					log.Printf("[INFO] %s %v action %s is scheduled", titledEntityType, id, action)
				} else if event.Status == EventFinished {
					log.Printf("[INFO] %s %v action %s is finished", titledEntityType, id, action)
					return &event, nil
				}
				// TODO(displague) can we bump the ticker to TimeRemaining/2 (>=1) when non-nil?
				log.Printf("[INFO] %s %v action %s is %s", titledEntityType, id, action, event.Status)
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("Error waiting for Event Status '%s' of %s %v action '%s': %s", EventFinished, titledEntityType, id, action, ctx.Err())
		}
	}
}
