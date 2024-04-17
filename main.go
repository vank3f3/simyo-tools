package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type LoginRequest struct {
	Password    string `json:"password"`
	PhoneNumber string `json:"phoneNumber"`
}

type ErrorResponse struct {
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	EsimCode string `json:"esimCode"`
}

func main() {
	// 设置路由，即当访问 "/" 时，调用 serveHomePage 函数
	http.HandleFunc("/", serveHomePage)
	http.HandleFunc("/commit", commitHandler)

	// 设置监听的端口，这里我们使用 8080 端口
	fmt.Println("Server is listening on http://localhost:80")
	http.ListenAndServe("0.0.0.0:80", nil)
}

// serveHomePage 函数用于提供首页
func serveHomePage(w http.ResponseWriter, r *http.Request) {
	// 直接提供位于同一目录下的 index.html 文件
	http.ServeFile(w, r, "index.html")
}

func commitHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Invalid request method")
		return
	}

	// 阶段一
	var loginData LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error parsing request body: %v", err)
		return
	}

	loginBody, _ := json.Marshal(loginData)
	sessionUrl := "https://appapi.simyo.nl/simyoapi/api/v1/sessions"

	headers := map[string]string{
		"X-Client-Token":    "e77b7e2f43db41bb95b17a2a11581a38",
		"X-Client-Platform": "android",
		"X-Client-Version":  "3.64.4",
		"X-Session-Token":   "",
		"User-Agent":        "MijnSimyo/3.64.4 (Linux; Android 13; Scale/2.75)",
		"Content-Type":      "application/json; charset=UTF-8",
		"Connection":        "Keep-Alive",
		"Accept-Encoding":   "gzip",
	}

	req, _ := http.NewRequest("POST", sessionUrl, bytes.NewBuffer(loginBody))
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(w, "Error in HTTP request: %v", err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		jsonResponse(w, 400, "Error in response structure", "")
		return
	}

	esimCode, _ := result["result"].(map[string]interface{})["esimCode"].(string)
	if esimCode != "" {
		jsonResponse(w, 200, "Success", esimCode)
		return
	}

	sessionToken, _ := result["result"].(map[string]interface{})["sessionToken"].(string)

	// 阶段二
	headers["X-Session-Token"] = sessionToken
	customer := "https://appapi.simyo.nl/simyoapi/api/v1/esim/get-by-customer"
	req, _ = http.NewRequest("GET", customer, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err = client.Do(req)
	if err != nil {
		fmt.Fprintf(w, "Error in HTTP request: %v", err)
		return
	}
	defer resp.Body.Close()

	body, _ = ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &result); err != nil {

		jsonResponse(w, 400, "Error in response structure", "")
		return
	}

	activationCode, _ := result["result"].(map[string]interface{})["activationCode"].(string)
	if activationCode != "" {
		jsonResponse(w, 200, "Success", activationCode)
		return
	}
}

func jsonResponse(w http.ResponseWriter, code int, msg, esimCode string) {
	response := ErrorResponse{
		Code:     code,
		Msg:      msg,
		EsimCode: esimCode,
	}
	respBody, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}
