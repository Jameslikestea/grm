package config

import "github.com/spf13/viper"

const (
	storageType = "storage.type"

	storageS3Endpoint  = "storage.s3.endpoint"
	storageS3AccessKey = "storage.s3.access_key"
	storageS3SecretKey = "storage.s3.secret_key"
	storageS3SSL       = "storage.s3.ssl"
	storageS3Bucket    = "storage.s3.bucket"

	storageCQLEndpoint = "storage.cql.endpoint"
	storageCQLUsername = "storage.cql.username"
	storageCQLPassword = "storage.cql.password"
	storageCQLKeyspace = "storage.cql.keyspace"

	storageSQLiteFile = "storage.sqlite.url"
)

func GetStorageType() string {
	return viper.GetString(storageType)
}

func GetStorageS3Endpoint() string {
	return viper.GetString(storageS3Endpoint)
}

func GetStorageS3AccessKey() string {
	return viper.GetString(storageS3AccessKey)
}

func GetStorageS3SecretKey() string {
	return viper.GetString(storageS3SecretKey)
}

func GetStorageS3SSL() bool {
	return viper.GetBool(storageS3SSL)
}

func GetStorageS3Bucket() string {
	return viper.GetString(storageS3Bucket)
}

func GetStorageCQLEndpoint() string {
	return viper.GetString(storageCQLEndpoint)
}

func GetStorageCQLUsername() string {
	return viper.GetString(storageCQLUsername)
}

func GetStorageCQLPassword() string {
	return viper.GetString(storageCQLPassword)
}

func GetStorageCQLKeyspace() string {
	return viper.GetString(storageCQLKeyspace)
}

func GetStorageSQLiteURL() string {
	return viper.GetString(storageSQLiteFile)
}
