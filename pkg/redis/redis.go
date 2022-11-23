package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// Config ...
type Config struct {
	MasterName       string
	Addrs            []string
	DB               int
	Username         string
	Password         string
	SentinelUsername string
	SentinelPassword string
	Prefix           string
	Separator        string
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
}

// Client ...
type Client struct {
	client redis.UniversalClient

	config Config
}

// New ...
func New(cfg Config) (*Client, error) {
	client := redis.NewUniversalClient(&redis.UniversalOptions{
		MasterName:       cfg.MasterName,
		Addrs:            cfg.Addrs,
		DB:               cfg.DB,
		Username:         cfg.Username,
		Password:         cfg.Password,
		SentinelUsername: cfg.SentinelUsername,
		SentinelPassword: cfg.SentinelPassword,
		ReadTimeout:      cfg.ReadTimeout,
		WriteTimeout:     cfg.WriteTimeout,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	return &Client{
		config: cfg,
		client: client,
	}, nil
}

// Set ...
func (c *Client) Set(ctx context.Context, key string, v interface{}, exp time.Duration) error {
	k := FormatKey(c.config.Separator, c.config.Prefix, key)

	data, err := Encode(v)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, k, data, exp).Err()
}

// Get ...
func (c *Client) Get(ctx context.Context, key string, o interface{}) error {
	k := FormatKey(c.config.Separator, c.config.Prefix, key)
	data, err := c.client.Get(ctx, k).Result()
	if err != nil {
		return err
	}

	return Decode([]byte(data), o)
}

// Exists ...
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	k := FormatKey(c.config.Separator, c.config.Prefix, key)
	res, err := c.client.Exists(ctx, k).Result()
	if err != nil {
		return false, err
	}

	return res > 0, nil
}