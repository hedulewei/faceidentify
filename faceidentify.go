package main

import (
	"github.com/widaT/golib/web/server"
	"math"
	"errors"
	"github.com/widaT/faceidentify/utils"
)
type FaceInfo struct {

}

var bigArr []FaceInfo

func identify( ctx *server.Context) string  {

	return ""
}

//求欧几里距离
func euclidean(infoA,infoB[]float64) (float64,error) {
	if len(infoA) != len(infoB) {
		return  0,errors.New("params err")
	}
	var distance float64
	for i,number := range infoA {
		distance += math.Pow(number - infoB[i],2)
	}
	return distance,nil
}

func main()  {
	server.AddRoute("/",identify)
	server.Run(":"+utils.Conf.GetString("base","port"))
}
