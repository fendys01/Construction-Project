package handler

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"panorama/lib/array"
	"panorama/lib/psql"
	"panorama/services/api/handler/request"
	"panorama/services/api/handler/response"
	"panorama/services/api/model"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

// GetUserListAct ...
func (h *Contract) GetUserListAct(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	param := map[string]interface{}{
		"keyword": "",
		"page":    1,
		"limit":   10,
		"offset":  0,
		"sort":    "desc",
		"order":   "u.name",
		"role":    "admin",
	}

	if page, ok := r.URL.Query()["page"]; ok && len(page[0]) > 0 {
		if p, err := strconv.Atoi(page[0]); err == nil && p > 1 {
			param["page"] = p
		}
	}

	if sort, ok := r.URL.Query()["sort"]; ok && len(sort[0]) > 0 && strings.ToLower(sort[0]) == "asc" {
		param["sort"] = "asc"
	}

	if order, ok := r.URL.Query()["order"]; ok && len(order[0]) > 0 {
		arrStr := new(array.ArrStr)
		if exist, _ := arrStr.InArray(order[0], []string{"id"}); exist {
			param["order"] = order[0]
		}
	}

	if keyword, ok := r.URL.Query()["keyword"]; ok && len(keyword[0]) > 0 {
		param["keyword"] = keyword[0]
	}

	if role, ok := r.URL.Query()["role"]; ok && len(role[0]) > 0 {
		param["role"] = role[0]
	}

	if limit, ok := r.URL.Query()["limit"]; ok {
		if l, err := strconv.Atoi(limit[0]); err == nil {
			param["limit"] = l
		}
	}

	param["offset"] = (param["page"].(int) - 1) * param["limit"].(int)

	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	m := model.Contract{App: h.App}
	if len(code) > 0 {
		// return single user (get by user code)
		u, err := m.GetUserByCode(db, ctx, code)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}

		var res response.UsersResponse
		res = res.Transform(u)

		h.SendSuccess(w, res, nil)
		return
	}

	u, err := m.GetUser(db, ctx, param)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	var listResponse []response.UsersResponse
	for _, a := range u {
		var res response.UsersResponse
		res = res.Transform(a)
		listResponse = append(listResponse, res)
	}

	h.SendSuccess(w, listResponse, param)
}

// GetDetailAdminAndTc ...
func (h *Contract) GetDetailAdminAndTc(w http.ResponseWriter, r *http.Request) {

	code := chi.URLParam(r, "code")

	param := map[string]interface{}{
		"role": "",
	}

	if role, ok := r.URL.Query()["role"]; ok && len(role[0]) > 0 {
		param["role"] = role[0]
	}

	if len(param["role"].(string)) <= 0 {
		h.SendBadRequest(w, "Parameter Role Required")
		return
	}

	if role, ok := r.URL.Query()["role"]; ok && role[0] == "admin" || role[0] == "tc" {
		if ok == false {
			h.SendBadRequest(w, "Role parameters between tc or admin")
			return
		}
	}

	if param["role"].(string) == "admin" {

		// get detail activity admin

		ctx := context.Background()
		db, err := h.DB.Acquire(ctx)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}
		defer db.Release()

		m := model.Contract{App: h.App}
		u, err := m.GetDetailAdminByCode(db, ctx, code)
		if u.ID == 0 {
			h.SendNotfound(w, fmt.Sprintf("Admin with code %s not found.", code))
			return
		}
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}

		log, err := m.GetListLogActivity(db, ctx, param["role"].(string), code)
		if err != nil && err != sql.ErrNoRows {
			h.SendBadRequest(w, err.Error())
			return
		}

		u.LogActivityUser = log

		var res response.AdminDetailResponse
		res = res.Transform(u)

		h.SendSuccess(w, res, param)

	} else if param["role"].(string) == "tc" {

		ctx := context.Background()
		db, err := h.DB.Acquire(ctx)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}
		defer db.Release()

		// get detail activity tc
		m := model.Contract{App: h.App}
		u, err := m.GetDetailTcByCode(db, ctx, code)
		if u.ID == 0 {
			h.SendNotfound(w, fmt.Sprintf("Tc with code %s not found.", code))
			return
		}
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}

		// get list log activity
		log, err := m.GetListLogActivity(db, ctx, param["role"].(string), code)
		if err != nil && err != sql.ErrNoRows {
			h.SendBadRequest(w, err.Error())
			return
		}
		u.LogActivityUser = log

		// get id user
		user, err := m.GetUserByCode(db, ctx, code)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}
		u.Email = user.Email

		// get list active client
		act, err := m.GetActiveClient(db, ctx, user.ID)
		if err != nil && err != sql.ErrNoRows {
			h.SendBadRequest(w, err.Error())
			return
		}
		u.ActiveClientConsultan = act

		var res response.TcDetailResponse
		res = res.Transform(u)

		h.SendSuccess(w, res, param)

	}

}

