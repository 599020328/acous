package service

import (
	common "acous/commom"
	"acous/web/conf"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func BeginFind(req BeginReq) error {
	timeLog = time.Now().Unix()
	workChain = strings.Split(req.WorkChain, "-")
	nowWorkNum = len(workChain) - 1

	antReq := RequestReq{
		SourceURL:       req.SourceURL,
		DestNumber:      workChain[nowWorkNum],
		Round:           0,
		AntNumber:       0,
		IsLastAnt:       false,
		IsLastRound:     false,
		Data:            "", //todo data size
		TimeStampNano:   time.Now().UnixNano(),
		Path:            configure.Number,
		PathDelay:       0,
		PathDelayDetail: "",
	}

	configure.UpdateList = make([]conf.UpdateInfo, configure.Ant)
	configure.PathLog = make(map[string]int)
	configure.PathTotalDelay = make(map[string]float64)

	_, err := requestNodeByUrl(antReq, req.SourceURL+configure.RequestRelativePath)
	if err != nil {
		log.Println("begin request error :", err)
		return err
	}

	return nil
}

func GetNeighborDelay(req GetNeighborDelayReq) error {
	configure.DataSize = req.DataSize
	if !(req.Round > configure.UpdateRound) {
		log.Println("neighbor delay already obtained!")
		return errors.New("this round get neighbor delay")
	}
	if req.Round-1 != configure.UpdateRound {
		log.Println("lose one update!")
	}

	configure.UpdateRound = req.Round
	for num, info := range configure.NeighborList {
		req.TimeStamp = time.Now().UnixNano()
		result, err := requestNodeByUrl(req, info.Url+configure.UpdateDelayRelativePath)
		if err != nil {
			log.Println("update node ", num, "neighbor delay error, error :", err)
		}

		var resp GetDelayResp
		ret, err := getResp(result, reflect.TypeOf(resp))
		if err != nil {
			log.Println(err.Error())
		}
		resp = *ret.(*GetDelayResp)

		if resp.Code != 0 {
			log.Println("get neighbor delay error :", resp)
		}
		temp := configure.NeighborList[num]
		temp.Delay = resp.TimeDelay
		if temp.Delay == 0 {
			temp.Delay = 1000 //todo init value
		}
		configure.NeighborList[num] = temp
	}

	return nil
}

func getResp(resp []byte, typ reflect.Type) (interface{}, error) {
	ret := reflect.New(typ).Interface()

	if err := json.Unmarshal(resp, ret); err != nil {
		return ret, err
	}

	return ret, nil
}

//开线程处理蚂蚁
func HandleRequest(req RequestReq) error {
	if configure.Number == req.DestNumber {
		log.Println("this ant find the dest!")
		err := sendResultToSourceNode(req)
		if err != nil {
			return err
		}

		return nil
	}

	nextNode, er := ChoseNextNode(configure.NeighborList, req.Path)
	if er != nil {
		err := sendResultToSourceNode(req)
		if err != nil {
			return err
		}

		return er
	}

	er = visitNextNode(req, nextNode)
	if er != nil {
		log.Println("visit next node :", nextNode, " error :", er)
		err := sendResultToSourceNode(req)
		if err != nil {
			return err
		}

		return er
	}

	return nil
}

//访问下一个节点
func visitNextNode(req RequestReq, nextNode string) error {
	log.Println("visit next node :", nextNode)
	newTimeStamp := time.Now().UnixNano()
	if newTimeStamp <= req.TimeStampNano {
		log.Println("time synchronise error!")
	}
	pathDelay := math.Abs(float64(newTimeStamp-req.TimeStampNano)) / 1e6

	req.Path += fmt.Sprintf("-%s", nextNode)
	req.PathDelay += pathDelay
	req.PathDelayDetail += fmt.Sprintf("%s-", strconv.FormatFloat(pathDelay, 'f', -1,
		64))
	req.TimeStampNano = newTimeStamp

	_, err := requestNodeByUrl(req, configure.NeighborList[nextNode].Url+configure.RequestRelativePath)
	return err
}

func sendResultToSourceNode(req RequestReq) error {
	log.Println("visit source node")
	isSuccess := false
	if configure.Number == req.DestNumber {
		isSuccess = true
	}

	reqData := DestNodeReq{
		SourceURL:       req.SourceURL,
		DestNumber:      req.DestNumber,
		Round:           req.Round,
		AntNumber:       req.AntNumber,
		IsLastAnt:       req.IsLastAnt,
		IsLastRound:     req.IsLastRound,
		Path:            req.Path,
		PathDelay:       req.PathDelay,
		PathDelayDetail: req.PathDelayDetail,
		IsSuccess:       isSuccess,
	}

	jsonData, er := json.Marshal(reqData)
	if er != nil {
		log.Println("marshal req data error, err :", er)
		return er
	}

	result, er := common.HTTPPostData(jsonData, req.SourceURL+configure.ResponseRelativePath)
	if er != nil {
		log.Println("request source node error, err :", er)
		return er
	}

	log.Println("send result to source node, result is :", result)
	return nil
}

func ChoseNextNode(NeighborList map[string]conf.Neighbor, path string) (string, error) {
	NNum := len(NeighborList)
	prob := make(map[string]float64, NNum)
	keyList := make([]string, NNum)
	visitedList := strings.Split(path, "-")

	total := 0.0
	nowNum := 0
	for number, info := range NeighborList {
		if isNodeInList(number, visitedList) {
			continue
		}
		prob[number] = math.Pow(info.Tau, configure.Alpha)*math.Pow(1/info.Delay, configure.Gamma) + 0.000001
		total += prob[number]
		keyList[nowNum] = number
		nowNum++
	}

	if nowNum < 1 {
		return "", errors.New("no new neighbor")
	}
	if nowNum == 1 {
		return keyList[0], nil
	}
	//fmt.Println("first prob is :", prob)
	//fmt.Println("before :", keyList)
	//sort.Strings(keyList)
	//fmt.Println("after :", keyList)

	lastProb := 0.0
	for i := 0; i < nowNum; i++ {
		//fmt.Println(keyList[i])
		prob[keyList[i]] = prob[keyList[i]]/total + lastProb
		lastProb = prob[keyList[i]]
	}
	//fmt.Println("second prob is :", prob)

	rand.Seed(time.Now().Unix())
	randNum := rand.Float64()
	//fmt.Println("rand num is :", randNum)
	lastNode := ""
	for i := 0; i < nowNum; i++ {
		fmt.Println(keyList[i])
		lastNode = keyList[i]
		if prob[keyList[i]] > randNum {
			break
		}
	}

	return lastNode, nil
}

//判断是否经过节点
func isNodeInList(node string, lists []string) bool {
	for _, value := range lists {
		if node == value {
			return true
		}
	}

	return false
}

//通过URL请求节点
func requestNodeByUrl(req interface{}, url string) ([]byte, error) {
	reqData, er := json.Marshal(req)
	if er != nil {
		log.Println("marshal req data error, err :", er)
		return nil, er
	}
	result, er := common.HTTPPostData(reqData, url)
	if er != nil {
		log.Println("request error, err :", er)
		return nil, er
	}

	log.Println("request", url, " result is :", string(result))
	return result, nil
}

func SourceNodeHandleResponse(req DestNodeReq) error {
	if !req.IsSuccess {
		log.Println("find path to", req.DestNumber, "error, detail in the dest node!")
		//return nil
	}

	nextReq := RequestReq{
		SourceURL:       req.SourceURL,
		DestNumber:      req.DestNumber,
		Round:           req.Round,
		AntNumber:       req.AntNumber,
		IsLastAnt:       false,
		IsLastRound:     req.IsLastRound,
		Data:            "", //todo data size
		TimeStampNano:   time.Now().UnixNano(),
		Path:            configure.Number,
		PathDelay:       0,
		PathDelayDetail: "",
	}

	er := RecordAntResult(req)
	if er != nil {
		log.Println("log result error, err :", er)
	}
	if req.IsSuccess {
		configure.UpdateList[req.AntNumber].Path = req.Path
		configure.UpdateList[req.AntNumber].PathCost = req.PathDelayDetail

		configure.PathLog[req.Path] += 1
		configure.PathTotalDelay[req.Path] += req.PathDelay
		if configure.PathLog[req.Path] >= configure.PathThreshold {
			log.Println("path convergence!")

			er := RecordPathResult(fmt.Sprintf("dest number: %s; round num: %d; ant num: %d; is success: %v; "+
				"is last round: %v; is last ant: %v \n 	path convergence : %s \n	average delay : %v \n 	"+
				"total time : %v s\n",
				req.DestNumber, req.Round, req.AntNumber, req.IsSuccess, req.IsLastRound, req.IsLastAnt, req.Path,
				configure.PathTotalDelay[req.Path]/float64(configure.PathLog[req.Path]), time.Now().Unix()-timeLog))
			if er != nil {
				log.Println("log path convergence result error, err :", er)
			}

			return beginNewNodeFound(nextReq)
		}
	} else {
		configure.UpdateList[req.AntNumber].Path = ""
		configure.UpdateList[req.AntNumber].PathCost = ""
	}
	if req.IsLastAnt {

		for idx := 0; idx < configure.Ant; idx++ {
			updateTauReq := UpdateTauReq{
				Path:     configure.UpdateList[idx].Path,
				PathCost: configure.UpdateList[idx].PathCost,
				Round:    req.Round,
				Rho:      configure.Rho,
				Q:        configure.Q,
			}
			//fmt.Println(updateTauReq)
			err := updateSourceNodeTau(updateTauReq)
			if err != nil {
				log.Println("update tau error, err :", err)
				//return err
			} else {
				err = updateTauOfNeighbor(updateTauReq)
				if err != nil {
					log.Println("update neighbor tau error, err :", err)
					//return err
				}
			}

		}
		configure.UpdateList = make([]conf.UpdateInfo, configure.Ant)

		nextReq.AntNumber = 0
		nextReq.Round += 1
		if req.IsLastRound {
			log.Println("last round and last ant, end!!!")

			pathFind := findMaxOccurTimesPath()
			er := RecordPathResult(fmt.Sprintf("dest number: %s; round num: %d; ant num: %d; is success: %v; "+
				"is last round: %v; is last ant: %v \n 	path convergence : %s \n	average delay : %v \n 	"+
				"total time : %v s\n",
				req.DestNumber, req.Round, req.AntNumber, req.IsSuccess, req.IsLastRound, req.IsLastAnt, req.Path,
				configure.PathTotalDelay[pathFind]/float64(configure.PathLog[pathFind]), time.Now().Unix()-timeLog))
			if er != nil {
				log.Println("log path convergence result error, err :", er)
			}

			UpdateData(req.DestNumber)					//todo send result to dest node and update the data
			return beginNewNodeFound(nextReq)
		} else {
			if req.Round+1 >= configure.Round {
				nextReq.IsLastRound = true
			}
		}
	} else {
		nextReq.AntNumber += 1
		if nextReq.AntNumber+1 >= configure.Ant {
			nextReq.IsLastAnt = true
		}
	}

	nextNode, err := ChoseNextNode(configure.NeighborList, "")
	if err != nil {
		log.Println("chose next node error :", err)
		return err
	}

	return visitNextNode(nextReq, nextNode)
}

//任务链往下走
func beginNewNodeFound(req RequestReq) error {
	nowWorkNum--
	if nowWorkNum < 0 {
		return errors.New("work china ended")
	}
	antReq := RequestReq{
		SourceURL:       req.SourceURL,
		DestNumber:      workChain[nowWorkNum],
		Round:           0,
		AntNumber:       0,
		IsLastAnt:       false,
		IsLastRound:     false,
		Data:            "", //todo data size
		TimeStampNano:   time.Now().UnixNano(),
		Path:            configure.Number,
		PathDelay:       0,
		PathDelayDetail: "",
	}
	_, err := requestNodeByUrl(antReq, antReq.SourceURL+configure.RequestRelativePath)
	if err != nil {
		log.Println("begin request error :", err)
		return err
	}
	return nil
}

func findMaxOccurTimesPath() string {
	nowMax := 0
	nowPath := ""

	for key, value := range configure.PathLog {
		if value > nowMax {
			nowMax = value
			nowPath = key
		}
	}

	return nowPath
}

//结果记录
func RecordAntResult(req DestNodeReq) error {
	f, err := os.OpenFile("ant_result.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Println("open log file error :", err)
	}

	defer func() {
		err := f.Close()
		if err != nil {
			log.Println("close file error :", err)
		}
	}()
		//todo update data
	result, err := f.WriteString(fmt.Sprintf("dest number: %s; round num: %d; ant num: %d; is success: %v; "+
		"is last round: %v; is last ant: %v \n 	path: %s\n 	path_cost: %f\n 	path_detail:%s\n", req.DestNumber,
		req.Round, req.AntNumber, req.IsSuccess, req.IsLastRound, req.IsLastAnt, req.Path, req.PathDelay,
		req.PathDelayDetail))
	if err != nil {
		log.Println("write to file error :", err)
	}
	log.Println("write to file, result :", result)

	return nil
}

//收敛结果记录
func RecordPathResult(content string) error {
	f, err := os.OpenFile("path_result.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Println("open log file error :", err)
	}

	defer func() {
		err := f.Close()
		if err != nil {
			log.Println("close file error :", err)
		}
	}()

	result, err := f.WriteString(content)
	if err != nil {
		log.Println("write to file error :", err)
	}
	log.Println("write to file, result :", result)

	return nil
}

//处理更新请求
func HandleUpdate(req UpdateTauReq) error {
	err := updateTauOverNode(req)
	if err != nil {
		log.Println("update node tau error :", err)
		return err
	}

	return updateTauOfNeighbor(req)
}

//更新所有邻居Tau
func updateTauOfNeighbor(req UpdateTauReq) error {
	configure.Lock()
	defer configure.Unlock()
	for node, value := range configure.NeighborList {
		_, err := requestNodeByUrl(req, value.Url+configure.UpdateRelativePath)
		if err != nil {
			log.Println("update node ", node, "error, error :", err)
		}
	}

	return nil
}

//更新节点对邻居tau
func updateTauOverNode(req UpdateTauReq) error {
	configure.Lock()
	defer configure.Unlock()
	if !(req.Round > configure.Round) {
		return errors.New("this round already updated")
	}

	if configure.Round != req.Round-1 {
		log.Println("now update round :", req.Round, ", last update round :", configure.Round)
	}

	for num, info := range configure.NeighborList {
		log.Println("neighbor ", num, "sub tau")
		configure.UpdateTau(num, info.Tau*(1-req.Rho))
	}

	if req.Path == "" {
		log.Println("no add")
		return nil
	}
	pathList := strings.Split(req.Path, "-")
	for idx, num := range pathList {
		if num == configure.Number {
			log.Println("path :", req.Path)
			log.Println("path cost", req.PathCost)
			costList := strings.Split(req.PathCost, "-")
			if idx+1 == len(pathList) {
				olderTau := configure.NeighborList[pathList[idx-1]].Tau
				cost, _ := strconv.ParseFloat(costList[idx-1], 64)
				newTau := olderTau + float64(req.Q)/cost
				configure.UpdateTau(pathList[idx-1], newTau)

				configure.Round = req.Round
				break
			}

			log.Println("pre node :", pathList[idx-1], ", cost :", costList[idx-1], "; next node :",
				pathList[idx+1], ", cost :", costList[idx])

			olderTau := configure.NeighborList[pathList[idx+1]].Tau
			cost, _ := strconv.ParseFloat(costList[idx], 64)
			newTau := olderTau + float64(req.Q)/cost
			configure.UpdateTau(pathList[idx+1], newTau)

			olderTau = configure.NeighborList[pathList[idx-1]].Tau
			cost, _ = strconv.ParseFloat(costList[idx-1], 64)
			newTau = olderTau + float64(req.Q)/cost
			configure.UpdateTau(pathList[idx-1], newTau)

			configure.Round = req.Round
			break
		}
	}

	return nil
}

//更新源节点对邻居tau
func updateSourceNodeTau(req UpdateTauReq) error {
	if !(req.Round > configure.Round) {
		return errors.New("this round already updated")
	}

	if configure.Round != req.Round-1 {
		log.Println("now update round :", req.Round, ", last update round :", configure.Round)
	}

	for num, info := range configure.NeighborList {
		log.Println("neighbor ", num, "sub tau")
		configure.UpdateTau(num, info.Tau*(1-req.Rho))
	}

	if !(req.Path != "") {
		log.Println("this ant failed")

		return nil
	}
	pathList := strings.Split(req.Path, "-")
	costList := strings.Split(req.PathCost, "-")

	olderTau := configure.NeighborList[pathList[1]].Tau
	cost, _ := strconv.ParseFloat(costList[0], 64)
	newTau := olderTau + float64(req.Q)/cost
	configure.UpdateTau(pathList[1], newTau)

	configure.Round = req.Round

	return nil
}

func startUpdateTask(freqOnMinute int32) {
	log.Printf("Start timing scan match at %d minutes late...", freqOnMinute)

	freq := time.Second * time.Duration(freqOnMinute)
	c := time.Tick(freq)

	timingNum := 0
	for range c {
		log.Println("now time :", timingNum)
		timingNum++

		if timingNum > 10 {
			break
		}
	}
}

//发送处理请求到目标节点
func UpdateData(destNum string) {
	updateDataReq := UpdateDataReq{DataDelta:""}			//todo get real world now node data, and data type

	_, err := requestNodeByUrl(updateDataReq, configure.NeighborList[destNum].Url+configure.UpdateDataPath)
	if err != nil {
		log.Println("update node ", destNum, "data error, error :", err)
	}
}

//更新目标节点数据
func HandleUpdateData(req UpdateDataReq) error {
	nowData := ""		//todo get real world now node data, and data type

	nowData = XOR(nowData, req.DataDelta)	//todo flush newest data to disk

	return nil
}

func XOR(data, delta string) string {
	return data + delta		//todo compute real world result
}
