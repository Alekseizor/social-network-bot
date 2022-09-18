package state

import (
	"context"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/api/params"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/enescakir/emoji"
	log "github.com/sirupsen/logrus"
	"main/internal/app/config"
	"main/internal/app/ds"
	"main/internal/app/model"
	"main/internal/app/redis"
	"main/internal/pkg/clients/bitop"
	"strconv"
	"strings"
	"time"
)

///////////////////////////////////////////////////////////
type ChatContext struct {
	User        *ds.User
	Vk          *api.VK
	RedisClient *redis.RedClient
	Ctx         *context.Context
	BitopClient *bitop.Client
	//получаем информацию о пользователе
	//используем для записи информации о выборе пользователя, на каком состоянии он находится
}

func (chc ChatContext) ChatID() int {
	return chc.User.VkID
}
func (chc ChatContext) Get(VkID int, Field string) string { //получаем информацию о пользователе(либо состояние, либо uuid)
	//в стрингу(входной параметр) будем писать нужный нам атрибут из БД, возвращаем
	var err error
	chc.User, err = chc.RedisClient.GetUser(*chc.Ctx, VkID)
	if err != nil {
		log.Println("Failed to get record")
		log.Error(err)
	}
	if Field == "State" {
		return chc.User.State
	}
	if Field == "GroupUUID" {
		return chc.User.GroupUUID
	}
	if Field == "IsNumerator" {
		return strconv.FormatBool(chc.User.IsNumerator)
	}

	return "not found"
}
func (chc ChatContext) Set() { //записываем информацию в бд
	err := chc.RedisClient.SetUser(*chc.Ctx, *chc.User)
	if err != nil {
		log.WithError(err).Error("cant set user")
		return
	}
}

type State interface {
	Name() string                      //получаем название состояния в виде строки, чтобы в дальнейшем куда-то записать(БД)
	Process(ChatContext, string) State //нужно взять контекст, посмотреть на каком состоянии сейчас пользователь, метод должен вернуть состояние
}

//////////////////////////////////////////////////////////
type StartState struct {
}

func (state StartState) Process(ctc ChatContext, messageText string) State {
	if messageText == "Узнать расписание" {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.Message("Укажите свою группу. Например: ИУ7-34Б")
		b.PeerID(ctc.User.VkID)
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Назад", "", "secondary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("Failed to get record")
			log.Error(err)
		}
		return &GroupState{}
	} else if messageText == "Задать вопрос" {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.Message("Уточните в чём суть Вашего вопроса")
		b.PeerID(ctc.User.VkID)
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Сотрудничество", "", "primary")
		k.AddRow()
		k.AddTextButton("Предложить идею", "", "primary")
		k.AddRow()
		k.AddTextButton("О приложении", "", "primary")
		k.AddRow()
		k.AddTextButton("Другое", "", "primary")
		k.AddRow()
		k.AddTextButton("Назад", "", "secondary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &QuestionState{}
	} else if messageText == "Хочу в команду" {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.Message("Мы рады, что Вы проявили интерес. По ссылке вся информация о вакасниях и стажировках \n \n https://docs.google.com/document/d/1XbXTUxYSZsno1oGBoF8lqxDFVqZooxzu0y3JZjK9spk/edit")
		b.PeerID(ctc.User.VkID)
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Главное меню", "", "primary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		b.Message("Если заинтересует определенная позиция, пишите @bond_nick_bond. Расскажем подробнее и ответим на все вопросы")
		_, err = ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &TeamState{}
	} else if messageText == "Сообщить о баге" {
		bugReport[ctc.User.VkID] = new(BugReport)
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.Message("Опишите проблему и прикрепите скриншоты или видео")
		b.PeerID(ctc.User.VkID)
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Назад", "", "secondary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}

		return &BugReportDescription{}
	} else {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.PeerID(ctc.User.VkID)
		b.Message("С помощью бота можно узнать расписание, сообщить о баге, задать вопрос или выяснить, как попасть к нам в команду" + emoji.WinkingFace.String())
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Узнать расписание", "", "primary")
		k.AddRow()
		k.AddTextButton("Задать вопрос", "", "primary")
		k.AddRow()
		k.AddTextButton("Сообщить о баге", "", "primary")
		k.AddRow()
		k.AddTextButton("Хочу в команду", "", "primary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &StartState{}
	}
}

func (state StartState) Name() string {
	return "StartState"
}

/////////////////////////////////////////////////////////
type GroupState struct {
}

