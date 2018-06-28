package ext

const (
	// CassandraQuery is the tag name used for cassandra queries.
	CassandraQuery = "cassandra.query"

	// CassandraConsistencyLevel is the tag name to set for consitency level.
	CassandraConsistencyLevel = "cassandra.consistency_level"

	// CassandraCluster specifies the tag name that is used to set the cluster.
	CassandraCluster = "cassandra.cluster"

	// CassandraRowCount specifies the tag name to use when settings the row count.
	CassandraRowCount = "cassandra.row_count"

	// CassandraKeyspace is used as tag name for setting the key space.
	CassandraKeyspace = "cassandra.keyspace"

	// CassandraPaginated specifies the tag name for paginated queries.
	CassandraPaginated = "cassandra.paginated"
)
