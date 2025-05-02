package personio

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type TimecardResponse struct {
	Timecards             []Timecard `json:"timecards"`
	Widgets               Widgets    `json:"widgets"`
	SupervisorPersonID    string     `json:"supervisor_person_id"`
	OwnerHasProposeRights bool       `json:"owner_has_propose_rights"`
}

type Timecard struct {
	DayID                      *uuid.UUID  `json:"day_id"`
	Date                       string      `json:"date"`
	State                      string      `json:"state"`
	IsOffDay                   bool        `json:"is_off_day"`
	AssignedWorkingScheduleID  string      `json:"assigned_working_schedule_id"`
	Periods                    []Period    `json:"periods"`
	BreakDurationMinutes       int         `json:"break_duration_minutes"`
	LegacyBreakDurationMinutes int         `json:"legacy_break_duration_minutes"`
	TargetHours                TargetHours `json:"target_hours"`
	Overtime                   *Overtime   `json:"overtime"`
	TimeOff                    *TimeOff    `json:"time_off"`
	Approval                   *Approval   `json:"approval"`
	Alerts                     []any       `json:"alerts"` // Empty array, unknown structure
	CanCreateTimeOffInLieu     bool        `json:"can_create_time_off_in_lieu"`
}

// PersonioTime is a wrapper around time.Time
// that marshals / unmarshals into JSON as YYYY-MM-DDTHH:MM:SS
type PersonioTime struct {
	time.Time
}

func (p *PersonioTime) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	t, err := time.Parse("2006-01-02T15:04:05", str)
	if err != nil {
		return err
	}
	p.Time = t
	return nil
}

func (p PersonioTime) MarshalJSON() ([]byte, error) {
	str := p.Time.Format("2006-01-02T15:04:05")
	return json.Marshal(str)
}

type Period struct {
	ID        uuid.UUID    `json:"id"`
	Start     PersonioTime `json:"start"`
	End       PersonioTime `json:"end"`
	ProjectID *int         `json:"project_id"`
	Comment   *string      `json:"comment"`
	Type      PeriodType   `json:"type"`
}

func (p Period) GetComment() string {
	if p.Comment == nil {
		return ""
	}
	return *p.Comment
}

func (p Period) GetProjectID() int {
	if p.ProjectID == nil {
		return 0
	}
	return *p.ProjectID
}

type PeriodType string

const (
	PeriodTypeWork  PeriodType = "work"
	PeriodTypeBreak PeriodType = "break"
)

type TargetHours struct {
	EffectiveWorkDurationMinutes    int     `json:"effective_work_duration_minutes"`
	EffectiveBreakDurationMinutes   *int    `json:"effective_break_duration_minutes"`
	ContractualWorkDurationMinutes  int     `json:"contractual_work_duration_minutes"`
	ContractualBreakDurationMinutes int     `json:"contractual_break_duration_minutes"`
	StartTime                       *string `json:"start_time"`
	EndTime                         *string `json:"end_time"`
}

type Approval struct {
	Status          string  `json:"status"`
	RejectionReason *string `json:"rejection_reason"`
}

type TimeOff struct {
	AggregatedDurationMinutes int           `json:"aggregated_duration_minutes"`
	Items                     []TimeOffItem `json:"items"`
}

type TimeOffItem struct {
	Type                     string     `json:"type"`
	Name                     string     `json:"name"`
	DurationMinutes          *int       `json:"duration_minutes"`
	Status                   *string    `json:"status"`
	CreatedAt                *time.Time `json:"created_at"`
	Color                    *string    `json:"color"`
	ColorFamily              *string    `json:"color_family"`
	MisconfiguredAbsenceType *bool      `json:"misconfigured_absence_type"`
	IsOffsiteWorkAbsence     *bool      `json:"is_offsite_work_absence"`
}

type Overtime struct {
	// All fields are nullable or optional in examples, so using pointers
	OvertimeMinutes      *int `json:"overtime_minutes"`
	PendingMinutes       *int `json:"pending_minutes"`
	CliffMinutes         *int `json:"cliff_minutes"`
	TotalOvertimeMinutes *int `json:"total_overtime_minutes"`
}

// Widgets Section

type Widgets struct {
	TrackedHours         TrackedHours          `json:"tracked_hours"`
	Overtime             WidgetOvertime        `json:"overtime"`
	TimeOff              WidgetTimeOff         `json:"time_off"`
	WorkingScheduleWeeks []WorkingScheduleWeek `json:"working_schedule_weeks"`
}

type TrackedHours struct {
	TrackedMinutes   int `json:"tracked_minutes"`
	ConfirmedMinutes int `json:"confirmed_minutes"`
	TargetMinutes    int `json:"target_minutes"`
	PendingMinutes   int `json:"pending_minutes"`
}

type WidgetOvertime struct {
	OvertimeMinutes      int  `json:"overtime_minutes"`
	PendingMinutes       *int `json:"pending_minutes"`
	CliffMinutes         int  `json:"cliff_minutes"`
	TotalOvertimeMinutes int  `json:"total_overtime_minutes"`
}

type WidgetTimeOff struct {
	ConfirmedMinutes int `json:"confirmed_minutes"`
	PendingMinutes   int `json:"pending_minutes"`
}

type WorkingScheduleWeek struct {
	WeeklyWorkMinutes   int           `json:"weekly_work_minutes"`
	WeeklyHourMismatch  bool          `json:"weekly_hour_mismatch"`
	WeekNumber          int           `json:"week_number"`
	IsCurrentActiveWeek bool          `json:"is_current_active_week"`
	Days                []ScheduleDay `json:"days"`
}

type ScheduleDay struct {
	DayOfWeek         int `json:"day_of_week"`
	TargetWorkMinutes int `json:"target_work_minutes"`
}

type SetAttendanceDayRequest struct {
	Periods    []RequestPeriod `json:"periods"`
	EmployeeID int             `json:"employee_id"`
}

type RequestPeriod struct {
	ID        uuid.UUID    `json:"id"`
	Start     PersonioTime `json:"start"`
	End       PersonioTime `json:"end"`
	ProjectID *int         `json:"project_id"`
	Comment   *string      `json:"comment"`
	Type      PeriodType   `json:"period_type"`
}
