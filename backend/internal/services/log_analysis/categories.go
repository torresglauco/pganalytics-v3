package log_analysis

type LogCategory string

const (
	CategoryDatabaseError        LogCategory = "database_error"
	CategoryConnectionError      LogCategory = "connection_error"
	CategoryAuthenticationError  LogCategory = "authentication_error"
	CategorySyntaxError          LogCategory = "syntax_error"
	CategoryConstraintError      LogCategory = "constraint_error"
	CategorySlowQuery            LogCategory = "slow_query"
	CategoryCheckpoint           LogCategory = "checkpoint"
	CategoryVacuum               LogCategory = "vacuum"
	CategoryLongTransaction      LogCategory = "long_transaction"
	CategoryLockTimeout          LogCategory = "lock_timeout"
	CategoryDeadlock             LogCategory = "deadlock"
	CategoryReplicationError     LogCategory = "replication_error"
	CategoryWALError             LogCategory = "wal_error"
	CategoryOutOfMemory          LogCategory = "out_of_memory"
	CategoryDiskFull             LogCategory = "disk_full"
	CategoryWarning              LogCategory = "warning"
	CategoryInfo                 LogCategory = "info"
)

var LogCategoryPatterns = map[LogCategory][]string{
	CategoryDatabaseError: {
		"database .* does not exist",
		"FATAL: database .* does not exist",
	},
	CategoryConnectionError: {
		"connection refused",
		"Connection refused",
		"FATAL: could not accept SSL connection",
	},
	CategoryAuthenticationError: {
		"FATAL: no pg_hba.conf entry",
		"FATAL: password authentication failed",
		"role .* does not exist",
	},
	CategorySyntaxError: {
		"syntax error",
		"ERROR: syntax error",
	},
	CategoryConstraintError: {
		"constraint",
		"UNIQUE constraint",
		"FOREIGN KEY constraint",
	},
	CategorySlowQuery: {
		"duration: \\d+\\.\\d+ ms",
		"slow query",
	},
	CategoryCheckpoint: {
		"LOG: checkpoint",
		"FATAL: checkpoint failed",
	},
	CategoryVacuum: {
		"automatic vacuum",
		"VACUUM",
	},
	CategoryLongTransaction: {
		"long transaction",
		"transaction \\d+ is still in progress",
	},
	CategoryLockTimeout: {
		"lock timeout",
		"could not obtain lock",
	},
	CategoryDeadlock: {
		"deadlock detected",
		"Deadlock found",
	},
	CategoryReplicationError: {
		"replication",
		"standby",
	},
	CategoryWALError: {
		"WAL",
		"wal",
	},
	CategoryOutOfMemory: {
		"out of memory",
		"OOM",
	},
	CategoryDiskFull: {
		"disk full",
		"No space left on device",
	},
	CategoryWarning: {
		"WARNING",
	},
}
