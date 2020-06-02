package service

import (
	common "acous/commom"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"math"
	"time"
)

func BeginHandler(c *gin.Context) {
	req := BeginReq{}
	if err := c.BindJSON(&req); err == nil {
		er := BeginFind(req)
		if er != nil {
			log.Println("begin finding error :", er)
		}

		c.JSON(200, gin.H{"code": 0, "message": "success"})
		return
	}
	c.JSON(200, gin.H{"code": -1, "message": "传入参数错误"})
}

func BeginUpdateDelayHandler(c *gin.Context) {
	req := GetNeighborDelayReq{}
	if err := c.BindJSON(&req); err == nil {
		req.TimeStamp = time.Now().UnixNano()
		req.Round = configure.Round + 1

		er := GetNeighborDelay(req)
		if er != nil {
			log.Println("begin update neighbor delay error :", er)
		}

		c.JSON(200, gin.H{"code": 0, "message": "success"})
		return
	}
	c.JSON(200, gin.H{"code": -1, "message": "传入参数错误"})
}

func GetNeighborDelayHandler(c *gin.Context) {
	req := GetNeighborDelayReq{}
	if err := c.BindJSON(&req); err == nil {
		go func() {
			er := GetNeighborDelay(req)
			if er != nil {
				log.Println("get neighbor delay error :", er)
			}
		}()

		c.JSON(200, gin.H{"code": 0, "message": "success", "time_delay": math.Abs(float64(time.Now().UnixNano()-req.TimeStamp)) / 1e6})
		return
	}
	c.JSON(200, gin.H{"code": -1, "message": "传入参数错误"})
}

func RequestHandler(c *gin.Context) {
	req := RequestReq{}
	if err := c.BindJSON(&req); err == nil {
		go func() {
			er := HandleRequest(req)
			if er != nil {
				log.Println("handle request error :", er)
			}
		}()

		c.JSON(200, gin.H{"code": 0, "message": "success"})
		return
	}

	log.Println("parse param error")
	err := sendResultToSourceNode(req)
	if err != nil {
		log.Println("request source node error, err :", err)
	}
	c.JSON(200, gin.H{"code": -1, "message": "传入参数错误"})
}

func ResultHandler(c *gin.Context) {
	req := DestNodeReq{}
	if err := c.BindJSON(&req); err == nil {
		go func() {
			er := SourceNodeHandleResponse(req)
			if er != nil {
				log.Println("handle result error :", er)
			}
		}()

		c.JSON(200, gin.H{"code": 0, "message": "success"})
		return
	}

	c.JSON(200, gin.H{"code": -1, "message": "传入参数错误"})
}

//更新tau
func UpdateHandler(c *gin.Context) {
	req := UpdateTauReq{}
	if err := c.BindJSON(&req); err == nil {
		go func() {
			er := HandleUpdate(req)
			if er != nil {
				log.Println("handle update error :", er)
			}
		}()

		c.JSON(200, gin.H{"code": 0, "message": "success"})
		return
	}

	c.JSON(200, gin.H{"code": -1, "message": "传入参数错误"})
}

//更新tau
func UpdateDataHandler(c *gin.Context) {
	req := UpdateDataReq{}
	if err := c.BindJSON(&req); err == nil {
		go func() {
			er := HandleUpdateData(req)
			if er != nil {
				log.Println("handle update error :", er)
			}
		}()

		c.JSON(200, gin.H{"code": 0, "message": "success"})
		return
	}

	c.JSON(200, gin.H{"code": -1, "message": "传入参数错误"})
}

func LogConfig(c *gin.Context) {
	reqData, er := json.Marshal(configure)
	if er != nil {
		log.Println("marshal req data error, err :", er)
	}
	log.Println(string(reqData))

	c.JSON(200, gin.H{"code": 0, "message": "success"})
}

func TestSt(c *gin.Context) {
	req := TestReq{}
	if err := c.BindJSON(&req); err == nil {
		go startUpdateTask(1)

		c.JSON(200, gin.H{"code": 0, "message": "success"})

		timeStamp := time.Now().Unix()
		timeStampNano := time.Now().UnixNano()

		fmt.Println("req time stamp :", req.TimeStamp, " now time stamp :", timeStamp, "; req time stamp nano :",
			req.TimeStampNano, "time stamp nano :", timeStampNano)

		reqData := TestReq{
			TimeStamp:     timeStamp,
			TimeStampNano: timeStampNano,
			NextURL:       "http://www.zcss-sportsculture.com:8000/user/loginByPassword",
		}
		jsonData, er := json.Marshal(reqData)
		if er != nil {
			log.Println("marshal req data error, err :", er)
		}

		result, er := common.HTTPPostData(jsonData, req.NextURL)
		if er != nil {
			log.Println("request source node error, err :", er)
		}
		fmt.Println(string(result))

		return
	}

	c.JSON(200, gin.H{"code": -1, "message": "传入参数错误"})
}
