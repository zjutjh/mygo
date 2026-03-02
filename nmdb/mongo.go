package nmdb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// New 以指定配置创建实例
func New(conf Config) (*mongo.Database, error) {
	// 设置客户端选项
	clientOptions := options.Client().
		SetHosts(conf.Host).
		SetConnectTimeout(conf.ConnectTimeout).
		SetSocketTimeout(conf.SocketTimeout).
		SetServerSelectionTimeout(conf.ServerSelectionTimeout).
		SetMaxPoolSize(conf.MaxPoolSize).
		SetMinPoolSize(conf.MinPoolSize).
		SetMaxConnecting(conf.MaxConnecting).
		SetMaxConnIdleTime(conf.MaxIdleTime).
		SetAppName(conf.AppName).
		SetReplicaSet(conf.ReplicaSet).
		SetDirect(conf.Direct).
		SetRetryWrites(conf.RetryWrites).
		SetRetryReads(conf.RetryReads).
		SetCompressors(conf.Compressors).
		SetHeartbeatInterval(conf.HeartbeatInterval).
		SetZstdLevel(conf.ZstdLevel)

	// 设置读写关注
	if conf.ReadConcern != "" {
		clientOptions.SetReadConcern(&readconcern.ReadConcern{Level: conf.ReadConcern})
	}
	if conf.WriteConcern != "" {
		clientOptions.SetWriteConcern(&writeconcern.WriteConcern{W: conf.WriteConcern})
	}
	if conf.ReadPreference != "" {
		mode, err := readpref.ModeFromString(conf.ReadPreference)
		if err != nil {
			mode = readpref.PrimaryMode
		}
		// 创建读取偏好
		rp, _ := readpref.New(mode)
		clientOptions.SetReadPreference(rp)

	}

	// 设置认证信息
	if conf.Username != "" && conf.Password != "" {
		clientOptions.SetAuth(options.Credential{
			Username:   conf.Username,
			Password:   conf.Password,
			AuthSource: conf.AuthDB,
		})
	}
	// 创建 MongoDB 客户端连接并初始化
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("创建MongoDB实例错误: %w", err)
	}

	//失败断开连接
	if err = client.Ping(context.Background(), readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("MongoDB连接测试失败: %w", err)
	}
	//从client获取指定的数据库实例
	var db *mongo.Database
	db = client.Database(conf.Database)

	return db, nil

}
