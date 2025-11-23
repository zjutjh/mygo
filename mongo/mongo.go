package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type DB struct {
	Client   *mongo.Client
	Config   Config
	Database *mongo.Database
}

func New(conf Config) (*DB, error) {

	uri := buildURI(conf)
	clientOptions := options.Client().ApplyURI(uri)
	applyConfig(clientOptions, conf)
	ctx, cancel := context.WithTimeout(context.Background(), conf.ConnectTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("创建MongoDB实例错误: %w", err)
	}

	// 测试连接
	ctx, cancel = context.WithTimeout(context.Background(), conf.ServerSelectionTimeout)
	defer cancel()

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("MongoDB连接测试失败: %w", err)
	}

	db := &DB{
		Client: client,
		Config: conf,
	}
	//初始化Database对象
	if conf.Database != "" {
		db.Database = client.Database(conf.Database)
	}

	return db, nil

}

func buildURI(conf Config) string {
	uri := fmt.Sprintf("mongodb://%s:%d", conf.Host, conf.Port)
	if conf.Username != "" && conf.Password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d", conf.Username, conf.Password, conf.Host, conf.Port)
		if conf.AuthDB != "" {
			uri = fmt.Sprintf("%s/?authSource=%s", uri, conf.AuthDB)
		}
	}
	return uri
}
func applyConfig(clientOptions *options.ClientOptions, conf Config) {
	clientOptions.SetConnectTimeout(conf.ConnectTimeout)
	clientOptions.SetSocketTimeout(conf.SocketTimeout)
	clientOptions.SetServerSelectionTimeout(conf.ServerSelectionTimeout)
	clientOptions.SetMaxPoolSize(conf.MaxPoolSize)
	clientOptions.SetMinPoolSize(conf.MinPoolSize)
	clientOptions.SetMaxConnecting(conf.MaxConnecting)
	clientOptions.SetMaxConnIdleTime(conf.MaxIdleTime)
	clientOptions.SetAppName(conf.AppName)
	clientOptions.SetReplicaSet(conf.ReplicaSet)
	clientOptions.SetDirect(conf.Direct)
	clientOptions.SetRetryWrites(conf.RetryWrites)
	clientOptions.SetRetryReads(conf.RetryReads)
	clientOptions.SetCompressors(conf.Compressors)

}