func (state GroupState) Process(ctc ChatContext, messageText string) State {
	if messageText == "Назад" {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.PeerID(ctc.User.VkID)
		b.Message("С помощью бота можно узнать расписание, сообщить о баге, задать вопрос или выяснить, как попасть к нам в команду" + emoji.WinkingFace.String())
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Узнать расписание", "", "primary")
		k.AddRow()
		k.AddTextButton("Задать вопрос", "", "primary")
		k.AddRow()
		k.AddTextButton("Сообщить о баге", "", "primary")
		k.AddRow()
		k.AddTextButton("Хочу в команду", "", "primary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &StartState{}
	}
	ctc.BitopClient = bitop.New(*ctc.Ctx)
	resp, _ := ctc.BitopClient.GetGroup(*ctc.Ctx, messageText)
	if resp.Total > 1 {
		for _, group := range resp.Items {
			if group.Caption == strings.ToUpper(messageText) {
				ctc.User.GroupUUID = resp.Items[0].Uuid
				b := params.NewMessagesSendBuilder()
				b.PeerID(ctc.User.VkID)
				b.RandomID(0)
				b.Message("Выберите тип недели")
				k := &object.MessagesKeyboard{}
				k.AddRow()
				k.AddTextButton("Числитель", "", "primary")
				k.AddRow()
				k.AddTextButton("Знаменатель", "", "primary")
				k.AddRow()
				k.AddTextButton("Назад", "", "secondary")
				b.Keyboard(k)
				_, err := ctc.Vk.MessagesSend(b.Params)
				if err != nil {
					log.Println("error sending message")
					log.Error(err)
				}
				return &WeekState{}
			}
		}
		b := params.NewMessagesSendBuilder()
		b.PeerID(ctc.User.VkID)
		b.RandomID(0)
		b.Message("Введите полное название группы")
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &GroupState{}
	}
	if resp.Total == 1 {
		ctc.User.GroupUUID = resp.Items[0].Uuid
		b := params.NewMessagesSendBuilder()
		b.PeerID(ctc.User.VkID)
		b.RandomID(0)
		b.Message("Выберите тип недели")
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Числитель", "", "primary")
		k.AddRow()
		k.AddTextButton("Знаменатель", "", "primary")
		k.AddRow()
		k.AddTextButton("Назад", "", "secondary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &WeekState{}
	}
	b := params.NewMessagesSendBuilder()
	b.RandomID(0)
	b.PeerID(ctc.User.VkID)
	b.Message("Введите нужную группу повторно, не удалось найти")
	_, err := ctc.Vk.MessagesSend(b.Params)
	if err != nil {
		log.Println("error sending message")
		log.Error(err)
	}
	return &GroupState{}
}

func (state GroupState) Name() string {
	return "GroupState"
}

//////////////////////////////////////////////////////////
type WeekState struct {
}

func (state WeekState) Process(ctc ChatContext, messageText string) State {
	if messageText == "Назад" {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.Message("Укажите свою группу. Например: ИУ7-34Б")
		b.PeerID(ctc.User.VkID)
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Назад", "", "secondary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &GroupState{}
	}
	if messageText == "Числитель" {
		ctc.User.IsNumerator = true
		b := params.NewMessagesSendBuilder()
		b.PeerID(ctc.User.VkID)
		b.RandomID(0)
		b.Message("Выберите нужный день недели")
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Понедельник", "", "primary")
		k.AddRow()
		k.AddTextButton("Вторник", "", "primary")
		k.AddRow()
		k.AddTextButton("Среда", "", "primary")
		k.AddRow()
		k.AddTextButton("Четверг", "", "primary")
		k.AddRow()
		k.AddTextButton("Пятница", "", "primary")
		k.AddRow()
		k.AddTextButton("Суббота", "", "primary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &DayState{}
	} else if messageText == "Знаменатель" {
		ctc.User.IsNumerator = false
		b := params.NewMessagesSendBuilder()
		b.PeerID(ctc.User.VkID)
		b.RandomID(0)
		b.Message("Выберите нужный день недели")
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Понедельник", "", "primary")
		k.AddRow()
		k.AddTextButton("Вторник", "", "primary")
		k.AddRow()
		k.AddTextButton("Среда", "", "primary")
		k.AddRow()
		k.AddTextButton("Четверг", "", "primary")
		k.AddRow()
		k.AddTextButton("Пятница", "", "primary")
		k.AddRow()
		k.AddTextButton("Суббота", "", "primary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &DayState{}
	} else {
		b := params.NewMessagesSendBuilder()
		b.PeerID(ctc.User.VkID)
		b.RandomID(0)
		b.Message("Выберите тип недели")
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Числитель", "", "primary")
		k.AddRow()
		k.AddTextButton("Знаменатель", "", "primary")
		k.AddRow()
		k.AddTextButton("Назад", "", "secondary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &WeekState{}
	}
}

func (state WeekState) Name() string {
	return "WeekState"
}

///////////////////////////////////////////////////////////

type DayState struct {
}

func (state DayState) Process(ctc ChatContext, messageText string) State {
	var v string
	if (messageText == "Понедельник") || (messageText == "Вторник") || (messageText == "Среда") || (messageText == "Четверг") || (messageText == "Пятница") || (messageText == "Суббота") {
		ctc.BitopClient = bitop.New(*ctc.Ctx)
		Schedule, err := ctc.BitopClient.GetSchedule(*ctc.Ctx, ctc.User.GroupUUID, ctc.User.IsNumerator, messageText)
		if err != nil {
			log.WithError(err).Error("failed to get schedule")
		}
		if Schedule == nil {
			v := "В этот день Вы отдыхаете!"
			b := params.NewMessagesSendBuilder()
			b.PeerID(ctc.User.VkID)
			b.RandomID(0)
			k := &object.MessagesKeyboard{}
			k.AddRow()
			k.AddTextButton("Выбрать день", "", "primary")
			k.AddRow()
			k.AddTextButton("Выбрать тип недели", "", "primary")
			k.AddRow()
			k.AddTextButton("Главное меню", "", "secondary")
			b.Message(v)
			b.Keyboard(k)
			_, err := ctc.Vk.MessagesSend(b.Params)
			if err != nil {
				log.Println("error sending message")
				log.Error(err)
			}
			return &DayState{}
		}
		var lessons []model.Lesson
		var less model.Lesson
		var teach model.Teacher
		var teachs model.Teachers
		for _, lesson := range Schedule.Lessons {
			less.Name = lesson.Name
			less.Cabinet = lesson.Cabinet
			less.Type = lesson.Type
			for _, teacher := range lesson.Teachers {
				teach.Name = teacher.Name
				teachs = append(teachs, teach)
			}
			less.Teachers = teachs
			teachs = nil
			less.StartAt = lesson.StartAt
			less.EndAt = lesson.EndAt
			less.Day = lesson.Day
			less.IsNumerator = lesson.IsNumerator
			lessons = append(lessons, less)
		}
		lessons = quickSort(&lessons)
		switch messageText {
		case "Понедельник":
			{
				v = emoji.Pushpin.String() + " Ваше расписание на понедельник:\n\n"
			}
		case "Вторник":
			{
				v = emoji.Pushpin.String() + " Ваше расписание на вторник:\n\n"
			}
		case "Среда":
			{
				v = emoji.Pushpin.String() + " Ваше расписание на среду:\n\n"
			}
		case "Четверг":
			{
				v = emoji.Pushpin.String() + " Ваше расписание на четверг:\n\n"
			}
		case "Пятница":
			{
				v = emoji.Pushpin.String() + " Ваше расписание на пятницу:\n\n"
			}
		case "Суббота":
			{
				v = emoji.Pushpin.String() + " Ваше расписание на субботу:\n\n"
			}
		}
		for _, lesson := range lessons {
			v += (emoji.Watch.String() + " " + lesson.StartAt[0:5] + " - " + lesson.EndAt[0:5] + "\n")
			v += (emoji.OpenBook.String() + " " + lesson.Name + "\n")
			if lesson.Type != "" {
				if lesson.Type == "сем" {
					v += (emoji.GreenCircle.String() + " Семинар" + "\n")
				} else if lesson.Type == "лек" {
					v += (emoji.RedCircle.String() + " Лекция" + "\n")
				} else {
					v += (emoji.YellowCircle.String() + " Лабораторная" + "\n")
				}
			}
			if (lesson.Cabinet) != "" {
				v += (emoji.Door.String() + " " + lesson.Cabinet + "\n")
			}
			for _, teacher := range lesson.Teachers {
				v += (emoji.ManScientist.String() + " " + teacher.Name + "\n")
			}

			v += "\n\n"
		}
		b := params.NewMessagesSendBuilder()
		b.PeerID(ctc.User.VkID)
		b.RandomID(0)
		b.Message(v)
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Выбрать день", "", "primary")
		k.AddRow()
		k.AddTextButton("Выбрать тип недели", "", "primary")
		k.AddRow()
		k.AddTextButton("Главное меню", "", "secondary")
		b.Keyboard(k)
		_, err = ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &DayState{}
	} else if messageText == "Главное меню" {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.PeerID(ctc.User.VkID)
		b.Message("С помощью бота можно узнать расписание, сообщить о баге, задать вопрос или выяснить, как попасть к нам в команду" + emoji.WinkingFace.String())
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Узнать расписание", "", "primary")
		k.AddRow()
		k.AddTextButton("Задать вопрос", "", "primary")
		k.AddRow()
		k.AddTextButton("Сообщить о баге", "", "primary")
		k.AddRow()
		k.AddTextButton("Хочу в команду", "", "primary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &StartState{}
	} else if messageText == "Выбрать день" {
		b := params.NewMessagesSendBuilder()
		b.PeerID(ctc.User.VkID)
		b.RandomID(0)
		b.Message("Выберите нужный день недели")
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Понедельник", "", "primary")
		k.AddRow()
		k.AddTextButton("Вторник", "", "primary")
		k.AddRow()
		k.AddTextButton("Среда", "", "primary")
		k.AddRow()
		k.AddTextButton("Четверг", "", "primary")
		k.AddRow()
		k.AddTextButton("Пятница", "", "primary")
		k.AddRow()
		k.AddTextButton("Суббота", "", "primary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &DayState{}
	} else if messageText == "Выбрать тип недели" {
		b := params.NewMessagesSendBuilder()
		b.PeerID(ctc.User.VkID)
		b.RandomID(0)
		b.Message("Выберите тип недели")
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Числитель", "", "primary")
		k.AddRow()
		k.AddTextButton("Знаменатель", "", "primary")
		k.AddRow()
		k.AddTextButton("Назад", "", "secondary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &WeekState{}
	} else {
		b := params.NewMessagesSendBuilder()
		v := "Проверьте правильность ввода введенного учебного дня"
		b.PeerID(ctc.User.VkID)
		b.RandomID(0)
		b.Message(v)
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Главное меню", "", "secondary")
		k.AddRow()
		k.AddTextButton("Выбрать день", "", "primary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &DayState{}
	}
}

func (state DayState) Name() string {
	return "DayState"
}

///////////////////////////////////////////////////////////
type QuestionState struct {
}

func (state QuestionState) Process(ctc ChatContext, messageText string) State {

	if messageText == "Сотрудничество" {
		b := params.NewMessagesSendBuilder()
		b.PeerID(ctc.User.VkID)
		b.RandomID(0)
		b.Message("Мы рады, что Вы выразили интерес к сотрудничеству с нами. Расскажите о Вашем предложении")
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Назад", "", "secondary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &CooperationState{}
	} else if messageText == "Предложить идею" {
		b := params.NewMessagesSendBuilder()
		b.PeerID(ctc.User.VkID)
		b.RandomID(0)
		b.Message("Мы всегда рады новым и интересным идеям. Расскажите о Вашем замысле")
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Назад", "", "secondary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &IdeaState{}
	} else if messageText == "О приложении" {
		b := params.NewMessagesSendBuilder()
		b.PeerID(ctc.User.VkID)
		b.RandomID(0)
		b.Message("Расскажите о сути Вашего вопроса")
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Назад", "", "secondary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &ApplicationState{}
	} else if messageText == "Другое" {
		b := params.NewMessagesSendBuilder()
		b.PeerID(ctc.User.VkID)
		b.RandomID(0)
		b.Message("Расскажите о сути Вашего вопроса")
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Назад", "", "secondary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &AnotherState{}
	} else if messageText == "Назад" {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.PeerID(ctc.User.VkID)
		b.Message("С помощью бота можно узнать расписание, сообщить о баге, задать вопрос или выяснить, как попасть к нам в команду" + emoji.WinkingFace.String())
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Узнать расписание", "", "primary")
		k.AddRow()
		k.AddTextButton("Задать вопрос", "", "primary")
		k.AddRow()
		k.AddTextButton("Сообщить о баге", "", "primary")
		k.AddRow()
		k.AddTextButton("Хочу в команду", "", "primary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &StartState{}
	} else {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.Message("Уточните в чём суть Вашего вопроса")
		b.PeerID(ctc.User.VkID)
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Сотрудничество", "", "primary")
		k.AddRow()
		k.AddTextButton("Предложить идею", "", "primary")
		k.AddRow()
		k.AddTextButton("О приложении", "", "primary")
		k.AddRow()
		k.AddTextButton("Другое", "", "primary")
		k.AddRow()
		k.AddTextButton("Назад", "", "secondary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &QuestionState{}
	}
}

func (state QuestionState) Name() string {
	return "QuestionState"
}

///////////////////////////////////////////////////////////
type CooperationReport struct {
	hashtag     string
	author      string
	description string
	chatLink    string
}

var сooperationReport CooperationReport

type CooperationState struct {
}

func (state CooperationState) Process(ctc ChatContext, messageText string) State {
	cfg := config.FromContext(*ctc.Ctx).Bot
	groupID := cfg.GroupID
	chatID := cfg.ChatID
	var userID int
	if messageText == "Главное меню" {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.PeerID(ctc.User.VkID)
		b.Message("С помощью бота можно узнать расписание, сообщить о баге, задать вопрос или выяснить, как попасть к нам в команду" + emoji.WinkingFace.String())
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Узнать расписание", "", "primary")
		k.AddRow()
		k.AddTextButton("Задать вопрос", "", "primary")
		k.AddRow()
		k.AddTextButton("Сообщить о баге", "", "primary")
		k.AddRow()
		k.AddTextButton("Хочу в команду", "", "primary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &StartState{}
	} else if messageText == "Назад" {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.Message("Уточните в чём суть Вашего вопроса")
		b.PeerID(ctc.User.VkID)
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Сотрудничество", "", "primary")
		k.AddRow()
		k.AddTextButton("Предложить идею", "", "primary")
		k.AddRow()
		k.AddTextButton("О приложении", "", "primary")
		k.AddRow()
		k.AddTextButton("Другое", "", "primary")
		k.AddRow()
		k.AddTextButton("Назад", "", "secondary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &QuestionState{}
	} else {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		users, err := ctc.Vk.UsersGet(api.Params{
			"user_ids": ctc.User.VkID,
		})
		сooperationReport.hashtag = "#сотрудничество"
		сooperationReport.author = users[0].FirstName + " " + users[0].LastName
		сooperationReport.description = messageText
		userID = ctc.User.VkID
		сooperationReport.chatLink = "Ссылка на чат: " + "https://vk.com/gim" + groupID + "?sel=" + strconv.Itoa(userID)
		b.PeerID(chatID)
		k := &object.MessagesKeyboard{}
		k.Buttons = make([][]object.MessagesKeyboardButton, 0)
		b.Keyboard(k)
		b.Message(сooperationReport.hashtag + "\n\n" + сooperationReport.author + "\n\n" + сooperationReport.description + "\n\n" + сooperationReport.chatLink + "\n")
		_, err = ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		b.RandomID(0)
		b.PeerID(ctc.User.VkID)
		b.Message("Спасибо за Ваше предложение! Совсем скоро мы вернемся с ответом и всё обсудим")
		k = &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Главное меню", "", "primary")
		b.Keyboard(k)
		b.RandomID(0)
		_, err = ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &CooperationState{}
	}
}

func (state CooperationState) Name() string {
	return "CooperationState"
}

///////////////////////////////////////////////////////////
type IdeaReport struct {
	hashtag     string
	author      string
	description string
	chatLink    string
}

var ideaReport IdeaReport

type IdeaState struct {
}

func (state IdeaState) Process(ctc ChatContext, messageText string) State {
	cfg := config.FromContext(*ctc.Ctx).Bot
	groupID := cfg.GroupID
	chatID := cfg.ChatID
	var userID int
	if messageText == "Главное меню" {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.PeerID(ctc.User.VkID)
		b.Message("С помощью бота можно узнать расписание, сообщить о баге, задать вопрос или выяснить, как попасть к нам в команду" + emoji.WinkingFace.String())
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Узнать расписание", "", "primary")
		k.AddRow()
		k.AddTextButton("Задать вопрос", "", "primary")
		k.AddRow()
		k.AddTextButton("Сообщить о баге", "", "primary")
		k.AddRow()
		k.AddTextButton("Хочу в команду", "", "primary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &StartState{}
	} else if messageText == "Назад" {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.Message("Уточните в чём суть Вашего вопроса")
		b.PeerID(ctc.User.VkID)
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Сотрудничество", "", "primary")
		k.AddRow()
		k.AddTextButton("Предложить идею", "", "primary")
		k.AddRow()
		k.AddTextButton("О приложении", "", "primary")
		k.AddRow()
		k.AddTextButton("Другое", "", "primary")
		k.AddRow()
		k.AddTextButton("Назад", "", "secondary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &QuestionState{}
	} else {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		users, err := ctc.Vk.UsersGet(api.Params{
			"user_ids": ctc.User.VkID,
		})
		ideaReport.hashtag = "#идея"
		ideaReport.author = users[0].FirstName + " " + users[0].LastName
		ideaReport.description = messageText
		userID = ctc.User.VkID
		ideaReport.chatLink = "Ссылка на чат: " + "https://vk.com/gim" + groupID + "?sel=" + strconv.Itoa(userID)
		b.PeerID(chatID)
		k := &object.MessagesKeyboard{}
		k.Buttons = make([][]object.MessagesKeyboardButton, 0)
		b.Keyboard(k)
		b.Message(ideaReport.hashtag + "\n\n" + ideaReport.author + "\n\n" + ideaReport.description + "\n\n" + ideaReport.chatLink + "\n")
		_, err = ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		b.RandomID(0)
		b.Message("Спасибо за идею! Совсем скоро мы вернемся с ответом и всё обсудим")
		b.PeerID(ctc.User.VkID)
		k = &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Главное меню", "", "primary")
		b.Keyboard(k)
		b.RandomID(0)
		_, err = ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &CooperationState{}
	}
}

func (state IdeaState) Name() string {
	return "IdeaState"
}

///////////////////////////////////////////////////////////
type ApplicationState struct {
}

type ApplicationReport struct {
	hashtag  string
	author   string
	question string
	chatLink string
}

var applicationReport ApplicationReport

func (state ApplicationState) Process(ctc ChatContext, messageText string) State {
	cfg := config.FromContext(*ctc.Ctx).Bot
	groupID := cfg.GroupID
	chatID := cfg.ChatID
	var userID int
	if messageText == "Главное меню" {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.PeerID(ctc.User.VkID)
		b.Message("С помощью бота можно узнать расписание, сообщить о баге, задать вопрос или выяснить, как попасть к нам в команду" + emoji.WinkingFace.String())
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Узнать расписание", "", "primary")
		k.AddRow()
		k.AddTextButton("Задать вопрос", "", "primary")
		k.AddRow()
		k.AddTextButton("Сообщить о баге", "", "primary")
		k.AddRow()
		k.AddTextButton("Хочу в команду", "", "primary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &StartState{}
	} else if messageText == "Назад" {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.Message("Уточните в чём суть Вашего вопроса")
		b.PeerID(ctc.User.VkID)
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Сотрудничество", "", "primary")
		k.AddRow()
		k.AddTextButton("Предложить идею", "", "primary")
		k.AddRow()
		k.AddTextButton("О приложении", "", "primary")
		k.AddRow()
		k.AddTextButton("Другое", "", "primary")
		k.AddRow()
		k.AddTextButton("Назад", "", "secondary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &QuestionState{}
	} else {
		users, err := ctc.Vk.UsersGet(api.Params{
			"user_ids": ctc.User.VkID,
		})
		applicationReport.hashtag = "#о_приложении"
		applicationReport.author = users[0].FirstName + " " + users[0].LastName
		applicationReport.question = messageText
		userID = ctc.User.VkID
		applicationReport.chatLink = "Ссылка на чат: " + "https://vk.com/gim" + groupID + "?sel=" + strconv.Itoa(userID)
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.Message("Спасибо за вопрос! Совсем скоро мы вернемся с ответом")
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Главное меню", "", "primary")
		b.Keyboard(k)
		b.RandomID(0)
		b.PeerID(ctc.User.VkID)
		_, err = ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		b.PeerID(chatID)
		k = &object.MessagesKeyboard{}
		k.Buttons = make([][]object.MessagesKeyboardButton, 0)
		b.Keyboard(k)
		b.Message(applicationReport.hashtag + "\n\n" + applicationReport.author + "\n\n" + applicationReport.question + "\n\n" + applicationReport.chatLink + "\n")
		_, err = ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &ApplicationState{}
	}
}

func (state ApplicationState) Name() string {
	return "ApplicationState"
}

///////////////////////////////////////////////////////////

type AnotherState struct {
}
type AnotherReport struct {
	hashtag  string
	author   string
	question string
	chatLink string
}

var anotherReport AnotherReport

func (state AnotherState) Process(ctc ChatContext, messageText string) State {
	cfg := config.FromContext(*ctc.Ctx).Bot
	groupID := cfg.GroupID
	chatID := cfg.ChatID
	var userID int
	if messageText == "Главное меню" {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.PeerID(ctc.User.VkID)
		b.Message("С помощью бота можно узнать расписание, сообщить о баге, задать вопрос или выяснить, как попасть к нам в команду" + emoji.WinkingFace.String())
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Узнать расписание", "", "primary")
		k.AddRow()
		k.AddTextButton("Задать вопрос", "", "primary")
		k.AddRow()
		k.AddTextButton("Сообщить о баге", "", "primary")
		k.AddRow()
		k.AddTextButton("Хочу в команду", "", "primary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &StartState{}
	} else if messageText == "Назад" {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.Message("Уточните в чём суть Вашего вопроса")
		b.PeerID(ctc.User.VkID)
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Сотрудничество", "", "primary")
		k.AddRow()
		k.AddTextButton("Предложить идею", "", "primary")
		k.AddRow()
		k.AddTextButton("О приложении", "", "primary")
		k.AddRow()
		k.AddTextButton("Другое", "", "primary")
		k.AddRow()
		k.AddTextButton("Назад", "", "secondary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &QuestionState{}
	} else {
		users, err := ctc.Vk.UsersGet(api.Params{
			"user_ids": ctc.User.VkID,
		})
		anotherReport.hashtag = "#другое"
		anotherReport.author = users[0].FirstName + " " + users[0].LastName
		anotherReport.question = messageText
		userID = ctc.User.VkID
		anotherReport.chatLink = "Ссылка на чат: " + "https://vk.com/gim" + groupID + "?sel=" + strconv.Itoa(userID)
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.Message("Спасибо за вопрос! Совсем скоро мы вернемся с ответом")
		b.PeerID(ctc.User.VkID)
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Главное меню", "", "primary")
		b.Keyboard(k)
		b.RandomID(0)
		_, err = ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		b.PeerID(chatID)
		k = &object.MessagesKeyboard{}
		k.Buttons = make([][]object.MessagesKeyboardButton, 0)
		b.Keyboard(k)
		b.Message(anotherReport.hashtag + "\n\n" + anotherReport.author + "\n\n" + anotherReport.question + "\n\n" + anotherReport.chatLink + "\n")
		_, err = ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &CooperationState{}
	}
}

func (state AnotherState) Name() string {
	return "AnotherState"
}

//////////////////////////////////////////////////////////
type BugReportDescription struct {
}
type BugReportAppVersion struct {
}
type BugReportOsVersion struct {
}
type BugReportChatLink struct {
}

type BugReport struct {
	hashtag     string
	author      string
	description string
	appVersion  string
	osVersion   string
	chatLink    string
}

var bugReport = map[int]*BugReport{}

func (state BugReportDescription) Process(ctc ChatContext, messageText string) State {
	if messageText == "Назад" {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.PeerID(ctc.User.VkID)
		b.Message("Рады приветствовать тебя у нас в сообществе, давай найдем твоё расписание!" + emoji.WinkingFace.String())
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Узнать расписание", "", "primary")
		k.AddRow()
		k.AddTextButton("Задать вопрос", "", "primary")
		k.AddRow()
		k.AddTextButton("Сообщить о баге", "", "primary")
		k.AddRow()
		k.AddTextButton("Хочу в команду", "", "primary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &StartState{}
	} else {
		users, err := ctc.Vk.UsersGet(api.Params{
			"user_ids": ctc.User.VkID,
		})
		bugReport[ctc.User.VkID].hashtag = "#баг"
		bugReport[ctc.User.VkID].author = users[0].FirstName + " " + users[0].LastName
		bugReport[ctc.User.VkID].description = "Описание проблемы: " + messageText
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.PeerID(ctc.User.VkID)
		k := &object.MessagesKeyboard{}
		k.Buttons = make([][]object.MessagesKeyboardButton, 0)
		b.Keyboard(k)
		b.Message("Укажите версию приложения.")
		_, err = ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		b = params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.PeerID(ctc.User.VkID)
		b.Message("Чтобы узнать версию, зайдите в приложение, перейдите в меню и нажмите \"О приложении\"")
		_, err = ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &BugReportAppVersion{}
	}
}

func (state BugReportDescription) Name() string {
	return "BugReportDescription"
}
func (state BugReportAppVersion) Process(ctc ChatContext, messageText string) State {
	bugReport[ctc.User.VkID].appVersion = "Версия приложения: " + messageText
	b := params.NewMessagesSendBuilder()
	b.RandomID(0)
	b.PeerID(ctc.User.VkID)
	b.Message("Назовите версию ОС. Например: IOS 15.4.1. \n \nНиже инструкция, как узнать версию ОС на вашем устройстве")
	_, err := ctc.Vk.MessagesSend(b.Params)
	if err != nil {
		log.Println("error sending message")
		log.Error(err)
	}
	b.Message("Для Android: Настройки(пролистайте вниз) > О телефоне > Версия Android. \n \nДля IOS: Настройки > Основные > Об этом устройстве > Версия ПО")
	_, err = ctc.Vk.MessagesSend(b.Params)
	if err != nil {
		log.Println("error sending message")
		log.Error(err)
	}
	return &BugReportOsVersion{}
}

func (state BugReportAppVersion) Name() string {
	return "BugReportAppVersion"
}
func (state BugReportOsVersion) Process(ctc ChatContext, messageText string) State {
	cfg := config.FromContext(*ctc.Ctx).Bot
	groupID := cfg.GroupID
	chatID := cfg.ChatID
	var userID int
	bugReport[ctc.User.VkID].osVersion = "Версия ОС: " + messageText
	userID = ctc.User.VkID
	bugReport[ctc.User.VkID].chatLink = "Ссылка на чат: " + "https://vk.com/gim" + groupID + "?sel=" + strconv.Itoa(userID)
	b := params.NewMessagesSendBuilder()
	b.RandomID(0)
	b.Message("Спасибо, что сообщили о проблеме! Совсем скоро мы вернемся с ответом и поможем Вам")
	b.PeerID(ctc.User.VkID)
	k := &object.MessagesKeyboard{}
	k.AddRow()
	k.AddTextButton("Главное меню", "", "primary")
	b.Keyboard(k)
	_, err := ctc.Vk.MessagesSend(b.Params)
	if err != nil {
		log.Println("error sending message")
		log.Error(err)
	}
	/*chat, err := ctc.Vk.MessagesSearchConversations(api.Params{
		"q":     "Backend Active",
		"count": 1,
	})
	chatId := chat.Items[0].Peer.LocalID*/
	b.PeerID(chatID)
	k = &object.MessagesKeyboard{}
	k.Buttons = make([][]object.MessagesKeyboardButton, 0)
	b.Keyboard(k)
	b.Message(bugReport[ctc.User.VkID].hashtag + "\n\n" + bugReport[ctc.User.VkID].author + "\n\n" + bugReport[ctc.User.VkID].description + "\n\n" + bugReport[ctc.User.VkID].appVersion + "\n\n" + bugReport[ctc.User.VkID].osVersion + "\n\n" + bugReport[ctc.User.VkID].chatLink + "\n")
	_, err = ctc.Vk.MessagesSend(b.Params)
	if err != nil {
		log.Println("error sending message")
		log.Error(err)
	}
	return &BugReportChatLink{}

}

func (state BugReportOsVersion) Name() string {
	return "BugReportOsVersion"
}

func (state BugReportChatLink) Process(ctc ChatContext, messageText string) State {
	if messageText == "Главное меню" {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.PeerID(ctc.User.VkID)
		b.Message("Рады приветствовать тебя у нас в сообществе, давай найдем твоё расписание!" + emoji.WinkingFace.String())
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Узнать расписание", "", "primary")
		k.AddRow()
		k.AddTextButton("Хочу в команду", "", "primary")
		k.AddRow()
		k.AddTextButton("Сообщить о баге", "", "primary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &StartState{}
	} else {
		//conversation, err := ctc.Vk.MessagesGetConversations(api.Params{
		//	"offset": 0,
		//})
		//conversation.Items[0].Conversation.Peer

		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.Message("Спасибо, что сообщили о проблеме! Совсем скоро мы вернемся с ответом и поможем Вам")
		b.PeerID(ctc.User.VkID)
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Главное меню", "", "primary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}

		return &BugReportChatLink{}
	}

}

func (state BugReportChatLink) Name() string {
	return "BugReportChatLink"
}

//////////////////////////////////////////////////////////
type TeamState struct {
}

func (state TeamState) Process(ctc ChatContext, messageText string) State {
	if messageText == "Главное меню" {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.PeerID(ctc.User.VkID)
		b.Message("Рады приветствовать тебя у нас в сообществе, давай найдем твоё расписание!" + emoji.WinkingFace.String())
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Узнать расписание", "", "primary")
		k.AddRow()
		k.AddTextButton("Хочу в команду", "", "primary")
		k.AddRow()
		k.AddTextButton("Сообщить о баге", "", "primary")
		b.Keyboard(k)
		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &StartState{}
	} else {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.Message("Мы рады, что Вы проявили интерес. По ссылке вся информация о вакасниях и стажировках \nhttps://docs.google.com/document/d/1XbXTUxYSZsno1oGBoF8lqxDFVqZooxzu0y3JZjK9spk/edit")
		b.PeerID(ctc.User.VkID)
		k := &object.MessagesKeyboard{}
		k.AddRow()
		k.AddTextButton("Главное меню", "", "primary")
		b.Keyboard(k)

		_, err := ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		b.Message("Если заинтересует определенная позиция, пишите @bond_nick_bond. Расскажем подробнее и ответим на все вопросы")
		_, err = ctc.Vk.MessagesSend(b.Params)
		if err != nil {
			log.Println("error sending message")
			log.Error(err)
		}
		return &TeamState{}
	}
}

func (state TeamState) Name() string {
	return "TeamState"
}

///////////////////////////////////////////////////////////
func quickSort(lessons *[]model.Lesson) []model.Lesson {
	var lessonl, lessone, lessonm []model.Lesson
	if (len(*lessons) == 1) || (len(*lessons) == 0) {
		return *lessons
	}
	randomTime := (*lessons)[0].StartAt
	randomTimetime, _ := time.Parse("15:04:05", randomTime)
	for _, lesson := range *lessons {
		Timetime, _ := time.Parse("15:04:05", lesson.StartAt)
		if Timetime.Before(randomTimetime) { //если ли Timetime раньше randomTimetime
			lessonl = append(lessonl, lesson)
		} else if Timetime.After(randomTimetime) {
			lessonm = append(lessonm, lesson)
		} else {
			lessone = append(lessone, lesson)
		}
	}
	finalLessonsl := quickSort(&lessonl)
	for _, lesson := range lessone {
		finalLessonsl = append(finalLessonsl, lesson)
	}
	finalLessonsm := quickSort(&lessonm)
	for _, lesson := range finalLessonsm {
		finalLessonsl = append(finalLessonsl, lesson)
	}
	return finalLessonsl
}
