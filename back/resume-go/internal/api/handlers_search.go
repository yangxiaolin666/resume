package api

import (
	"net/http"

	"resume-backend/internal/model"
)

func (a *API) searchStudents(w http.ResponseWriter, r *http.Request) {
	var query model.StudentQuery
	if err := decodeJSON(r, &query); err != nil {
		a.fail(w, err)
		return
	}
	page, err := a.Service.SearchStudents(query)
	if err != nil {
		a.fail(w, err)
		return
	}
	ok(w, page)
}

func (a *API) searchStudentDetails(w http.ResponseWriter, r *http.Request) {
	var query model.StudentQuery
	if err := decodeJSON(r, &query); err != nil {
		a.fail(w, err)
		return
	}
	page, err := a.Service.SearchStudentDetails(query)
	if err != nil {
		a.fail(w, err)
		return
	}
	ok(w, page)
}
