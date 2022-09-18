package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/longpoll-bot"
	log "github.com/sirupsen/logrus"
	vk_client "main/internal/app/button"
	"main/internal/app/config"
	"main/internal/app/ds"
	"main/internal/app/redis"
	"main/internal/app/state"
	"main/internal/pkg/clients/bitop"
	"net/http"
	"strings"
	"time"
)

var chatContext state.ChatContext

type App struct {
	// корневой контекст
	ctx context.Context
	vk  *api.VK
	lp  *longpoll.LongPoll

	vkClient    *vk_client.VkClient
	redisClient *redis.RedClient
	bitopClient *bitop.Client
}

func New(ctx context.Context) (*App, error) {
	cfg := config.FromContext(ctx)
	vk := api.NewVK(cfg.VKToken)
	group, err := vk.GroupsGetByID(nil)
	if err != nil {
		log.WithError(err).Error("cant get groups by id")

		return nil, err
	}

	log.WithField("group_id", group[0].ID).Info("init such group")

	c, err := redis.New(ctx)
	if err != nil {
		return nil, err
	}

	vkClient, err := vk_client.New(ctx)
	if err != nil {
		return nil, err
	}

	//starting long poll
	lp, err := longpoll.NewLongPoll(vk, group[0].ID)
	if err != nil {
		log.Println("error on request")
		log.Error(err)
	}

	app := &App{
		ctx:         ctx,
		vk:          vk,
		lp:          lp,
		vkClient:    vkClient,
		redisClient: c,
	}

	return app, nil
}

func (a *App) Run(ctx context.Context) error {
	var err error
	go func() error {
		if err = InitSysRoutes(ctx); err != nil {
			log.WithError(err).Error("can't InitSysRoute")
			return err
		}
		return nil
	}()

	var ScheduleUser *ds.User
	a.lp.MessageNew(func(_ context.Context, obj events.MessageNewObject) {
		messageText := obj.Message.Text
		ScheduleUser, err = a.redisClient.GetUser(ctx, obj.Message.PeerID)
		if err != nil {
			log.WithError(err).Error("cant set user")
			return
		}
		//if the user writes for the first time, add to the database
		if ScheduleUser == nil {
			ScheduleUser = &ds.User{}
			ScheduleUser.VkID = obj.Message.PeerID
			ScheduleUser.State = "StartState"
			err := a.redisClient.SetUser(ctx, *ScheduleUser)
			if err != nil {
				log.WithError(err).Error("cant set user")
				return
			}
		} else if ScheduleUser.State == "" {
			ScheduleUser.State = "StartState"
			err := a.redisClient.SetUser(ctx, *ScheduleUser)
			if err != nil {
				log.WithError(err).Error("cant set user")
				return
			}
		}

		if strings.EqualFold(messageText, "Главное меню") {
			ScheduleUser.State = "StartState"
			err := a.redisClient.SetUser(ctx, *ScheduleUser)
			if err != nil {
				log.WithError(err).Error("cant set user")
				return
			}
		}
		strInState := map[string]state.State{
			(&(state.StartState{})).Name():           &(state.StartState{}),
			(&(state.GroupState{})).Name():           &(state.GroupState{}),
			(&(state.WeekState{})).Name():            &(state.WeekState{}),
			(&(state.DayState{})).Name():             &(state.DayState{}),
			(&(state.BugReportAppVersion{})).Name():  &(state.BugReportAppVersion{}),
			(&(state.BugReportDescription{})).Name(): &(state.BugReportDescription{}),
			(&(state.BugReportOsVersion{})).Name():   &(state.BugReportOsVersion{}),
			(&(state.BugReportChatLink{})).Name():    &(state.BugReportChatLink{}),
			(&(state.TeamState{})).Name():            &(state.TeamState{}),
			(&(state.QuestionState{})).Name():        &(state.QuestionState{}),
			(&(state.AnotherState{})).Name():         &(state.AnotherState{}),
			(&(state.ApplicationState{})).Name():     &(state.ApplicationState{}),
			(&(state.IdeaState{})).Name():            &(state.IdeaState{}),
			(&(state.CooperationState{})).Name():     &(state.CooperationState{}),
		}
		ctc := state.ChatContext{
			ScheduleUser,
			a.vk,
			a.redisClient,
			&ctx,
			a.bitopClient,
		}
		cfg := config.FromContext(*ctc.Ctx).Bot
		chatID := cfg.ChatID
		if obj.Message.PeerID == chatID {
			return
		}
		step := strInState[ScheduleUser.State]
		nextStep := step.Process(ctc, messageText)
		ScheduleUser.State = nextStep.Name()
		err = a.redisClient.SetUser(ctx, *ScheduleUser)
		if err != nil {
			log.WithError(err).Error("cant set user")
			return
		}

	})
	log.Println("Start Long Poll")
	if err := a.lp.Run(); err != nil {
		log.Println("error when starting handler")
		log.Error(err)
		return nil
	}
	return nil
}

const (
	sysHTTPDefaultTimeout = 5 * time.Minute
)

func InitSysRoutes(ctx context.Context) error {

	mux := http.NewServeMux()
	{
		mux.HandleFunc("/ready", ReadyHandler)
		mux.HandleFunc("/live", LiveHandler)
	}

	port := "80"

	s := &http.Server{
		Addr:         "0.0.0.0:" + port,
		WriteTimeout: sysHTTPDefaultTimeout,
		ReadTimeout:  sysHTTPDefaultTimeout,
		IdleTimeout:  sysHTTPDefaultTimeout,
		Handler:      mux,
	}
	err := s.ListenAndServe()
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		fmt.Println(err)
	}
	return err
}

func ReadyHandler(w http.ResponseWriter, _ *http.Request) {
	httpStatus := http.StatusOK
	w.WriteHeader(httpStatus)
	enc := json.NewEncoder(w)
	_ = enc.Encode(map[string]bool{
		"ready": true,
	})
}

func LiveHandler(w http.ResponseWriter, _ *http.Request) {
	httpStatus := http.StatusOK
	w.WriteHeader(httpStatus)
	enc := json.NewEncoder(w)
	_ = enc.Encode(map[string]bool{
		"live": true,
	})
}
