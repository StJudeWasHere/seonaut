package user

import (
	"testing"
)

func TestMaxAllowedProjects(t *testing.T) {
	regularUser := User{Advanced: false}
	if regularUser.GetMaxAllowedProjects() != 3 {
		t.Error("regularUser.GetMaxAllowedProjects() != 3")
	}

	advancedUser := User{Advanced: true}
	if advancedUser.GetMaxAllowedProjects() != 6 {
		t.Error("advancedUser.GetMaxAllowedProjects() != 6")
	}
}
