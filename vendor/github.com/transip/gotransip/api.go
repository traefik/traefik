package gotransip

// CancellationTime represents the possible ways of canceling a contract
type CancellationTime string

var (
	// CancellationTimeEnd specifies to cancel the contract when the contract was
	// due to end anyway
	CancellationTimeEnd CancellationTime = "end"
	// CancellationTimeImmediately specifies to cancel the contract immediately
	CancellationTimeImmediately CancellationTime = "immediately"
)
