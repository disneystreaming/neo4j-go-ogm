package gogm

//VersionParts represents the part of a version string that is affected by a given change.
//
//Based on format: 1.0.0
type VersionParts int //@name VersionParts
//Enum represents the part of a version string that is affected by a given change.
const (
	Major VersionParts = 0
	Minor VersionParts = 1
	Patch VersionParts = 2
)

//DataOperationType represents the type of data operation neo4j should perform.
//
type DataOperationType int //@name VersionParts
//Enum represents the part of a version string that is affected by a given change.
const (
	Upsert  DataOperationType = 0
	Delete  DataOperationType = 1
	Unknown DataOperationType = 3
)

type LocateTarget string

const (
	StartNode LocateTarget = `startNode`
	EndNode   LocateTarget = `endNode`
	Entity    LocateTarget = `entity`
)

//LookupValueSource represents the origin of a lookup value.
type LookupValueSource string

const (
	From   = `from`   //Represents a value that was provided by the From (start) node on a relationship.
	To     = `to`     //Represents a value that was provided by the To (end) node on a relationship.
	Unique = `unique` //Represents a value that should be referenced as a standalone value. Either for individual entities or rich relationships
)

//Interface definition allowing for conversion between entities and unified reference behavior
//for all implementers.
type Member interface {
	GetAndTrySetInnerRef(any) any
	CopyDataToMember(data map[string]any) Member
	TransitionMemberState(*[]Member)
	GetLookupValue() (int, map[LookupValueSource][]string)
	GetLookupKey() *string
	Persist(session *SessionImpl) error
	GetDescriptor() *string
	IsDirty() bool
	SetIsDirty(value bool)
	Clean()
	GetDataOperationType() DataOperationType
	RequiredVersionUpdate() *VersionParts
	GetID() *int64
	Locate(key string, value string, targetComponent LocateTarget) string
	GetSuppressedVersionUpdates() bool
}
