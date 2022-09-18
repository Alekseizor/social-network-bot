package model

type RequestBody struct {
	ParentUuid string `json:"parent_uuid"`
	Query      string `json:"query"`
	Type       string `json:"type"`
}

type RequestBodySchedule struct {
	ParentUuid string `json:"parent_uuid"`
}

type ResponseBody struct {
	Items []struct {
		Additional string `json:"additional"`
		Caption    string `json:"caption"`
		Type       string `json:"type"`
		Uuid       string `json:"uuid"`
	} `json:"items"`
	Total int `json:"total"`
}

type ResponseBodySchedule struct {
	Lessons []struct {
		Name     string `json:"name"`
		Cabinet  string `json:"cabinet"`
		Type     string `json:"type"`
		Teachers []struct {
			Name string `json:"name"`
		} `json:"teachers"`
		StartAt     string `json:"start_at"`
		EndAt       string `json:"end_at"`
		Day         int    `json:"day"`
		IsNumerator bool   `json:"is_numerator"`
	} `json:"lessons"`
}

type BitopBody struct {
	Token   string
	Request RequestBody
}

type Lesson struct {
	Name     string
	Cabinet  string
	Type     string
	Teachers []struct {
		Name string
	}
	StartAt     string
	EndAt       string
	Day         int
	IsNumerator bool
}
type Teacher struct {
	Name string
}
type Teachers []struct {
	Name string
}