// AddUserAct ...
func (h *Contract) AddUserAct(w http.ResponseWriter, r *http.Request) {
	var err error
	req := request.NewUserReq{}
	if err = h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	if err = h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	pass, err := bcrypt.GenerateFromPassword([]byte(req.Pass), 10)
	if err != nil {
		h.SendBadRequest(w, "Error when generate password")
		return
	}

	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	m := model.Contract{App: h.App}
	tx, err := db.Begin(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	user, err := m.AddUser(tx, ctx, model.UserEnt{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: string(pass),
		Role:     req.Role,
		Img:      sql.NullString{String: req.Img, Valid: true},
		IsActive: req.IsActive,
	})
	if err != nil {
		h.SendBadRequest(w, psql.ParseErr(err))
		return
	}

	err = m.AddNewLogVisitApp(tx, ctx, user.ID, req.Role)
	if err != nil {
		tx.Rollback(ctx)
		h.SendBadRequest(w, err.Error())
		return
	}
	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	var res response.UsersResponse
	res = res.Transform(user)

	h.SendSuccess(w, res, nil)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// UpdateUserAct ...
func (h *Contract) UpdateUserAct(w http.ResponseWriter, r *http.Request) {
	var err error
	req := request.UserReqUpdate{}
	if err = h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	if err = h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	code := chi.URLParam(r, "code")
	if len(code) == 0 {
		h.SendBadRequest(w, "invalid code")
		return
	}

	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	m := model.Contract{App: h.App}
	tx, err := db.Begin(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	data, err := m.GetUserByCode(db, ctx, code)
	if err == sql.ErrNoRows {
		h.SendNotfound(w, err.Error())
		return
	}
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// check condition if email null
	if req.Email ==  data.Email{
		data = req.Transform(data)
		err = m.UpdateUser(tx, ctx, code, data)
		if err != nil {
			tx.Rollback(ctx)
			h.SendBadRequest(w, psql.ParseErr(err))
			return
		}
	} else if req.Email == "" {
		data = req.Transform(data)
		err = m.UpdateUser(tx, ctx, code, data)
		if err != nil {
			tx.Rollback(ctx)
			h.SendBadRequest(w, psql.ParseErr(err))
			return
		}
	} else if req.Email != "" { // check condition if email not null
		// request Password
		password := req.Pass

		// checking Password
		match := CheckPasswordHash(password, data.Password)
		log.Println("Match:   ", match)

		if match {
			data = req.Transform(data)
			err = m.UpdateUser(tx, ctx, code, data)
			if err != nil {
				tx.Rollback(ctx)
				h.SendBadRequest(w, psql.ParseErr(err))
				return
			}
		} else if !match {
			h.SendNotfound(w, fmt.Sprintf("Password %s not match.", password))
			return
		}
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	var res response.UsersResponse
	res = res.Transform(data)

	h.SendSuccess(w, res, nil)
}

// UpdateUserPassAct ...
func (h *Contract) UpdateUserPassAct(w http.ResponseWriter, r *http.Request) {
	var err error
	req := request.PassReq{}
	if err = h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	if err = h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	code := chi.URLParam(r, "code")
	if len(code) == 0 {
		h.SendBadRequest(w, "invalid code")
		return
	}

	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	m := model.Contract{App: h.App}
	// userCode := h.GetUserCode(r.Context())
	user, err := m.GetUserByCode(db, ctx, code)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// request old Password
	password := req.OldPassword

	// checking old Password
	match := CheckPasswordHash(password, user.Password)
	log.Println("Match:   ", match)

	if match {
		pass, err := bcrypt.GenerateFromPassword([]byte(req.Pass), 10)
		if err != nil {
			h.SendBadRequest(w, "Error when generate password")
			return
		}

		err = m.UpdateUserPass(db, ctx, code, string(pass))
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}
	} else if !match {
		h.SendNotfound(w, fmt.Sprintf("Password %s not found.", password))
		return
	}

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}

// DeleteUser ...
func (h *Contract) DeleteUser(w http.ResponseWriter, r *http.Request) {
	var err error
	req := request.UserReqUpdate{}
	code := chi.URLParam(r, "code")
	if len(code) == 0 {
		h.SendBadRequest(w, "invalid code")
		return
	}

	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	m := model.Contract{App: h.App}
	tx, err := db.Begin(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	data, err := m.GetUserByCode(db, ctx, code)
	if err == sql.ErrNoRows {
		h.SendNotfound(w, err.Error())
		return
	}
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	data = req.Transform(data)
	data.IsActive = false

	// edit is active to false
	err = m.UpdateUser(tx, ctx, code, data)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	if data.Role == "tc" {

		// find tc id yang plg sedikit orderannya
		id, err := m.GetTcIDLeastWorkByID(db, ctx, data.ID)
		if err != nil && err != sql.ErrNoRows {
			h.SendBadRequest(w, err.Error())
			return
		}

		if id > 0 {

			// find member itin berdasarkan tc id sebelumnya
			arrItinID, err := m.GetMemberItinByTcID(db, ctx, data.ID)
			if err != nil && err != sql.ErrNoRows {
				h.SendBadRequest(w, err.Error())
				return
			}

			// insert to table itin change
			for _, v := range arrItinID {
				err = m.ChangeTc(db, ctx, tx, model.MemberItinChangesEnt{MemberItinID: v, ChangedBy: "admin", ChangedUserID: id})
				if err != nil {
					h.SendBadRequest(w, err.Error())
					return
				}

				err = m.UpdateTcIdOrder(db, ctx, tx, id, v)
				if err != nil {
					h.SendBadRequest(w, err.Error())
					return
				}

			}

		}

	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}
