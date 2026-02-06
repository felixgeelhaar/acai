package meeting

// ParticipantRole represents a participant's role in a meeting.
type ParticipantRole string

const (
	RoleHost     ParticipantRole = "host"
	RoleAttendee ParticipantRole = "attendee"
)

// Participant is an immutable value object representing a meeting attendee.
// Value objects are compared by their attribute values, not by identity.
type Participant struct {
	name  string
	email string
	role  ParticipantRole
}

func NewParticipant(name, email string, role ParticipantRole) Participant {
	return Participant{
		name:  name,
		email: email,
		role:  role,
	}
}

func (p Participant) Name() string          { return p.name }
func (p Participant) Email() string         { return p.email }
func (p Participant) Role() ParticipantRole { return p.role }

func (p Participant) Equals(other Participant) bool {
	return p.name == other.name && p.email == other.email && p.role == other.role
}
