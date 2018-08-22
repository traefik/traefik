package sarama

type PartitionMetadata struct {
	Err             KError
	ID              int32
	Leader          int32
	Replicas        []int32
	Isr             []int32
	OfflineReplicas []int32
}

func (pm *PartitionMetadata) decode(pd packetDecoder, version int16) (err error) {
	tmp, err := pd.getInt16()
	if err != nil {
		return err
	}
	pm.Err = KError(tmp)

	pm.ID, err = pd.getInt32()
	if err != nil {
		return err
	}

	pm.Leader, err = pd.getInt32()
	if err != nil {
		return err
	}

	pm.Replicas, err = pd.getInt32Array()
	if err != nil {
		return err
	}

	pm.Isr, err = pd.getInt32Array()
	if err != nil {
		return err
	}

	if version >= 5 {
		pm.OfflineReplicas, err = pd.getInt32Array()
		if err != nil {
			return err
		}
	}

	return nil
}

func (pm *PartitionMetadata) encode(pe packetEncoder, version int16) (err error) {
	pe.putInt16(int16(pm.Err))
	pe.putInt32(pm.ID)
	pe.putInt32(pm.Leader)

	err = pe.putInt32Array(pm.Replicas)
	if err != nil {
		return err
	}

	err = pe.putInt32Array(pm.Isr)
	if err != nil {
		return err
	}

	if version >= 5 {
		err = pe.putInt32Array(pm.OfflineReplicas)
		if err != nil {
			return err
		}
	}

	return nil
}

type TopicMetadata struct {
	Err        KError
	Name       string
	IsInternal bool // Only valid for Version >= 1
	Partitions []*PartitionMetadata
}

func (tm *TopicMetadata) decode(pd packetDecoder, version int16) (err error) {
	tmp, err := pd.getInt16()
	if err != nil {
		return err
	}
	tm.Err = KError(tmp)

	tm.Name, err = pd.getString()
	if err != nil {
		return err
	}

	if version >= 1 {
		tm.IsInternal, err = pd.getBool()
		if err != nil {
			return err
		}
	}

	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}
	tm.Partitions = make([]*PartitionMetadata, n)
	for i := 0; i < n; i++ {
		tm.Partitions[i] = new(PartitionMetadata)
		err = tm.Partitions[i].decode(pd, version)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tm *TopicMetadata) encode(pe packetEncoder, version int16) (err error) {
	pe.putInt16(int16(tm.Err))

	err = pe.putString(tm.Name)
	if err != nil {
		return err
	}

	if version >= 1 {
		pe.putBool(tm.IsInternal)
	}

	err = pe.putArrayLength(len(tm.Partitions))
	if err != nil {
		return err
	}

	for _, pm := range tm.Partitions {
		err = pm.encode(pe, version)
		if err != nil {
			return err
		}
	}

	return nil
}

type MetadataResponse struct {
	Version        int16
	ThrottleTimeMs int32
	Brokers        []*Broker
	ClusterID      *string
	ControllerID   int32
	Topics         []*TopicMetadata
}

func (r *MetadataResponse) decode(pd packetDecoder, version int16) (err error) {
	r.Version = version

	if version >= 3 {
		r.ThrottleTimeMs, err = pd.getInt32()
		if err != nil {
			return err
		}
	}

	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	r.Brokers = make([]*Broker, n)
	for i := 0; i < n; i++ {
		r.Brokers[i] = new(Broker)
		err = r.Brokers[i].decode(pd, version)
		if err != nil {
			return err
		}
	}

	if version >= 2 {
		r.ClusterID, err = pd.getNullableString()
		if err != nil {
			return err
		}
	}

	if version >= 1 {
		r.ControllerID, err = pd.getInt32()
		if err != nil {
			return err
		}
	} else {
		r.ControllerID = -1
	}

	n, err = pd.getArrayLength()
	if err != nil {
		return err
	}

	r.Topics = make([]*TopicMetadata, n)
	for i := 0; i < n; i++ {
		r.Topics[i] = new(TopicMetadata)
		err = r.Topics[i].decode(pd, version)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *MetadataResponse) encode(pe packetEncoder) error {
	err := pe.putArrayLength(len(r.Brokers))
	if err != nil {
		return err
	}
	for _, broker := range r.Brokers {
		err = broker.encode(pe, r.Version)
		if err != nil {
			return err
		}
	}

	if r.Version >= 1 {
		pe.putInt32(r.ControllerID)
	}

	err = pe.putArrayLength(len(r.Topics))
	if err != nil {
		return err
	}
	for _, tm := range r.Topics {
		err = tm.encode(pe, r.Version)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *MetadataResponse) key() int16 {
	return 3
}

func (r *MetadataResponse) version() int16 {
	return r.Version
}

func (r *MetadataResponse) requiredVersion() KafkaVersion {
	switch r.Version {
	case 1:
		return V0_10_0_0
	case 2:
		return V0_10_1_0
	case 3, 4:
		return V0_11_0_0
	case 5:
		return V1_0_0_0
	default:
		return MinVersion
	}
}

// testing API

func (r *MetadataResponse) AddBroker(addr string, id int32) {
	r.Brokers = append(r.Brokers, &Broker{id: id, addr: addr})
}

func (r *MetadataResponse) AddTopic(topic string, err KError) *TopicMetadata {
	var tmatch *TopicMetadata

	for _, tm := range r.Topics {
		if tm.Name == topic {
			tmatch = tm
			goto foundTopic
		}
	}

	tmatch = new(TopicMetadata)
	tmatch.Name = topic
	r.Topics = append(r.Topics, tmatch)

foundTopic:

	tmatch.Err = err
	return tmatch
}

func (r *MetadataResponse) AddTopicPartition(topic string, partition, brokerID int32, replicas, isr []int32, err KError) {
	tmatch := r.AddTopic(topic, ErrNoError)
	var pmatch *PartitionMetadata

	for _, pm := range tmatch.Partitions {
		if pm.ID == partition {
			pmatch = pm
			goto foundPartition
		}
	}

	pmatch = new(PartitionMetadata)
	pmatch.ID = partition
	tmatch.Partitions = append(tmatch.Partitions, pmatch)

foundPartition:

	pmatch.Leader = brokerID
	pmatch.Replicas = replicas
	pmatch.Isr = isr
	pmatch.Err = err

}
