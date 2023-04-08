package proto

const (
	RegisterVoterSuccess	= 0
	RegisterVoterExists	= 1
	RegisterVoterUnknown	= 2

	UnregisterVoterSuccess	= 0
	UnregisterVoterNotFound	= 1
	UnregisterVoterUnknown	= 2

	CreateElectionSuccess	= 0
	CreateElectionUnauthn	= 1
	CreateElectionNoSpec	= 2
	CreateElectionUnknown	= 3

	CastVoteSuccess		= 0
	CastVoteUnauthn		= 1
	CastVoteNotFound	= 2
	CastVoteUnauthz		= 3
	CastVoteAlready		= 4

	GetResultSuccess	= 0
	GetResultNotFound	= 1
	GetResultNotYet		= 2
)
