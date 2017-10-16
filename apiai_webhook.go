package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	apiTokens []string
)

func init() {
	apiTokens = nil
	http.HandleFunc("/intent", postHandler)
}

func mustGetenv(ctx context.Context, k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Criticalf(ctx, "%s environment variable not set.", k)
	}
	return v
}

func VerifyApiToken(token string) error {
	for _, x := range apiTokens {
		if x == token {
			return nil
		}
	}
	return errors.New("invalid api token.")
}

type RequestIntent struct {
	Result struct {
		Metadata struct {
			IntentName string `json:"intentName"`
		} `json:"metadata"`
	} `json:"result"`
}

type BlocksInvokeResponse struct {
	Result bool `json:"result"`
	JobId  int  `json:"job_id"`
}

type Response struct {
	Speech string `json:"speech"`
}

func postBlocksFlow(ctx context.Context, blocks_url, blocks_api_token, intent string, data []byte) (int, error) {
	client := urlfetch.Client(ctx)
	req, err := http.NewRequest("POST", blocks_url+"/flows/"+intent+".json", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+blocks_api_token)
	res, err := client.Do(req)
	if err == nil {
		defer res.Body.Close()
	}
	if err != nil {
		return 0, err
	}
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	jid := &BlocksInvokeResponse{}
	log.Infof(ctx, "BLOCKS flow invoke response = %s", buf)
	err = json.Unmarshal(buf, &jid)
	if err != nil {
		return 0, err
	}
	return jid.JobId, nil
}

func getBlocksFlowResult(ctx context.Context, blocks_url, blocks_api_token, intent string, job_id int) (string, error) {
	for true {
		client := urlfetch.Client(ctx)
		req, err := http.NewRequest("GET", blocks_url+"/flows/"+intent+"/jobs/"+strconv.Itoa(job_id)+"/status.txt", bytes.NewBuffer([]byte("")))
		req.Header.Set("Authorization", "Bearer "+blocks_api_token)
		res, err := client.Do(req)
		if err != nil {
			return "", err
		}
		buf, err := ioutil.ReadAll(res.Body)
		status := string(buf)
		res.Body.Close()
		if err != nil {
			return "", err
		}
		if status == "finished" {
			req, err := http.NewRequest("GET", blocks_url+"/flows/"+intent+"/jobs/"+strconv.Itoa(job_id)+"/variable.json?variable=_",
				bytes.NewBuffer([]byte("")))
			req.Header.Set("Authorization", "Bearer "+blocks_api_token)
			res, err := client.Do(req)
			if err != nil {
				return "", err
			}
			buf, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				return "", err
			}
			return string(buf), nil
		} else if status == "failed" || status == "canceled" {
			return "", errors.New("flow execution failed.")
		}
		time.Sleep(500000000) // 0.5 sec
	}
	return "", nil
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	//projectId := appengine.AppID(ctx)
	response := Response{"something wrong."}
	rawResponseJson := []byte(nil)
	code := 500

	defer func() {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if code == 200 {
			fmt.Fprint(w, string(rawResponseJson))
		} else {
			outjson, e := json.Marshal(response)
			if e != nil {
				log.Errorf(ctx, e.Error())
			}
			http.Error(w, string(outjson), code)
		}
	}()

	if r.Method != "POST" {
		response.Speech = "only POST method method was accepted"
		code = 404
		return
	}

	// Check API Token
	api_key := r.Header.Get("X-APIAI-TOKEN")
	if apiTokens == nil {
		apiTokens = strings.Split(mustGetenv(ctx, "API_TOKEN"), ",")
	}
	err := VerifyApiToken(api_key)
	if err != nil {
		response.Speech = err.Error()
		code = 401
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		response.Speech = err.Error()
		code = 500
		return
	}
	log.Infof(ctx, "%s", body)
	intent := &RequestIntent{}
	err = json.Unmarshal(body, &intent)
	if err != nil {
		log.Errorf(ctx, "Error occur during decode API.AI request: %s", err.Error())
		response.Speech = err.Error()
		code = 500
		return
	}
	intentName := strings.Replace(intent.Result.Metadata.IntentName, "..", "", -1)
	log.Infof(ctx, "intentName = %s", intentName)
	blocks_url := os.Getenv("BLOCKS_URL")
	blocks_api_token := os.Getenv("BLOCKS_API_TOKEN")
	job_id, err := postBlocksFlow(ctx, blocks_url, blocks_api_token, intentName, body)
	if err != nil {
		log.Errorf(ctx, "Error occur during BLOCKS Job invocation: %s", err.Error())
		response.Speech = "Error occur during BLOCKS Job invocation."
		code = 500
		return
	}

	result, err := getBlocksFlowResult(ctx, blocks_url, blocks_api_token, intentName, job_id)

	if err != nil {
		log.Errorf(ctx, "get BLOCKS job result failed: %", err.Error())
		response.Speech = "Error occur during getting BLOCKS Job result." + err.Error()
		code = 500
		return
	}

	log.Infof(ctx, "Response = %s", result)
	rawResponseJson = []byte(result)
	code = 200
	return
}
