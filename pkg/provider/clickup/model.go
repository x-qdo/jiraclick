package clickup

import "regexp"

type CustomFieldKey string

const (
	ApprovedBy       CustomFieldKey = "065a1567-0655-4a7e-aefe-e179f7983069"
	BillableHours    CustomFieldKey = "074e1387-e7b8-41c6-92db-fbada8f8486c"
	JiraLink         CustomFieldKey = "349fbec4-f71f-4cee-9861-c112e253a6e1"
	SlackLink        CustomFieldKey = "517a450f-ce8b-4683-b34a-616d5c3b0fb4"
	DoneNotification CustomFieldKey = "86477c9c-b494-423b-8dc0-3a49734b8b28"
	Synced           CustomFieldKey = "926d35a9-5f70-4f54-bc07-d11b82d4cf21"
	RequestedBy      CustomFieldKey = "eb30f61c-dbad-4ad4-896d-15d2a239cb69"
)

type TaskStatus string

type Task struct {
	ID           string        `json:"id"`
	CustomID     string        `json:"custom_id,omitempty"`
	Name         string        `json:"name"`
	Description  string        `json:"description,omitempty"`
	Status       TaskStatus    `json:"status.status"`
	DateCreated  string        `json:"date_created"`
	DateUpdated  string        `json:"date_updated"`
	DateClosed   interface{}   `json:"date_closed,omitempty"`
	Creator      User          `json:"creator"`
	Priority     interface{}   `json:"priority,omitempty"`
	DueDate      interface{}   `json:"due_date,omitempty"`
	StartDate    interface{}   `json:"start_date,omitempty"`
	TimeEstimate interface{}   `json:"time_estimate,omitempty"`
	TimeSpent    interface{}   `json:"time_spent,omitempty"`
	URL          string        `json:"url"`
	Archived     bool          `json:"archived"`
	TeamID       string        `json:"team_id"`
	CustomFields []CustomField `json:"custom_fields,omitempty"`
	Assignees    []User        `json:"assignees"`

	List struct {
		ID string `json:"id"`
	} `json:"list"`
	Folder struct {
		ID string `json:"id"`
	} `json:"folder"`
	Space struct {
		ID string `json:"id"`
	} `json:"space"`
}

type User struct {
	ID             int    `json:"id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	Color          string `json:"color,omitempty"`
	Initials       string `json:"initials,omitempty"`
	ProfilePicture string `json:"profilePicture,omitempty"`
}

type CustomField struct {
	ID    CustomFieldKey `json:"id"`
	Name  string         `json:"name"`
	Value interface{}    `json:"value"`
}

func (t *Task) GetSlackChannel() string {
	for _, field := range t.CustomFields {
		if field.ID == SlackLink {
			link, ok := field.Value.(string)
			if !ok {
				return ""
			}
			reg := regexp.MustCompile(`.*archives/(\w+)/.*`)
			result := reg.FindStringSubmatch(link)
			if len(result) == 2 {
				return result[1]
			}
		}
	}
	return ""
}
