package api

import (
	"net/http"

	"toad/model"
	"toad/permission"
)

type LogAPI bool

func (api LogAPI) Enable() bool {
	return bool(api)
}

func (api LogAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/log", Next: api.getEventLogEndpoint, Method: "GET", Auth: false, Group: permission.All},
	}
}

func (api *LogAPI) getEventLogEndpoint(w http.ResponseWriter, req *http.Request) {

	logM := model.GetEventLogModel(di)
	// queryVar := util.GetQueryValue(req, []string{"branch"}, true)
	// branch := (*queryVar)["branch"].(string)
	// if branch == "" {
	// 	branch = "all"
	// }

	//data, err := json.Marshal(result)
	logM.GetEventLogData()
	data, err := logM.Json("")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
