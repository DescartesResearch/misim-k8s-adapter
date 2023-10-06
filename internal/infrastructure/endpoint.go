package infrastructure

import (
	"net/http"
)

type Endpoint func(w http.ResponseWriter, r *http.Request)
