package user

const (
	MaxProjects         = 3
	AdvancedMaxProjects = 6
)

type User struct {
	Id       int
	Email    string
	Password string
	Advanced bool
}

func (u *User) GetMaxAllowedProjects() int {
	if u.Advanced {
		return AdvancedMaxProjects
	}

	return MaxProjects
}
