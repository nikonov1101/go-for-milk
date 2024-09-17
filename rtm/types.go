package rtm

import (
	"encoding/xml"
	"time"
)

const (
	methodGetFrom        = "rtm.auth.getFrob"
	methodGetAuthToken   = "rtm.auth.getToken"
	methodCheckAuthToken = "rtm.auth.checkToken"
	methodAddTask        = "rtm.tasks.add"
	methodListTasks      = "rtm.tasks.getList"

	authURL = "https://www.rememberthemilk.com/services/auth/"
	apiURL  = "https://api.rememberthemilk.com/services/rest/"
)

// Task is a minimal yet useful representation of RTM's task
type Task struct {
	ID          string
	Name        string
	Priority    int
	Tags        []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   time.Time
	CompletedAt time.Time
}

func (t Task) Visible() bool {
	return t.DeletedAt.IsZero() && t.CompletedAt.IsZero()
}

// stuff below are generated from
// the real API responses using famous
// https://xml-to-go.github.io

type frobResponse struct {
	XMLName xml.Name `xml:"rsp"`
	Text    string   `xml:",chardata"`
	Stat    string   `xml:"stat,attr"`
	Frob    string   `xml:"frob"`
}

type authTokenResponse struct {
	XMLName xml.Name `xml:"rsp"`
	Text    string   `xml:",chardata"`
	Stat    string   `xml:"stat,attr"`
	Auth    struct {
		Text  string `xml:",chardata"`
		Token string `xml:"token"`
		Perms string `xml:"perms"`
		User  struct {
			Text     string `xml:",chardata"`
			ID       string `xml:"id,attr"`
			Username string `xml:"username,attr"`
			Fullname string `xml:"fullname,attr"`
		} `xml:"user"`
	} `xml:"auth"`
}

type listTasksResponse struct {
	XMLName xml.Name `xml:"rsp"`
	Text    string   `xml:",chardata"`
	Stat    string   `xml:"stat,attr"`
	Tasks   struct {
		Text string `xml:",chardata"`
		Rev  string `xml:"rev,attr"`
		List []struct {
			Text       string `xml:",chardata"`
			ID         string `xml:"id,attr"`
			Taskseries []struct {
				Text       string `xml:",chardata"`
				ID         string `xml:"id,attr"`
				Created    string `xml:"created,attr"`
				Modified   string `xml:"modified,attr"`
				Name       string `xml:"name,attr"`
				Source     string `xml:"source,attr"`
				URL        string `xml:"url,attr"`
				LocationID string `xml:"location_id,attr"`
				Tags       struct {
					Text string   `xml:",chardata"`
					Tag  []string `xml:"tag"`
				} `xml:"tags"`
				Participants string `xml:"participants"`
				Notes        string `xml:"notes"`
				Task         []struct {
					Text       string `xml:",chardata"`
					ID         string `xml:"id,attr"`
					Due        string `xml:"due,attr"`
					HasDueTime string `xml:"has_due_time,attr"`
					Added      string `xml:"added,attr"`
					Completed  string `xml:"completed,attr"`
					Deleted    string `xml:"deleted,attr"`
					Priority   string `xml:"priority,attr"`
					Postponed  string `xml:"postponed,attr"`
					Estimate   string `xml:"estimate,attr"`
				} `xml:"task"`
			} `xml:"taskseries"`
		} `xml:"list"`
	} `xml:"tasks"`
}

func (x listTasksResponse) intoTasks() []Task {
	list := make([]Task, 0, 100)
	for _, ls := range x.Tasks.List {
		for _, ser := range ls.Taskseries {
			// we don't care about repeated tasks and their statuses (yet),
			// so inspect only the first item.
			meta := ser.Task[0]

			list = append(list, Task{
				ID:          ser.ID,
				Name:        ser.Name,
				Priority:    convertPriority(meta.Priority),
				Tags:        ser.Tags.Tag,
				CreatedAt:   timeFromXML(ser.Created),
				UpdatedAt:   timeFromXML(ser.Modified),
				DeletedAt:   timeFromXML(meta.Deleted),
				CompletedAt: timeFromXML(meta.Completed),
			})
		}
	}

	return list
}

func timeFromXML(s string) time.Time {
	if len(s) == 0 {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

// convert priorities to int to make it sortable
func convertPriority(p string) int {
	switch p {
	case "N":
		return 0
	case "1":
		return 1
	case "2":
		return 2
	case "3":
		return 3
	default:
		return 0
	}
}

type cachedToken struct {
	Token     string    `json:"token"`
	UpdatedAt time.Time `json:"updated_at"`
}
