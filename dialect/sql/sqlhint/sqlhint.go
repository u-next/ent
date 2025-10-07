package sqlhint

import (
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
)

// TODO: support hints by dialect, currently only for Spanner.

// HintKey represents a hint key type.
type HintKey interface {
	JoinHintKey | StatementHintKey | TableHintKey
}

// HintValue represents a hint value type.
type HintValue interface {
	~bool | ~string | ~int
}

// Hint represents a Graph query hint.
type Hint[K HintKey] map[K]any

// Query returns the hint representation.
func (h Hint[HintKey]) Write(b *sql.Builder) {
	b.WriteString("@")
	b.WrapBraces(func(b *sql.Builder) {
		i := 0
		for k, v := range h {
			if i > 0 {
				b.Comma()
			}
			var val string
			switch v := v.(type) {
			case bool:
				val = strings.ToUpper(fmt.Sprintf("%t", v))
			case string:
				val = v
			case int:
				val = fmt.Sprintf("%d", v)
			default:
				val = fmt.Sprintf("%v", v)
			}
			b.WriteString(fmt.Sprintf("%s=%s", k, val))
			i++
		}
	})
}

// TableHintKey represents a table hint key type
type TableHintKey string

// TableHint represents a table hint.
type TableHint Hint[TableHintKey]

func (h TableHint) Write(b *sql.Builder) {
	Hint[TableHintKey](h).Write(b)
}

// StatementHintKey represents a statement hint key type
type StatementHintKey string

// StatementHint represents a statement hint.
type StatementHint Hint[StatementHintKey]

func (h StatementHint) Write(b *sql.Builder) {
	Hint[StatementHintKey](h).Write(b)
}

// JoinHintKey represents a join hint key type
type JoinHintKey string

// JoinHint represents a join hint.
type JoinHint Hint[JoinHintKey]

func (h JoinHint) Write(b *sql.Builder) {
	Hint[JoinHintKey](h).Write(b)
}

// Statement hint keys
const (
	UseAdditionalParallelism        StatementHintKey = "USE_ADDITIONAL_PARALLELISM"
	OptimizerVersion                StatementHintKey = "OPTIMIZER_VERSION"
	OptimizerStatisticsPackage      StatementHintKey = "OPTIMIZER_STATISTICS_PACKAGE"
	AllowDistributedMerge           StatementHintKey = "ALLOW_DISTRIBUTED_MERGE"
	LockScannedRanges               StatementHintKey = "LOCK_SCANNED_RANGES"
	ScanMethod                      StatementHintKey = "SCAN_METHOD"
	ExecutionMethod                 StatementHintKey = "EXECUTION_METHOD"
	UseUnenforcedForeignKey         StatementHintKey = "USE_UNENFORCED_FOREIGN_KEY"
	AllowTimestampPredicatePushdown StatementHintKey = "ALLOW_TIMESTAMP_PREDICATE_PUSHDOWN"
)

// OptimizerVersionValue enum for OPTIMIZER_VERSION hint
type OptimizerVersionValue string

const (
	OptimizerVersionLatest  OptimizerVersionValue = "latest_version"
	OptimizerVersionDefault OptimizerVersionValue = "default_version"
)

// OptimizerStatisticsPackageValue enum for OPTIMIZER_STATISTICS_PACKAGE hint
type OptimizerStatisticsPackageValue string

const (
	OptimizerStatisticsPackageLatest OptimizerStatisticsPackageValue = "latest"
)

// LockScannedRangesValue enum for LOCK_SCANNED_RANGES hint
type LockScannedRangesValue string

const (
	LockScannedRangesExclusive LockScannedRangesValue = "exclusive"
	LockScannedRangesShared    LockScannedRangesValue = "shared"
)

// ScanMethodValue enum for SCAN_METHOD hint
type ScanMethodValue string

const (
	ScanMethodAuto  ScanMethodValue = "AUTO"
	ScanMethodBatch ScanMethodValue = "BATCH"
	ScanMethodRow   ScanMethodValue = "ROW"
)

// ExecutionMethodValue enum for EXECUTION_METHOD hint
type ExecutionMethodValue string

const (
	ExecutionMethodDefault ExecutionMethodValue = "DEFAULT"
	ExecutionMethodBatch   ExecutionMethodValue = "BATCH"
	ExecutionMethodRow     ExecutionMethodValue = "ROW"
)

// Table hint keys
const (
	ForceIndex              TableHintKey = "FORCE_INDEX"
	GroupbyScanOptimization TableHintKey = "GROUPBY_SCAN_OPTIMIZATION"
	TableScanMethod         TableHintKey = "SCAN_METHOD"
	IndexStrategy           TableHintKey = "INDEX_STRATEGY"
	SeekableKeySize         TableHintKey = "SEEKABLE_KEY_SIZE"
)

// ForceIndexValue enum for FORCE_INDEX hint
type ForceIndexValue string

const (
	ForceIndexBaseTable ForceIndexValue = "_BASE_TABLE"
)

// IndexStrategyValue enum for INDEX_STRATEGY hint
type IndexStrategyValue string

const (
	IndexStrategyForceIndexUnion IndexStrategyValue = "FORCE_INDEX_UNION"
)

// Join hint keys
const (
	ForceJoinOrder    JoinHintKey = "FORCE_JOIN_ORDER"
	JoinMethod        JoinHintKey = "JOIN_METHOD"
	HashJoinBuildSide JoinHintKey = "HASH_JOIN_BUILD_SIDE"
	BatchMode         JoinHintKey = "BATCH_MODE"
	HashJoinExecution JoinHintKey = "HASH_JOIN_EXECUTION"
)

// ForceJoinOrder enum for FORCE_JOIN_ORDER hint
type ForceJoinOrderValue string

const (
	ForceJoinOrderTrue  ForceJoinOrderValue = "TRUE"
	ForceJoinOrderFalse ForceJoinOrderValue = "FALSE"
)

// JoinMethod enum for JOIN_METHOD hint
type JoinMethodValue string

const (
	JoinMethodHashJoin              JoinMethodValue = "HASH_JOIN"
	JoinMethodApplyJoin             JoinMethodValue = "APPLY_JOIN"
	JoinMethodMergeJoin             JoinMethodValue = "MERGE_JOIN"
	JoinMethodPushBroadcastHashJoin JoinMethodValue = "PUSH_BROADCAST_HASH_JOIN"
)

// HashJoinBuildSide enum for HASH_JOIN_BUILD_SIDE hint
type HashJoinBuildSideValue string

const (
	HashJoinBuildSideLeft  HashJoinBuildSideValue = "BUILD_LEFT"
	HashJoinBuildSideRight HashJoinBuildSideValue = "BUILD_RIGHT"
)

// BatchMode enum for BATCH_MODE hint
type BatchModeValue string

const (
	BatchModeTrue  BatchModeValue = "TRUE"
	BatchModeFalse BatchModeValue = "FALSE"
)

// HashJoinExecution enum for HASH_JOIN_EXECUTION hint
type HashJoinExecutionValue string

const (
	HashJoinExecutionMultiPass HashJoinExecutionValue = "MULTI_PASS"
	HashJoinExecutionOnePass   HashJoinExecutionValue = "ONE_PASS"
)
