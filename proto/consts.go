package proto

const (
	RegisterVoterSuccess	int32 = 0
	RegisterVoterExists	int32 = 1
	RegisterVoterUnknown	int32 = 2

	UnregisterVoterSuccess	int32 = 0
	UnregisterVoterNotFound	int32 = 1
	UnregisterVoterUnknown	int32 = 2

	CreateElectionSuccess	int32 = 0
	CreateElectionUnauthn	int32 = 1
	CreateElectionNoSpec	int32 = 2
	CreateElectionUnknown	int32 = 3

	CastVoteSuccess		int32 = 0
	CastVoteUnauthn		int32 = 1
	CastVoteNotFound	int32 = 2
	CastVoteUnauthz		int32 = 3
	CastVoteAlready		int32 = 4

	GetResultSuccess	int32 = 0
	GetResultNotFound	int32 = 1
	GetResultNotYet		int32 = 2
)
