package app

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

type Page struct {
	Title string
}

type Task struct {
	Id           int64 `datastore:"-"` // instructs datastore to ignore
	Title        string
	Details      string
	Category     string
	State        string
	Dt_completed int64
	Dt_created   int64
}

var cached_templates = template.Must(template.ParseGlob("app/templates/*.html"))

func init() {
	http.HandleFunc("/api/", router)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/", home)
}

func router(res http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case "GET":
		get(res, req)
	case "POST":
		post(res, req)
	case "PUT":
		put(res, req)
	case "DELETE":
		archive(res, req)
	default:
		fmt.Fprintf(res, "{}")
	}
}

func home(res http.ResponseWriter, req *http.Request) {
	page := Page{}
	page.Title = "bongo app"
	renderTemplate(res, "index.html", &page)
}

func logout(res http.ResponseWriter, req *http.Request) {
	c := appengine.NewContext(req)
	logout_url, err := user.LogoutURL(c, "/")
	if error_check(res, err) {
		return
	}
	http.Redirect(res, req, logout_url, http.StatusTemporaryRedirect)
}

func get(res http.ResponseWriter, req *http.Request) {
	c := appengine.NewContext(req)

	q := datastore.NewQuery("Task").Filter("State =", "active").Order("Title").Limit(50)
	tasks := make([]Task, 0, 50)
	keys, err := q.GetAll(c, &tasks)
	if error_check(res, err) {
		return
	}

	for i, key := range keys {
		tasks[i].Id = key.IntID()
	}

	jsonResponse(res, tasks)
}

func post(res http.ResponseWriter, req *http.Request) {
	c := appengine.NewContext(req)
	var task Task
	var model = req.FormValue("model")
	json.Unmarshal([]byte(model), &task)

	postKey, err := datastore.Put(c, datastore.NewIncompleteKey(c, "Task", nil), &task)
	if (error_check(res, err)) {
		return
	}

	task.Id = postKey.IntID()
	postKey, err = datastore.Put(c, postKey, &task)
	if (error_check(res, err)) {
		return
	}

	// return object
	jsonResponse(res, task)
}

func put(res http.ResponseWriter, req *http.Request) {
	c := appengine.NewContext(req)
	var task Task
	var model = req.FormValue("model")
	json.Unmarshal([]byte(model), &task)

	key := datastore.NewKey(c, "Task", "", task.Id, nil)
	_, err := datastore.Put(c, key, &task)
	if error_check(res, err) {
		return
	}

	// return object
	jsonResponse(res, task)
}

func archive(res http.ResponseWriter, req *http.Request) {

	c := appengine.NewContext(req)
	taskId, _ := strconv.ParseInt(strings.Replace(req.URL.Path, "/api/", "", 1), 10, 64)

	key := datastore.NewKey(c, "Task", "", taskId, nil)
	task := new(Task)
	err := datastore.Get(c, key, task)
	if error_check(res, err) {
		return
	}

	task.State = "archived"
	_, err2 := datastore.Put(c, key, task)
	if error_check(res, err2) {
		return
	}

	// return object
	jsonResponse(res, task)
}

func renderTemplate(res http.ResponseWriter, template string, p *Page) {
	err := cached_templates.ExecuteTemplate(res, template, p)
	error_check(res, err)
}

func jsonResponse(res http.ResponseWriter, data interface{}) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")

	payload, err := json.Marshal(data)
	if (error_check(res, err)) {
		return
	}

	fmt.Fprintf(res, string(payload))
}

func error_check(res http.ResponseWriter, err error) bool {
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return true
	}
	return false
}
