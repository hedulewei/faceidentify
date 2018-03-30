package main

import (
	"errors"
	"net/rpc"
	"net"
	"fmt"
	"net/http"
	"github.com/widaT/faceidentify/utils"
	"github.com/boltdb/bolt"
	"log"
	"github.com/vmihailenco/msgpack"
	"sort"
)

var db *bolt.DB
func init()  {
	var err error
	db, err = bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal("create db err")
	}
}

type UserInfo struct {
	UID string	`json:"uid"`
	Features [][]float64 `json:"features"`
	GroupId string `json:"group_id"`
	ActionType string `json:"-"`
	Distance []float64 `json:"-"`
}

type FaceInfo struct {
	Users []UserInfo
}

type ResultInfo struct {
	Distance float64
	User UserInfo
}

type ResultSlice []ResultInfo
func (p ResultSlice) Len() int           { return len(p) }
func (p ResultSlice) Less(i, j int) bool { return p[i].Distance < p[j].Distance }
func (p ResultSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type Arg struct {
	Feature []float64
	GroupId string
}

func (f *FaceInfo) Identify(arg Arg, result *[]ResultInfo) error {
	return db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("group_"+arg.GroupId))
		if b == nil {
			//不存在的group
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var user UserInfo
			msgpack.Unmarshal(v,&user)
			for _,feature := range user.Features{
				distance,_ := euclidean(arg.Feature,feature)
				user.Distance = append(user.Distance,distance)
				sort.Float64s(user.Distance)
			}
			user.Features = nil
			if len(*result) < 5 {
				*result = append(*result,ResultInfo{user.Distance[0],user})
				sort.Sort(ResultSlice(*result))
			}else{
				if user.Distance[0]< (*result)[4].Distance {
					(*result)[4] = ResultInfo{user.Distance[0],user}
				}
				sort.Sort(ResultSlice(*result))
			}
		}
		return nil
	})
}

func (f *FaceInfo) GetInfo(arg int, result * string) error {
	*result = fmt.Sprintf("user length %d",len(f.Users))
	return nil
}

func (f *FaceInfo) AddUser(user UserInfo,result *bool)error  {
	err := db.Update(func(tx *bolt.Tx) error {
		b ,_:= tx.CreateBucketIfNotExists([]byte("group_"+user.GroupId))
		if user.ActionType == "replace" {
			buf, err := msgpack.Marshal(user)
			if err != nil {
				return err
			}
			return b.Put([]byte(user.UID), buf)
		}
		btye := b.Get([]byte(user.UID))
		if btye != nil {
			var userB UserInfo
			msgpack.Unmarshal(btye,&userB)
			userB.Features = append(userB.Features,user.Features[0])
			buf, err := msgpack.Marshal(userB)
			if err != nil {
				return err
			}
			return b.Put([]byte(user.UID), buf)
		}else {
			buf, err := msgpack.Marshal(user)
			if err != nil {
				return err
			}
			return b.Put([]byte(user.UID), buf)
		}
		return nil
	})
	if err == nil {
		*result = true
		return nil
	}
	*result = false
	return err
}

func (f *FaceInfo) DelGroup(group string,result *bool)error  {
	err := db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte("group_"+group))
	})
	if err == nil {
		*result = true
		return nil
	}
	*result = false
	return err
}
//求欧几里距离
func euclidean(infoA, infoB []float64) (float64, error) {
	if len(infoA) != len(infoB) {
		return 0, errors.New("params err")
	}
	var distance float64
	for i, number := range infoA {
		//distance += math.Pow(number-infoB[i], 2)
		a:= number-infoB[i]
		distance += a * a
	}
	return distance, nil
}

func main() {
	faceInfo := new(FaceInfo)
	rpc.Register(faceInfo)
	rpc.HandleHTTP()
	port := utils.Conf.GetString("base","port")
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("监听失败，端口可能已经被占用")
	}
	fmt.Println("正在监听"+port+"端口")
	http.Serve(l, nil)
	db.Close()
}