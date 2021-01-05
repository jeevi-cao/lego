package mongo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

//document @see https://godoc.org/go.mongodb.org/mongo-driver
//mongodb uri:  @see https://docs.mongodb.com/manual/reference/connection-string/
// usage:
//
//	hosts := "10.136.158.10:27000,10.136.158.10:27001,10.136.158.10:27002"
//	replset := "rs_image"
//
//	setting := Setting{
//		Hosts: hosts,
//		ReplSet: replset,
//	}
//	Mongo, err := NewMongo(&setting)
//	if err != nil {
//		return
//	}
//
//	c := Mongo.Client.Database("test").Collection("user")
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	cur, err := c.Find(ctx, bson.D{})
//	if err != nil {
//		t.Error("find collection error:", err)
//		return
//	}
//	defer cur.Close(ctx)
//	for cur.Next(ctx) {
//		var result bson.M
//		err := cur.Decode(&result)
//		if err != nil {
//		}
//	}

var readPreferenceMap = map[string]*readpref.ReadPref{
	"primary":            readpref.Primary(),
	"primaryPreferred":   readpref.PrimaryPreferred(),
	"secondary":          readpref.Secondary(),
	"secondaryPreferred": readpref.PrimaryPreferred(),
	"nearest":            readpref.Nearest(),
}

type Mongo struct {
	Client  *mongo.Client
	Setting *Setting
}

type Setting struct {
	//Uri 形式
	Uri string
	//此四个参数和Uri 互斥
	Hosts    string
	ReplSet  string
	Username string
	Password string

	//max conn size default: 100
	MaxPoolSize uint64
	//min conn size
	MinPoolSize uint64
	//unit second
	MaxIdleTime int
	//primary (Default)
	//primaryPreferred
	//secondary
	//secondaryPreferred
	//nearest
	ReadPreference string
}

//初始化数据
func NewMongo(setting *Setting) (*Mongo, error) {
	opts := buildOptions(setting)
	cli, err := mongo.NewClient(opts)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("new mongodb error:%s", err.Error()))
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = cli.Connect(ctx)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("connect mongodb error:%s", err.Error()))
	}
	return &Mongo{Client: cli, Setting: setting}, nil
}

func (m *Mongo) GetClient() *mongo.Client {
	return m.Client
}

func (m *Mongo) Close() {
	ctx := context.Background()
	_ = m.Client.Disconnect(ctx)
}

func buildOptions(setting *Setting) *options.ClientOptions {
	opts := options.Client()
	//first uri
	if len(setting.Uri) > 0 {
		opts.ApplyURI(setting.Uri)
	} else {
		//hosts
		if len(setting.Hosts) > 0 {
			opts.SetHosts(strings.Split(setting.Hosts, ","))
		}
		//Username
		if len(setting.Username) > 0 {
			auth := options.Credential{
				Username: setting.Username,
				Password: setting.Password,
			}
			opts.SetAuth(auth)
		}
		//replySet
		opts.SetReplicaSet(setting.ReplSet)
	}
	//maxsize
	if setting.MaxPoolSize > 0 {
		opts.SetMaxPoolSize(setting.MaxPoolSize)
	}
	//min pool size
	if setting.MinPoolSize > 0 {
		opts.SetMinPoolSize(setting.MinPoolSize)
	}

	if setting.MaxIdleTime > 0 {
		opts.SetMaxConnIdleTime(time.Duration(setting.MaxIdleTime) * time.Second)
	}

	//readPreference
	if len(setting.ReadPreference) > 0 {
		if v, ok := readPreferenceMap[setting.ReadPreference]; ok {
			opts.SetReadPreference(v)
		}
	}
	return opts
}
