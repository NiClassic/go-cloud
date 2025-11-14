package handler

import "github.com/NiClassic/go-cloud/config"

type baseHandler struct {
	cfg *config.Config
	r   *Renderer
}

func newBaseHandler(cfg *config.Config, r *Renderer) *baseHandler {
	return &baseHandler{cfg, r}
}
