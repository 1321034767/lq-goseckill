package main

import (
	"GoSecKill/common"
	"GoSecKill/internal/config"
	"GoSecKill/pkg/log"
	"GoSecKill/pkg/models"
	"GoSecKill/pkg/mq"
	"GoSecKill/pkg/util"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"go.uber.org/zap"
)

var hostArray = []string{"127.0.0.1", "127.0.0.1"}

var localHost = "127.0.0.1"

var GetOneIp = "127.0.0.1"

var GetOnePort = "8084"

var port = "8083"

var hashConsistent *common.Consistent

var rabbitMqValidate *mq.RabbitMQ

type AccessControl struct {
	// slice of host
	sourceArray map[int]interface{}
	// rw mutex
	sync.RWMutex
}

var accessControl = &AccessControl{sourceArray: make(map[int]interface{})}

func (m *AccessControl) GetNewRecord(uid int) interface{} {
	m.RLock()
	defer m.RUnlock()
	return m.sourceArray[uid]
}

func (m *AccessControl) SetNewRecord(uid int) {
	m.Lock()
	m.sourceArray[uid] = uid
	m.Unlock()
}

func (m *AccessControl) GetDistributedRight(req *http.Request) bool {
	uid, err := req.Cookie("uid")
	if err != nil {
		return false
	}
	hostRequest, err := hashConsistent.Get(uid.Value)
	if err != nil {
		return false
	}
	if hostRequest == localHost {
		return m.GetDataFromMap(uid.Value)
	}
	return m.GetDataFromOtherMap(hostRequest, req)
}

func (m *AccessControl) GetDataFromMap(uid string) bool {
	uidInt, err := strconv.Atoi(uid)
	if err != nil {
		return false
	}
	data := m.GetNewRecord(uidInt)

	if data != nil {
		return true
	}
	return false
}

func (m *AccessControl) GetDataFromOtherMap(host string, request *http.Request) bool {
	hostUrl := "http://" + host + ":" + port + "/checkRight"
	response, body, err := GetCurl(hostUrl, request)
	if err != nil {
		return false
	}
	if response.StatusCode == 200 && string(body) == "true" {
		return true
	}
	return false
}

func GetCurl(hostUrl string, request *http.Request) (response *http.Response, body []byte, err error) {
	uidPre, err := request.Cookie("uid")
	if err != nil {
		return
	}

	uidSign, err := request.Cookie("sign")
	if err != nil {
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", hostUrl, nil)
	if err != nil {
		return
	}

	cookieUid := &http.Cookie{Name: "uid", Value: uidPre.Value, Path: "/"}
	cookieSign := &http.Cookie{Name: "sign", Value: uidSign.Value, Path: "/"}

	req.AddCookie(cookieUid)
	req.AddCookie(cookieSign)

	response, err = client.Do(req)
	defer response.Body.Close()
	if err != nil {
		return
	}
	body, err = io.ReadAll(response.Body)
	return
}

func CheckRight(w http.ResponseWriter, r *http.Request) {
	right := accessControl.GetDistributedRight(r)
	if !right {
		w.Write([]byte("false"))
		return
	}
	w.Write([]byte("true"))
	return
}

func Check(rw http.ResponseWriter, req *http.Request) {
	zap.L().Info("receive request")
	queryForm, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil || len(queryForm["productID"]) <= 0 {
		_, _ = rw.Write([]byte("false"))
		return
	}
	productString := queryForm["productID"][0]
	fmt.Println(productString)

	userCookie, err := req.Cookie("uid")
	if err != nil {
		_, _ = rw.Write([]byte("false"))
		return
	}

	right := accessControl.GetDistributedRight(req)
	if right == false {
		_, _ = rw.Write([]byte("false"))
		return
	}

	hostUrl := "http://" + GetOneIp + ":" + GetOnePort + "/getOne"
	responseValidate, validateBody, err := GetCurl(hostUrl, req)
	if err != nil {
		_, _ = rw.Write([]byte("false"))
		return
	}

	if responseValidate.StatusCode == 200 {
		if string(validateBody) == "true" {
			productID, err := strconv.ParseInt(productString, 10, 64)
			if err != nil {
				_, _ = rw.Write([]byte("false"))
				return
			}

			userID, err := strconv.ParseInt(userCookie.Value, 10, 64)
			if err != nil {
				_, _ = rw.Write([]byte("false"))
				return
			}

			message := models.NewMessage(userID, productID)
			byteMessage, err := json.Marshal(message)
			if err != nil {
				_, _ = rw.Write([]byte("false"))
				return
			}

			rabbitMqValidate.PublishSimple(string(byteMessage))
			_, _ = rw.Write([]byte("true"))
			return
		}
	}
	_, _ = rw.Write([]byte("false"))
	return
}

func Auth(rw http.ResponseWriter, req *http.Request) error {
	err := CheckUserInfo(req)
	if err != nil {
		return err
	}
	return nil
}

func CheckUserInfo(req *http.Request) error {
	uidCookie, err := req.Cookie("uid")
	if err != nil {
		return err
	}
	signCookie, err := req.Cookie("sign")
	if err != nil {
		return err
	}
	signByte, err := util.Decrypt([]byte(signCookie.Value))
	if err != nil {
		return err
	}
	if uidCookie.Value != string(signByte) {
		return errors.New("validate sign failed")
	}
	return nil
}

func main() {
	// Load application configuration
	if err := config.LoadConfig("./config"); err != nil {
		panic(err)
	}

	// Initialize logger
	log.InitLogger()
	zap.L().Info("log init success")

	hashConsistent = common.NewConsistent()

	for _, v := range hostArray {
		hashConsistent.Add(v)
	}

	localIp, err := common.GetIntranceIp()
	if err != nil {
		zap.L().Error("get local ip failed", zap.Error(err))
	}
	localHost = localIp

	rabbitMqValidate = mq.NewRabbitMQSimple("go_seckill")
	defer rabbitMqValidate.Destroy()

	http.Handle("/html", http.StripPrefix("/html", http.FileServer(http.Dir("./web/server/htmlProductShow"))))
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./web/server/assets"))))

	filter := common.NewFilter()

	filter.RegisterFilterUri("/check", Auth)
	filter.RegisterFilterUri("/checkRight", Auth)

	http.HandleFunc("/check", filter.Handle(Check))
	http.HandleFunc("/checkRight", filter.Handle(CheckRight))

	_ = http.ListenAndServe(":8080", nil)
}
