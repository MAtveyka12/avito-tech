package teams

type Team struct {
	Name    string
	Members []*TeamMember
}

type TeamMember struct {
	UserID   string
	Username string
	IsActive bool
}

