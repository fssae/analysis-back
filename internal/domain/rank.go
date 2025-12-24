package domain

type RankRequest struct {
	ClassName  string `json:"className" form:"className"`
	CourseName string `json:"courseName" form:"courseName"`
	Sort       string `json:"sort" form:"sort"`
	Order      string `json:"order" form:"order"`
	Page       int    `json:"page" form:"page"`
	PageSize   int    `json:"pageSize" form:"pageSize"`
}
type RankItem struct {
	List  []*Rank `json:"list"`
	Total int64   `json:"total"`
}
type Rank struct {
	ClassName  string  `json:"className"`
	CourseName string  `json:"courseName"`
	FocusAvg   float64 `json:"focusAvg"`
}
