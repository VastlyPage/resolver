package hlbaby

import (
	"io"
	"net/http"
)

var enableCaching = true

func PipeURLToResponse(method, url string, w http.ResponseWriter) error {
	if enableCaching {
		if cachedContent, cachedHeaders, statusCode, found := GetCachedContent(url, method); found {
			writeHeaders(w, cachedHeaders)
			w.WriteHeader(statusCode)
			_, err := w.Write(cachedContent)
			return err
		}
	}

	client := GetClient()
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	writeHeaders(w, flattenHeaders(resp.Header))
	w.WriteHeader(resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if enableCaching {
		CacheContent(url, method, body, flattenHeaders(resp.Header), resp.StatusCode)
	}

	_, err = w.Write(body)
	return err
}

func writeHeaders(w http.ResponseWriter, headers map[string]string) {
	for key, value := range headers {
		w.Header().Add(key, value)
	}
}

func flattenHeaders(header http.Header) map[string]string {
	flattened := make(map[string]string)
	for key, values := range header {
		if len(values) > 0 {
			flattened[key] = values[0]
		}
	}
	return flattened
}
