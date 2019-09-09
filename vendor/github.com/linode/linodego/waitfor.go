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

	ticker := time.NewTicker(client.millisecondsPerPoll * time.Millisecond)
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

// WaitForInstanceDiskStatus waits for the Linode instance disk to reach the desired state
// before returning. It will timeout with an error after timeoutSeconds.
func (client Client) WaitForInstanceDiskStatus(ctx context.Context, instanceID int, diskID int, status DiskStatus, timeoutSeconds int) (*InstanceDisk, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	ticker := time.NewTicker(client.millisecondsPerPoll * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// GetInstanceDisk will 404 on newly created disks. use List instead.
			// disk, err := client.GetInstanceDisk(ctx, instanceID, diskID)
			disks, err := client.ListInstanceDisks(ctx, instanceID, nil)
			if err != nil {
				return nil, err
			}
			for _, disk := range disks {
				disk := disk
				if disk.ID == diskID {
					complete := (disk.Status == status)
					if complete {
						return &disk, nil
					}
					break
				}
			}

		case <-ctx.Done():
			return nil, fmt.Errorf("Error waiting for Instance %d Disk %d status %s: %s", instanceID, diskID, status, ctx.Err())
		}
	}
}

// WaitForVolumeStatus waits for the Volume to reach the desired state
// before returning. It will timeout with an error after timeoutSeconds.
func (client Client) WaitForVolumeStatus(ctx context.Context, volumeID int, status VolumeStatus, timeoutSeconds int) (*Volume, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	ticker := time.NewTicker(client.millisecondsPerPoll * time.Millisecond)
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

	ticker := time.NewTicker(client.millisecondsPerPoll * time.Millisecond)
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

	ticker := time.NewTicker(client.millisecondsPerPoll * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			volume, err := client.GetVolume(ctx, volumeID)
			if err != nil {
				return volume, err
			}

			switch {
			case linodeID == nil && volume.LinodeID == nil:
				return volume, nil
			case linodeID == nil || volume.LinodeID == nil:
				// continue waiting
			case *volume.LinodeID == *linodeID:
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
	filterStruct := map[string]interface{}{
		// Nor is action
		//"action": action,

		// Created is not correctly filtered by the API
		// We'll have to verify these values manually, for now.
		//"created": map[string]interface{}{
		//	"+gte": minStart.Format(time.RFC3339),
		//},

		// With potentially 1000+ events coming back, we should filter on something
		// Warning: This optimization has the potential to break if users are clearing
		// events before we see them.
		"seen": false,

		// Float the latest events to page 1
		"+order_by": "created",
		"+order":    "desc",
	}

	// Optimistically restrict results to page 1.  We should remove this when more
	// precise filtering options exist.
	pages := 1

	// The API has limitted filtering support for Event ID and Event Type
	// Optimize the list, if possible
	switch entityType {
	case EntityDisk, EntityLinode, EntityDomain, EntityNodebalancer:
		// All of the filter supported types have int ids
		filterableEntityID, err := strconv.Atoi(fmt.Sprintf("%v", id))
		if err != nil {
			return nil, fmt.Errorf("Error parsing Entity ID %q for optimized WaitForEventFinished EventType %q: %s", id, entityType, err)
		}
		filterStruct["entity.id"] = filterableEntityID
		filterStruct["entity.type"] = entityType

		// TODO: are we conformatable with pages = 0 with the event type and id filter?
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	if deadline, ok := ctx.Deadline(); ok {
		duration := time.Until(deadline)
		log.Printf("[INFO] Waiting %d seconds for %s events since %v for %s %v", int(duration.Seconds()), action, minStart, titledEntityType, id)
	}

	ticker := time.NewTicker(client.millisecondsPerPoll * time.Millisecond)

	// avoid repeating log messages
	nextLog := ""
	lastLog := ""
	lastEventID := 0

	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if lastEventID > 0 {
				filterStruct["id"] = map[string]interface{}{
					"+gte": lastEventID,
				}
			}

			filter, err := json.Marshal(filterStruct)
			if err != nil {
				return nil, err
			}
			listOptions := NewListOptions(pages, string(filter))

			events, err := client.ListEvents(ctx, listOptions)
			if err != nil {
				return nil, err
			}

			// If there are events for this instance + action, inspect them
			for _, event := range events {
				event := event

				if event.Action != action {
					// log.Println("action mismatch", event.Action, action)
					continue
				}
				if event.Entity == nil || event.Entity.Type != entityType {
					// log.Println("type mismatch", event.Entity.Type, entityType)
					continue
				}

				var entID string

				switch id := event.Entity.ID.(type) {
				case float64, float32:
					entID = fmt.Sprintf("%.f", id)
				case int:
					entID = strconv.Itoa(id)
				default:
					entID = fmt.Sprintf("%v", id)
				}

				var findID string
				switch id := id.(type) {
				case float64, float32:
					findID = fmt.Sprintf("%.f", id)
				case int:
					findID = strconv.Itoa(id)
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

				// This is the event we are looking for. Save our place.
				if lastEventID == 0 {
					lastEventID = event.ID
				}

				switch event.Status {
				case EventFailed:
					return &event, fmt.Errorf("%s %v action %s failed", titledEntityType, id, action)
				case EventFinished:
					log.Printf("[INFO] %s %v action %s is finished", titledEntityType, id, action)
					return &event, nil
				}
				// TODO(displague) can we bump the ticker to TimeRemaining/2 (>=1) when non-nil?
				nextLog = fmt.Sprintf("[INFO] %s %v action %s is %s", titledEntityType, id, action, event.Status)
			}

			// de-dupe logging statements
			if nextLog != lastLog {
				log.Print(nextLog)
				lastLog = nextLog
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("Error waiting for Event Status '%s' of %s %v action '%s': %s", EventFinished, titledEntityType, id, action, ctx.Err())
		}
	}
}
