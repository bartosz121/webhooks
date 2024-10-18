package health

import (
	"net/http"
)

// @Summary	Read health
// @Tags		health
// @Accept		json
// @Produce	json
// @Success	200
// @Router		/v1/health [get]
func Healthcheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"msg": "ok"}`))
}
