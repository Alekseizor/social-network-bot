package bitop

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
	"main/internal/app/config"
	"main/internal/app/model"
	"net/http"
	"net/url"
	"strconv"
)

type Client struct {
	ctx    context.Context
	body   model.BitopBody
	client *http.Client
}

func New(ctx context.Context) *Client {
	cfg := config.FromContext(ctx)
	return &Client{
		ctx: ctx,
		body: model.BitopBody{
			Token: cfg.BITOPToken,
		},
		client: &http.Client{},
	}
}

// GetBranch get info about branch from request
func (c *Client) GetBranch(ctx context.Context, branch string) (*model.ResponseBody, error) {
	cfg := config.FromContext(ctx).BITOP

	//creating url
	url := url.URL{
		Scheme: cfg.Protocol,
		Host:   cfg.SiteAdress,
		Path:   cfg.PathSearch,
	}

	log.Info("url created", url.String())

	//create request body
	reqBody, _ := json.Marshal(model.RequestBody{
		"",
		branch,
		"branch",
	})

	//create request
	reqToApi, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(reqBody))
	if err != nil {
		log.WithError(err).Error("cant create request")
		return nil, err
	}

	//create request headers
	reqToApi.Header = http.Header{
		"x-bb-token": {c.body.Token},
	}

	//do request
	rawResp, err := c.client.Do(reqToApi)
	if err != nil {
		log.WithError(err).Error("cant do request")
		return nil, err
	}

	var resp model.ResponseBody

	//status code check
	if rawResp.StatusCode != 200 {
		errLog := "status code is" + strconv.Itoa(rawResp.StatusCode)
		log.Error(errLog)
		return nil, errors.New(errLog)
	}

	//read response body
	body, err := io.ReadAll(rawResp.Body)
	if err != nil {
		log.WithError(err).Error("cant read response")
		return nil, err
	}

	//unmarshall response body
	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.WithError(err).Error("cant unmarshal response")
		return nil, err
	}

	return &resp, err
}

// GetFaculty get info about faculty from parent uuid
func (c *Client) GetFaculty(ctx context.Context, parentUUID string) (*model.ResponseBody, error) {
	cfg := config.FromContext(ctx).BITOP

	url := url.URL{
		Scheme: cfg.Protocol,
		Host:   cfg.SiteAdress,
		Path:   cfg.PathSearch,
	}

	reqBody, _ := json.Marshal(model.RequestBody{
		parentUUID,
		"",
		"faculty",
	})

	reqToApi, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(reqBody))
	if err != nil {
		log.WithError(err).Error("cant create request")
	}

	reqToApi.Header = http.Header{
		"x-bb-token": {c.body.Token},
	}

	rawResp, err := c.client.Do(reqToApi)
	if err != nil {
		log.WithError(err).Error("cant do request")
		return nil, err
	}

	if rawResp.StatusCode != 200 {
		errLog := "status code is" + strconv.Itoa(rawResp.StatusCode)
		log.Error(errLog)
		return nil, errors.New(errLog)
	}

	body, err := io.ReadAll(rawResp.Body)
	if err != nil {
		log.WithError(err).Error("cant read response")
		return nil, err
	}
	var resp model.ResponseBody

	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.WithError(err).Error("cant unmarshal response")
		return nil, err
	}

	return &resp, err
}

// GetDepartment get info about department from from parent uuid
func (c *Client) GetDepartment(ctx context.Context, parentUUID string) (*model.ResponseBody, error) {
	cfg := config.FromContext(ctx).BITOP

	url := url.URL{
		Scheme: cfg.Protocol,
		Host:   cfg.SiteAdress,
		Path:   cfg.PathSearch,
	}

	reqBody, _ := json.Marshal(model.RequestBody{
		parentUUID,
		"",
		"department",
	})

	reqToApi, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(reqBody))
	if err != nil {
		log.WithError(err).Error("cant create request")
	}

	reqToApi.Header = http.Header{
		"x-bb-token": {c.body.Token},
	}

	rawResp, err := c.client.Do(reqToApi)
	if err != nil {
		log.WithError(err).Error("cant do request")
		return nil, err
	}

	if rawResp.StatusCode != 200 {
		errLog := "status code is" + strconv.Itoa(rawResp.StatusCode)
		log.Error(errLog)
		return nil, errors.New(errLog)
	}

	body, err := io.ReadAll(rawResp.Body)
	if err != nil {
		log.WithError(err).Error("cant read response")
		return nil, err
	}
	var resp model.ResponseBody

	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.WithError(err).Error("cant unmarshal response")
		return nil, err
	}

	return &resp, err
}

// GetGroup get info about group, from parent uuid
func (c *Client) GetGroup(ctx context.Context, groupName string) (*model.ResponseBody, error) {
	cfg := config.FromContext(ctx).BITOP

	url := url.URL{
		Scheme: cfg.Protocol,
		Host:   cfg.SiteAdress,
		Path:   cfg.PathSearch,
	}

	reqBody, _ := json.Marshal(model.RequestBody{
		"",
		groupName,
		"group",
	})

	reqToApi, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(reqBody))
	if err != nil {
		log.WithError(err).Error("cant create request")
	}

	reqToApi.Header = http.Header{
		"x-bb-token": {c.body.Token},
	}

	rawResp, err := c.client.Do(reqToApi)
	if err != nil {
		log.WithError(err).Error("cant do request")
		return nil, err
	}

	if rawResp.StatusCode != 200 {
		errLog := "status code is" + strconv.Itoa(rawResp.StatusCode)
		log.Error(errLog)
		return nil, errors.New(errLog)
	}

	body, err := io.ReadAll(rawResp.Body)
	if err != nil {
		log.WithError(err).Error("cant read response")
		return nil, err
	}
	var resp model.ResponseBody

	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.WithError(err).Error("cant unmarshal response")
		return nil, err
	}

	return &resp, err
}

func (c *Client) GetSchedule(ctx context.Context, parentUUID string, IsNumerator bool, message string) (*model.ResponseBodySchedule, error) {
	weekdays := map[string]int{
		"Понедельник": 1,
		"Вторник":     2,
		"Среда":       3,
		"Четверг":     4,
		"Пятница":     5,
		"Суббота":     6,
	}

	cfg := config.FromContext(ctx).BITOP

	url := url.URL{
		Scheme: cfg.Protocol,
		Host:   cfg.SiteAdress,
		Path:   cfg.PathPath,
	}

	urlS := url.String() + parentUUID

	reqToApi, err := http.NewRequest("GET", urlS, nil)
	if err != nil {
		log.WithError(err).Error("cant create request")
	}

	reqToApi.Header = http.Header{
		"x-bb-token": {c.body.Token},
	}

	rawResp, err := c.client.Do(reqToApi)
	if err != nil {
		log.WithError(err).Error("cant do request")
		return nil, err
	}

	if rawResp.StatusCode != 200 {
		errLog := "status code is" + strconv.Itoa(rawResp.StatusCode)
		log.Error(errLog)
		return nil, errors.New(errLog)
	}

	body, err := io.ReadAll(rawResp.Body)
	if err != nil {
		log.WithError(err).Error("cant read response")
		return nil, err
	}
	var resp model.ResponseBodySchedule

	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.WithError(err).Error("cant unmarshal response")
		return nil, err
	}

	var result model.ResponseBodySchedule
	k := 0
	for _, item := range resp.Lessons {
		if item.Day == weekdays[message] && item.IsNumerator == IsNumerator {
			k++
			result.Lessons = append(result.Lessons, item)
		}
	}
	if k == 0 {
		return nil, nil
	}
	return &result, err
}
