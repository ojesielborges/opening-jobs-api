package v1

import "openingjobs/internal/business/opening/domain"

type openingRequest struct {
	Role     string `json:"role" binding:"required,max=200"`
	Company  string `json:"company" binding:"required,max=200"`
	Location string `json:"location" binding:"max=200"`
	Remote   bool   `json:"remote"`
	Link     string `json:"link" binding:"max=2048"`
	Salary   int64  `json:"salary"`
}

func (r openingRequest) toDomain() domain.Opening {
	return domain.Opening{Role: r.Role, Company: r.Company, Location: r.Location, Remote: r.Remote, Link: r.Link, Salary: r.Salary}
}
