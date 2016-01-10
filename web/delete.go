package web

import (
  "net/http"
)

func handleDelete(resp http.ResponseWriter, req *http.Request) {
  path := safePath(req.URL.Path)
  err := theServer.DeleteNode(path)
  if err != nil {
    handleCommonErrors(resp, req, err)
    return
  }

  resp.WriteHeader(http.StatusOK)
}
