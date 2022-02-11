package personio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type WorkingTimes []struct {
	ID         string      `json:"id"`
	EmployeeID int         `json:"employee_id"`
	Start      time.Time   `json:"start"`
	End        time.Time   `json:"end"`
	ActivityID interface{} `json:"activity_id"`
	Comment    string      `json:"comment"`
	ProjectID  interface{} `json:"project_id"`
}

type Personio struct {
	Username   string
	Password   string
	baseURL    string
	client     http.Client
	EmployeeID int
}

func NewPersonio(baseURL, user, pwd string) *Personio {
	p := &Personio{baseURL: baseURL, Username: user, Password: pwd}
	options := cookiejar.Options{}
	jar, err := cookiejar.New(&options)
	if err != nil {
		log.Fatal(err)
	}
	p.client = http.Client{Jar: jar}
	return p
}

func (p *Personio) LoginToPersonio() {
	params := url.Values{}
	params.Add("email", p.Username)
	params.Add("password", p.Password)
	body := strings.NewReader(params.Encode())

	log.Println("Login to personio....")
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/login/index", p.baseURL), body)
	if err != nil {
		// handle err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := p.client.PostForm(fmt.Sprintf("%slogin/index", p.baseURL), params)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Login response ", resp.StatusCode)

	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	//log.Println(string(data)) // print whole html of user profile data
	re, _ := regexp.Compile(`EMPLOYEE.*{.*id:\s*(\d+),`)
	res := re.FindStringSubmatch(strings.ReplaceAll(string(data), "\n", ""))
	if len(res) > 1 {
		p.EmployeeID, _ = strconv.Atoi(res[1])
		log.Printf("Found Employee ID %d", p.EmployeeID)
	}

}

func (p *Personio) SetWorkingTimes(from, to time.Time) {
	path := p.baseURL + "api/v1/attendances/periods"

	type Payload []struct {
		ID         string      `json:"id"`
		EmployeeID int         `json:"employee_id"`
		Start      string      `json:"start"`
		End        string      `json:"end"`
		ActivityID interface{} `json:"activity_id"`
		Comment    string      `json:"comment"`
		ProjectID  interface{} `json:"project_id"`
	}

	data := Payload{
		{
			ID:         uuid.New().String(),
			EmployeeID: p.EmployeeID,
			Start:      from.Format("2006-01-02T15:04:05Z"),
			End:        to.Format("2006-01-02T15:04:05Z"),
		},
	}

	payloadBytes, _ := json.Marshal(data)
	body := bytes.NewReader(payloadBytes)

	req, _ := http.NewRequest("POST", path, body)
	req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Accept", "application/json, text/plain, */*")

	response, err := p.client.Do(req)
	if err != nil {
		log.Printf("cannot post workingtimes %v\n", err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Printf("Received %d response code %s", response.StatusCode, path)
	}

	dataRes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("cannot read body %v\n", err)
	}
	//	log.Println(string(dataRes))
	var res PersonioPeriodsResult
	json.Unmarshal(dataRes, &res)
	//	pretty.Println(res)
	if !res.Success {
		log.Printf("Error %s", res.Error.Message)
	}
}

type PersonioPeriodsResult struct {
	Success bool `json:"success"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Details struct {
			OverlappingExisting []struct {
				ID         string `json:"id"`
				Type       string `json:"type"`
				Attributes struct {
					LegacyID       interface{} `json:"legacy_id"`
					LegacyStatus   string      `json:"legacy_status"`
					Start          time.Time   `json:"start"`
					End            time.Time   `json:"end"`
					Comment        string      `json:"comment"`
					LegacyBreakMin int         `json:"legacy_break_min"`
					Origin         string      `json:"origin"`
					CreatedAt      time.Time   `json:"created_at"`
					UpdatedAt      time.Time   `json:"updated_at"`
					DeletedAt      interface{} `json:"deleted_at"`
				} `json:"attributes"`
				Relationships struct {
					Project struct {
						Data struct {
							ID interface{} `json:"id"`
						} `json:"data"`
					} `json:"project"`
					Employee struct {
						Data struct {
							ID int `json:"id"`
						} `json:"data"`
					} `json:"employee"`
					Company struct {
						Data struct {
							ID int `json:"id"`
						} `json:"data"`
					} `json:"company"`
					AttendanceDay struct {
						Data struct {
							ID string `json:"id"`
						} `json:"data"`
					} `json:"attendance_day"`
					CreatedBy struct {
						Data struct {
							ID int `json:"id"`
						} `json:"data"`
					} `json:"created_by"`
					UpdatedBy struct {
						Data struct {
							ID int `json:"id"`
						} `json:"data"`
					} `json:"updated_by"`
				} `json:"relationships"`
			} `json:"overlapping_existing"`
		} `json:"details"`
	} `json:"error"`

	Data []struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			LegacyID       int         `json:"legacy_id"`
			LegacyStatus   string      `json:"legacy_status"`
			Start          time.Time   `json:"start"`
			End            time.Time   `json:"end"`
			Comment        string      `json:"comment"`
			LegacyBreakMin int         `json:"legacy_break_min"`
			Origin         string      `json:"origin"`
			CreatedAt      time.Time   `json:"created_at"`
			UpdatedAt      time.Time   `json:"updated_at"`
			DeletedAt      interface{} `json:"deleted_at"`
		} `json:"attributes"`
		Relationships struct {
			Project struct {
				Data struct {
					ID interface{} `json:"id"`
				} `json:"data"`
			} `json:"project"`
			Employee struct {
				Data struct {
					ID int `json:"id"`
				} `json:"data"`
			} `json:"employee"`
			Company struct {
				Data struct {
					ID int `json:"id"`
				} `json:"data"`
			} `json:"company"`
			AttendanceDay struct {
				Data struct {
					ID string `json:"id"`
				} `json:"data"`
			} `json:"attendance_day"`
			CreatedBy struct {
				Data struct {
					ID int `json:"id"`
				} `json:"data"`
			} `json:"created_by"`
			UpdatedBy struct {
				Data struct {
					ID int `json:"id"`
				} `json:"data"`
			} `json:"updated_by"`
		} `json:"relationships"`
	} `json:"data"`
}

type PersonioPeriods struct {
	Success bool `json:"success"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`

	Data []struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Start           time.Time `json:"start"`
			End             time.Time `json:"end"`
			LegacyBreakMin  int       `json:"legacy_break_min"`
			Comment         string    `json:"comment"`
			PeriodType      string    `json:"period_type"`
			CreatedAt       time.Time `json:"created_at"`
			UpdatedAt       time.Time `json:"updated_at"`
			EmployeeID      int       `json:"employee_id"`
			CreatedBy       int       `json:"created_by"`
			AttendanceDayID string    `json:"attendance_day_id"`
		} `json:"attributes"`
	} `json:"data"`
}

func (p *Personio) GetWorkingTimes(from, to time.Time) *PersonioPeriods {
	path := p.baseURL + "api/v1/attendances/periods"

	req, _ := http.NewRequest("GET", path, nil)
	req.Header.Set("accept", "application/json")
	//req.Header.Set("Accept", "application/json, text/plain, */*")

	//?filter[startDate]=2022-01-31&filter[endDate]=2022-03-06&filter[employee]=991824
	q := req.URL.Query()
	q.Add("filter[startDate]", from.Format("2006-01-02"))
	q.Add("filter[endDate]", to.Format("2006-01-02"))
	q.Add("filter[employee]", fmt.Sprintf("%d", p.EmployeeID))
	req.URL.RawQuery = q.Encode()

	response, err := p.client.Do(req)
	if err != nil {
		log.Printf("cannot get workingtimes %v\n", err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Printf("Received %d response code %s", response.StatusCode, path)
	}

	dataRes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("cannot read body %v\n", err)
	}
	//log.Println(string(dataRes))
	var res PersonioPeriods
	json.Unmarshal(dataRes, &res)
	for k, _ := range res.Data {
		res.Data[k].Attributes.Start = ConvertTime(res.Data[k].Attributes.Start)
		res.Data[k].Attributes.End = ConvertTime(res.Data[k].Attributes.End)
		res.Data[k].Attributes.CreatedAt = ConvertTime(res.Data[k].Attributes.CreatedAt)
		res.Data[k].Attributes.UpdatedAt = ConvertTime(res.Data[k].Attributes.UpdatedAt)
	}
	//pretty.Println(res)
	if !res.Success {
		log.Printf("Error %s", res.Error.Message)
	}
	return &res
}

func ConvertTime(t time.Time) time.Time {
	startTime := t
	loc, _ := time.LoadLocation(time.Now().Location().String())
	return time.Date(startTime.Year(), startTime.Month(), startTime.Day(), startTime.Hour(), startTime.Minute(), startTime.Second(), startTime.Nanosecond(), loc)
}
