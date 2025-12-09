package ngodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func New(conf Config) (*mongo.Database, error) {
	clientOptions := options.Client().
		SetHosts([]string{fmt.Sprintf("%s:%d", conf.Host, conf.Port)}).
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
		SetCompressors(conf.Compressors)

	// 设置认证信息
	if conf.Username != "" && conf.Password != "" {
		clientOptions.SetAuth(options.Credential{
			Username:   conf.Username,
			Password:   conf.Password,
			AuthSource: conf.AuthDB,
		})
	}

	ctx := context.Background()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("创建MongoDB实例错误: %w", err)
	}

	// 测试连接
	ctx = context.Background()

	//失败断开连接
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("MongoDB连接测试失败: %w", err)
	}

	var db *mongo.Database

	if conf.Database != "" {
		db = client.Database(conf.Database)
	}

	return db, nil

}
