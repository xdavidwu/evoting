package proto

import "errors"

func RegisterVoterToError(s *Status) error {
	switch *s.Code {
	case RegisterVoterSuccess:
		return nil
	case RegisterVoterExists:
		return errors.New("Voter with the same name already exists")
	default:
		return errors.New("Undefined error")
	}
}

func UnregisterVoterToError(s *Status) error {
	switch *s.Code {
	case UnregisterVoterSuccess:
		return nil
	case UnregisterVoterNotFound:
		return errors.New("No voter with the name exists on the server")
	default:
		return errors.New("Undefined error")
	}
}

func CreateElectionToError(s *Status) error {
	switch *s.Code {
	case CreateElectionSuccess:
		return nil
	case CreateElectionUnauthn:
		return errors.New("Invalid authentication token")
	case CreateElectionNoSpec:
		return errors.New("Missing groups or choices specification")
	default:
		return errors.New("Unknown error")
	}
}

func CastVoteToError(s *Status) error {
	switch *s.Code {
	case CastVoteSuccess:
		return nil
	case CastVoteUnauthn:
		return errors.New("Invalid authentication token")
	case CastVoteNotFound:
		return errors.New("Invalid election name")
	case CastVoteUnauthz:
		return errors.New("The voter’s group is not allowed in the election")
	case CastVoteAlready:
		return errors.New("A previous vote has been cast")
	default:
		return errors.New("Undefined error")
	}
}

func GetResultToError(e *ElectionResult) (*ElectionResult, error) {
	switch *e.Status {
	case GetResultSuccess:
		return e, nil
	case GetResultNotFound:
		return nil, errors.New("Non-existent election")
	case GetResultNotYet:
		return nil, errors.New("The election is still ongoing. Election result is not available yet.")
	default:
		return nil, errors.New("Undefined error")
	}
}
