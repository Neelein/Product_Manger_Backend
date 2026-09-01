package http

import (
	"os"
	"strings"
	"testing"
)

func TestAdministrativeRoutesUseEmployeeBoundary(t *testing.T) {
	source, err := os.ReadFile("router.go")
	if err != nil {
		t.Fatal(err)
	}
	s := string(source)
	for _, route := range []string{
		"h.CreateCategory", "h.UpdateCategory", "h.DeleteCategory",
		"h.CreateAnnouncement", "h.UpdateAnnouncement", "h.DeleteAnnouncement",
		"h.CreateEvent", "h.UpdateEvent", "h.DeleteEvent", "h.AddViewer", "h.RemoveViewer",
		"h.CreateRoom", "h.UpdateRoom", "h.DeleteRoom", "h.AddMembers", "h.RemoveMember", "h.SendMessage",
		"h.DeleteImage",
	} {
		if !strings.Contains(s, "employeeAuth(http.HandlerFunc("+route+"))") {
			t.Errorf("%s is not protected by employee authorization", route)
		}
	}
}
