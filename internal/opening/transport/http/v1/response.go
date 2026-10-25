package v1

import "openingjobs/internal/opening/domain"

type openingResponse struct {
	ID       string `json:"id"`
	Role     string `json:"role"`
	Company  string `json:"company"`
	Location string `json:"location"`
	Remote   bool   `json:"remote"`
	Link     string `json:"link"`
	Salary   int64  `json:"salary"`
}

func toResponse(o domain.Opening) openingResponse {
	return openingResponse{ID: o.ID.String(), Role: o.Role, Company: o.Company, Location: o.Location, Remote: o.Remote, Link: o.Link, Salary: o.Salary}
}
